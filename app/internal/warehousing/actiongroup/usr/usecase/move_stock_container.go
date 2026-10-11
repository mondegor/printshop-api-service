package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/errors"
	"github.com/mondegor/go-core/mrevent"

	"print-shop-back/internal/warehousing/actiongroup/usr"
	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/enum/locationkind"
)

type (
	// MoveStockContainer - перемещение остатков контейнера на другое место хранения.
	MoveStockContainer struct {
		storageStock            usr.StockStorage
		serviceStore            usr.StoreService
		useCaseRefreshLocations refreshLocationsUseCase
		eventEmitter            mrevent.Emitter
		errorWrapper            errors.Wrapper
		errorNotFoundWrapper    errors.Wrapper
	}
)

// NewMoveStockContainer - создаёт объект MoveStockContainer.
func NewMoveStockContainer(
	storageStock usr.StockStorage,
	serviceStore usr.StoreService,
	useCaseRefreshLocations refreshLocationsUseCase,
	eventEmitter mrevent.Emitter,
) *MoveStockContainer {
	return &MoveStockContainer{
		storageStock:            storageStock,
		serviceStore:            serviceStore,
		useCaseRefreshLocations: useCaseRefreshLocations,
		eventEmitter:            eventEmitter,
		errorWrapper:            errors.NewServiceOperationFailedWrapper(),
		errorNotFoundWrapper:    errors.NewServiceRecordNotFoundWrapper(),
	}
}

// Execute - перемещает указанное количество контейнеров остатка (0 - все) на новое место хранения
// и возвращает ID остатка на новом месте.
func (uc *MoveStockContainer) Execute(ctx context.Context, item dto.MoveStockContainer) (movedStockID uint64, err error) {
	if item.AccountID == uuid.Nil {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("accountId is empty")
	}

	if item.StockID == 0 {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("stockId is zero")
	}

	if item.LocationID == 0 {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("locationId is zero")
	}

	stock, err := uc.storageStock.FetchOne(ctx, item.AccountID, item.StockID)
	if err != nil {
		return 0, uc.errorNotFoundWrapper.Wrap(err)
	}

	if stock.ContainerID == item.LocationID {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("stock.ContainerId and item.LocationId are equal")
	}

	if item.Quantity > stock.ContainerQuantity {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("item.ExemplarQuantity > stock.Quantity") // user  usecase error
	}

	if item.Quantity == 0 {
		item.Quantity = stock.ContainerQuantity
	}

	if stock.LocationID == item.LocationID {
		return stock.ID, nil
	}

	if err = uc.serviceStore.CheckAvailability(ctx, item.AccountID, item.LocationID); err != nil {
		return 0, uc.errorWrapper.Wrap(err) // user  usecase error
	}

	// ограничение на вложение групповых контейнеров
	if locationkind.Is(stock.ContainerID, locationkind.Group) && !locationkind.Is(item.LocationID, locationkind.Store) {
		return 0, errors.ErrInternalIncorrectInputData.WithDetails("G/ container add error")
	}

	newStock := stock
	newStock.ID = 0
	newStock.LocationID = item.LocationID
	newStock.ContainerQuantity = item.Quantity
	stock.ContainerQuantity -= item.Quantity

	// TODO: транзакция

	if newStock.ID, err = uc.storageStock.InsertOrUpdate(ctx, newStock); err != nil {
		return 0, uc.errorWrapper.Wrap(err)
	}

	if stock.ContainerQuantity > 0 {
		if stock.ID, err = uc.storageStock.UpdateQuantity(ctx, stock.AccountID, stock.ID, stock.ContainerQuantity); err != nil {
			return 0, uc.errorWrapper.Wrap(err)
		}
	} else {
		if err = uc.storageStock.Delete(ctx, stock.AccountID, stock.ID); err != nil {
			return 0, uc.errorWrapper.Wrap(err)
		}
	}

	// TODO: событие отправляется до пересчёта локаций (в AddMoreStockContainer - после),
	// поэтому при ошибке пересчёта оно уже ушло; выбрать единый порядок.
	uc.eventEmitter.Emit(ctx, "Move", "accountId", item.AccountID, "stockId", stock.ID, "newStockId", newStock.ID)

	// TODO: через событие сделать
	if err = uc.useCaseRefreshLocations.Execute(ctx, []uint64{newStock.LocationID, stock.LocationID}); err != nil {
		return 0, uc.errorWrapper.Wrap(err)
	}

	return newStock.ID, nil
}
