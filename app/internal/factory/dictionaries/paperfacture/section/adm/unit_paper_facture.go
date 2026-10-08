package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm/controller/httpv1"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm/entity"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm/repository"
	"print-shop-back/internal/dictionaries/paperfacture/section/adm/usecase"
	"print-shop-back/pkg/transport/validate/adm"
)

func initPaperFactureController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *adm.Parser,
	responseSender mrserver.ResponseSender,
	pageSizeMax int,
) (mrserver.HttpController, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.PaperFacture{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewPaperFacturePostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	useCase := usecase.NewPaperFacture(storage, eventEmitter)

	controller := httpv1.NewPaperFacture(
		requestParser,
		responseSender,
		useCase,
		entityMeta.MetaOrderBy(),
	)

	return controller, nil
}
