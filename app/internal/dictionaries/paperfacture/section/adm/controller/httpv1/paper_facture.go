package httpv1

import (
	"net/http"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/paperfacture/module"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm/entity"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/transport/model"
	admvalidate "print-shop-back/pkg/transport/validate/adm"
)

const (
	paperFactureListURL             = "/v1/dictionaries/paper-factures"
	paperFactureItemURL             = "/v1/dictionaries/paper-factures/{id}"
	paperFactureItemChangeStatusURL = "/v1/dictionaries/paper-factures/{id}/status"
)

type (
	// PaperFacture - comment struct.
	PaperFacture struct {
		parser       admvalidate.RequestParser
		sender       mrserver.ResponseSender
		useCase      adm.PaperFactureUseCase
		listSorter   mrtype.ListSorter
		errorWrapper errors.CustomWrapper
	}
)

// NewPaperFacture - создаёт контроллер PaperFacture.
func NewPaperFacture(
	parser admvalidate.RequestParser,
	sender mrserver.ResponseSender,
	useCase adm.PaperFactureUseCase,
	listSorter mrtype.ListSorter,
) *PaperFacture {
	return &PaperFacture{
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

// Handlers - возвращает обработчики контроллера PaperFacture.
func (ht *PaperFacture) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: paperFactureListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: paperFactureListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: paperFactureItemURL, Func: ht.Get},
		{Method: http.MethodPut, URL: paperFactureItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: paperFactureItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: paperFactureItemChangeStatusURL, Func: ht.ChangeStatus},
	}
}

// GetList - comment method.
func (ht *PaperFacture) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		PaperFactureListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *PaperFacture) listParams(r *http.Request) entity.PaperFactureParams {
	return entity.PaperFactureParams{
		Filter: entity.PaperFactureListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *PaperFacture) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *PaperFacture) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreatePaperFactureRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperFacture{
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
func (ht *PaperFacture) Save(w http.ResponseWriter, r *http.Request) error {
	req := StorePaperFactureRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperFacture{
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
func (ht *PaperFacture) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.PaperFacture{
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
func (ht *PaperFacture) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return err
	}

	return ht.sender.SendNoContent(w)
}

func (ht *PaperFacture) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *PaperFacture) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *PaperFacture) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return api.ErrPaperFactureNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
