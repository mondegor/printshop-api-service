package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/catalog/box/section/pub/controller/httpv1"
	"print-shop-back/internal/catalog/box/section/pub/repository"
	"print-shop-back/internal/catalog/box/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initBoxController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewBoxPostgres(
		dbConnManager,
	)

	useCase := usecase.NewBox(storage)

	controller := httpv1.NewBox(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
