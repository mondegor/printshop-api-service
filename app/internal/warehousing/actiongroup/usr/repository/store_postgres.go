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
	// StorePostgres - comment struct.
	StorePostgres struct {
		client     mrstorage.DBConnManager
		sqlBuilder mrstorage.SQLConditionBuilder
	}
)

// NewStorePostgres - создаёт объект StorePostgres.
func NewStorePostgres(client mrstorage.DBConnManager) *StorePostgres {
	return &StorePostgres{
		client:     client,
		sqlBuilder: part.NewSQLConditionBuilder(),
	}
}

// FetchByCondition - возвращает места хранения аккаунта по фильтру, начиная за позицией params.Cursor
// (не более params.Cursor.Limit), и признак наличия записей за ней.
// -- отображение мест на территории (предварительно проверить, что территория принадлежит аккаунту)
// -- (территории редко меняются, они загружаются отдельно).
func (re *StorePostgres) FetchByCondition(ctx context.Context, params dto.StoreParams) (rows []entity.Store, hasNext bool, err error) {
	condition := re.sqlBuilder.BuildFunc(
		func(c mrstorage.SQLConditionHelper) mrstorage.SQLPartFunc {
			conds := []mrstorage.SQLPartFunc{
				c.Equal("account_id", params.AccountID),
				c.Expr("deleted_at IS NULL"),
			}

			if len(params.Filter.SearchTerritories) > 0 {
				conds = append(conds, c.FilterAnyOf("territory_id", params.Filter.SearchTerritories))
			}

			if params.Filter.SearchCode != "" {
				conds = append(conds, c.FilterLikePrefix("store_code", params.Filter.SearchCode))
			}

			if params.Cursor.TerritoryID > 0 {
				conds = append(
					conds,
					c.Expr("(territory_id, store_code) > (%s, %s)", params.Cursor.TerritoryID, params.Cursor.Code),
				)
			}

			return c.JoinAnd(conds...)
		},
	)

	whereStr, whereArgs := condition.ToSQL()

	sql := `
		SELECT
			store_id,
			tag_version,
			territory_id,
			store_kind,
			store_code,
			store_volume,
			activity_status,
			containers_volume,
			created_at,
			updated_at
		FROM
			` + module.DBTableNameStores + `
		WHERE
			` + whereStr + `
		ORDER BY
			territory_id, store_code
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

		var row entity.Store

		err = cursor.Scan(
			&row.ID,
			&row.TagVersion,
			&row.TerritoryID,
			&row.Kind,
			&row.Code,
			&row.Volume,
			&row.Status,
			&row.ContainersVolume,
			&row.CreatedAt,
			&row.UpdatedAt,
		)
		if err != nil {
			return nil, false, err
		}

		// :TODO: перенести
		if row.Volume.Calc() > 0 {
			row.Fullness = 100 * row.ContainersVolume / row.Volume.Calc()
		}

		if rows == nil {
			rows = make([]entity.Store, 0, params.Cursor.Limit)
		}

		rows = append(rows, row)
	}

	return rows, hasNext, cursor.Err()
}

// FetchOne - comment method.
func (re *StorePostgres) FetchOne(ctx context.Context, accountID uuid.UUID, rowID uint64) (row entity.Store, err error) {
	sql := `
		SELECT
			store_id,
			tag_version,
			territory_id,
			store_kind,
			store_code,
			store_volume,
			activity_status,
			containers_volume,
			created_at,
			updated_at
		FROM
			` + module.DBTableNameStores + `
		WHERE
			store_id = $1 AND account_id = $2 AND deleted_at IS NULL
		FETCH FIRST 1 ROW ONLY;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
	).Scan(
		&row.ID,
		&row.TagVersion,
		&row.TerritoryID,
		&row.Kind,
		&row.Code,
		&row.Volume,
		&row.Status,
		&row.ContainersVolume,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	return row, err
}

// FetchState - comment method.
func (re *StorePostgres) FetchState(ctx context.Context, accountID uuid.UUID, rowID uint64) (row entity.StoreState, err error) {
	sql := `
		SELECT
			store_id,
			tag_version,
			territory_id,
			store_kind,
			store_code,
			activity_status
		FROM
			` + module.DBTableNameStores + `
		WHERE
			store_id = $1 AND account_id = $2 AND deleted_at IS NULL
		FETCH FIRST 1 ROW ONLY;`

	err = re.client.Conn(ctx).QueryRow(
		ctx,
		sql,
		rowID,
		accountID,
	).Scan(
		&row.ID,
		&row.TagVersion,
		&row.TerritoryID,
		&row.Kind,
		&row.Code,
		&row.Status,
	)

	return row, err
}
