package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/laminate/section/pub/controller/httpv1"
	"print-shop-back/internal/catalog/laminate/section/pub/repository"
	"print-shop-back/internal/catalog/laminate/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initLaminateController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewLaminatePostgres(
		dbConnManager,
	)

	useCase := usecase.NewLaminate(storage)

	controller := httpv1.NewLaminate(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
