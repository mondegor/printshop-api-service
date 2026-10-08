package repository

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/mrpostgres/builder/part"
	"github.com/mondegor/go-core/mrstorage"

	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/module"
)

type (
	// StockPostgres - comment struct.
	StockPostgres struct {
		client     mrstorage.DBConnManager
		sqlBuilder mrstorage.SQLConditionBuilder
	}
)

// NewStockPostgres - создаёт объект StockPostgres.
func NewStockPostgres(client mrstorage.DBConnManager) *StockPostgres {
	return &StockPostgres{
		client:     client,
		sqlBuilder: part.NewSQLConditionBuilder(),
	}
}

// FetchByCondition - возвращает остатки аккаунта по фильтру по возрастанию stock_id, начиная за позицией
// params.Cursor (не более params.Cursor.Limit), и признак наличия записей за ней.
//
// cs.container_id IN(1000000000000000005):
// -- отображение местоположения указанных контейнеров (территории редко меняются, они загружаются отдельно)
//
// cs.location_id IN(11, 15):
// -- отображение занятости указанных мест (складов)
// -- (территории редко меняются, они загружаются отдельно)
// -- store_volume >= container_quantity * container_volume - ok, else перегружен.
func (re *StockPostgres) FetchByCondition(ctx context.Context, params dto.StockParams) (rows []entity.Stock, hasNext bool, err error) {
	condition := re.sqlBuilder.BuildFunc(
		func(c mrstorage.SQLConditionHelper) mrstorage.SQLPartFunc {
			conds := []mrstorage.SQLPartFunc{
				c.Equal("account_id", params.AccountID),
				c.Greater("stock_id", params.Cursor.AfterID()),
			}

			if len(params.Filter.SearchLocations) > 0 {
				conds = append(conds, c.FilterAnyOf("location_id", params.Filter.SearchLocations))
			}

			if len(params.Filter.SearchContainers) > 0 {
				conds = append(conds, c.FilterAnyOf("container_id", params.Filter.SearchContainers))
			}

			return c.JoinAnd(conds...)
		},
	)

	whereStr, whereArgs := condition.ToSQL()

	sql := `
		SELECT
			stock_id,
			account_id,
			container_id,
			location_id,
			container_quantity,
			container_volume,
			created_at
		FROM
			` + module.DBTableNameStocks + `
		WHERE
			` + whereStr + `
		ORDER BY
			stock_id
		FETCH FIRST ` + strconv.Itoa(params.Cursor.Limit+1) + ` ROWS ONLY;`

	cursor, err := re.client.Conn(ctx).Query(
		ctx,
		sql,
		whereArgs...,
	)
	if err != nil {
		return nil, false, err
	}

	defer cursor.Close()

	for cursor.Next() {
		// лишняя строка сверх limit лишь сообщает, что за текущей страницей есть записи
		if len(rows) == params.Cursor.Limit {
			hasNext = true

			break
		}

		var row entity.Stock

		err = cursor.Scan(
			&row.ID,
			&row.AccountID,
			&row.ContainerID,
			&row.LocationID,
			&row.ContainerQuantity,
			&row.ContainerVolume,
			&row.CreatedAt,
		)
		if err != nil {
			return nil, false, err
		}

		if rows == nil {
			rows = make([]entity.Stock, 0, params.Cursor.Limit)
		}

		rows = append(rows, row)
	}

	return rows, hasNext, cursor.Err()
}

// FetchOne - comment method.
func (re *StockPostgres) FetchOne(ctx context.Context, accountID uuid.UUID, rowID uint64) (row entity.Stock, err error) {
	sql := `
		SELECT
			stock_id,
			account_id,
			container_id,
			location_id,
			container_quantity,
			container_volume,
			created_at
		FROM
			` + module.DBTableNameStocks + `
		WHERE
			stock_id = $1 AND account_id = $2
		FETCH FIRST 1 ROW ONLY;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
	).Scan(
		&row.ID,
		&row.AccountID,
		&row.ContainerID,
		&row.LocationID,
		&row.ContainerQuantity,
		&row.ContainerVolume,
		&row.CreatedAt,
	)

	return row, err
}

// CheckLocationAvailability - comment method.
func (re *StockPostgres) CheckLocationAvailability(ctx context.Context, accountID uuid.UUID, locationID uint64) error {
	sql := `
		SELECT
			1
		FROM
			` + module.DBTableNameStocks + `
		WHERE
			account_id = $1 AND location_id = $2
		FETCH FIRST 1 ROW ONLY;`

	var value int

	err := re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		accountID,
		locationID,
	).Scan(
		&value,
	)
	if err != nil {
		return err
	}

	if value == 1 {
		return module.ErrLocationIsOccupied
	}

	return nil
}

// InsertOrUpdate - comment method.
func (re *StockPostgres) InsertOrUpdate(ctx context.Context, row entity.Stock) (rowID uint64, err error) {
	sql := `
		INSERT INTO ` + module.DBTableNameStocks + ` AS t1
			(
				account_id,
				container_id,
				location_id,
				container_quantity,
				container_volume
			)
		VALUES
			($1, $2, $3, $4, $5)
		ON CONFLICT (container_id, location_id) DO UPDATE
		SET
			stock_id = EXCLUDED.stock_id,
			container_quantity = t1.container_quantity + EXCLUDED.container_quantity,
			container_volume = EXCLUDED.container_volume,
			created_at = NOW()
		RETURNING
			t1.stock_id;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		row.AccountID,
		row.ContainerID,
		row.LocationID,
		row.ContainerQuantity,
		row.ContainerVolume,
	).Scan(
		&rowID,
	)

	return rowID, err
}

// UpdateQuantity - comment method.
func (re *StockPostgres) UpdateQuantity(ctx context.Context, accountID uuid.UUID, rowID uint64, quantity int) (newRowID uint64, err error) {
	sql := `
		UPDATE
			` + module.DBTableNameStocks + `
		SET
			stock_id = nextval('` + module.DBTableNameStocks + `_stock_id_seq'),
			container_quantity = $3,
			created_at = NOW()
		WHERE
			stock_id = $1 AND account_id = $2
		RETURNING
			stock_id;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
		quantity,
	).Scan(
		&newRowID,
	)

	return newRowID, err
}

// Delete - comment method.
func (re *StockPostgres) Delete(ctx context.Context, accountID uuid.UUID, rowID uint64) error {
	sql := `
		DELETE FROM
			` + module.DBTableNameStocks + `
		WHERE
			stock_id = $1 AND account_id = $2;`

	return re.client.Conn(ctx).Exec(
		ctx,
		sql,
		rowID,
		accountID,
	)
}

//
// // по ID стоков возвращаются все склады (без контейнеров)
// func (re *StockPostgres) FetchStoresByStockIDs(ctx context.Context, stockIDs []uint64) (storeIDs []uint64, err error) {
// 	sql := `
// 		SELECT
// 			location_id
// 		FROM
//     		` + module.DBTableNameStocks + `
// 		WHERE
//     		stock_id = ANY($1) AND location_id < 1000000000000000000
// 		GROUP BY
// 			location_id;`
//
// 	cursor, err := re.client.Conn(ctx).Query(
// 		ctx,
// 		sql,
// 		stockIDs,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	defer cursor.Close()
//
// 	storeIDs = make([]uint64, 0, len(stockIDs))
//
// 	for cursor.Next() {
// 		var (
// 			storeID uint64
// 		)
//
// 		err = cursor.Scan(
// 			&storeID,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		storeIDs = append(storeIDs, storeID)
// 	}
//
// 	return storeIDs, cursor.Err()
// }
