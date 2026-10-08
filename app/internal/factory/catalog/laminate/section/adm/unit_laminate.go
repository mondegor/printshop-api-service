package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/catalog/laminate/section/adm/controller/httpv1"
	"print-shop-back/internal/catalog/laminate/section/adm/entity"
	"print-shop-back/internal/catalog/laminate/section/adm/repository"
	"print-shop-back/internal/catalog/laminate/section/adm/usecase"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/transport/validate/adm"
)

func initLaminateController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *adm.Parser,
	responseSender mrserver.ResponseSender,
	materialTypeAPI api.MaterialTypeAvailability,
	pageSizeMax int,
) (mrserver.HttpController, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.Laminate{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewLaminatePostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLSetMetaEntity(entityMeta.MetaUpdate()),
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	useCase := usecase.NewLaminate(storage, materialTypeAPI, eventEmitter)

	controller := httpv1.NewLaminate(
		requestParser,
		responseSender,
		useCase,
		entityMeta.MetaOrderBy(),
	)

	return controller, nil
}
