package adm

import (
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrpostgres/builder"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrstorage/mrsql"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/provideraccounts/section/adm/controller/httpv1"
	"print-shop-back/internal/provideraccounts/section/adm/entity"
	"print-shop-back/internal/provideraccounts/section/adm/repository"
	"print-shop-back/internal/provideraccounts/section/adm/usecase"
	"print-shop-back/internal/provideraccounts/section/adm/validate"
)

func initCompanyPageController(
	logger log.Logger,
	dbConnManager mrstorage.DBConnManager,
	requestModuleParser *validate.Parser,
	responseSender mrserver.ResponseSender,
	logoURLBuilder mrpath.Builder,
	pageSizeMax int,
) (mrserver.HttpController, error) {
	entityMeta, err := mrsql.ParseEntity(logger, entity.CompanyPage{})
	if err != nil {
		return nil, err
	}

	storage := repository.NewCompanyPagePostgres(
		dbConnManager,
		builder.NewSQL(
			builder.WithSQLOrderByDefaultSort(entityMeta.MetaOrderBy().DefaultSort()),
			builder.WithSQLLimitMaxSize(pageSizeMax),
		),
	)

	useCase := usecase.NewCompanyPage(storage, logoURLBuilder)

	controller := httpv1.NewCompanyPage(
		requestModuleParser,
		responseSender,
		useCase,
		entityMeta.MetaOrderBy(),
	)

	return controller, nil
}
