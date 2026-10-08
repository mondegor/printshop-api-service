package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/paperfacture/section/pub/controller/httpv1"
	"print-shop-back/internal/dictionaries/paperfacture/section/pub/repository"
	"print-shop-back/internal/dictionaries/paperfacture/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initPaperFactureController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewPaperFacturePostgres(
		dbConnManager,
	)

	useCase := usecase.NewPaperFacture(storage)

	controller := httpv1.NewPaperFacture(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
