package usr

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	repository2 "print-shop-back/internal/warehousing/actiongroup/back/repository"
	usecase3 "print-shop-back/internal/warehousing/actiongroup/back/usecase"
	"print-shop-back/internal/warehousing/actiongroup/usr/repository"
	"print-shop-back/internal/warehousing/actiongroup/usr/service"
	"print-shop-back/internal/warehousing/actiongroup/usr/transport/httpv1"
	"print-shop-back/internal/warehousing/actiongroup/usr/usecase"
	"print-shop-back/pkg/transport/validate/usr"
)

func initContainerController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *usr.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewContainerPostgres(
		dbConnManager,
	)

	// :TODO: временно, вынести на уровень выше
	storage2 := repository.NewStorePostgres(
		dbConnManager,
	)

	// :TODO: временно, вынести на уровень выше
	storageStock := repository.NewStockPostgres(
		dbConnManager,
	)

	// :TODO: временно, вынести на уровень выше
	usecase2 := usecase3.NewRefreshStoreContainersVolume(
		repository2.NewStorePostgres(dbConnManager),
		repository2.NewStockPostgres(dbConnManager),
		logger,
	)

	usecase4 := usecase3.NewRefreshGroupContainers(
		repository2.NewContainerPostgres(dbConnManager),
		repository2.NewStockPostgres(dbConnManager),
		logger,
	)

	usecase5 := usecase3.NewRefreshLocations(
		usecase2,
		usecase4,
	)

	service2 := service.NewAccountSequence(
		repository.NewAccountSequencePostgres(dbConnManager),
	)

	useCase := usecase.NewCreateStockContainer(
		dbConnManager,
		storage,
		service2,
		storageStock,
		service.NewStore(
			storage2,
			storage,
		),
		usecase5,
		eventEmitter,
	)

	service1 := service.NewContainer(storage, eventEmitter)

	controller := httpv1.NewContainer(
		requestParser,
		responseSender,
		service1,
		useCase,
	)

	return controller, nil
}
