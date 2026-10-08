package pub

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/queryhistory/module"
	"print-shop-back/pkg/transport/validate/pub"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initQueryHistoryController(
						eventEmitter,
						dbConnManager,
						requestParser,
						responseSender,
					)
				},
			},
		},
	}
}
