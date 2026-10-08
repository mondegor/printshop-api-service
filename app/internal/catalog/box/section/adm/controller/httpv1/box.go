package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/box/module"
	"print-shop-back/internal/catalog/box/section/adm"
	"print-shop-back/internal/catalog/box/section/adm/entity"
	"print-shop-back/pkg/mrcalc/measure"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	boxListURL             = "/v1/catalog/boxes"
	boxItemURL             = "/v1/catalog/boxes/{id}"
	boxItemChangeStatusURL = "/v1/catalog/boxes/{id}/status"
)

type (
	// Box - comment struct.
	Box struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.BoxUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewBox - создаёт контроллер Box.
func NewBox(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.BoxUseCase,
	listSorter mrtype.ListSorter,
) *Box {
	return &Box{
		parser:     parser,
		sender:     sender,
		useCase:    useCase,
		listSorter: listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
			module.ErrBoxArticleAlreadyExists.Code(), "article",
		),
	}
}

// Handlers - возвращает обработчики контроллера Box.
func (ht *Box) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: boxListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: boxListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: boxItemURL, Func: ht.Get},
		{Method: http.MethodPatch, URL: boxItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: boxItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: boxItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *Box) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		BoxListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *Box) listParams(r *http.Request) entity.BoxParams {
	return entity.BoxParams{
		Filter: entity.BoxListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Length:     measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterLengthRange).Transform(measure.OneThousandth)),    // mm -> m
			Width:      measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterWidthRange).Transform(measure.OneThousandth)),     // mm -> m
			Height:     measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterHeightRange).Transform(measure.OneThousandth)),    // mm -> m
			Weight:     measure.RangeKilogram(ht.parser.FilterRangeInt64(r, module.ParamNameFilterWeightRange).Transform(measure.OneThousandth)), // g -> kg
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *Box) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *Box) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateBoxRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Box{
		Article: req.Article,
		Caption: req.Caption,
		Length:  measure.Meter(req.Length * measure.OneThousandth),
		Width:   measure.Meter(req.Width * measure.OneThousandth),
		Height:  measure.Meter(req.Height * measure.OneThousandth),
		Weight:  measure.Kilogram(req.Weight * measure.OneThousandth),
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
func (ht *Box) Save(w http.ResponseWriter, r *http.Request) error {
	req := StoreBoxRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Box{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Article:    req.Article,
		Caption:    req.Caption,
		Length:     measure.Meter(req.Length * measure.OneThousandth),
		Width:      measure.Meter(req.Width * measure.OneThousandth),
		Height:     measure.Meter(req.Height * measure.OneThousandth),
		Weight:     measure.Kilogram(req.Weight * measure.OneThousandth),
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *Box) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Box{
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
func (ht *Box) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *Box) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *Box) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *Box) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrBoxNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
