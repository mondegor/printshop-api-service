package rest

import (
	"net/http"

	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-webcore/mrcore/initing"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrresp"

	"print-shop-back/internal/app"
	controlselementtemplatevalidate "print-shop-back/internal/controls/elementtemplate/section/adm/validate"
	controlssubmitformvalidate "print-shop-back/internal/controls/submitform/section/adm/validate"
	catalogbox "print-shop-back/internal/factory/catalog/box/section/adm"
	cataloglaminate "print-shop-back/internal/factory/catalog/laminate/section/adm"
	catalogpaper "print-shop-back/internal/factory/catalog/paper/section/adm"
	controlssubmitformapi "print-shop-back/internal/factory/controls/elementtemplate/api/header"
	controlselementtemplate "print-shop-back/internal/factory/controls/elementtemplate/section/adm"
	controlssubmitform "print-shop-back/internal/factory/controls/submitform/section/adm"
	dictionariesmaterialtype "print-shop-back/internal/factory/dictionaries/materialtype/section/adm"
	dictionariespapercolor "print-shop-back/internal/factory/dictionaries/papercolor/section/adm"
	dictionariespaperfacture "print-shop-back/internal/factory/dictionaries/paperfacture/section/adm"
	dictionariesprintformat "print-shop-back/internal/factory/dictionaries/printformat/section/adm"
	provideraccounts "print-shop-back/internal/factory/provideraccounts/section/adm"
	provideraccountsvalidate "print-shop-back/internal/provideraccounts/section/adm/validate"
	pkgcontrolsvalidate "print-shop-back/pkg/controls/validate"
	pkgprovideraccountsvalidate "print-shop-back/pkg/provideraccounts/validate"
)

// RegisterRestRouterAdmHandlers - регистрирует в указанном роутере обработчики секции AdminAPI.
func RegisterRestRouterAdmHandlers(
	router mrserver.HttpRouter,
	opts app.Options,
	actionGroup mraccess.ActionGroup,
	userProvider mraccess.UserProvider,
) error {
	router.HandlerFunc(http.MethodGet, actionGroup.BasePath, mrresp.HandlerGetStatusOkAsJSON(opts.Logger))

	controllers, err := initing.CreateHttpControllers(
		opts.Logger,
		getAdminAPIControllers(opts),
		initing.WithCheckAccessMiddleware(opts.Logger, actionGroup, userProvider, opts.PermsProvider),
	)
	if err != nil {
		return err
	}

	router.Register(controllers...)

	return nil
}

func getAdminAPIControllers(opts app.Options) []initing.HttpModule {
	return []initing.HttpModule{
		catalogbox.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			app.PageSizeMax,
		),
		cataloglaminate.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			opts.DictionariesMaterialTypeAPI,
			app.PageSizeMax,
		),
		catalogpaper.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			opts.DictionariesMaterialTypeAPI,
			opts.DictionariesPaperColorAPI,
			opts.DictionariesPaperFactureAPI,
			app.PageSizeMax,
		),
		controlselementtemplate.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			controlselementtemplatevalidate.NewParser(
				opts.RequestParsers.AdmParser,
				opts.RequestParsers.FileJson,
				pkgcontrolsvalidate.NewDetailingParser(opts.Logger),
			),
			opts.ResponseSenders.FileSender,
			app.PageSizeMax,
		),
		controlssubmitform.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.Locker,
			controlssubmitformvalidate.NewParser(
				opts.RequestParsers.AdmParser,
				opts.RequestParsers.FileJson,
				pkgcontrolsvalidate.NewDetailingParser(opts.Logger),
			),
			opts.ResponseSenders.Sender,
			opts.ResponseSenders.FileSender,
			controlssubmitformapi.NewElementTemplate(
				opts.PostgresConnManager,
				opts.Tracer,
			),
			app.PageSizeMax,
		),
		dictionariesmaterialtype.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			app.PageSizeMax,
		),
		dictionariespapercolor.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			app.PageSizeMax,
		),
		dictionariespaperfacture.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			app.PageSizeMax,
		),
		dictionariesprintformat.InitHttpModule(
			opts.Logger,
			opts.EventEmitter,
			opts.PostgresConnManager,
			opts.RequestParsers.AdmParser,
			opts.ResponseSenders.Sender,
			app.PageSizeMax,
		),
		provideraccounts.InitHttpModule(
			opts.Logger,
			opts.PostgresConnManager,
			provideraccountsvalidate.NewParser(
				opts.RequestParsers.AdmParser,
				pkgprovideraccountsvalidate.NewPublicStatusParser(opts.Logger),
			),
			opts.ResponseSenders.Sender,
			opts.ImageURLBuilder,
			app.PageSizeMax,
		),
	}
}
