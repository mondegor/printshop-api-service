package httpv1

import (
	"context"
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/warehousing/actiongroup/usr"
	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/actiongroup/usr/transport/model"
	"print-shop-back/internal/warehousing/module"
	pkgmodel "print-shop-back/pkg/transport/model"
	usrvalidate "print-shop-back/pkg/transport/validate/usr"
)

const (
	stockListURL     = "/v1/warehousing/stocks"
	stockAddMoreURL  = "/v1/warehousing/stocks/add-more"
	stockMoveURL     = "/v1/warehousing/stocks/move"
	stockTransferURL = "/v1/warehousing/stocks/transfer"
)

type (
	// Stock - comment struct.
	Stock struct {
		parser               usrvalidate.RequestParser
		sender               mrserver.ResponseSender
		serviceStock         usr.StockService
		useCaseMoveStock     moveStockUseCase
		useCaseAddMoreStock  addStockUseCase
		useCaseTransferStock transferStockUseCase
	}

	moveStockUseCase interface {
		Execute(ctx context.Context, item dto.MoveStockContainer) (movedStockID uint64, err error)
	}

	addStockUseCase interface {
		Execute(ctx context.Context, item dto.AddMoreStockContainer) (stockID uint64, err error)
	}

	transferStockUseCase interface {
		Execute(ctx context.Context, item dto.TransferStockContainers) error
	}
)

// NewStock - создаёт контроллер Stock.
func NewStock(
	parser usrvalidate.RequestParser,
	sender mrserver.ResponseSender,
	serviceStock usr.StockService,
	useCaseMoveStock moveStockUseCase,
	useCaseAddStock addStockUseCase,
	useCaseTransferStock transferStockUseCase,
) *Stock {
	return &Stock{
		parser:               parser,
		sender:               sender,
		serviceStock:         serviceStock,
		useCaseMoveStock:     useCaseMoveStock,
		useCaseAddMoreStock:  useCaseAddStock,
		useCaseTransferStock: useCaseTransferStock,
	}
}

// Handlers - возвращает обработчики контроллера Stock.
func (ht *Stock) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: stockListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: stockAddMoreURL, Func: ht.AddMore},
		{Method: http.MethodPost, URL: stockMoveURL, Func: ht.Move},
		{Method: http.MethodPost, URL: stockTransferURL, Func: ht.Transfer},
	}
}

// GetList - comment method.
func (ht *Stock) GetList(w http.ResponseWriter, r *http.Request) error {
	items, hasNext, err := ht.serviceStock.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		model.StockListResponse{
			Items:   items,
			Cursor:  entity.CreateStockCursorValue(items),
			HasNext: hasNext,
		},
	)
}

func (ht *Stock) listParams(r *http.Request) dto.StockParams {
	return dto.StockParams{
		AccountID: ht.parser.UserID(r),
		Filter: dto.StockListFilter{
			SearchContainers: ht.parser.FilterUint64List(r, module.ParamNameFilterSearchStockContainerIDs),
			SearchLocations:  ht.parser.FilterUint64List(r, module.ParamNameFilterSearchStockLocationIDs),
		},
		Cursor: mrstorage.NewIDCursor(ht.parser.CursorParams(r)),
	}
}

// AddMore - comment method.
func (ht *Stock) AddMore(w http.ResponseWriter, r *http.Request) error {
	req := model.AddMoreStockRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := dto.AddMoreStockContainer{
		AccountID:   ht.parser.UserID(r),
		ContainerID: req.ContainerID,
		LocationID:  req.LocationID,
		Quantity:    req.Quantity,
	}

	itemID, err := ht.useCaseAddMoreStock.Execute(r.Context(), item)
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(
		w,
		http.StatusCreated,
		pkgmodel.SuccessCreatedItemUintResponse{
			ItemID: itemID,
		},
	)
}

// Move - comment method.
func (ht *Stock) Move(w http.ResponseWriter, r *http.Request) error {
	req := model.MoveStockRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := dto.MoveStockContainer{
		AccountID:  ht.parser.UserID(r),
		StockID:    req.StockID,
		LocationID: req.LocationID,
		Quantity:   req.Quantity,
	}

	itemID, err := ht.useCaseMoveStock.Execute(r.Context(), item)
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(
		w,
		http.StatusCreated,
		pkgmodel.SuccessCreatedItemUintResponse{
			ItemID: itemID,
		},
	)
}

// Transfer - comment method.
func (ht *Stock) Transfer(w http.ResponseWriter, r *http.Request) error {
	req := model.TransferStockRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := dto.TransferStockContainers{
		AccountID: ht.parser.UserID(r),
		Stocks: []dto.StockQuantity{
			{
				StockID:  req.StockID,
				Quantity: req.Quantity,
			},
		},
	}

	if err := ht.useCaseTransferStock.Execute(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *Stock) getItemValue(r *http.Request) string {
	return ht.parser.PathParamString(r, "value")
}

func (ht *Stock) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrStockNotFound.Wrap(err, ht.getItemValue(r))
	}

	return err
}
