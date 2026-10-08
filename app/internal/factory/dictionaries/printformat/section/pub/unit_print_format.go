package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/printformat/section/pub/controller/httpv1"
	"print-shop-back/internal/dictionaries/printformat/section/pub/repository"
	"print-shop-back/internal/dictionaries/printformat/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initPrintFormatController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewPrintFormatPostgres(
		dbConnManager,
	)

	useCase := usecase.NewPrintFormat(storage)

	controller := httpv1.NewPrintFormat(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
