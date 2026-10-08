package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/algo/section/pub/sheet/insideoutside/controller/httpv1"
	"print-shop-back/internal/calculations/algo/section/pub/sheet/insideoutside/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initSheetInsideOutsideController(
	eventEmitter mrevent.Emitter,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	useCase := usecase.NewSheetInsideOutside(eventEmitter)

	controller := httpv1.NewSheetInsideOutside(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
