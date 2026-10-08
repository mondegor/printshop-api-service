package httpv1

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/errors"
	mrmodel "github.com/mondegor/go-core/mrmodel/media"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/controls/submitform/module"
	"print-shop-back/internal/controls/submitform/section/adm"
	"print-shop-back/internal/controls/submitform/section/adm/entity"
	"print-shop-back/internal/controls/submitform/section/adm/validate"
	"print-shop-back/pkg/transport/model"
)

const (
	submitFormListURL             = "/v1/controls/submit-forms"
	submitFormItemURL             = "/v1/controls/submit-forms/{id}"
	submitFormItemChangeStatusURL = "/v1/controls/submit-forms/{id}/status"

	submitFormVersionItemJsonURL    = "/v1/controls/submit-forms/{id}/versions/{version}/json"
	submitFormItemPrepareForTestURL = "/v1/controls/submit-forms/{id}/prepare-for-test"
	submitFormItemPublishURL        = "/v1/controls/submit-forms/{id}/publish"
)

type (
	// SubmitForm - comment struct.
	SubmitForm struct {
		parser         validate.RequestSubmitFormParser
		sender         mrserver.FileResponseSender
		useCase        adm.SubmitFormUseCase
		useCaseVersion adm.FormVersionUseCase
		listSorter     mrtype.ListSorter
		errorWrapper   errors.CustomWrapper
	}
)

// NewSubmitForm - создаёт контроллер SubmitForm.
func NewSubmitForm(
	parser validate.RequestSubmitFormParser,
	sender mrserver.FileResponseSender,
	useCase adm.SubmitFormUseCase,
	useCaseVersion adm.FormVersionUseCase,
	listSorter mrtype.ListSorter,
) *SubmitForm {
	return &SubmitForm{
		parser:         parser,
		sender:         sender,
		useCase:        useCase,
		useCaseVersion: useCaseVersion,
		listSorter:     listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
			module.ErrSubmitFormRewriteNameAlreadyExists.Code(), "rewriteName",
			module.ErrSubmitFormParamNameAlreadyExists.Code(), "paramName",
		),
	}
}

// Handlers - возвращает обработчики контроллера SubmitForm.
func (ht *SubmitForm) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: submitFormListURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: submitFormListURL, Func: ht.Create},

		{Method: http.MethodGet, URL: submitFormItemURL, Func: ht.Get},
		{Method: http.MethodPatch, URL: submitFormItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: submitFormItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: submitFormItemChangeStatusURL, Func: ht.ChangeStatus},

		{Method: http.MethodGet, URL: submitFormVersionItemJsonURL, Func: ht.GetVersionJson},
		{Method: http.MethodPatch, URL: submitFormItemPrepareForTestURL, Func: ht.PrepareForTest},
		{Method: http.MethodPatch, URL: submitFormItemPublishURL, Permission: "modControlsSubmitFormToPublish", Func: ht.Publish},
	}
}

// GetList - comment method.
func (ht *SubmitForm) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		SubmitFormListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *SubmitForm) listParams(r *http.Request) entity.SubmitFormParams {
	return entity.SubmitFormParams{
		Filter: entity.SubmitFormListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Detailing:  ht.parser.FilterElementDetailingList(r, module.ParamNameFilterElementDetailing),
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *SubmitForm) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *SubmitForm) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateSubmitFormRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.SubmitForm{
		RewriteName: req.RewriteName,
		ParamName:   req.ParamName,
		Caption:     req.Caption,
		Detailing:   req.Detailing,
	}

	itemID, err := ht.useCase.Create(r.Context(), item)
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(
		w,
		http.StatusCreated,
		model.SuccessCreatedItemResponse{
			ItemID: itemID.String(),
		},
	)
}

// Save - comment method.
func (ht *SubmitForm) Save(w http.ResponseWriter, r *http.Request) error {
	req := StoreSubmitFormRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.SubmitForm{
		ID:          ht.getItemID(r),
		TagVersion:  req.TagVersion,
		RewriteName: req.RewriteName,
		ParamName:   req.ParamName,
		Caption:     req.Caption,
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *SubmitForm) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.SubmitForm{
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
func (ht *SubmitForm) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return err
	}

	return ht.sender.SendNoContent(w)
}

// GetVersionJson - comment method.
func (ht *SubmitForm) GetVersionJson(w http.ResponseWriter, r *http.Request) error {
	primary := entity.PrimaryKey{
		FormID:  ht.getItemID(r),
		Version: ht.getItemVersion(r),
	}

	body, err := ht.useCaseVersion.GetItemJson(r.Context(), primary, true)
	if err != nil {
		return err
	}

	return ht.sender.SendAttachmentFile(
		r.Context(),
		w,
		mrmodel.File{
			FileInfo: mrmodel.FileInfo{
				ContentType:  "application/json",
				OriginalName: fmt.Sprintf(module.JsonFileNamePattern, primary.FormID, primary.Version),
				Size:         int64(len(body)),
			},
			Body: io.NopCloser(bytes.NewReader(body)),
		},
	)
}

// PrepareForTest - comment method.
func (ht *SubmitForm) PrepareForTest(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCaseVersion.PrepareForTest(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// Publish - comment method.
func (ht *SubmitForm) Publish(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCaseVersion.Publish(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *SubmitForm) getItemID(r *http.Request) uuid.UUID {
	return ht.parser.PathParamUUID(r, "id")
}

func (ht *SubmitForm) getItemVersion(r *http.Request) int32 {
	ver := ht.parser.PathParamUint64(r, "version")
	if ver > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(ver)
}

func (ht *SubmitForm) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *SubmitForm) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrSubmitFormNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
