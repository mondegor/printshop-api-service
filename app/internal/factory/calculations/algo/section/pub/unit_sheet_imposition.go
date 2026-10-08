package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/calculations/algo/section/pub/sheet/imposition/controller/httpv1"
	"print-shop-back/internal/calculations/algo/section/pub/sheet/imposition/usecase"
	"print-shop-back/pkg/mrcalc/algo/sheet/imposition"
	"print-shop-back/pkg/transport/validate/pub"
)

func initSheetImpositionController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	algoComponent := imposition.New(logger)

	useCase := usecase.NewSheetImposition(algoComponent, eventEmitter)

	controller := httpv1.NewSheetImposition(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
