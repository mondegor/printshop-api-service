package httpv1

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mraccess"
	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/calculations/queryhistory/module"
	"print-shop-back/internal/calculations/queryhistory/section/pub"
	"print-shop-back/internal/calculations/queryhistory/section/pub/entity"
	"print-shop-back/pkg/transport/model"
	pubvalidate "print-shop-back/pkg/transport/validate/pub"
)

const (
	queryHistoryURL     = "/v1/calculations/query-history"
	queryHistoryItemURL = "/v1/calculations/query-history/{id}"
)

type (
	// QueryHistory - comment struct.
	QueryHistory struct {
		parser  pubvalidate.RequestParser
		sender  mrserver.ResponseSender
		useCase pub.QueryResultUseCase
	}
)

// NewQueryHistory - создаёт контроллер QueryHistory.
func NewQueryHistory(parser pubvalidate.RequestParser, sender mrserver.ResponseSender, useCase pub.QueryResultUseCase) *QueryHistory {
	return &QueryHistory{
		parser:  parser,
		sender:  sender,
		useCase: useCase,
	}
}

// Handlers - возвращает обработчики контроллера QueryHistory.
func (ht *QueryHistory) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: queryHistoryItemURL, Permission: mraccess.PermissionEveryone, Func: ht.Get},
		{Method: http.MethodPost, URL: queryHistoryURL, Permission: mraccess.PermissionEveryone, Func: ht.Create},
	}
}

// Get - comment method.
func (ht *QueryHistory) Get(w http.ResponseWriter, r *http.Request) error {
	item, err := ht.useCase.GetItem(r.Context(), ht.getItemID(r))
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(w, http.StatusOK, item)
}

// Create - comment method.
func (ht *QueryHistory) Create(w http.ResponseWriter, r *http.Request) error {
	req := CreateQueryHistoryRequest{}

	if err := ht.parser.Validate(r, &req); err != nil {
		return err
	}

	item := entity.QueryHistoryItem{
		Caption: req.Caption,
		Params:  req.Params,
		Result:  req.Result,
	}

	itemID, err := ht.useCase.Create(r.Context(), item)
	if err != nil {
		return ht.wrapError(err, r)
	}

	return ht.sender.Send(
		w,
		http.StatusCreated,
		model.SuccessCreatedItemResponse{
			ItemID: itemID.String(),
		},
	)
}

func (ht *QueryHistory) getItemID(r *http.Request) uuid.UUID {
	return ht.parser.PathParamUUID(r, "id")
}

func (ht *QueryHistory) getRawItemID(r *http.Request) string {
	return ht.parser.PathParamString(r, "id")
}

func (ht *QueryHistory) wrapError(err error, r *http.Request) error {
	if errors.Is(err, errors.ErrRecordNotFound) {
		return module.ErrQueryHistoryNotFound.Wrap(err, ht.getRawItemID(r))
	}

	return err
}
