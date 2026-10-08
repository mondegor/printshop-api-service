package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/box/section/pub"
	"print-shop-back/internal/catalog/box/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	boxListURL = "/v1/catalog/boxes"
)

type (
	// Box - comment struct.
	Box struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.BoxUseCase
	}
)

// NewBox - создаёт контроллер Box.
func NewBox(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.BoxUseCase) *Box {
	return &Box{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера Box.
func (ht *Box) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: boxListURL, Func: ht.GetList},
	}
}

// GetList - comment method.
func (ht *Box) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.parser.Localizer(r), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *Box) listParams(_ *http.Request) entity.BoxParams {
	return entity.BoxParams{}
}
