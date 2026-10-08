package pub

import (
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/controls/submitform/section/pub/controller/httpv1"
	"print-shop-back/internal/controls/submitform/section/pub/repository"
	"print-shop-back/internal/controls/submitform/section/pub/usecase"
	"print-shop-back/pkg/transport/validate/pub"
)

func initSubmitFormController(
	dbConnManager mrstorage.DBConnManager,
	requestParser *pub.Parser,
	responseSender mrserver.ResponseSender,
) (mrserver.HttpController, error) {
	storage := repository.NewSubmitFormPostgres(
		dbConnManager,
	)

	useCase := usecase.NewSubmitForm(storage)

	controller := httpv1.NewSubmitForm(
		requestParser,
		responseSender,
		useCase,
	)

	return controller, nil
}
