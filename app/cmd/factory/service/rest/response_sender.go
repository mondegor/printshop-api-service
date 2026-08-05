package rest

import (
	"github.com/mondegor/go-components/wire/mrauth/mapping"
	"github.com/mondegor/go-core/errors/runtime/hint"
	"github.com/mondegor/go-webcore/mrserver"
	"github.com/mondegor/go-webcore/mrserver/mrjson"
	"github.com/mondegor/go-webcore/mrserver/mrresp"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/app"
)

// CreateResponseSenders - создаются и возвращаются компоненты для отправки ответа клиенту.
func CreateResponseSenders(logger log.Logger) (app.ResponseSenders, error) {
	log.Info(logger, "Create and init base response senders")

	sender := mrresp.NewSender(mrjson.NewEncoder())

	return app.ResponseSenders{
		Sender:     sender,
		FileSender: mrresp.NewFileSender(sender, logger),
	}, nil
}

// NewErrorResponseSender - создаёт объект mrresp.ErrorSender.
func NewErrorResponseSender(opts app.Options) (*mrresp.ErrorSender, error) {
	log.Info(opts.Logger, "Create and init error response sender")

	statusMapper, err := mrserver.NewHttpErrorStatusMapper(
		int(opts.Cfg.UnexpectedErrorHttpStatus),
		mapping.OptionErrorCodeToHttpStatus()...,
	)
	if err != nil {
		return nil, err
	}

	return mrresp.NewErrorSender(
		mrjson.NewEncoder(),
		opts.ErrorHandler,
		func(err error) string {
			return hint.Extract(err).ErrorID()
		},
		opts.Logger,
		opts.RequestParsers.Locale,
		statusMapper,
		opts.DebugFunc,
	), nil
}
