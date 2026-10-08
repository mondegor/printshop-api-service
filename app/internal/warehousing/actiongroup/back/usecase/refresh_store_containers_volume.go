package usecase

import (
	"context"
	"math"

	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/util/slices/ordered"

	"print-shop-back/internal/adapter/log"
	"print-shop-back/internal/warehousing/actiongroup/back/dto"
	"print-shop-back/internal/warehousing/actiongroup/back/util/locationstock"
	"print-shop-back/internal/warehousing/enum/locationkind"
	"print-shop-back/internal/warehousing/module"
)

type (
	// RefreshStoreContainersVolume - comment struct.
	RefreshStoreContainersVolume struct {
		storeStorage storeStorage
		storageStock stockStorage
		logger       log.Logger
		errorWrapper errors.Wrapper
	}

	storeStorage interface {
		UpdateContainersVolume(ctx context.Context, rows []dto.LocationContainersVolume) error
	}
)

// NewRefreshStoreContainersVolume - создаёт объект RefreshStoreContainersVolume.
func NewRefreshStoreContainersVolume(
	storageStore storeStorage,
	storageStock stockStorage,
	logger log.Logger,
) *RefreshStoreContainersVolume {
	return &RefreshStoreContainersVolume{
		storeStorage: storageStore,
		storageStock: storageStock,
		logger:       logger,
		errorWrapper: errors.NewServiceOperationFailedWrapper(),
	}
}

// Execute - comment method.
func (uc *RefreshStoreContainersVolume) Execute(ctx context.Context, storeIDs []uint64) error {
	storeIDs = ordered.FilterFunc(
		storeIDs,
		func(el uint64) bool {
			return locationkind.Is(el, locationkind.Store)
		},
	)

	// TODO: склады разбить на чанки, чтобы уменьшить кол-во получаемых стоков
	if len(storeIDs) == 0 {
		return nil
	}

	storeIDs = ordered.SortedUnique(storeIDs)

	stocks, hasNext, err := uc.storageStock.FetchByLocationIDs(
		ctx,
		storeIDs,
		mrstorage.IDCursor{
			Limit: min(module.GroupContainersMax*len(storeIDs), math.MaxInt16),
		},
	)
	if err != nil {
		return uc.errorWrapper.Wrap(err)
	}

	if hasNext {
		uc.logger.Error(
			ctx,
			"there are more containers in stores than allowed value",
			"store_ids", storeIDs,
			"limit", len(stocks),
		)
	}

	stores := make([]dto.LocationContainersVolume, 0, len(storeIDs))

	for s := range locationstock.ChunkByLocationID(stocks) {
		stores = append(
			stores,
			dto.LocationContainersVolume{
				LocationID:  s[0].LocationID, // существование гарантирует метод ChunkByLocationID()
				TotalVolume: locationstock.CalcContainersVolume(s),
			},
		)
	}

	// помечаются все склады со стоками обработанными
	for i := range stocks {
		if index := ordered.BinaryIndex(storeIDs, stocks[i].LocationID); index >= 0 {
			storeIDs[index] = 0
		}
	}

	for _, storeID := range storeIDs {
		if storeID == 0 {
			continue
		}

		// все необработанные склады не содержат стоков
		stores = append(
			stores,
			dto.LocationContainersVolume{
				LocationID: storeID,
			},
		)
	}

	return uc.storeStorage.UpdateContainersVolume(ctx, stores)
}
