package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/queryhistory/section/pub/controller/httpv1"
	"print-shop-back/internal/calculations/queryhistory/section/pub/repository"
	"print-shop-back/internal/calculations/queryhistory/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initQueryHistoryController(
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewQueryHistoryPostgres(
		dbConnManager,
	)

	useCase := usecase.NewQueryHistory(storage, eventEmitter)

	controller := httpv1.NewQueryHistory(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
