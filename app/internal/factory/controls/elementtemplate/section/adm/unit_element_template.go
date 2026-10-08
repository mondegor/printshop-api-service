package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/controls/elementtemplate/section/adm/controller/httpv1"
	"print-shop-back/internal/controls/elementtemplate/section/adm/entity"
	"print-shop-back/internal/controls/elementtemplate/section/adm/repository"
	"print-shop-back/internal/controls/elementtemplate/section/adm/usecase"
	"print-shop-back/internal/controls/elementtemplate/section/adm/validate"
)

func initElementTemplateController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *validate.Parser,
	responseFileSender mrserver.FileResponseSender,
	pageSizeMax int,
) (mrserver.HttpController, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.ElementTemplate{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewElementTemplatePostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLSetMetaEntity(entityMeta.MetaUpdate()),
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	useCase := usecase.NewElementTemplate(storage, eventEmitter)

	controller := httpv1.NewElementTemplate(
		requestParser,
		responseFileSender,
		useCase,
		entityMeta.MetaOrderBy(),
	)

	return controller, nil
}
