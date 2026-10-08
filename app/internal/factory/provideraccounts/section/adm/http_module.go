package adm

import (
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/provideraccounts/module"
	"print-shop-back/internal/provideraccounts/section/adm/validate"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	dbConnManager mrstorage.DBConnManager,
	requestModuleParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	logoURLBuilder mrpath.Builder,
	pageSizeMax int,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initCompanyPageController(
						logger,
						dbConnManager,
						requestModuleParser,
						responseSender,
						logoURLBuilder,
						pageSizeMax,
					)
				},
			},
		},
	}
}
