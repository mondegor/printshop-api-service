package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/paper/module"
	"print-shop-back/internal/catalog/paper/section/adm"
	"print-shop-back/internal/catalog/paper/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/mrcalc/measure"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	paperListURL             = "/v1/catalog/papers"
	paperItemURL             = "/v1/catalog/papers/{id}"
	paperItemChangeStatusURL = "/v1/catalog/papers/{id}/status"
)

type (
	// Paper - comment struct.
	Paper struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.PaperUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewPaper - создаёт контроллер Paper.
func NewPaper(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.PaperUseCase,
	listSorter mrtype.ListSorter,
) *Paper {
	return &Paper{
		parser:     parser,
		sender:     sender,
		useCase:    useCase,
		listSorter: listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
			module.ErrPaperArticleAlreadyExists.Code(), "article",
			api.ErrMaterialTypeRequired.Code(), "typeId",
			api.ErrMaterialTypeNotFound.Code(), "typeId",
			api.ErrPaperColorRequired.Code(), "colorId",
			api.ErrPaperColorNotFound.Code(), "colorId",
			api.ErrPaperFactureRequired.Code(), "factureId",
			api.ErrPaperFactureNotFound.Code(), "factureId",
		),
	}
}

// Handlers - возвращает обработчики контроллера Paper.
func (ht *Paper) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: paperListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: paperListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: paperItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: paperItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: paperItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: paperItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *Paper) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		PaperListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *Paper) listParams(r *http.Request) entity.PaperParams {
	return entity.PaperParams{
		Filter: entity.PaperListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			TypeIDs:    ht.parser.FilterUint64List(r, module.ParamNameFilterCatalogPaperTypeIDs),
			ColorIDs:   ht.parser.FilterUint64List(r, module.ParamNameFilterCatalogPaperColorIDs),
			FactureIDs: ht.parser.FilterUint64List(r, module.ParamNameFilterCatalogPaperFactureIDs),
			Width:      measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterLengthRange).Transform(measure.OneThousandth)), // mm -> m
			Height:     measure.RangeMeter(ht.parser.FilterRangeInt64(r, module.ParamNameFilterHeightRange).Transform(measure.OneThousandth)), // mm -> m
			Density: measure.RangeKilogramPerMeter2(
				ht.parser.FilterRangeInt64(r, module.ParamNameFilterDensityRange).Transform(measure.OneThousandth), // g/m2 -> kg/m2
			),
			Statuses: ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *Paper) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *Paper) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreatePaperRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Paper{
		Article:   req.Article,
		Caption:   req.Caption,
		TypeID:    req.TypeID,
		ColorID:   req.ColorID,
		FactureID: req.FactureID,
		Width:     measure.Meter(req.Length * measure.OneThousandth),
		Height:    measure.Meter(req.Height * measure.OneThousandth),
		Thickness: measure.Meter(req.Thickness * measure.OneMillionth),
		Density:   measure.KilogramPerMeter2(req.Density * measure.OneThousandth),
		Sides:     req.Sides,
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
func (ht *Paper) Save(w http.ResponseWriter, r *http.Request) error {
	req := StorePaperRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Paper{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Article:    req.Article,
		Caption:    req.Caption,
		TypeID:     req.TypeID,
		ColorID:    req.ColorID,
		FactureID:  req.FactureID,
		Width:      measure.Meter(req.Length * measure.OneThousandth),
		Height:     measure.Meter(req.Height * measure.OneThousandth),
		Thickness:  measure.Meter(req.Thickness * measure.OneMillionth),
		Density:    measure.KilogramPerMeter2(req.Density * measure.OneThousandth),
		Sides:      req.Sides,
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *Paper) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.Paper{
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
func (ht *Paper) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *Paper) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *Paper) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *Paper) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrPaperNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
