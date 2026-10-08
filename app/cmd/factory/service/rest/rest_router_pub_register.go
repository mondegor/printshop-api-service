package rest

import (
	"net/http"

	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-core/mrpath"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrresp"

	"print-shop-back/internal/app"
	calculationsalgo "print-shop-back/internal/factory/calculations/algo/section/pub"
	calculationsquery "print-shop-back/internal/factory/calculations/queryhistory/section/pub"
	catalogbox "print-shop-back/internal/factory/catalog/box/section/pub"
	cataloglaminate "print-shop-back/internal/factory/catalog/laminate/section/pub"
	catalogpaper "print-shop-back/internal/factory/catalog/paper/section/pub"
	controlssubmitform "print-shop-back/internal/factory/controls/submitform/section/pub"
	dictionariesmaterialtype "print-shop-back/internal/factory/dictionaries/materialtype/section/pub"
	dictionariespapercolor "print-shop-back/internal/factory/dictionaries/papercolor/section/pub"
	dictionariespaperfacture "print-shop-back/internal/factory/dictionaries/paperfacture/section/pub"
	dictionariesprintformat "print-shop-back/internal/factory/dictionaries/printformat/section/pub"
	filestation "print-shop-back/internal/factory/filestation/section/pub"
	provideraccount "print-shop-back/internal/factory/provideraccounts/section/pub"
)

// RegisterRestRouterPubHandlers - регистрирует в указанном роутере обработчики секции PublicAPI.
func RegisterRestRouterPubHandlers(
	router mrserver.HttpRouter,
	opts app.Options,
	actionGroup mraccess.ActionGroup,
	userProvider mraccess.UserProvider,
) error {
	router.HandlerFunc(http.MethodGet, actionGroup.BasePath, mrresp.HandlerGetStatusOkAsJSON(opts.Logger))

	controllers, err := initing.CreateHttpControllers(
		opts.Logger,
		getPublicAPIControllers(opts),
		initing.WithCheckAccessMiddleware(opts.Logger, actionGroup, userProvider, opts.PermsProvider),
	)
	if err != nil {
		return err
	}

	router.Register(controllers...)

	return nil
}

func getPublicAPIControllers(opts app.Options) []initing.HttpModule {
	return []initing.HttpModule{
		catalogbox.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		cataloglaminate.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		catalogpaper.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		controlssubmitform.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		dictionariesmaterialtype.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		dictionariespapercolor.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		dictionariespaperfacture.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		dictionariesprintformat.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		provideraccount.InitHttpModule(
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
			opts.ImageURLBuilder,
		),
		filestation.InitHttpModule(
			opts.RequestParsers.String,
			opts.ResponseSenders.FileSender,
			func() (mrstorage.FileProviderAPI, mrpath.Builder, error) {
				fileAPI, err := opts.FileProviderPool.ProviderAPI(
					opts.Cfg.ModuleSettings.FileStation.ImageProxyProvider,
				)
				if err != nil {
					return nil, nil, err
				}

				basePath, err := mrpath.NewPlaceholder(
					opts.Cfg.ModuleSettings.FileStation.ImageProxyBasePath,
					mrpath.Placeholder,
				)
				if err != nil {
					return nil, nil, err
				}

				return fileAPI, basePath, nil
			},
		),
		calculationsquery.InitHttpModule(
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
		calculationsalgo.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.RequestParsers.PubParser,
			opts.ResponseSenders.Sender,
		),
	}
}
