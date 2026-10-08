package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/algo/section/pub/sheet/cutting/controller/httpv1"
	"print-shop-back/internal/calculations/algo/section/pub/sheet/cutting/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initBoxSheetCuttingController(
	eventEmitter mrevent.Emitter,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	useCase := usecase.NewSheetCutting(eventEmitter)

	controller := httpv1.NewSheetCutting(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
