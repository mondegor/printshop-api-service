package integration

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/util/xio"
	"github.com/mondegor/go-core/wire/mrlog"
	"github.com/mondegor/go-core/wire/mrtrace"
	"github.com/mondegor/go-storage/mrtests/pgtest"
	"github.com/mondegor/go-storage/mrtests/redistest"
	"github.com/mondegor/go-webcore/mrtests/helpers"
	"github.com/stretchr/testify/require"

	"print-shop-back/cmd/factory"
	"print-shop-back/cmd/factory/service/rest"
	"print-shop-back/config"
	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/adapter/trace"
	"print-shop-back/internal/app"
	"print-shop-back/tests"
)

type (
	// HttpHandlerTester - вспомогательный объект для тестирования http обработчиков.
	HttpHandlerTester struct {
		parentT *testing.T
		ctx     context.Context
		opts    app.Options
		router  http.Handler
		fpool   *mrstorage.FileProviderPool
	}
)

// NewHandlerTester - создаёт объект HttpHandlerTester.
func NewHandlerTester(t *testing.T) *HttpHandlerTester {
	t.Helper()

	ctx := context.Background()
	cfg, err := config.Create(
		config.CmdArgs{
			WorkDir:     tests.AppWorkDir(),
			Environment: "tests",
		},
		os.Stdout,
	)
	require.NoError(t, err)

	skipIfS3Unavailable(t, cfg)

	logger := log.NopLogger()
	tracer := trace.NopTracer()
	traceManager, err := mrtrace.InitTraceContextManager(mrlog.DefaultProcessIDs(), logger)
	require.NoError(t, err)

	pgt := pgtest.NewTester(t, tests.DBSchemas(), tests.ExcludedDBTables())
	pgt.ApplyMigrations(t, tests.AppMigrationsDir())

	rds := redistest.NewTester(t)

	fpool, err := factory.InitFileProviderPool(logger, tracer, cfg)
	require.NoError(t, err)

	opts, err := factory.InitAppEnvironment(
		app.Options{
			Cfg:                 cfg,
			Logger:              logger,
			Tracer:              tracer,
			TraceManager:        traceManager,
			OpenedResources:     xio.NewCloseManager(logger),
			PostgresConnManager: pgt.ConnManager(),
			RedisAdapter:        rds.Conn(),
			FileProviderPool:    fpool,
		},
	)
	require.NoError(t, err)

	router, err := rest.InitRestRouterWithHandlers(opts)
	require.NoError(t, err)

	return &HttpHandlerTester{
		parentT: t,
		ctx:     ctx,
		opts:    opts,
		router:  router,
		fpool:   fpool,
	}
}

// Context - возвращает текущий контекст.
func (t *HttpHandlerTester) Context() context.Context {
	return t.ctx
}

// Options - возвращает опции приложения.
func (t *HttpHandlerTester) Options() app.Options {
	return t.opts
}

// Router - возвращает текущий роутер.
func (t *HttpHandlerTester) Router() http.Handler {
	return t.router
}

// ExecRequest - исполняет текущий запрос.
func (t *HttpHandlerTester) ExecRequest(r *helpers.HttpRequest, structResponse any) (statusCode int, err error) {
	return r.Exec(t.router, structResponse)
}

// Clean - очищает ресурсы приложения после завершения тестирования обработчика
// (контейнеры Postgres и Redis освобождаются через t.Cleanup теста-владельца).
func (t *HttpHandlerTester) Clean() {
	t.opts.OpenedResources.Close()

	err := t.fpool.Close()
	require.NoError(t.parentT, err)
}

// skipIfS3Unavailable - пропускает тест, если S3-хранилище недоступно (например, в CI).
func skipIfS3Unavailable(t *testing.T, cfg config.Config) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(cfg.S3Host, cfg.S3Port), time.Second)
	if err != nil {
		t.Skipf("S3 storage is unavailable, test skipped: %v", err)
	}

	_ = conn.Close()
}
