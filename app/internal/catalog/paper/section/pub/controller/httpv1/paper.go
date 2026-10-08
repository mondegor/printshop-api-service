package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/paper/section/pub"
	"print-shop-back/internal/catalog/paper/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	paperListURL        = "/v1/catalog/papers"
	paperTypeListURL    = "/v1/catalog/papers/types"
	paperColorListURL   = "/v1/catalog/papers/colors"
	paperDensityListURL = "/v1/catalog/papers/densities"
	paperFactureListURL = "/v1/catalog/papers/factures"
)

type (
	// Paper - comment struct.
	Paper struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.PaperUseCase
	}
)

// NewPaper - создаёт контроллер Paper.
func NewPaper(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.PaperUseCase) *Paper {
	return &Paper{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера Paper.
func (ht *Paper) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: paperListURL, Func: ht.GetList},
		{Method: http.MethodGet, URL: paperTypeListURL, Func: ht.GetTypeList},
		{Method: http.MethodGet, URL: paperColorListURL, Func: ht.GetColorList},
		{Method: http.MethodGet, URL: paperDensityListURL, Func: ht.GetDensityList},
		{Method: http.MethodGet, URL: paperFactureListURL, Func: ht.GetFactureList},
	}
}

// GetList - comment method.
func (ht *Paper) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.parser.Localizer(r), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *Paper) listParams(_ *http.Request) entity.PaperParams {
	return entity.PaperParams{}
}

// GetTypeList - comment method.
func (ht *Paper) GetTypeList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetTypeList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

// GetColorList - comment method.
func (ht *Paper) GetColorList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetColorList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

// GetDensityList - comment method.
func (ht *Paper) GetDensityList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetDensityList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

// GetFactureList - comment method.
func (ht *Paper) GetFactureList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetFactureList(r.Context())
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}
