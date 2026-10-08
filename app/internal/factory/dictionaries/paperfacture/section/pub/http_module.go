package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/dictionaries/paperfacture/module"
	"print-shop-back/pkg/transport/validate/pub"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
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
					return initPaperFactureController(
						dbConnManager,
						requestParser,
						responseSender,
					)
				},
			},
		},
	}
}
