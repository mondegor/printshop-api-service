package model

import (
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/xtype"
)

type (
	// CreateContainerRequest - запрос на создание контейнера.
	CreateContainerRequest struct {
		Code             string       `json:"code,omitempty"`
		Volume           xtype.Volume `json:"volume,omitempty"`
		Tags             []string     `json:"tags,omitempty"`
		Images           []string     `json:"images,omitempty"`
		LocationID       uint64       `json:"location_id"`
		ExemplarQuantity int          `json:"exemplar_quantity,omitempty"`
	}

	// CreateContainerGroupRequest - запрос на создание группы контейнеров.
	CreateContainerGroupRequest struct {
		Code       string       `json:"code,omitempty"`
		Volume     xtype.Volume `json:"volume"`
		LocationID uint64       `json:"location_id"`
	}

	// SaveContainerTagsRequest - запрос на обновление тегов контейнера.
	SaveContainerTagsRequest struct {
		ContainerID uint64   `json:"container_id"`
		TagVersion  uint32   `json:"tag_version" validate:"required,gte=1"`
		Tags        []string `json:"tags"`
	}

	// ContainerListResponse - страница списка контейнеров.
	ContainerListResponse struct {
		Items  []entity.Container `json:"containers"`
		Cursor string             `json:"cursor,omitempty"`
	}

	// SuccessCreatedContainerResponse - результат создания контейнера или группы контейнеров.
	SuccessCreatedContainerResponse struct {
		ContainerID uint64 `json:"container_id"`
		Code        string `json:"code"`
		Marker      uint16 `json:"marker"`
		StockID     uint64 `json:"stock_id"`
	}
)
