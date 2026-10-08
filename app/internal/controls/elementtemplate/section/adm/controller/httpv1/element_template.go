package httpv1

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/mondegor/go-core/errors"
	mrmodel "github.com/mondegor/go-core/mrmodel/media"
	"github.com/mondegor/go-core/mrtype"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/controls/elementtemplate/module"
	"print-shop-back/internal/controls/elementtemplate/section/adm"
	"print-shop-back/internal/controls/elementtemplate/section/adm/entity"
	"print-shop-back/internal/controls/elementtemplate/section/adm/validate"
	"print-shop-back/pkg/controls/api"
	"print-shop-back/pkg/transport/model"
)

const (
	elementListTemplateURL             = "/v1/controls/element-templates"
	elementTemplateItemURL             = "/v1/controls/element-templates/{id}"
	elementTemplateItemChangeStatusURL = "/v1/controls/element-templates/{id}/status"
	elementTemplateItemJsonURL         = "/v1/controls/element-templates/{id}/json"

	maxElementTemplateSize = 512 * 1024 // 5120 Kb limit
)

type (
	// ElementTemplate - comment struct.
	ElementTemplate struct {
		parser           validate.RequestElementTemplateParser
		sender           mrserver.FileResponseSender
		useCase          adm.ElementTemplateUseCase
		listSorter       mrtype.ListSorter
		errorWrapper     errors.CustomWrapper
		fileErrorWrapper errors.CustomWrapper
	}
)

// NewElementTemplate - создаёт контроллер ElementTemplate.
func NewElementTemplate(
	parser validate.RequestElementTemplateParser,
	sender mrserver.FileResponseSender,
	useCase adm.ElementTemplateUseCase,
	listSorter mrtype.ListSorter,
) *ElementTemplate {
	return &ElementTemplate{
		parser:     parser,
		sender:     sender,
		useCase:    useCase,
		listSorter: listSorter,
		errorWrapper: errors.NewCustomWrapper(
			errors.ErrRecordVersionConflict.Code(), "tagVersion",
			errors.ErrSwitchStatusRejected.Code(), "status",
		),
		fileErrorWrapper: errors.NewDownloadFileWrapper(module.ParamNameElementTemplateAttachment),
	}
}

// Handlers - возвращает обработчики контроллера ElementTemplate.
func (ht *ElementTemplate) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: elementListTemplateURL, Func: ht.GetList},
		{Method: http.MethodPost, URL: elementListTemplateURL, Func: ht.Create},

		{Method: http.MethodGet, URL: elementTemplateItemURL, Func: ht.Get},
		{Method: http.MethodPatch, URL: elementTemplateItemURL, Func: ht.Save},
		{Method: http.MethodDelete, URL: elementTemplateItemURL, Func: ht.Remove},

		{Method: http.MethodPatch, URL: elementTemplateItemChangeStatusURL, Func: ht.ChangeStatus},

		{Method: http.MethodGet, URL: elementTemplateItemJsonURL, Func: ht.GetJson},
	}
}

// GetList - comment method.
func (ht *ElementTemplate) GetList(w http.ResponseWriter, r *http.Request) error {
	items, totalItems, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		ElementTemplateListResponse{
			Items: items,
			Total: totalItems,
		},
	)
}

func (ht *ElementTemplate) listParams(r *http.Request) entity.ElementTemplateParams {
	return entity.ElementTemplateParams{
		Filter: entity.ElementTemplateListFilter{
			SearchText: ht.parser.FilterString(r, module.ParamNameFilterSearchText),
			Detailing:  ht.parser.FilterElementDetailingList(r, module.ParamNameFilterElementDetailing),
			Statuses:   ht.parser.FilterStatusList(r, module.ParamNameFilterStatuses),
		},
		Sorter: ht.parser.SortParams(r, ht.listSorter),
		Pager:  ht.parser.PageParams(r),
	}
}

// Get - comment method.
func (ht *ElementTemplate) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// GetJson - comment method.
func (ht *ElementTemplate) GetJson(w http.ResponseWriter, r *http.Request) error {
	itemID := ht.getItemID(r)

	body, err := ht.useCase.GetItemJson(r.Context(), itemID, true)
	if err != nil {
		return err
	}

	return ht.sender.SendAttachmentFile(
		r.Context(),
		w,
		mrmodel.File{
			FileInfo: mrmodel.FileInfo{
				ContentType:  "application/json",
				OriginalName: fmt.Sprintf(module.JsonFileNamePattern, itemID),
				Size:         int64(len(body)),
			},
			Body: io.NopCloser(bytes.NewReader(body)),
		},
	)
}

// Create - comment method.
func (ht *ElementTemplate) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateElementTemplateRequest{}
	r.Body = http.MaxBytesReader(w, r.Body, maxElementTemplateSize)
	rawElementTemplate := []byte(r.FormValue(module.ParamNameElementTemplateObject))

	if err := ht.parser.ValidateContent(r.Context(), rawElementTemplate, &req); err != nil {
		return err
	}

	file, err := ht.parser.FormFileContent(r, module.ParamNameElementTemplateAttachment)
	if err != nil {
		return ht.fileErrorWrapper.Wrap(err)
	}

	item := entity.ElementTemplate{
		ParamName: req.ParamName,
		Caption:   req.Caption,
		Type:      req.Type,
		Detailing: req.Detailing,
		Body:      file.Body,
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
func (ht *ElementTemplate) Save(w http.ResponseWriter, r *http.Request) error {
	req := StoreElementTemplateRequest{}
	r.Body = http.MaxBytesReader(w, r.Body, maxElementTemplateSize)
	rawElementTemplate := []byte(r.FormValue(module.ParamNameElementTemplateObject))

	if err := ht.parser.ValidateContent(r.Context(), rawElementTemplate, &req); err != nil {
		return err
	}

	file, err := ht.parser.FormFileContent(r, module.ParamNameElementTemplateAttachment)
	if err != nil {
		// указывать файл необязательно
		if !errors.Is(err, errors.ErrHttpFileUpload) {
			return ht.fileErrorWrapper.Wrap(err)
		}

		file.Body = nil
	}

	item := entity.ElementTemplate{
		ID:         ht.getItemID(r),
		TagVersion: req.TagVersion,
		ParamName:  req.ParamName,
		Caption:    req.Caption,
		Body:       file.Body,
	}

	if err := ht.useCase.Save(r.Context(), item); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

// ChangeStatus - comment method.
func (ht *ElementTemplate) ChangeStatus(w http.ResponseWriter, r *http.Request) error {
	req := model.ChangeItemStatusRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.ElementTemplate{
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
func (ht *ElementTemplate) Remove(w http.ResponseWriter, r *http.Request) error {
	if err := ht.useCase.Remove(r.Context(), ht.getItemID(r)); err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.SendNoContent(w)
}

func (ht *ElementTemplate) getItemID(r *http.Request) uint64 {
	return ht.parser.PathParamUint64(r, "id")
}

func (ht *ElementTemplate) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *ElementTemplate) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return api.ErrElementTemplateNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return ht.errorWrapper.Wrap(err)
}
