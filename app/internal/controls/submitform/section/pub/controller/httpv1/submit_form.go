package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/controls/submitform/section/pub"
	"print-shop-back/internal/controls/submitform/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	submitFormListURL = "/v1/controls/submit-forms"
	submitFormItemURL = "/v1/controls/submit-forms/{rewriteName}"
)

type (
	// SubmitForm - comment struct.
	SubmitForm struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.SubmitFormUseCase
	}
)

// NewSubmitForm - создаёт контроллер SubmitForm.
func NewSubmitForm(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.SubmitFormUseCase) *SubmitForm {
	return &SubmitForm{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера SubmitForm.
func (ht *SubmitForm) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: submitFormListURL, Func: ht.GetList},
		{Method: http.MethodGet, URL: submitFormItemURL, Func: ht.Get},
	}
}

// GetList - comment method.
func (ht *SubmitForm) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *SubmitForm) listParams(_ *http.Request) entity.SubmitFormParams {
	return entity.SubmitFormParams{}
}

// Get - comment method.
func (ht *SubmitForm) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItemByRewriteName(r.Context(), ht.parser.PathParamString(r, "rewriteName"))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, item)
}
