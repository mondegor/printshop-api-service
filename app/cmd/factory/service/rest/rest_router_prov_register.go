package rest

import (
	"net/http"

	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrresp"

	"print-shop-back/internal/app"
	provideraccounts "print-shop-back/internal/factory/provideraccounts/section/prov"
	provideraccountsvalidate "print-shop-back/internal/provideraccounts/section/prov/validate"
)

// RegisterRestRouterProvHandlers - регистрирует в указанном роутере обработчики секции ProvidersAPI.
func RegisterRestRouterProvHandlers(
	router mrserver.HttpRouter,
	opts app.Options,
	actionGroup mraccess.ActionGroup,
	userProvider mraccess.UserProvider,
) error {
	router.HandlerFunc(http.MethodGet, actionGroup.BasePath, mrresp.HandlerGetStatusOkAsJSON(opts.Logger))

	controllers, err := initing.CreateHttpControllers(
		opts.Logger,
		getProviderAPIControllers(opts),
		initing.WithCheckAccessMiddleware(opts.Logger, actionGroup, userProvider, opts.PermsProvider),
	)
	if err != nil {
		return err
	}

	router.Register(controllers...)

	return nil
}

func getProviderAPIControllers(opts app.Options) []initing.HttpModule {
	return []initing.HttpModule{
		provideraccounts.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.Locker,
			provideraccountsvalidate.NewParser(
				opts.RequestParsers.ProvParser,
				opts.RequestParsers.ImageLogo,
			),
			opts.ResponseSenders.Sender,
			func() (mrstorage.FileProviderAPI, error) {
				return opts.FileProviderPool.ProviderAPI(
					opts.Cfg.ModuleSettings.ProviderAccount.CompanyPageLogoProvider,
				)
			},
			opts.ImageURLBuilder,
		),
	}
}
