package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/printformat/module"
	"print-shop-back/internal/dictionaries/printformat/section/adm"
	"print-shop-back/internal/dictionaries/printformat/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/mrcalc/measure"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	printFormatListURL             = "/v1/dictionaries/print-formats"
	printFormatItemURL             = "/v1/dictionaries/print-formats/{id}"
	printFormatItemChangeStatusURL = "/v1/dictionaries/print-formats/{id}/status"
)

type (
	// PrintFormat - comment struct.
	PrintFormat struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.PrintFormatUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewPrintFormat - создаёт контроллер PrintFormat.
func NewPrintFormat(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.PrintFormatUseCase,
	listSorter mrtype.ListSorter,
) *PrintFormat {
	return &PrintFormat{
		parser:     parser,
		sender:     sender,
		useCase:    useCase,
		listSorter: listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
		),
	}
}

// Handlers - возвращает обработчики контроллера PrintFormat.
func (ht *PrintFormat) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: printFormatListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: printFormatListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: printFormatItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: printFormatItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: printFormatItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: printFormatItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *PrintFormat) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		PrintFormatListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *PrintFormat) listParams(r *http.Request) entity.PrintFormatParams {
	return entity.PrintFormatParams{
		Filter: entity.PrintFormatListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Width:      measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterWidthRange).Transform(measure.OneThousandth)),  // mm -> m
			Height:     measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterHeightRange).Transform(measure.OneThousandth)), // mm -> m
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *PrintFormat) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *PrintFormat) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreatePrintFormatRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PrintFormat{
		Caption: req.Caption,
		Width:   measure.Meter(req.Width * measure.OneThousandth),
		Height:  measure.Meter(req.Height * measure.OneThousandth),
	}

	itemID, err := ht.useCase.Create(r.Context(), item)
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(
		w,
		http.StatusCreated,
		model.SuccessCreatedItemUintResponse{
			ItemID: itemID,
		},
	)
}

// Save - comment method.
func (ht *PrintFormat) Save(w http.ResponseWriter, r *http.Request) error {
	req := StorePrintFormatRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PrintFormat{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Caption:    req.Caption,
		Width:      measure.Meter(req.Width * measure.OneThousandth),
		Height:     measure.Meter(req.Height * measure.OneThousandth),
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *PrintFormat) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PrintFormat{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Status:     req.Status,
	}

	if err := ht.useCase.ChangeStatus(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// Remove - comment method.
func (ht *PrintFormat) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return err
	}

	return ht.sender.SendNoContent(w)
}

func (ht *PrintFormat) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *PrintFormat) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *PrintFormat) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return api.ErrPrintFormatNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
