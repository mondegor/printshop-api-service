package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/algo/section/pub/sheet/packinstack/controller/httpv1"
	"print-shop-back/internal/calculations/algo/section/pub/sheet/packinstack/usecase"
	"print-shop-back/pkg/mrcalc/algo/sheet/packinstack"
	"print-shop-back/pkg/transport/validate/pub"
)

func initSheetPackInStackController(
	eventEmitter mrevent.Emitter,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	packInStackAlgoSheet := packinstack.New()

	useCase := usecase.NewSheetPackInStack(packInStackAlgoSheet, eventEmitter)

	controller := httpv1.NewSheetPackInStack(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
