package config

import (
	"io"
	"time"

	authcfg "github.com/mondegor/go-components/wire/mrauth/config"
	maliercfg "github.com/mondegor/go-components/wire/mrmailer/config"
	notifiercfg "github.com/mondegor/go-components/wire/mrnotifier/config"
	modelcfg "github.com/mondegor/go-core/mrmodel/config"
	processcfg "github.com/mondegor/go-core/mrprocess/config"
	"github.com/mondegor/go-core/util/mime"
)

type (
	// Config - comment struct.
	Config struct {
		WorkDir     string // cmd ARG:work-dir
		ConfigPaths []string
		DotEnvPath  string // cmd ARG:dot-env-path
		StartedAt   time.Time
		Stdout      io.Writer

		AppName      string   `yaml:"app_name" env:"APPX_NAME" env-required:"true"`
		AppVersion   string   `yaml:"app_version" env:"APPX_VER" env-default:"v0.0.0"`    // v0.0.0 - autodetect
		Environment  string   `env:"APPX_ENV" env-required:"true"`                        // ENV or cmd ARG:environment (local, dev, test, prod)
		AppLanguages []string `yaml:"app_languages" env:"APPX_LANGS" env-required:"true"` // language by "," separated
		AppTimeZones []string `yaml:"app_time_zones"`                                     // IANA names by "," separated; UTC is always registered

		// Log - настройки логирования и отладки
		LogLevel          string `yaml:"log_level"` // YAML or cmd ARG:log-level
		LogTimeFormat     string `yaml:"log_time_format"`
		LogJsonFormat     bool   `yaml:"log_json_format"`
		LogColorMode      bool   `yaml:"log_color_mode"`
		LogTraceIsEnabled bool   `yaml:"log_trace_is_enabled"`
		DebugIsEnabled    bool   `yaml:"debug_is_enabled"`

		// StackTrace - настройки стека вызовов формируемого в runtime ошибках
		StackTraceIsEnabled   bool     `yaml:"stack_trace_is_enabled"`
		StackTraceDepth       uint8    `yaml:"stack_trace_depth"`
		StackTraceShowFunc    bool     `yaml:"stack_trace_show_func"`
		StackTraceUpperBounds []string `yaml:"stack_trace_upper_bounds"`

		// Http статус вместо http.StatusInternalServerError
		// для необработанной ошибки (явно не обёрнутую в runtime ошибку)
		UnexpectedErrorHttpStatus uint16 `yaml:"unexpected_error_http_status"`

		// Sentry - настройки мониторинга и отслеживание ошибок [OPTIONAL]
		SentryDSN              string        `yaml:"sentry_dsn" env:"APPX_SENTRY_SENTRY_DSN"`
		SentryTracesSampleRate float64       `yaml:"sentry_traces_sample_rate"`
		SentryFlushTimeout     time.Duration `yaml:"sentry_flush_timeout"`

		// Postgres storage
		DBHost            string        `yaml:"db_host" env:"APPX_DB_HOST" env-required:"true"`
		DBPort            string        `yaml:"db_port" env:"APPX_DB_PORT" env-required:"true"`
		DBUsername        string        `yaml:"db_username" env:"APPX_DB_USER" env-required:"true"`
		DBPassword        string        `yaml:"db_password" env:"APPX_DB_PASSWORD" env-required:"true"`
		DBDatabase        string        `yaml:"db_database" env:"APPX_DB_NAME" env-required:"true"`
		DBMigrationsDir   string        `yaml:"db_migrations_dir"`
		DBMigrationsTable string        `yaml:"db_migrations_table" env:"APPX_DB_MIGRATIONS_TABLE"`
		DBMaxPoolSize     uint16        `yaml:"db_max_pool_size"`
		DBMaxConnLifetime time.Duration `yaml:"db_max_conn_lifetime"`
		DBMaxConnIdleTime time.Duration `yaml:"db_max_conn_idle_time"`
		DBTimeout         time.Duration `yaml:"db_timeout"`

		// Redis storage
		RedisHost         string        `yaml:"redis_host" env:"APPX_REDIS_HOST" env-required:"true"`
		RedisPort         string        `yaml:"redis_port" env:"APPX_REDIS_PORT" env-required:"true"`
		RedisPassword     string        `yaml:"redis_password" env:"APPX_REDIS_PASSWORD" env-required:"true"`
		RedisReadTimeout  time.Duration `yaml:"redis_read_timeout"`
		RedisWriteTimeout time.Duration `yaml:"redis_write_timeout"`

		// FileSystem storage (disabled)
		FSDirMode    uint32 `yaml:"fs_dir_mode"`
		FSCreateDirs bool   `yaml:"fs_create_dirs"`

		// S3 storage
		S3Host          string `yaml:"s3_host" env:"APPX_S3_HOST" env-required:"true"`
		S3Port          string `yaml:"s3_port" env:"APPX_S3_PORT" env-required:"true"`
		S3UseSSL        bool   `yaml:"s3_use_ssl" env:"APPX_S3_USE_SSL"`
		S3Username      string `yaml:"s3_username" env:"APPX_S3_USER" env-required:"true"`
		S3Password      string `yaml:"s3_password" env:"APPX_S3_PASSWORD" env-required:"true"`
		S3CreateBuckets bool   `yaml:"s3_create_buckets"`

		// FileProviders - провайдеры хранения файлов и изображений
		FileProviders struct {
			// S3
			ImageStorageName       string `yaml:"image_storage_name"`
			ImageStorageBucketName string `yaml:"image_storage_bucket_name"`

			// FileSystem (disabled)
			ImageStorage2Name    string `yaml:"image_storage2_name"`
			ImageStorage2RootDir string `yaml:"image_storage2_root_dir" env:"APPX_IMAGE_STORAGE2_ROOT_DIR" env-required:"true"`
		} `yaml:"file_providers"`

		// Mail - настройки службы отправки электронных писем
		// MailDefaultFrom - e-mail от которого отправляются письма [OPTIONAL].
		MailDefaultFrom string `yaml:"mail_default_from" env:"APPX_MAIL_DEFAULT_FROM" env-required:"true"`
		// MailSmtpHost - SMTP сервер для отправки почты (домен или IP).
		MailSmtpHost string `yaml:"mail_smtp_host" env:"APPX_MAIL_SMTP_HOST" env-required:"true"`
		// MailSmtpPort - порт SMTP сервера для отправки почты.
		MailSmtpPort string `yaml:"mail_smtp_port" env:"APPX_MAIL_SMTP_PORT" env-required:"true"`
		// MailSmtpUserName - адрес почтового ящика на SMTP сервере.
		MailSmtpUserName string `yaml:"mail_smtp_user_name" env:"APPX_MAIL_SMTP_USERNAME" env-required:"true"`
		// MailSmtpPassword - пароль почтового ящика на SMTP сервере.
		MailSmtpPassword string `yaml:"mail_smtp_password" env:"APPX_MAIL_SMTP_PASSWORD" env-required:"true"`

		// Telegram - настройки службы отправки сообщений [OPTIONAL]
		TelegramChannelName  string `yaml:"telegram_channel_name" env:"APPX_TELEGRAM_CHANNEL_NAME"`
		TelegramChannelToken string `yaml:"telegram_channel_token" env:"APPX_TELEGRAM_CHANNEL_TOKEN"`

		// API HTTP server - настройки HTTP сервера, принимающего запросы из вне
		HttpServerReadTimeout     time.Duration `yaml:"http_server_read_timeout"`
		HttpServerWriteTimeout    time.Duration `yaml:"http_server_write_timeout"`
		HttpServerShutdownTimeout time.Duration `yaml:"http_server_shutdown_timeout"`
		HttpServerBindIP          string        `yaml:"http_server_bind_ip" env:"APPX_HTTP_SERVER_BIND_IP" env-required:"true"`
		HttpServerPort            string        `yaml:"http_server_port" env:"APPX_HTTP_SERVER_PORT" env-required:"true"`

		// Monitoring HTTP server - настройки HTTP сервера для мониторинга системы
		MonitoringServerReadTimeout     time.Duration `yaml:"monitoring_server_read_timeout"`
		MonitoringServerWriteTimeout    time.Duration `yaml:"monitoring_server_write_timeout"`
		MonitoringServerShutdownTimeout time.Duration `yaml:"monitoring_server_shutdown_timeout"`
		MonitoringServerBindIP          string        `yaml:"monitoring_server_bind_ip" env:"APPX_MONITORING_SERVER_BIND_IP" env-required:"true"`
		MonitoringServerPort            string        `yaml:"monitoring_server_port" env:"APPX_MONITORING_SERVER_PORT" env-required:"true"`

		// Cors - настройки Cross-Origin Resource Sharing
		CorsAllowedOrigins   []string `yaml:"cors_allowed_origins" env:"APPX_CORS_ALLOWED_ORIGINS" env-required:"true"` // items by "," separated
		CorsAllowedMethods   []string `yaml:"cors_allowed_methods"`
		CorsAllowedHeaders   []string `yaml:"cors_allowed_headers"`
		CorsExposedHeaders   []string `yaml:"cors_exposed_headers"`
		CorsAllowCredentials bool     `yaml:"cors_allow_credentials"`

		// RefreshCookie - настройки cookie с refresh токеном (web-версия)
		RefreshCookie authcfg.RefreshCookie `yaml:"refresh_cookie"`

		// AccessControl - настройки доступа к системе
		AccessControl authcfg.AccessControl `yaml:"access_control"`

		// JWT - крипто ключи для шифрования JWT токенов [OPTIONAL]
		JWT authcfg.JWT `yaml:"jwt"`

		// Module settings - настройки модулей системы
		ModuleSettings struct {
			ProviderAccount struct {
				CompanyPageLogoProvider string `yaml:"company_page_logo_provider"` // FileProviders.ImageStorageName
			} `yaml:"provider_account"`

			FileStation struct {
				ImageProxyHost     string `yaml:"image_proxy_host" env:"APPX_IMAGE_PROXY_HOST" env-required:"true"`
				ImageProxyBasePath string `yaml:"image_proxy_base_path"`
				ImageProxyProvider string `yaml:"image_proxy_provider"` // FileProviders.ImageStorageName
			} `yaml:"file_station"`
		} `yaml:"module_settings"`

		// Validation - настройки валидации типов файлов и изображений
		ValidationFilesJson  modelcfg.FileType  `yaml:"validation_files_json"`
		ValidationImagesLogo modelcfg.ImageType `yaml:"validation_images_logo"`
		AllowedMimeTypes     []mime.Type        `yaml:"allowed_mime_types"`

		// TaskSchedule - настройки задач, запускаемых по расписанию
		TaskScheduleSettings struct {
			// Caption        string        `yaml:"caption"`
			ReloadSettings              processcfg.SchedulerTask `yaml:"reload_settings"`
			NotificationCheckConnPeriod time.Duration            `yaml:"notification_check_conn_period"`
			DefaultPeriodRatio          float64                  `yaml:"default_period_ratio"`
		} `yaml:"task_schedule_settings"`

		// TaskSchedule Auth - настройки задач модуля Auth, запускаемых по расписанию
		TaskScheduleAuth authcfg.TaskSchedule `yaml:"task_schedule_auth"`

		// TaskSchedule Mailer - настройки задач модуля Mailer, запускаемых по расписанию
		TaskScheduleMailer maliercfg.TaskSchedule `yaml:"task_schedule_mailer"`

		// TaskSchedule Notifier - настройки задач модуля Notifier, запускаемых по расписанию
		TaskScheduleNotifier notifiercfg.TaskSchedule `yaml:"task_schedule_notifier"`

		// TestUser - тестовый пользователь с указанными разрешениями [OPTIONAL]
		TestUserID       string `env:"TEST_USER_ID"`
		TestUserRealm    string `env:"TEST_USER_REALM"`
		TestUserKind     string `env:"TEST_USER_KIND"`
		TestUserLangCode string `env:"TEST_USER_LANG_CODE"`
	}
)
