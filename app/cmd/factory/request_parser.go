package factory

import (
	"fmt"

	mrauthconf "github.com/mondegor/go-components/wire/mrauth/config"
	"github.com/mondegor/go-components/wire/mrauth/mapping"
	"github.com/mondegor/go-core/util/mime"
	"github.com/mondegor/go-webcore/mrserver/mrchi"
	"github.com/mondegor/go-webcore/mrserver/mrjson"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	"github.com/mondegor/go-webcore/mrview"
	"github.com/mondegor/go-webcore/mrview/mrplayvalidator"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/app"
	mrcalcvalidate "print-shop-back/pkg/mrcalc/validate"
	validate2 "print-shop-back/pkg/transport/validate"
)

const (
	langURLParam = "lang"
	tzURLParam   = "tz"
)

// InitRequestParsers - создаются и возвращаются парсеры запросов клиента.
func InitRequestParsers(opts app.Options) (app.RequestParsers, error) {
	log.Info(opts.Logger, "Create and init base request parsers")

	validator, err := initUserValidator(opts)
	if err != nil {
		return app.RequestParsers{}, err
	}

	// WARNING: функция использует контекст роутера chi,
	// поэтому её можно менять только при смене самого роутера
	pathFunc := mrchi.URLPathParam

	registeredMimeTypes := mime.NewTypeList(opts.Cfg.AllowedMimeTypes)

	jsonMimeTypes, err := registeredMimeTypes.TypesByExts(opts.Cfg.ValidationFilesJson.Extensions)
	if err != nil {
		return app.RequestParsers{}, err
	}

	logoMimeTypes, err := registeredMimeTypes.TypesByExts(opts.Cfg.ValidationImagesLogo.File.Extensions)
	if err != nil {
		return app.RequestParsers{}, err
	}

	parsers := app.RequestParsers{
		// Bool:       parser.NewBool(),
		// DateTime:   parser.NewDateTime(),
		Int64:      parser.NewInt64(opts.Logger),
		ItemStatus: parser.NewItemStatus(opts.Logger),
		Uint64:     parser.NewUint64(pathFunc, opts.Logger),
		ListCursor: parser.NewListCursor(
			opts.Logger,
			parser.ListCursorOptions{
				LimitMax:     int(opts.Cfg.ModuleSettings.General.PageSizeMax),
				LimitDefault: int(opts.Cfg.ModuleSettings.General.PageSizeDefault),
			},
		),
		ListPager: parser.NewListPager(
			opts.Logger,
			parser.ListPagerOptions{
				PageSizeMax:     int(opts.Cfg.ModuleSettings.General.PageSizeMax),
				PageSizeDefault: int(opts.Cfg.ModuleSettings.General.PageSizeDefault),
			},
		),
		ListSorter: parser.NewListSorter(opts.Logger, parser.ListSorterOptions{}),
		String:     parser.NewString(pathFunc, opts.Logger),
		UUID:       parser.NewUUID(pathFunc, opts.Logger),
		Validator:  parser.NewValidator(mrjson.NewDecoder(), validator),
		ClientIP:   parser.NewClientIP(opts.Logger),
		User:       parser.NewUser(opts.Logger),
		Locale:     parser.NewLocale(opts.LocalePool, opts.Logger, langURLParam),
		TimeZone:   parser.NewTimeZone(opts.LocationList, opts.Logger, tzURLParam),
		FileJson: parser.NewFile(
			opts.Logger,
			parser.WithFileMinSize(int64(opts.Cfg.ValidationFilesJson.MinSize)),
			parser.WithFileMaxSize(int64(opts.Cfg.ValidationFilesJson.MaxSize)),
			parser.WithFileMaxFiles(int(opts.Cfg.ValidationFilesJson.MaxFiles)),
			parser.WithFileCheckRequestContentType(opts.Cfg.ValidationFilesJson.CheckRequestContentType),
			parser.WithFileAllowedMimeTypes(jsonMimeTypes),
		),
		ImageLogo: parser.NewImage(
			opts.Logger,
			parser.WithImageMaxWidth(int32(opts.Cfg.ValidationImagesLogo.MaxWidth)),
			parser.WithImageMaxHeight(int32(opts.Cfg.ValidationImagesLogo.MaxHeight)),
			parser.WithImageCheckBody(opts.Cfg.ValidationImagesLogo.CheckBody),
			parser.WithImageFileOpts(
				parser.WithFileMinSize(int64(opts.Cfg.ValidationImagesLogo.File.MinSize)),
				parser.WithFileMaxSize(int64(opts.Cfg.ValidationImagesLogo.File.MaxSize)),
				parser.WithFileMaxFiles(int(opts.Cfg.ValidationImagesLogo.File.MaxFiles)),
				parser.WithFileCheckRequestContentType(opts.Cfg.ValidationImagesLogo.File.CheckRequestContentType),
				parser.WithFileAllowedMimeTypes(logoMimeTypes),
			),
		),
	}

	parsers.Parser = validate2.NewParser(
		parsers.Int64,
		parsers.Uint64,
		parsers.String,
		parsers.UUID,
		parsers.Validator,
		parsers.ClientIP,
		parsers.User,
		parsers.Locale,
		parsers.ListCursor,
	)

	parsers.ExtendParser = validate2.NewExtendParser(
		parsers.Parser,
		parsers.ItemStatus,
		parsers.ListPager,
		parsers.ListSorter,
	)

	return parsers, nil
}

// initUserValidator - создаёт валидатор структур запроса с зарегистрированными
// пользовательскими тегами.
//
// Списки языков и часовых поясов берутся у самих пула и списка поясов, а не у конфигурации,
// из которой те созданы: пул приводит коды языков к каноничному виду ("en_US" -> "en-US"),
// а список поясов всегда регистрирует UTC, даже если его не указывали. Взяв конфигурацию,
// теги отвергали бы значения, которые приложение само же и отдаёт.
func initUserValidator(opts app.Options) (*mrplayvalidator.ValidatorAdapter, error) {
	log.Info(opts.Logger, "Create and init data validator")

	customTags := []mrview.Tag{
		mrview.TagArticle(),
		mrauthconf.TagEmail(),
		mrauthconf.TagPhone(),
		mrauthconf.TagEmailPhone(),
		mrauthconf.TagLang(
			opts.LocalePool.Languages(),
		),
		mrauthconf.TagTimeZone(
			opts.LocationList.TimeZones(),
		),
		mrview.TagVariable(),
		mrview.TagName(),
		mrview.TagRewriteName(),
		mrview.TagPassword(),
		mrauthconf.TagRealm(
			mapping.OptionUserRealmsToStringRealms(opts.Cfg.AccessControl.Realms),
		),
		{
			Name:         "tag_2d_size",
			ValidateFunc: mrcalcvalidate.Size2d,
		},
		{
			Name:         "tag_3d_size",
			ValidateFunc: mrcalcvalidate.Size3d,
		},
	}

	validator := mrplayvalidator.New(opts.Logger)

	for _, tag := range customTags {
		if err := validator.Register(tag.Name, tag.ValidateFunc); err != nil {
			return nil, fmt.Errorf("tag %s: %w", tag.Name, err)
		}
	}

	return validator, nil
}
