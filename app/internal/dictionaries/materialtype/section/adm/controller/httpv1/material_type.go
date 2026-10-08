package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/materialtype/module"
	"print-shop-back/internal/dictionaries/materialtype/section/adm"
	"print-shop-back/internal/dictionaries/materialtype/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	materialTypeListURL             = "/v1/dictionaries/material-types"
	materialTypeItemURL             = "/v1/dictionaries/material-types/{id}"
	materialTypeItemChangeStatusURL = "/v1/dictionaries/material-types/{id}/status"
)

type (
	// MaterialType - comment struct.
	MaterialType struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.MaterialTypeUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewMaterialType - создаёт контроллер MaterialType.
func NewMaterialType(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.MaterialTypeUseCase,
	listSorter mrtype.ListSorter,
) *MaterialType {
	return &MaterialType{
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

// Handlers - возвращает обработчики контроллера MaterialType.
func (ht *MaterialType) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: materialTypeListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: materialTypeListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: materialTypeItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: materialTypeItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: materialTypeItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: materialTypeItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *MaterialType) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		MaterialTypeListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *MaterialType) listParams(r *http.Request) entity.MaterialTypeParams {
	return entity.MaterialTypeParams{
		Filter: entity.MaterialTypeListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *MaterialType) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *MaterialType) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateMaterialTypeRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.MaterialType{
		Caption: req.Caption,
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
func (ht *MaterialType) Save(w http.ResponseWriter, r *http.Request) error {
	req := StoreMaterialTypeRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.MaterialType{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		Caption:    req.Caption,
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *MaterialType) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.MaterialType{
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
func (ht *MaterialType) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return err
	}

	return ht.sender.SendNoContent(w)
}

func (ht *MaterialType) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *MaterialType) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *MaterialType) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return api.ErrMaterialTypeNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
