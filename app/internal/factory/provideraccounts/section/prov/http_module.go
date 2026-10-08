package prov

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrlock"
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/provideraccounts/module"
	"print-shop-back/internal/provideraccounts/section/prov/validate"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	locker mrlock.Locker,
	requestModuleParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	logoFileAPIFunc func() (mrstorage.FileProviderAPI, error),
	logoURLBuilder mrpath.Builder,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initCompanyPageController(
						eventEmitter,
						dbConnManager,
						requestModuleParser,
						responseSender,
						logoURLBuilder,
					)
				},
			},
			{
				Create: func() (mrserver.HttpController, error) {
					logoFileAPI, err := logoFileAPIFunc()
					if err != nil {
						return nil, err
					}

					return initCompanyPageLogoController(
						logger,
						eventEmitter,
						dbConnManager,
						locker,
						requestModuleParser,
						responseSender,
						logoFileAPI,
					)
				},
			},
		},
	}
}
