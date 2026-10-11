package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mondegor/go-core/errors"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"print-shop-back/internal/warehousing/actiongroup/usr/dto"
	"print-shop-back/internal/warehousing/actiongroup/usr/entity"
	"print-shop-back/internal/warehousing/actiongroup/usr/usecase"
	"print-shop-back/internal/warehousing/actiongroup/usr/usecase/mock"
)

//go:generate mockgen -source=create_stock_container.go -destination=mock/create_stock_container.go -package=mock

const (
	moveStockID       uint64 = 100
	moveToStoreID     uint64 = 8 // вид локации: склад
	moveStockQuantity        = 10
)

type MoveStockContainerSuite struct {
	suite.Suite

	ctrl             *gomock.Controller
	ctx              context.Context
	accountID        uuid.UUID
	storageStock     *mock.MockStockStorage
	serviceStore     *mock.MockStoreService
	refreshLocations *mock.MockrefreshLocationsUseCase
	eventEmitter     *mock.MockEmitter
	uc               *usecase.MoveStockContainer
}

func TestMoveStockContainerSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(MoveStockContainerSuite))
}

func (s *MoveStockContainerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.accountID = uuid.New()
	s.storageStock = mock.NewMockStockStorage(s.ctrl)
	s.serviceStore = mock.NewMockStoreService(s.ctrl)
	s.refreshLocations = mock.NewMockrefreshLocationsUseCase(s.ctrl)
	s.eventEmitter = mock.NewMockEmitter(s.ctrl)
	s.uc = usecase.NewMoveStockContainer(
		s.storageStock,
		s.serviceStore,
		s.refreshLocations,
		s.eventEmitter,
	)
}

// stock - остаток контейнера на складе testStoreID.
func (s *MoveStockContainerSuite) stock() entity.Stock {
	return entity.Stock{
		ID:                moveStockID,
		AccountID:         s.accountID,
		ContainerID:       testContainerID,
		LocationID:        testStoreID,
		ContainerQuantity: moveStockQuantity,
		ContainerVolume:   24,
	}
}

// item - запрос на перенос quantity контейнеров остатка на склад moveToStoreID.
func (s *MoveStockContainerSuite) item(quantity int) dto.MoveStockContainer {
	return dto.MoveStockContainer{
		AccountID:  s.accountID,
		StockID:    moveStockID,
		LocationID: moveToStoreID,
		Quantity:   quantity,
	}
}

// movedStock - запись, которая должна появиться на новом месте: ID ей назначит хранилище.
func (s *MoveStockContainerSuite) movedStock(quantity int) entity.Stock {
	row := s.stock()
	row.ID = 0
	row.LocationID = moveToStoreID
	row.ContainerQuantity = quantity

	return row
}

// expectStockOnAvailableStore - остаток найден, а склад назначения доступен для размещения.
func (s *MoveStockContainerSuite) expectStockOnAvailableStore(stock entity.Stock) {
	s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(stock, nil)
	s.serviceStore.EXPECT().CheckAvailability(gomock.Any(), s.accountID, moveToStoreID).Return(nil)
}

// TestExecutePartial - при частичном переносе на новое место уходит ровно переносимое количество,
// а у исходного остатка остаётся разница; суммарное количество контейнеров не меняется.
// Перенос ровно половины проверяется отдельно: остаток там совпадает с переносимым количеством.
func (s *MoveStockContainerSuite) TestExecutePartial() {
	type testCase struct {
		name      string
		quantity  int
		remainder int
	}

	tests := []testCase{
		{name: "3 of 10", quantity: 3, remainder: 7},
		{name: "half: 5 of 10", quantity: 5, remainder: 5},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			const (
				newStockID     uint64 = 200
				updatedStockID uint64 = 101 // при обновлении количества хранилище выдаёт остатку новый ID
			)

			s.expectStockOnAvailableStore(s.stock())
			s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), s.movedStock(tt.quantity)).Return(newStockID, nil)
			s.storageStock.EXPECT().UpdateQuantity(gomock.Any(), s.accountID, moveStockID, tt.remainder).Return(updatedStockID, nil)
			s.eventEmitter.EXPECT().Emit(gomock.Any(), "Move", "accountId", s.accountID, "stockId", updatedStockID, "newStockId", newStockID)
			s.refreshLocations.EXPECT().Execute(gomock.Any(), []uint64{moveToStoreID, testStoreID}).Return(nil)

			stockID, err := s.uc.Execute(s.ctx, s.item(tt.quantity))
			s.Require().NoError(err)
			s.Equal(newStockID, stockID)
		})
	}
}

// TestExecuteFull - при переносе всего остатка (явно или нулевым количеством)
// на новое место уходят все контейнеры, а исходная запись удаляется.
func (s *MoveStockContainerSuite) TestExecuteFull() {
	type testCase struct {
		name     string
		quantity int
	}

	tests := []testCase{
		{name: "10 of 10", quantity: moveStockQuantity},
		{name: "zero quantity means all", quantity: 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			const newStockID uint64 = 200

			s.expectStockOnAvailableStore(s.stock())
			s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), s.movedStock(moveStockQuantity)).Return(newStockID, nil)
			s.storageStock.EXPECT().Delete(gomock.Any(), s.accountID, moveStockID).Return(nil)
			s.eventEmitter.EXPECT().Emit(gomock.Any(), "Move", "accountId", s.accountID, "stockId", moveStockID, "newStockId", newStockID)
			s.refreshLocations.EXPECT().Execute(gomock.Any(), []uint64{moveToStoreID, testStoreID}).Return(nil)

			stockID, err := s.uc.Execute(s.ctx, s.item(tt.quantity))
			s.Require().NoError(err)
			s.Equal(newStockID, stockID)
		})
	}
}

// TestExecuteSameLocation - перенос на текущее место ничего не меняет
// и возвращает ID исходного остатка.
func (s *MoveStockContainerSuite) TestExecuteSameLocation() {
	item := s.item(3)
	item.LocationID = testStoreID

	s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(s.stock(), nil)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().NoError(err)
	s.Equal(moveStockID, stockID)
}

// TestExecuteIncorrectRequest - запрос без обязательных полей отклоняется до обращения к хранилищу
// (моки без EXPECT: любой вызов провалит тест).
func (s *MoveStockContainerSuite) TestExecuteIncorrectRequest() {
	type testCase struct {
		name   string
		modify func(item *dto.MoveStockContainer)
	}

	tests := []testCase{
		{name: "empty accountId", modify: func(item *dto.MoveStockContainer) { item.AccountID = uuid.Nil }},
		{name: "zero stockId", modify: func(item *dto.MoveStockContainer) { item.StockID = 0 }},
		{name: "zero locationId", modify: func(item *dto.MoveStockContainer) { item.LocationID = 0 }},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			item := s.item(3)
			tt.modify(&item)

			stockID, err := s.uc.Execute(s.ctx, item)
			s.Require().ErrorIs(err, errors.ErrInternalIncorrectInputData)
			s.Zero(stockID)
		})
	}
}

// TestExecuteIncorrectMove - перенос, противоречащий остатку, отклоняется без записи:
// контейнер нельзя положить в самого себя и нельзя перенести больше, чем есть.
func (s *MoveStockContainerSuite) TestExecuteIncorrectMove() {
	type testCase struct {
		name string
		item dto.MoveStockContainer
	}

	intoItself := s.item(3)
	intoItself.LocationID = testContainerID

	tests := []testCase{
		{name: "into itself", item: intoItself},
		{name: "more than stock", item: s.item(moveStockQuantity + 1)},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(s.stock(), nil)

			stockID, err := s.uc.Execute(s.ctx, tt.item)
			s.Require().ErrorIs(err, errors.ErrInternalIncorrectInputData)
			s.Zero(stockID)
		})
	}
}

// TestExecuteGroupContainerOnlyToStore - групповой контейнер размещается только на складе,
// вложить его в другой контейнер нельзя.
func (s *MoveStockContainerSuite) TestExecuteGroupContainerOnlyToStore() {
	stock := s.stock()
	stock.ContainerID = testGroupContainerID

	item := s.item(3)
	item.LocationID = testContainerID

	s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(stock, nil)
	s.serviceStore.EXPECT().CheckAvailability(gomock.Any(), s.accountID, testContainerID).Return(nil)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().ErrorIs(err, errors.ErrInternalIncorrectInputData)
	s.Zero(stockID)
}

// TestExecuteFetchError - ошибка чтения остатка доходит до вызывающего.
func (s *MoveStockContainerSuite) TestExecuteFetchError() {
	errStorage := errors.New("stock storage is unavailable")

	s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(entity.Stock{}, errStorage)

	stockID, err := s.uc.Execute(s.ctx, s.item(3))
	s.Require().ErrorIs(err, errStorage)
	s.Zero(stockID)
}

// TestExecuteStoreUnavailable - на недоступный склад остаток не переносится.
func (s *MoveStockContainerSuite) TestExecuteStoreUnavailable() {
	errUnavailable := errors.New("store is unavailable")

	s.storageStock.EXPECT().FetchOne(gomock.Any(), s.accountID, moveStockID).Return(s.stock(), nil)
	s.serviceStore.EXPECT().CheckAvailability(gomock.Any(), s.accountID, moveToStoreID).Return(errUnavailable)

	stockID, err := s.uc.Execute(s.ctx, s.item(3))
	s.Require().ErrorIs(err, errUnavailable)
	s.Zero(stockID)
}

// TestExecuteInsertError - если запись на новом месте не создана, исходный остаток не трогается.
func (s *MoveStockContainerSuite) TestExecuteInsertError() {
	errStorage := errors.New("stock storage is unavailable")

	s.expectStockOnAvailableStore(s.stock())
	s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), gomock.Any()).Return(uint64(0), errStorage)

	stockID, err := s.uc.Execute(s.ctx, s.item(3))
	s.Require().ErrorIs(err, errStorage)
	s.Zero(stockID)
}

// TestExecuteSourceUpdateError - ошибка изменения исходного остатка доходит до вызывающего,
// событие о переносе не отправляется.
func (s *MoveStockContainerSuite) TestExecuteSourceUpdateError() {
	type testCase struct {
		name     string
		quantity int
		expect   func(err error)
	}

	errStorage := errors.New("stock storage is unavailable")

	tests := []testCase{
		{
			name:     "update quantity",
			quantity: 3,
			expect: func(err error) {
				s.storageStock.EXPECT().UpdateQuantity(gomock.Any(), s.accountID, moveStockID, 7).Return(uint64(0), err)
			},
		},
		{
			name:     "delete",
			quantity: moveStockQuantity,
			expect: func(err error) {
				s.storageStock.EXPECT().Delete(gomock.Any(), s.accountID, moveStockID).Return(err)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.expectStockOnAvailableStore(s.stock())
			s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), gomock.Any()).Return(uint64(200), nil)
			tt.expect(errStorage)

			stockID, err := s.uc.Execute(s.ctx, s.item(tt.quantity))
			s.Require().ErrorIs(err, errStorage)
			s.Zero(stockID)
		})
	}
}

// TestExecuteRefreshLocationsError - ошибка пересчёта локаций доходит до вызывающего.
func (s *MoveStockContainerSuite) TestExecuteRefreshLocationsError() {
	errRefresh := errors.New("refresh locations failed")

	s.expectStockOnAvailableStore(s.stock())
	s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), gomock.Any()).Return(uint64(200), nil)
	s.storageStock.EXPECT().UpdateQuantity(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(101), nil)
	s.eventEmitter.EXPECT().Emit(gomock.Any(), "Move", gomock.Any())
	s.refreshLocations.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(errRefresh)

	stockID, err := s.uc.Execute(s.ctx, s.item(3))
	s.Require().ErrorIs(err, errRefresh)
	s.Zero(stockID)
}
