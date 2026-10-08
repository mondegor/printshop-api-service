package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/laminate/section/pub"
	"print-shop-back/internal/catalog/laminate/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	laminateListURL          = "/v1/catalog/laminates"
	laminateTypeListURL      = "/v1/catalog/laminates/types"
	laminateThicknessListURL = "/v1/catalog/laminates/thicknesses"
)

type (
	// Laminate - comment struct.
	Laminate struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.LaminateUseCase
	}
)

// NewLaminate - создаёт контроллер Laminate.
func NewLaminate(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.LaminateUseCase) *Laminate {
	return &Laminate{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера Laminate.
func (ht *Laminate) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: laminateListURL, Func: ht.GetList},
		{Method: http.MethodGet, URL: laminateTypeListURL, Func: ht.GetTypeList},
		{Method: http.MethodGet, URL: laminateThicknessListURL, Func: ht.GetThicknessList},
	}
}

// GetList - comment method.
func (ht *Laminate) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.parser.Localizer(r), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *Laminate) listParams(_ *http.Request) entity.LaminateParams {
	return entity.LaminateParams{}
}

// GetTypeList - comment method.
func (ht *Laminate) GetTypeList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetTypeList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

// GetThicknessList - comment method.
func (ht *Laminate) GetThicknessList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetThicknessList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}
