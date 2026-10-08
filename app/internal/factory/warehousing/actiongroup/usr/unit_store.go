package usr

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/warehousing/actiongroup/usr/repository"
	"print-shop-back/internal/warehousing/actiongroup/usr/service"
	"print-shop-back/internal/warehousing/actiongroup/usr/transport/httpv1"
	"print-shop-back/pkg/transport/validate/usr"
)

func initStoreController(
	// logger log.Logger,
	dbConnManager mrstorage.DBConnManager,
	requestParser *usr.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewStorePostgres(
		dbConnManager,
	)

	container := repository.NewContainerPostgres(
		dbConnManager,
	)

	serviceStore := service.NewStore(storage, container)

	controller := httpv1.NewStore(
		requestParser,
		responseSender,
		serviceStore,
	)

	return controller, nil
}
