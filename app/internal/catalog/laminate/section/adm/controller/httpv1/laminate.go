package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/laminate/module"
	"print-shop-back/internal/catalog/laminate/section/adm"
	"print-shop-back/internal/catalog/laminate/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/mrcalc/measure"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	laminateListURL             = "/v1/catalog/laminates"
	laminateItemURL             = "/v1/catalog/laminates/{id}"
	laminateItemChangeStatusURL = "/v1/catalog/laminates/{id}/status"
)

type (
	// Laminate - comment struct.
	Laminate struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.LaminateUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewLaminate - создаёт контроллер Laminate.
func NewLaminate(parser admvalidate.RequestParser, sender mrserver.ResponseSender, useCase adm.LaminateUseCase, listSorter mrtype.ListSorter) *Laminate {
	return &Laminate{
		parser:     parser,
		sender:     sender,
		useCase:    useCase,
		listSorter: listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
			module.ErrLaminateArticleAlreadyExists.Code(), "article",
			api.ErrMaterialTypeRequired.Code(), "typeId",
			api.ErrMaterialTypeNotFound.Code(), "typeId",
		),
	}
}

// Handlers - возвращает обработчики контроллера Laminate.
func (ht *Laminate) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: laminateListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: laminateListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: laminateItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: laminateItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: laminateItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: laminateItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *Laminate) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		LaminateListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *Laminate) listParams(r *http.Request) entity.LaminateParams {
	return entity.LaminateParams{
		Filter: entity.LaminateListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			TypeIDs:    ht.parser.FilterUint64List(r, module.ParamNameFilterCatalogMaterialTypeIDs),
			Length:     measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterLengthRange).Transform(measure.OneThousandth)), // mm -> m
			Width:      measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterWidthRange).Transform(measure.OneThousandth)),  // mm -> m
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *Laminate) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *Laminate) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateLaminateRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Laminate{
		Article:   req.Article,
		Caption:   req.Caption,
		TypeID:    req.TypeID,
		Length:    req.Length,
		Width:     measure.Meter(req.Width * measure.OneThousandth),
		Thickness: measure.Meter(req.Thickness * measure.OneMillionth),
		WeightM2:  measure.KilogramPerMeter2(req.WeightM2 * measure.OneThousandth),
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
func (ht *Laminate) Save(w http.ResponseWriter, r *http.Request) error {
	req := StoreLaminateRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Laminate{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Article:    req.Article,
		Caption:    req.Caption,
		TypeID:     req.TypeID,
		Length:     req.Length,
		Width:      measure.Meter(req.Width * measure.OneThousandth),
		Thickness:  measure.Meter(req.Thickness * measure.OneMillionth),
		WeightM2:   measure.KilogramPerMeter2(req.WeightM2 * measure.OneThousandth),
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *Laminate) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Laminate{
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
func (ht *Laminate) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *Laminate) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *Laminate) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *Laminate) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrLaminateNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
