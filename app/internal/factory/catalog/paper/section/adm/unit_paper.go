package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/catalog/paper/section/adm/controller/httpv1"
	"print-shop-back/internal/catalog/paper/section/adm/entity"
	"print-shop-back/internal/catalog/paper/section/adm/repository"
	"print-shop-back/internal/catalog/paper/section/adm/usecase"
	"print-shop-back/pkg/dictionaries/api"
	"print-shop-back/pkg/transport/validate/adm"
)

func initPaperController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *adm.Parser,
	responseSender mrserver.ResponseSender,
	materialTypeAPI api.MaterialTypeAvailability,
	paperColorAPI api.PaperColorAvailability,
	paperFactureAPI api.PaperFactureAvailability,
	pageSizeMax int,
) (mrserver.HttpController, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.Paper{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewPaperPostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLSetMetaEntity(entityMeta.MetaUpdate()),
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	useCase := usecase.NewPaper(
		storage,
		materialTypeAPI,
		paperColorAPI,
		paperFactureAPI,
		eventEmitter,
	)

	controller := httpv1.NewPaper(
		requestParser,
		responseSender,
		useCase,
		entityMeta.MetaOrderBy(),
	)

	return controller, nil
}
