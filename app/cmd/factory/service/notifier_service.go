package service

import (
	"github.com/mondegor/go-components/mrnotifier/notifier/entity"
	"github.com/mondegor/go-components/mrnotifier/notifier/service/produce"
	"github.com/mondegor/go-components/wire/mrmailer"
	"github.com/mondegor/go-components/wire/mrnotifier/processor"
	"github.com/mondegor/go-components/wire/mrnotifier/producer"
	"github.com/mondegor/go-components/wire/mrnotifier/scheduler"
	"github.com/mondegor/go-core/mrprocess"
	"github.com/mondegor/go-core/mrprocess/consume"
	"github.com/mondegor/go-core/mrprocess/job/task"
	"github.com/mondegor/go-core/mrprocess/schedule"
	"github.com/mondegor/go-core/mrstorage/mrsql"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/app"
)

const (
	serviceNotifierTableName             = "printshop_global.notifier_notices"
	serviceNotifierQueueTableName        = "printshop_global.notifier_queue"
	serviceNotifierPrimaryKey            = "notice_id"
	serviceNotifierTemplatesTableName    = "printshop_global.notifier_templates"
	serviceNotifierTemplateVarsTableName = "printshop_global.notifier_template_vars"
)

// InitNotifierAPI - создаёт отправителя сообщений получателям.
func InitNotifierAPI(opts app.Options) *produce.NoteProducer {
	log.Info(opts.Logger, "Create and init notifier sender API")

	return producer.InitService(
		opts.PostgresConnManager,
		opts.TraceManager,
		mrsql.DBTableInfo{
			Name:       serviceNotifierTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		mrsql.DBTableInfo{
			Name:       serviceNotifierQueueTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		produce.WithRetryAttempts(int16(opts.Cfg.TaskScheduleNotifier.SendRetryAttempts)),
	)
}

// InitNotifierProcessorService - создаёт обработчик очереди уведомлений
// (формирует письма по шаблонам и передаёт их в рассылку).
func InitNotifierProcessorService(opts app.Options) *consume.MessageProcessor[entity.Note] {
	log.Info(opts.Logger, "Create and init notice processor service")

	return processor.InitService(
		opts.PostgresConnManager,
		mrmailer.NoticeToMessageAdapterFunc(opts.MailerAPI.Send),
		opts.ErrorHandler,
		opts.Logger,
		opts.TraceManager,
		mrsql.DBTableInfo{
			Name:       serviceNotifierTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		mrsql.DBTableInfo{
			Name:       serviceNotifierQueueTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		serviceNotifierTemplatesTableName,
		serviceNotifierTemplateVarsTableName,
		processor.WithDefaultLang(opts.LocalePool.Localizer().Language()),
		processor.WithNoticeProcessorOpts(
			consume.WithCaptionPrefix[entity.Note]("Notifier/"),
			consume.WithReadyTimeout[entity.Note](opts.Cfg.TaskScheduleNotifier.NoticeProcessor.ReadyTimeout),
			consume.WithReadPeriodStrategy[entity.Note](
				mrprocess.NewDoubleDelayedStartStrategy(
					opts.Cfg.TaskScheduleNotifier.NoticeProcessor.ReadPeriod,
					opts.Cfg.TaskScheduleSettings.DefaultPeriodRatio,
				),
			),
			consume.WithConsumerTimeout[entity.Note](
				opts.Cfg.TaskScheduleNotifier.NoticeProcessor.ConsumerReadTimeout,
				opts.Cfg.TaskScheduleNotifier.NoticeProcessor.ConsumerWriteTimeout,
			),
			consume.WithHandlerTimeout[entity.Note](opts.Cfg.TaskScheduleNotifier.NoticeProcessor.HandlerTimeout),
			consume.WithQueueSize[entity.Note](int(opts.Cfg.TaskScheduleNotifier.NoticeProcessor.QueueSize)),
			consume.WithWorkersCount[entity.Note](int(opts.Cfg.TaskScheduleNotifier.NoticeProcessor.WorkersCount)),
			consume.WithSignalExecuteHandler[entity.Note](
				opts.PostgresNotificationService.MustFind(opts.Cfg.TaskScheduleNotifier.NoticeProcessor.NotificationChannel),
			),
		),
	)
}

// InitNotifierSchedulerService - создаёт планировщик задач уведомлений
// (перенос/повтор уведомлений и очистка очереди).
func InitNotifierSchedulerService(opts app.Options) *schedule.TaskScheduler {
	log.Info(opts.Logger, "Create and init notice scheduler service")

	return scheduler.InitService(
		opts.PostgresConnManager,
		opts.EventEmitter,
		opts.ErrorHandler,
		opts.Logger,
		opts.TraceManager,
		mrsql.DBTableInfo{
			Name:       serviceNotifierTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		mrsql.DBTableInfo{
			Name:       serviceNotifierQueueTableName,
			PrimaryKey: serviceNotifierPrimaryKey,
		},
		scheduler.WithCaptionPrefix("Notifier/"),
		scheduler.WithChangeBatchSize(int(opts.Cfg.TaskScheduleNotifier.ChangeQueueBatchSize)),
		scheduler.WithChangeRetryTimeout(opts.Cfg.TaskScheduleNotifier.ChangeRetryTimeout),
		scheduler.WithChangeRetryDelayed(opts.Cfg.TaskScheduleNotifier.ChangeRetryDelayed),
		scheduler.WithCleanBatchSize(int(opts.Cfg.TaskScheduleNotifier.CleanQueueBatchSize)),
		scheduler.WithTaskChangeFromToRetryOpts(
			task.WithCaptionPrefix("Notifier/"),
			task.WithStartup(false),
			task.WithPeriodStrategy(
				mrprocess.NewDoubleDelayedStartStrategy(
					opts.Cfg.TaskScheduleNotifier.ChangeFromToRetry.Period,
					opts.Cfg.TaskScheduleSettings.DefaultPeriodRatio,
				),
			),
			task.WithTimeout(opts.Cfg.TaskScheduleNotifier.ChangeFromToRetry.Timeout),
		),
		scheduler.WithTaskCleanNoticesOpts(
			task.WithCaptionPrefix("Notifier/"),
			task.WithStartup(false),
			task.WithPeriodStrategy(
				mrprocess.NewQuadQuickStartStrategy(
					opts.Cfg.TaskScheduleNotifier.CleanQueue.Period,
					opts.Cfg.TaskScheduleSettings.DefaultPeriodRatio,
				),
			),
			task.WithTimeout(opts.Cfg.TaskScheduleNotifier.CleanQueue.Timeout),
		),
	)
}
