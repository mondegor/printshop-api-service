package rest

import (
	"github.com/mondegor/go-components/mrauth/infra/adapter/collect"
	"github.com/mondegor/go-components/wire/mrauth/mapping"
	"github.com/mondegor/go-core/wire/mraccess"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/middleware"
	"github.com/mondegor/go-webcore/mrserver/mrchi"
	"github.com/mondegor/go-webcore/mrserver/mrprometheus"
	"github.com/mondegor/go-webcore/mrserver/mrresp"
	"github.com/mondegor/go-webcore/mrserver/mrrscors"
	"github.com/mondegor/go-webcore/mrserver/stat"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/app"
)

// InitRestRouterWithHandlers - создаёт объект mrchi.RouterAdapter и регистрирует в нём http обработчики.
func InitRestRouterWithHandlers(opts app.Options) (*mrchi.RouterAdapter, error) {
	router, err := initRestRouter(opts)
	if err != nil {
		return nil, err
	}

	name2group := mraccess.InitActionGroups(opts.Logger, opts.Cfg.AccessControl.ActionGroups)

	err = RegisterRestRouterAuthHandlers(router, opts, name2group["auth-api"], opts.RealmUserProviders["*"])
	if err != nil {
		return nil, err
	}

	err = RegisterRestRouterAdmHandlers(router, opts, name2group["admin-api"], opts.RealmUserProviders["admin.printshop/backend"])
	if err != nil {
		return nil, err
	}

	err = RegisterRestRouterProvHandlers(router, opts, name2group["provider-api"], opts.RealmUserProviders["printshop/providers"])
	if err != nil {
		return nil, err
	}

	err = RegisterRestRouterPubHandlers(router, opts, name2group["public-api"], opts.RealmUserProviders["printshop/*"])
	if err != nil {
		return nil, err
	}

	err = RegisterRestRouterUsrHandlers(router, opts, name2group["user-api"], opts.RealmUserProviders["printshop/users"])
	if err != nil {
		return nil, err
	}

	return router, nil
}

func initRestRouter(opts app.Options) (*mrchi.RouterAdapter, error) {
	corsOptions := mrrscors.Options{
		AllowedOrigins:   opts.Cfg.CorsAllowedOrigins,
		AllowedMethods:   opts.Cfg.CorsAllowedMethods,
		AllowedHeaders:   opts.Cfg.CorsAllowedHeaders,
		ExposedHeaders:   opts.Cfg.CorsExposedHeaders,
		AllowCredentials: opts.Cfg.CorsAllowCredentials,
		Logger:           log.WithAttrs(opts.Logger, "middleware", "cors"),
	}

	errorSender, err := NewErrorResponseSender(opts)
	if err != nil {
		return nil, err
	}

	requestStat := mrserver.NewRequestContainer(
		stat.NewRequestMetrics(initPrometheusRequestObserve(opts)),
		stat.NewRequestTracer(opts.Tracer),
		stat.NewRequestLogger(opts.Logger),
		collect.NewUserRequest(
			opts.UserStatRequestCollectorService, // TODO: заменить на API
			opts.Logger,
			opts.RequestParsers.ClientIP,
			opts.RequestParsers.User,
			mapping.OptionUserRealmsToRealmRegistry(opts.Cfg.AccessControl.Realms),
		),
	)

	router := mrchi.New(
		log.WithAttrs(opts.Logger, "router", "chi"),
		middleware.HandlerAdapter(errorSender),
		mrresp.HandlerGetNotFoundAsJSON(opts.Logger),
		mrresp.HandlerGetMethodNotAllowedAsJSON(opts.Logger),
	)

	router.RegisterMiddleware(
		middleware.RecoverHandler(
			opts.Logger,
			opts.Cfg.DebugIsEnabled,
			mrresp.HandlerGetFatalErrorAsJSON(opts.Logger),
		),
		middleware.RequestIDHandler(opts.Logger, opts.TraceManager),
		mrrscors.Middleware(corsOptions),
		middleware.ObserverHandler(opts.Logger, requestStat),
	)

	return router, nil
}

func initPrometheusRequestObserve(opts app.Options) mrserver.RequestObserve {
	if opts.Prometheus == nil {
		log.Warn(opts.Logger, "Collector Rest Stat is disabled")

		return mrserver.NopRequestObserve()
	}

	requestMetrics := mrprometheus.NewObserveRequest("rest_api", "go")
	opts.Prometheus.Add(requestMetrics.Collectors()...)

	return requestMetrics
}
