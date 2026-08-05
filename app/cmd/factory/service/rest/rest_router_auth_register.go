package rest

import (
	"net/http"

	authvalidate "github.com/mondegor/go-components/mrauth/validate"
	auth "github.com/mondegor/go-components/wire/mrauth/infra/pub"
	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrresp"

	"print-shop-back/internal/app"
)

// TODO: дублирование названий таблиц.
const (
	serviceAuthTokensTableName      = "printshop_auth.auth_tokens" //nolint:gosec
	serviceSecureOperationTableName = "printshop_auth.secure_operations"
	// serviceSecureOperationLogTableName = "printshop_auth.secure_operations_log".
	serviceSessionsTableName            = "printshop_auth.sessions"
	serviceSessionsExcessQueueTableName = "printshop_auth.sessions_excess_queue"
	serviceUsersTableName               = "printshop_auth.users"
	// serviceUsersActivityLogTableName = "printshop_auth.users_activity_log".
	serviceUsersActivityStatTableName = "printshop_auth.users_activity_stat"
	serviceUsersAuth2faTableName      = "printshop_auth.users_auth_2fa"
	serviceUsersRealmsTableName       = "printshop_auth.users_realms"
)

// RegisterRestRouterAuthHandlers - регистрирует в указанном роутере обработчики секции AuthAPI.
func RegisterRestRouterAuthHandlers(
	router mrserver.HttpRouter,
	opts app.Options,
	actionGroup mraccess.ActionGroup,
	userProvider mraccess.UserProvider,
) error {
	router.HandlerFunc(http.MethodGet, actionGroup.BasePath, mrresp.HandlerGetStatusOkAsJSON(opts.Logger))

	controllers, err := initing.CreateHttpControllers(
		opts.Logger,
		getAuthAPIControllers(opts),
		initing.WithCheckAccessMiddleware(opts.Logger, actionGroup, userProvider, opts.PermsProvider),
	)
	if err != nil {
		return err
	}

	router.Register(controllers...)

	return nil
}

func getAuthAPIControllers(opts app.Options) []initing.HttpModule {
	return []initing.HttpModule{
		auth.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.Locker,
			// opts.RequestParsers.Parser,
			authvalidate.NewParser( // TODO: объединить со стандартным Parser или сделать свой? Может там нужно меньше парсеров
				opts.RequestParsers.Int64,
				opts.RequestParsers.Uint64,
				opts.RequestParsers.String,
				opts.RequestParsers.UUID,
				opts.RequestParsers.Validator,
				opts.RequestParsers.ClientIP,
				opts.RequestParsers.User,
				opts.RequestParsers.Locale,
				opts.RequestParsers.TimeZone,
			),
			opts.ResponseSenders.Sender,
			opts.ResponseSenders.FileSender,
			opts.NotifierAPI,
			opts.SecureOperationLogCollectorService,
			opts.Cfg.AccessControl.Realms,
			opts.Cfg.AccessControl.DefaultOperationConfirm,
			opts.Cfg.AccessControl.Auth2FA,
			opts.Cfg.JWT,
			opts.Cfg.RefreshCookie,
			opts.Cfg.AccessControl.SessionSoftThreshold,
			opts.Cfg.AccessControl.SessionHardThreshold,
			opts.Cfg.TaskScheduleAuth.TrimSessions.Period,
			nil, // appResolver
			nil, // locationResolver
			serviceAuthTokensTableName,
			serviceSecureOperationTableName,
			// serviceSecureOperationLogTableName,
			serviceSessionsTableName,
			serviceSessionsExcessQueueTableName,
			serviceUsersTableName,
			// serviceUsersActivityLogTableName,
			serviceUsersActivityStatTableName,
			serviceUsersAuth2faTableName,
			serviceUsersRealmsTableName,
			opts.DebugFunc,
		),
	}
}
