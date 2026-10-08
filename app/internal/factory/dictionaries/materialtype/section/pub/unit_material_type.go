package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/materialtype/section/pub/controller/httpv1"
	"print-shop-back/internal/dictionaries/materialtype/section/pub/repository"
	"print-shop-back/internal/dictionaries/materialtype/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initMaterialTypeController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewMaterialTypePostgres(
		dbConnManager,
	)

	useCase := usecase.NewMaterialType(storage)

	controller := httpv1.NewMaterialType(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
