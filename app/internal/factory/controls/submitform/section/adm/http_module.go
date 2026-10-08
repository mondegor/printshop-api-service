package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrlock"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/controls/submitform/module"
	"print-shop-back/internal/controls/submitform/section/adm/repository"
	"print-shop-back/internal/controls/submitform/section/adm/validate"
	"print-shop-back/pkg/controls/api"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	locker mrlock.Locker,
	requestModuleParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	responseFileSender mrserver.FileResponseSender,
	elementTemplateAPI api.ElementTemplateHeader,
	pageSizeMax int,
) initing.HttpModule {
	// переменные, которые должны быть инициализированы в InitSharedComponents модуля
	var (
		storageSubmitForm    *repository.SubmitFormPostgres
		entityMetaSubmitForm *mrsql.EntityMetaOrderBy
		storageFormElement   *repository.FormElementPostgres
	)

	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		InitSharedComponents: func() (err error) {
			storageSubmitForm, entityMetaSubmitForm, err = initSubmitFormStorage(logger, dbConnManager, pageSizeMax)
			if err != nil {
				return err
			}

			storageFormElement, err = initFormElementStorage(logger, dbConnManager)

			return err
		},
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initSubmitFormController(
						eventEmitter,
						dbConnManager,
						storageSubmitForm,
						storageFormElement,
						entityMetaSubmitForm,
						locker,
						requestModuleParser,
						responseFileSender,
					)
				},
			},
			{
				Caption:    module.UnitFormElementName,
				Permission: module.UnitFormElementPermission,
				Create: func() (mrserver.HttpController, error) {
					return initFormElementController(
						logger,
						eventEmitter,
						dbConnManager,
						storageSubmitForm,
						storageFormElement,
						requestModuleParser,
						responseSender,
						elementTemplateAPI,
					)
				},
			},
		},
	}
}
