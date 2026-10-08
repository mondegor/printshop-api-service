package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/calculations/algo/section/pub/box/packinbox/controller/httpv1"
	"print-shop-back/internal/calculations/algo/section/pub/box/packinbox/usecase"
	"print-shop-back/pkg/mrcalc/algo/box/packinbox"
	"print-shop-back/pkg/mrcalc/algo/sheet/imposition"
	"print-shop-back/pkg/transport/validate/pub"
)

func initBoxPackInBoxController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	impAlgo := imposition.New(logger)
	packInBoxAlgo := packinbox.New(impAlgo)

	useCase := usecase.NewBoxPackInBox(packInBoxAlgo, logger, eventEmitter)

	controller := httpv1.NewBoxPackInBox(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
