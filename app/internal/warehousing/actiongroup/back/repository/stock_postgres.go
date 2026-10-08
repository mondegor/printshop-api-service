package repository

import (
	"context"
	"strconv"

	"github.com/mondegor/go-core/mrstorage"

	"print-shop-back/internal/warehousing/actiongroup/back/dto"
	"print-shop-back/internal/warehousing/module"
)

type (
	// StockPostgres - comment struct.
	StockPostgres struct {
		client mrstorage.DBConnManager
	}
)

// NewStockPostgres - создаёт объект StockPostgres.
func NewStockPostgres(client mrstorage.DBConnManager) *StockPostgres {
	return &StockPostgres{
		client: client,
	}
}

// FetchByLocationIDs - возвращает остатки указанных мест хранения по возрастанию stock_id, начиная
// за позицией stockCursor (не более stockCursor.Limit), и признак наличия записей за ней.
func (re *StockPostgres) FetchByLocationIDs(
	ctx context.Context,
	locationIDs []uint64,
	stockCursor mrstorage.IDCursor,
) (rows []dto.LocationStock, hasNext bool, err error) {
	sql := `
		SELECT
			location_id,
			container_id,
		    container_quantity,
		    container_volume
		FROM
    		` + module.DBTableNameStocks + `
		WHERE
    		stock_id > $1 AND location_id = ANY($2)
		ORDER BY
			stock_id
		FETCH FIRST ` + strconv.Itoa(stockCursor.Limit+1) + ` ROWS ONLY;`

	cursor, err := re.client.Conn(ctx).Query(
		ctx,
		sql,
		stockCursor.AfterID(),
		locationIDs,
	)
	if err != nil {
		return nil, false, err
	}

	defer cursor.Close()

	for cursor.Next() {
		// лишняя строка сверх limit лишь сообщает, что за текущей страницей есть записи
		if len(rows) == stockCursor.Limit {
			hasNext = true

			break
		}

		var row dto.LocationStock

		err = cursor.Scan(
			&row.LocationID,
			&row.ContainerID,
			&row.ContainerQuantity,
			&row.ContainerVolume,
		)
		if err != nil {
			return nil, false, err
		}

		if rows == nil {
			rows = make([]dto.LocationStock, 0, stockCursor.Limit)
		}

		rows = append(rows, row)
	}

	return rows, hasNext, cursor.Err()
}

// FetchContainerIDsWithStocks - возвращает только те ID контейнеров, у которых имеются стоки.
// -- при удалении последнего стока, можно отправлять событие, на удаление контейнера
// -- после этого проверять, если контейнер не привязан, то его удалять
// -- проверка, числится ли где-то указанные контейнеры на складе, если да, удалить:
// containerIDs = slices.DeleteFunc(
// containerIDs,
//
//	func(n uint64) bool {
//		return n == containerID
//	},
//
// )
// TODO: заюзать.
func (re *StockPostgres) FetchContainerIDsWithStocks(ctx context.Context, candidateContainerIDs []uint64) (ids []uint64, err error) {
	sql := `
		SELECT
		    container_id -- контейнеры, которые удалять не нужно
		FROM
    		` + module.DBTableNameStocks + `
		WHERE
    		container_id = ANY($1)
		GROUP BY
    		container_id;`

	cursor, err := re.client.Conn(ctx).Query(
		ctx,
		sql,
		candidateContainerIDs,
	)
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	for cursor.Next() {
		var containerID uint64

		err = cursor.Scan(
			&containerID,
		)
		if err != nil {
			return nil, err
		}

		if ids == nil {
			ids = make([]uint64, 0, len(candidateContainerIDs))
		}

		ids = append(ids, containerID)
	}

	return ids, cursor.Err()
}
