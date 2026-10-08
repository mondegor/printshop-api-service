package adm

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/controls/elementtemplate/module"
	"print-shop-back/internal/controls/elementtemplate/section/adm/validate"
)

// InitHttpModule - создаются все компоненты модуля и возвращаются к нему контролеры.
func InitHttpModule(
	logger log.Logger,
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestParser *validate.Parser,
	responseFileSender mrserver.FileResponseSender,
	pageSizeMax int,
) initing.HttpModule {
	return initing.HttpModule{
		Caption:    module.Name,
		Permission: module.Permission,
		Controllers: []initing.HttpController{
			{
				Create: func() (mrserver.HttpController, error) {
					return initElementTemplateController(
						logger,
						eventEmitter,
						dbConnManager,
						requestParser,
						responseFileSender,
						pageSizeMax,
					)
				},
			},
		},
	}
}
