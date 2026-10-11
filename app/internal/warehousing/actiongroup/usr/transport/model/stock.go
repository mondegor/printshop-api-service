package model

import (
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
)

type (
	// AddMoreStockRequest - запрос на прибавление остатков контейнеру на месте хранения.
	AddMoreStockRequest struct {
		ContainerID uint64 `json:"container_id"`
		LocationID  uint64 `json:"location_id"`
		Quantity    int    `json:"quantity"`
	}

	// MoveStockRequest - запрос на перемещение остатков контейнера на другое место хранения.
	MoveStockRequest struct {
		StockID    uint64 `json:"stock_id"`
		LocationID uint64 `json:"location_id"`
		// TODO: проверять, что Quantity >= 0 (в контракте minimum: 0), сейчас отрицательное значение доходит до usecase.
		Quantity int `json:"quantity,omitempty"`
	}

	// TransferStockRequest - запрос на списание остатков.
	TransferStockRequest struct {
		StockID  uint64 `json:"stock_id"`
		Quantity int    `json:"quantity"`
	}

	// StockListResponse - страница списка остатков контейнеров.
	StockListResponse struct {
		Items  []entity.Stock `json:"stocks"`
		Cursor string         `json:"cursor,omitempty"`
	}
)
