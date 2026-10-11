package model

import (
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
)

type (
	// StoreListResponse - страница списка мест хранения.
	StoreListResponse struct {
		Items  []entity.Store `json:"stores"`
		Cursor string         `json:"cursor,omitempty"`
	}
)
