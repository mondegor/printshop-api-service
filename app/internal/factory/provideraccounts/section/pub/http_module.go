package pub

import (
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/provideraccounts/module"
	"print-shop-back/pkg/transport/validate/pub"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	dbConnManager mrstorage.DBConnManager,
	requestModuleParser *pub.Parser,
	responseSender mrserver.ResponseSender,
	logoURLBuilder mrpath.Builder,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initCompanyPageController(
						dbConnManager,
						requestModuleParser,
						responseSender,
						logoURLBuilder,
					)
				},
			},
		},
	}
}
