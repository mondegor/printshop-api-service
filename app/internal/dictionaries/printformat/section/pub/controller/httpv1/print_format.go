package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/printformat/section/pub"
	"print-shop-back/internal/dictionaries/printformat/section/pub/entity"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	printFormatListURL = "/v1/dictionaries/print-formats"
)

type (
	// PrintFormat - comment struct.
	PrintFormat struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.PrintFormatUseCase
	}
)

// NewPrintFormat - создаёт контроллер PrintFormat.
func NewPrintFormat(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.PrintFormatUseCase) *PrintFormat {
	return &PrintFormat{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера PrintFormat.
func (ht *PrintFormat) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: printFormatListURL, Func: ht.GetList},
	}
}

// GetList - comment method.
func (ht *PrintFormat) GetList(w http.ResponseWriter, r *http.Request) error {
	items, err := ht.useCase.GetList(r.Context(), ht.parser.Localizer(r), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(w, http.StatusOK, items)
}

func (ht *PrintFormat) listParams(_ *http.Request) entity.PrintFormatParams {
	return entity.PrintFormatParams{}
}
