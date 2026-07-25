package factory

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	authcfg "github.com/mondegor/go-components/wire/mrauth/config"
	wireauth "github.com/mondegor/go-components/wire/mrauth/initing"
	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrpostgres/listennotify"
	"github.com/mondegor/go-core/mrrun"
	"github.com/mondegor/go-core/util/xio"
	wireerrors "github.com/mondegor/go-core/wire/errors"
	wireaccess "github.com/mondegor/go-core/wire/mraccess"
	redislocker "github.com/mondegor/go-storage/mrredis/locker"

	dictionariesapi "print-shop-back/cmd/factory/api/dictionaries"
	"print-shop-back/cmd/factory/service"
	"print-shop-back/cmd/factory/service/rest"
	"print-shop-back/config"
	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/app"
)

// TODO: дублирование названий таблиц.
const (
	serviceAuthTokensTableName = "printshop_auth.auth_tokens" //nolint:gosec
)

// InitApp - настраивает конфигурацию и внешнее окружение приложения, после чего создаёт
// его модули и компоненты.
func InitApp(args []string, stdout io.Writer) (app.Options, error) {
	parsedArgs, err := config.ParseCmdArgs(args)
	if err != nil {
		return app.Options{}, err
	}

	cfg, err := config.Create(parsedArgs, stdout)
	if err != nil {
		return app.Options{}, fmt.Errorf("factory.InitApp(): %w", err)
	}

	logger, tracer, err := InitLoggerAndTracer(cfg)
	if err != nil {
		return app.Options{}, err
	}

	traceManager, err := InitTraceContextManager(cfg, logger)
	if err != nil {
		return app.Options{}, err
	}

	return InitAppEnvironment(
		app.Options{
			Cfg:             cfg,
			Logger:          logger,
			Tracer:          tracer,
			TraceManager:    traceManager,
			OpenedResources: xio.NewCloseManager(logger),
			DebugFunc:       InitDebugInfo(cfg.DebugIsEnabled),
		},
	)
}

// InitAppEnvironment - настраивает внешнее окружение приложения на основе переданной конфигурации,
// после чего создаёт модули и компоненты приложения.
// Часть компонентов можно задать заранее (актуально для тестов): opts.PostgresConnManager,
// opts.RedisAdapter, opts.FileProviderPool - если они уже заданы, повторно не создаются.
func InitAppEnvironment(opts app.Options) (app.Options, error) {
	logger := opts.Logger

	// show head info about started app
	log.Info(logger, opts.Cfg.AppName, "environment", opts.Cfg.Environment, "version", opts.Cfg.AppVersion)

	if opts.Cfg.WorkDir != "" {
		log.Debug(logger, "WORK DIR: "+opts.Cfg.WorkDir)
	}

	log.Debug(logger, "CONFIG PATHs: "+strings.Join(opts.Cfg.ConfigPaths, ", "))

	if opts.Cfg.DotEnvPath != "" {
		log.Debug(logger, ".ENV PATH: "+opts.Cfg.DotEnvPath)
	}

	if opts.Cfg.DebugIsEnabled {
		log.Info(logger, "DEBUG MODE: ON")
	}

	log.Info(logger, "LOG LEVEL: "+opts.Cfg.LogLevel)

	opts, err := createAppEnvironment(opts)
	if err != nil {
		return app.Options{}, err
	}

	// Shared APIs init section (!!! only after create environment)
	opts = createSharedAPI(opts)

	// Shared service's options (!!! only after create modules)
	if opts, err = createSharedServices(opts); err != nil {
		return app.Options{}, err
	}

	return opts, nil
}

// createAppEnvironment - создаёт, и настраивает внешнее окружение приложения.
func createAppEnvironment(opts app.Options) (enrichedOpts app.Options, err error) {
	if sentry, err := InitSentry(opts.Logger, opts.Cfg); err != nil {
		if !errors.Is(err, errSentryDisabled) {
			return app.Options{}, err
		}
	} else {
		opts.OpenedResources.Register(sentry)
		opts.Sentry = sentry
	}

	opts.MonitoringRouter = http.NewServeMux()

	if opts.Prometheus == nil {
		opts.Prometheus = InitPrometheus(opts)
	}

	// !!! only after init Sentry and Prometheus
	wireerrors.InitErrors(
		wireerrors.ErrorConfig{
			HasCaller:         opts.Cfg.StackTraceIsEnabled,
			CallerDepth:       opts.Cfg.StackTraceDepth,
			CallerShowFunc:    opts.Cfg.StackTraceShowFunc,
			CallerUpperBounds: opts.Cfg.StackTraceUpperBounds,
		},
	)

	opts.EventEmitter = InitEventEmitter(opts)
	opts.ErrorHandler = wireerrors.InitErrorHandler(opts.Logger)
	opts.AppHealth = mrrun.NewAppHealth()

	if opts.PostgresConnManager == nil {
		postgresAdapter, err := InitPostgres(opts.Logger, opts.Tracer, opts.Cfg)
		if err != nil {
			return app.Options{}, err
		}

		opts.OpenedResources.Register(postgresAdapter)
		opts.PostgresConnManager = InitPostgresConnManager(postgresAdapter, opts.Logger)

		if err = InitPrometheusStatPostgres(opts); err != nil {
			return app.Options{}, err
		}

		if err = ApplyPostgresMigrations(opts.Logger, opts.PostgresConnManager, opts.Cfg); err != nil {
			return app.Options{}, err
		}
	}

	if opts.RedisAdapter == nil {
		redisAdapter, err := InitRedis(opts.Logger, opts.Tracer, opts.Cfg)
		if err != nil {
			return app.Options{}, err
		}

		opts.OpenedResources.Register(redisAdapter)
		opts.RedisAdapter = redisAdapter
	}

	if opts.FileProviderPool == nil {
		opts.FileProviderPool, err = InitFileProviderPool(opts.Logger, opts.Tracer, opts.Cfg)
		if err != nil {
			return app.Options{}, err
		}

		opts.OpenedResources.Register(opts.FileProviderPool)
	}

	redisCli, err := opts.RedisAdapter.Cli()
	if err != nil {
		return app.Options{}, err
	}

	opts.Locker = redislocker.NewAdapter(
		redisCli,
		opts.Logger,
		opts.Tracer,
	)

	if opts.LocalePool, err = InitLocalePool(opts.Logger, opts.Cfg); err != nil {
		return app.Options{}, err
	}

	opts.LocationList = InitTimeZones(opts.Logger, opts.Cfg)

	if opts.RequestParsers, err = InitRequestParsers(opts); err != nil {
		return app.Options{}, err
	}

	if opts.ResponseSenders, err = rest.CreateResponseSenders(opts.Logger); err != nil {
		return app.Options{}, err
	}

	if opts.PermsProvider, err = wireaccess.InitPermsProvider(
		opts.Logger,
		opts.Cfg.AccessControl.RolesDirPath,
		opts.Cfg.AccessControl.Roles,
	); err != nil {
		return app.Options{}, err
	}

	userGroupRights, err := wireauth.InitRealmKindRights(opts.Logger, opts.Cfg.AccessControl.Realms, opts.PermsProvider)
	if err != nil {
		return app.Options{}, err
	}

	if opts.RealmUserProviders, err = wireauth.InitUserProviders(
		opts.Logger,
		opts.PostgresConnManager,
		userGroupRights,
		opts.Cfg.AccessControl.Realms,
		authcfg.TestUser{
			ID:       opts.Cfg.TestUserID,
			Realm:    opts.Cfg.TestUserRealm,
			Kind:     opts.Cfg.TestUserKind,
			LangCode: opts.Cfg.TestUserLangCode,
		},
		opts.Cfg.JWT.Verifier,
		serviceAuthTokensTableName,
	); err != nil {
		return app.Options{}, err
	}

	if opts.ImageURLBuilder, err = InitImageURLBuilder(opts.Cfg); err != nil {
		return app.Options{}, err
	}

	if err = RegisterSystemHandlers(opts); err != nil {
		return app.Options{}, err
	}

	if opts.Prometheus != nil {
		if err = opts.Prometheus.Register(); err != nil {
			return app.Options{}, err
		}
	}

	return opts, nil
}

func createSharedAPI(opts app.Options) app.Options {
	opts.PostgresNotificationService = listennotify.NewProcessWaitForNotification(
		opts.PostgresConnManager.ConnAdapter(),
		opts.Logger,
		[]string{
			opts.Cfg.TaskScheduleNotifier.NoticeProcessor.NotificationChannel,
			opts.Cfg.TaskScheduleMailer.MessageProcessor.NotificationChannel,
			opts.Cfg.TaskScheduleSettings.ReloadSettings.NotificationChannel,
		},
		opts.Cfg.TaskScheduleSettings.NotificationCheckConnPeriod,
	)

	// create settings module
	{
		getter, reloadScheduler := service.InitSettingsGetterAPI(opts)
		opts.SettingsGetterAPI = getter
		opts.TaskSchedulerServices = append(opts.TaskSchedulerServices, reloadScheduler)

		opts.SettingsSetterAPI = service.InitSettingsSetterAPI(opts)
	}

	opts.MailerAPI = service.InitMailerAPI(opts)
	opts.NotifierAPI = service.InitNotifierAPI(opts)
	opts.DictionariesMaterialTypeAPI = dictionariesapi.InitMaterialTypeAvailabilityAPI(opts)
	opts.DictionariesPaperColorAPI = dictionariesapi.InitPaperColorAvailabilityAPI(opts)
	opts.DictionariesPaperFactureAPI = dictionariesapi.InitPaperFactureAvailabilityAPI(opts)
	opts.DictionariesPrintFormatAPI = dictionariesapi.InitPrintFormatAvailabilityAPI(opts)

	return opts
}

func createSharedServices(opts app.Options) (enrichedOpts app.Options, err error) {
	opts.UserStatRequestCollectorService = service.InitUserStatRequestCollectorService(opts)
	opts.SecureOperationLogCollectorService = service.InitSecureOperationLogCollectorService(opts)

	opts.MailProcessorService, err = service.InitMailerProcessorService(opts)
	if err != nil {
		return app.Options{}, fmt.Errorf("factory.InitMailerProcessorService(): %w", err)
	}

	opts.NoticeProcessorService = service.InitNotifierProcessorService(opts)

	opts.HttpServer, err = rest.InitRestServer(opts)
	if err != nil {
		return app.Options{}, fmt.Errorf("factory.InitRestServer(): %w", err)
	}

	opts.HttpMonitoringServer = InitMonitoringServer(opts)

	opts.TaskSchedulerServices = append(
		opts.TaskSchedulerServices,
		service.InitAuthSchedulerService(opts),
		service.InitMailerSchedulerService(opts),
		service.InitNotifierSchedulerService(opts),
	)

	return opts, nil
}
