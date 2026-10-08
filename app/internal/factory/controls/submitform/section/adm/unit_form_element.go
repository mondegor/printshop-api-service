package adm

import (
	"github.com/mondegor/go-components/wire/mrordering"
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/controls/submitform/module"
	"print-shop-back/internal/controls/submitform/section/adm/controller/httpv1"
	"print-shop-back/internal/controls/submitform/section/adm/entity"
	"print-shop-back/internal/controls/submitform/section/adm/repository"
	"print-shop-back/internal/controls/submitform/section/adm/usecase"
	"print-shop-back/internal/controls/submitform/section/adm/validate"
	"print-shop-back/pkg/controls/api"
)

func initFormElementController(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	storageSubmitForm *repository.SubmitFormPostgres,
	storageFormElement *repository.FormElementPostgres,
	requestParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	elementTemplateAPI api.ElementTemplateHeader,
) (mrserver.HttpController, error) {
	useCase := usecase.NewFormElement(
		storageFormElement,
		storageSubmitForm,
		elementTemplateAPI,
		mrordering.InitServiceMover(
			dbConnManager,
			eventEmitter,
			mrsql.DBTableInfo{
				Name:       module.DBTableNameSubmitFormElements,
				PrimaryKey: "element_id",
			},
		),
		eventEmitter,
		logger,
	)

	controller := httpv1.NewFormElement(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}

func initFormElementStorage(
	logger log.Logger,
	dbConnManager mrstorage.DBConnManager,
) (*repository.FormElementPostgres, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.FormElement{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewFormElementPostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLSetMetaEntity(entityMeta.MetaUpdate()),
		),
	)

	return storage, nil
}
