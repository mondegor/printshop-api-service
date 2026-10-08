package usr

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	repository2 "print-shop-back/internal/warehousing/actiongroup/back/repository"
	usecase3 "print-shop-back/internal/warehousing/actiongroup/back/usecase"
	"print-shop-back/internal/warehousing/actiongroup/usr/repository"
	service2 "print-shop-back/internal/warehousing/actiongroup/usr/service"
	"print-shop-back/internal/warehousing/actiongroup/usr/transport/httpv1"
	"print-shop-back/internal/warehousing/actiongroup/usr/usecase"
	"print-shop-back/pkg/transport/validate/usr"
)

func initStockController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *usr.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	// :TODO: временно, вынести на уровень выше
	storage3 := repository.NewContainerPostgres(
		dbConnManager,
	)

	stock := repository.NewStockPostgres(
		dbConnManager,
	)

	service := service2.NewStock(stock)

	// :TODO: временно, вынести на уровень выше
	storage2 := repository.NewStorePostgres(
		dbConnManager,
	)

	// :TODO: временно, вынести на уровень выше
	usecase1 := usecase.NewAddMoreStockContainer(
		storage3,
		stock,
		service2.NewStore(storage2, storage3),
		usecase3.NewRefreshStoreContainersVolume(
			repository2.NewStorePostgres(dbConnManager),
			repository2.NewStockPostgres(dbConnManager),
			logger,
		),
		eventEmitter,
	)

	// :TODO: временно, вынести на уровень выше
	usecase2 := usecase.NewMoveStockContainer(
		stock,
		service2.NewStore(storage2, storage3),
		usecase3.NewRefreshLocations(
			usecase3.NewRefreshStoreContainersVolume(
				repository2.NewStorePostgres(dbConnManager),
				repository2.NewStockPostgres(dbConnManager),
				logger,
			),
			usecase3.NewRefreshGroupContainers(
				repository2.NewContainerPostgres(dbConnManager),
				repository2.NewStockPostgres(dbConnManager),
				logger,
			),
		),
		eventEmitter,
	)

	// :TODO: временно, вынести на уровень выше
	usecase4 := usecase.NewTransferStockContainer(
		stock,
		usecase3.NewRefreshStoreContainersVolume(
			repository2.NewStorePostgres(dbConnManager),
			repository2.NewStockPostgres(dbConnManager),
			logger,
		),
		eventEmitter,
	)

	controller := httpv1.NewStock(
		requestParser,
		responseSender,
		service,
		usecase2,
		usecase1,
		usecase4,
	)

	return controller, nil
}
