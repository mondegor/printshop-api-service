package pub

import (
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/provideraccounts/section/pub/controller/httpv1"
	"print-shop-back/internal/provideraccounts/section/pub/repository"
	"print-shop-back/internal/provideraccounts/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initCompanyPageController(
	dbConnManager mrstorage.DBConnManager,
	requestModuleParser *pub.Parser,
	responseSender mrserver.ResponseSender,
	logoURLBuilder mrpath.Builder,
) (mrserver.HttpController, error) {
	storage := repository.NewCompanyPagePostgres(dbConnManager)
	useCase := usecase.NewCompanyPage(storage)
	controller := httpv1.NewCompanyPage(
		requestModuleParser,
		responseSender,
		useCase,
		logoURLBuilder,
	)

	return controller, nil
}
