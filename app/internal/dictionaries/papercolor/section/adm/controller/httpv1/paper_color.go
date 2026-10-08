package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/papercolor/module"
	"print-shop-back/internal/dictionaries/papercolor/section/adm"
	"print-shop-back/internal/dictionaries/papercolor/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	paperColorListURL             = "/v1/dictionaries/paper-colors"
	paperColorItemURL             = "/v1/dictionaries/paper-colors/{id}"
	paperColorItemChangeStatusURL = "/v1/dictionaries/paper-colors/{id}/status"
)

type (
	// PaperColor - comment struct.
	PaperColor struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.PaperColorUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewPaperColor - создаёт контроллер PaperColor.
func NewPaperColor(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.PaperColorUseCase,
	listSorter mrtype.ListSorter,
) *PaperColor {
	return &PaperColor{
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

// Handlers - возвращает обработчики контроллера PaperColor.
func (ht *PaperColor) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: paperColorListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: paperColorListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: paperColorItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: paperColorItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: paperColorItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: paperColorItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *PaperColor) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		PaperColorListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *PaperColor) listParams(r *http.Request) entity.PaperColorParams {
	return entity.PaperColorParams{
		Filter: entity.PaperColorListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *PaperColor) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *PaperColor) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreatePaperColorRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperColor{
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
func (ht *PaperColor) Save(w http.ResponseWriter, r *http.Request) error {
	req := StorePaperColorRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperColor{
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
func (ht *PaperColor) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperColor{
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
func (ht *PaperColor) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return err
	}

	return ht.sender.SendNoContent(w)
}

func (ht *PaperColor) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *PaperColor) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *PaperColor) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return api.ErrPaperColorNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
