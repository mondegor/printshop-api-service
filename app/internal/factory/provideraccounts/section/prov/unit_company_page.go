package prov

import (
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/provideraccounts/section/prov/controller/httpv1"
	"print-shop-back/internal/provideraccounts/section/prov/repository"
	"print-shop-back/internal/provideraccounts/section/prov/usecase"
	"print-shop-back/internal/provideraccounts/section/prov/validate"
)

func initCompanyPageController(
	eventEmitter mrevent.Emitter,
	dbConnManager mrstorage.DBConnManager,
	requestModuleParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	logoURLBuilder mrpath.Builder,
) (mrserver.HttpController, error) {
	storage := repository.NewCompanyPagePostgres(dbConnManager)

	useCase := usecase.NewCompanyPage(
		dbConnManager,
		storage,
		logoURLBuilder,
		eventEmitter,
	)

	controller := httpv1.NewCompanyPage(
		requestModuleParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
