package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/papercolor/section/pub/controller/httpv1"
	"print-shop-back/internal/dictionaries/papercolor/section/pub/repository"
	"print-shop-back/internal/dictionaries/papercolor/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initPaperColorController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewPaperColorPostgres(
		dbConnManager,
	)

	useCase := usecase.NewPaperColor(storage)

	controller := httpv1.NewPaperColor(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
