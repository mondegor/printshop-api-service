package usr

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/warehousing/module"
	"print-shop-back/pkg/transport/validate/usr"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *usr.Parser,
	responseSender mrserver.ResponseSender,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initContainerController(
						logger,
						eventEmitter,
						dbConnManager,
						requestParser,
						responseSender,
					)
				},
			},
			{
				Create: func() (mrserver.HttpController, error) {
					return initStockController(
						logger,
						eventEmitter,
						dbConnManager,
						requestParser,
						responseSender,
					)
				},
			},
			{
				Create: func() (mrserver.HttpController, error) {
					return initStoreController(
						dbConnManager,
						requestParser,
						responseSender,
					)
				},
			},
		},
	}
}
