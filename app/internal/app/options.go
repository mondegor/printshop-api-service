package app

import (
	"net/http"

	"github.com/mondegor/go-components/mrauth/dto"
	authentity "github.com/mondegor/go-components/mrauth/entity"
	"github.com/mondegor/go-components/mrmailer"
	mailerentity "github.com/mondegor/go-components/mrmailer/entity"
	"github.com/mondegor/go-components/mrnotifier"
	notifierentity "github.com/mondegor/go-components/mrnotifier/notifier/entity"
	"github.com/mondegor/go-components/mrsettings"
	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-core/mraccess/provider/filestorage"
	"github.com/mondegor/go-core/mrevent"
	"github.com/mondegor/go-core/mrlocale"
	"github.com/mondegor/go-core/mrlock"
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrpostgres"
	"github.com/mondegor/go-core/mrpostgres/listennotify"
	"github.com/mondegor/go-core/mrprocess/collect"
	"github.com/mondegor/go-core/mrprocess/consume"
	"github.com/mondegor/go-core/mrrun"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/util/timezone"
	"github.com/mondegor/go-core/util/xio"
	"github.com/mondegor/go-storage/mrredis"
	"github.com/mondegor/go-webcore/mrclient/sentry"
	"github.com/mondegor/go-webcore/mrcore/mrinit"
	"github.com/mondegor/go-webcore/mrserver/httpserver"
	"github.com/mondegor/go-webcore/mrserver/mrresp"
	"github.com/mondegor/go-webcore/mrserver/request/parser"

	"print-shop-back/config"
	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/adapter/trace"
	"print-shop-back/pkg/dictionaries/api"
	validate2 "print-shop-back/pkg/transport/validate"
)

type (
	// Options - comment struct.
	Options struct {
		Cfg             config.Config
		Logger          log.Logger
		Tracer          trace.Tracer
		TraceManager    trace.ContextManager
		OpenedResources *xio.CloseManager
		DebugFunc       func(value any) string

		MonitoringRouter *http.ServeMux
		Sentry           *sentry.Adapter
		Prometheus       *mrinit.Prometheus
		EventEmitter     mrevent.Emitter
		ErrorHandler     errors.Handler
		AppHealth        *mrrun.AppHealth

		PostgresConnManager         *mrpostgres.ConnManager
		PostgresNotificationService *listennotify.ProcessWaitForNotification
		RedisAdapter                *mrredis.ConnAdapter
		FileProviderPool            *mrstorage.FileProviderPool
		Locker                      mrlock.Locker
		LocalePool                  *mrlocale.Pool
		LocationList                *timezone.LocationList
		RequestParsers              RequestParsers
		ResponseSenders             ResponseSenders
		PermsProvider               *filestorage.PermsProvider
		RealmUserProviders          map[string]mraccess.UserProvider
		ImageURLBuilder             mrpath.Builder

		// API section
		DictionariesMaterialTypeAPI api.MaterialTypeAvailability
		DictionariesPaperColorAPI   api.PaperColorAvailability
		DictionariesPaperFactureAPI api.PaperFactureAvailability
		DictionariesPrintFormatAPI  api.PrintFormatAvailability
		MailerAPI                   mrmailer.MessageProducer
		NotifierAPI                 mrnotifier.NoteProducer
		SettingsGetterAPI           mrsettings.MustGetter
		SettingsSetterAPI           mrsettings.Setter

		// Services and Servers section
		UserStatRequestCollectorService    *collect.MessageCollector[dto.UserActivityLogMessage]
		SecureOperationLogCollectorService *collect.MessageCollector[authentity.SecureOperationLog]
		MailProcessorService               *consume.MessageProcessor[mailerentity.Message]
		NoticeProcessorService             *consume.MessageProcessor[notifierentity.Note]
		HttpServer                         *httpserver.Adapter
		HttpMonitoringServer               *httpserver.Adapter
		TaskSchedulerServices              []mrrun.Process // можно добавлять начиная с формирования API
	}

	// RequestParsers - comment struct.
	RequestParsers struct {
		// Bool       *parser.Bool
		// DateTime   *parser.DateTime
		Int64      *parser.Int64
		ItemStatus *parser.ItemStatus
		Uint64     *parser.Uint64
		ListCursor *parser.ListCursor
		ListPager  *parser.ListPager
		ListSorter *parser.ListSorter
		String     *parser.String
		UUID       *parser.UUID
		Validator  *parser.Validator
		Locale     *parser.Locale
		TimeZone   *parser.TimeZone
		ClientIP   *parser.ClientIP
		User       *parser.User
		FileJson   *parser.File
		ImageLogo  *parser.Image

		Parser       *validate2.Parser
		ExtendParser *validate2.ExtendParser
	}

	// ResponseSenders - comment struct.
	ResponseSenders struct {
		Sender     *mrresp.Sender
		FileSender *mrresp.FileSender
	}
)
