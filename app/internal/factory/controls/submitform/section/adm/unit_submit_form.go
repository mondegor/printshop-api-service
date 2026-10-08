package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrlock"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/controls/submitform/section/adm/controller/httpv1"
	"print-shop-back/internal/controls/submitform/section/adm/entity"
	"print-shop-back/internal/controls/submitform/section/adm/repository"
	"print-shop-back/internal/controls/submitform/section/adm/usecase"
	"print-shop-back/internal/controls/submitform/section/adm/validate"
)

func initSubmitFormController(
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	storageSubmitForm *repository.SubmitFormPostgres,
	storageFormElement *repository.FormElementPostgres,
	entityMetaSubmitForm *mrsql.EntityMetaOrderBy,
	locker mrlock.Locker,
	requestParser *validate.Parser,
	responseFileSender mrserver.FileResponseSender,
) (mrserver.HttpController, error) {
	storageFormVersion := repository.NewFormVersionPostgres(
		dbConnManager,
	)

	useCase := usecase.NewSubmitForm(
		storageSubmitForm,
		storageFormElement,
		storageFormVersion,
		eventEmitter,
	)

	useCaseVersion := usecase.NewFormVersion(
		storageFormVersion,
		useCase,
		usecase.NewFormCompilerJson(),
		locker,
		eventEmitter,
	)

	controller := httpv1.NewSubmitForm(
		requestParser,
		responseFileSender,
		useCase,
		useCaseVersion,
		entityMetaSubmitForm,
	)

	return controller, nil
}

func initSubmitFormStorage(
	logger log.Logger,
	dbConnManager mrstorage.DBConnManager,
	pageSizeMax int,
) (*repository.SubmitFormPostgres, *mrsql.EntityMetaOrderBy, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.SubmitForm{})
	if err != nil {
		return nil, nil, err
	}

	storage := repository.NewSubmitFormPostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLSetMetaEntity(entityMeta.MetaUpdate()),
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	return storage, entityMeta.MetaOrderBy(), nil
}
