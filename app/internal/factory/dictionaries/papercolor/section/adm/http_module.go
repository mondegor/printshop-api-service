package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/dictionaries/papercolor/module"
	"print-shop-back/pkg/transport/validate/adm"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *adm.Parser,
	responseSender mrserver.ResponseSender,
	pageSizeMax int,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initPaperColorController(
						logger,
						eventEmitter,
						dbConnManager,
						requestParser,
						responseSender,
						pageSizeMax,
					)
				},
			},
		},
	}
}
