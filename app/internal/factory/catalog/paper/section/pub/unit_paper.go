package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/paper/section/pub/controller/httpv1"
	"print-shop-back/internal/catalog/paper/section/pub/repository"
	"print-shop-back/internal/catalog/paper/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initPaperController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewPaperPostgres(
		dbConnManager,
	)

	useCase := usecase.NewPaper(storage)

	controller := httpv1.NewPaper(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
