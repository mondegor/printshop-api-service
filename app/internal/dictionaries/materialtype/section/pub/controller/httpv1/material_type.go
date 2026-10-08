package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/materialtype/section/pub"
	"print-shop-back/internal/dictionaries/materialtype/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	materialTypeListURL = "/v1/dictionaries/material-types"
)

type (
	// MaterialType - comment struct.
	MaterialType struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.MaterialTypeUseCase
	}
)

// NewMaterialType - создаёт контроллер MaterialType.
func NewMaterialType(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.MaterialTypeUseCase) *MaterialType {
	return &MaterialType{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера MaterialType.
func (ht *MaterialType) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: materialTypeListURL, Func: ht.GetList},
	}
}

// GetList - comment method.
func (ht *MaterialType) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.parser.Localizer(r), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *MaterialType) listParams(_ *http.Request) entity.MaterialTypeParams {
	return entity.MaterialTypeParams{}
}
