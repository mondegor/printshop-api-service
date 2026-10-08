package httpv1

import (
	"net/http"

	"github.com/mondegor/go-webcore/mrserver"

	"print-shop-back/internal/warehousing/actiongroup/usr"
	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/actiongroup/usr/transport/model"
	"print-shop-back/internal/warehousing/module"
	"print-shop-back/internal/warehousing/xtype"
	usrvalidate "print-shop-back/pkg/transport/validate/usr"
)

const (
	storeListURL = "/v1/warehousing/stores"
)

type (
	// Store - comment struct.
	Store struct {
		parser       usrvalidate.RequestParser
		sender       mrserver.ResponseSender
		serviceStore usr.StoreService
	}
)

// NewStore - создаёт контроллер Store.
func NewStore(
	parser usrvalidate.RequestParser,
	sender mrserver.ResponseSender,
	serviceStore usr.StoreService,
) *Store {
	return &Store{
		parser:       parser,
		sender:       sender,
		serviceStore: serviceStore,
	}
}

// Handlers - возвращает обработчики контроллера Store.
func (ht *Store) Handlers() []mrserver.HttpHandler {
	return []mrserver.HttpHandler{
		{Method: http.MethodGet, URL: storeListURL, Func: ht.GetList},
	}
}

// GetList - comment method.
func (ht *Store) GetList(w http.ResponseWriter, r *http.Request) error {
	items, hasNext, err := ht.serviceStore.GetList(r.Context(), ht.listParams(r))
	if err != nil {
		return err
	}

	return ht.sender.Send(
		w,
		http.StatusOK,
		model.StoreListResponse{
			Items:   items,
			Cursor:  entity.CreateStoreCursorValue(items),
			HasNext: hasNext,
		},
	)
}

func (ht *Store) listParams(r *http.Request) dto.StoreParams {
	return dto.StoreParams{
		AccountID: ht.parser.UserID(r),
		Filter: dto.StoreListFilter{
			SearchCode:        ht.parser.FilterString(r, module.ParamNameFilterSearchStoreCode),
			SearchTerritories: ht.parser.FilterUint64List(r, module.ParamNameFilterSearchTerritoryIDs),
		},
		Cursor: xtype.NewStoreCursor(ht.parser.CursorParams(r)),
	}
}
