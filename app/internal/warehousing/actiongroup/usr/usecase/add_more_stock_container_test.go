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
	"print-shop-back/internal/warehousing/xtype"
)

//go:generate mockgen -source=add_more_stock_container.go -destination=mock/add_more_stock_container.go -package=mock

const testStockID uint64 = 501

type AddMoreStockContainerSuite struct {
	suite.Suite

	ctrl             *gomock.Controller
	ctx              context.Context
	storageContainer *mock.MockContainerStorage
	storageStock     *mock.MockStockStorage
	serviceStore     *mock.MockStoreService
	refreshStores    *mock.MockrefreshStoresUseCase
	eventEmitter     *mock.MockEmitter
	uc               *usecase.AddMoreStockContainer
}

func TestAddMoreStockContainerSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(AddMoreStockContainerSuite))
}

func (s *AddMoreStockContainerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.ctx = context.Background()
	s.storageContainer = mock.NewMockContainerStorage(s.ctrl)
	s.storageStock = mock.NewMockStockStorage(s.ctrl)
	s.serviceStore = mock.NewMockStoreService(s.ctrl)
	s.refreshStores = mock.NewMockrefreshStoresUseCase(s.ctrl)
	s.eventEmitter = mock.NewMockEmitter(s.ctrl)
	s.uc = usecase.NewAddMoreStockContainer(
		s.storageContainer,
		s.storageStock,
		s.serviceStore,
		s.refreshStores,
		s.eventEmitter,
	)
}

// okItem - корректный запрос на добавление контейнеров на склад.
func okItem() dto.AddMoreStockContainer {
	return dto.AddMoreStockContainer{
		AccountID:   uuid.New(),
		ContainerID: testContainerID,
		LocationID:  testStoreID,
		Quantity:    3,
	}
}

// expectContainerOnAvailableStore - контейнер найден, а склад доступен для размещения.
func (s *AddMoreStockContainerSuite) expectContainerOnAvailableStore(item dto.AddMoreStockContainer) {
	s.storageContainer.EXPECT().
		FetchOne(gomock.Any(), item.AccountID, item.ContainerID).
		Return(entity.Container{ID: item.ContainerID, Volume: xtype.Volume{Length: 2, Width: 3, Height: 4}}, nil)
	s.serviceStore.EXPECT().
		CheckAvailability(gomock.Any(), item.AccountID, item.LocationID).
		Return(nil)
}

// TestExecute - на склад записывается остаток с объёмом одного контейнера,
// рассчитанным по его габаритам. Событие отправляется только после пересчёта склада,
// чтобы подписчики видели уже обновлённое состояние.
func (s *AddMoreStockContainerSuite) TestExecute() {
	item := okItem()

	s.expectContainerOnAvailableStore(item)
	s.storageStock.EXPECT().
		InsertOrUpdate(gomock.Any(), entity.Stock{
			AccountID:         item.AccountID,
			ContainerID:       item.ContainerID,
			LocationID:        item.LocationID,
			ContainerQuantity: item.Quantity,
			ContainerVolume:   24,
		}).
		Return(testStockID, nil)
	gomock.InOrder(
		s.refreshStores.EXPECT().Execute(gomock.Any(), []uint64{item.LocationID}).Return(nil),
		s.eventEmitter.EXPECT().Emit(gomock.Any(), "AddMore", "accountId", item.AccountID, "stockId", testStockID),
	)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().NoError(err)
	s.Equal(testStockID, stockID)
}

// TestExecuteIncorrectInput - некорректный запрос отклоняется до обращения к хранилищам
// (моки без EXPECT: любой вызов провалит тест). На склад можно добавить только контейнер,
// поэтому ID другого вида локации - тоже ошибка входных данных.
func (s *AddMoreStockContainerSuite) TestExecuteIncorrectInput() {
	type testCase struct {
		name   string
		modify func(item *dto.AddMoreStockContainer)
	}

	tests := []testCase{
		{name: "empty accountId", modify: func(item *dto.AddMoreStockContainer) { item.AccountID = uuid.Nil }},
		{name: "zero containerId", modify: func(item *dto.AddMoreStockContainer) { item.ContainerID = 0 }},
		{name: "zero locationId", modify: func(item *dto.AddMoreStockContainer) { item.LocationID = 0 }},
		{name: "zero quantity", modify: func(item *dto.AddMoreStockContainer) { item.Quantity = 0 }},
		{name: "containerId is not container", modify: func(item *dto.AddMoreStockContainer) { item.ContainerID = testStoreID }},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			item := okItem()
			tt.modify(&item)

			stockID, err := s.uc.Execute(s.ctx, item)
			s.Require().ErrorIs(err, errors.ErrInternalIncorrectInputData)
			s.Zero(stockID)
		})
	}
}

// TestExecuteContainerFetchError - ошибка получения контейнера доходит до вызывающего,
// а склад и остатки при этом не трогаются.
func (s *AddMoreStockContainerSuite) TestExecuteContainerFetchError() {
	item := okItem()
	errStorage := errors.New("container storage is unavailable")

	s.storageContainer.EXPECT().FetchOne(gomock.Any(), item.AccountID, item.ContainerID).Return(entity.Container{}, errStorage)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().ErrorIs(err, errStorage)
	s.Zero(stockID)
}

// TestExecuteStoreUnavailable - на недоступный склад контейнеры не размещаются:
// остаток не записывается.
func (s *AddMoreStockContainerSuite) TestExecuteStoreUnavailable() {
	item := okItem()
	errUnavailable := errors.New("store is unavailable")

	s.storageContainer.EXPECT().FetchOne(gomock.Any(), item.AccountID, item.ContainerID).Return(entity.Container{}, nil)
	s.serviceStore.EXPECT().CheckAvailability(gomock.Any(), item.AccountID, item.LocationID).Return(errUnavailable)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().ErrorIs(err, errUnavailable)
	s.Zero(stockID)
}

// TestExecuteStockStorageError - если остаток не записан, склад не пересчитывается
// и событие о добавлении не отправляется.
func (s *AddMoreStockContainerSuite) TestExecuteStockStorageError() {
	item := okItem()
	errStorage := errors.New("stock storage is unavailable")

	s.expectContainerOnAvailableStore(item)
	s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), gomock.Any()).Return(uint64(0), errStorage)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().ErrorIs(err, errStorage)
	s.Zero(stockID)
}

// TestExecuteRefreshStoresError - ошибка пересчёта склада доходит до вызывающего,
// и событие о добавлении не отправляется.
func (s *AddMoreStockContainerSuite) TestExecuteRefreshStoresError() {
	item := okItem()
	errRefresh := errors.New("refresh stores failed")

	s.expectContainerOnAvailableStore(item)
	s.storageStock.EXPECT().InsertOrUpdate(gomock.Any(), gomock.Any()).Return(testStockID, nil)
	s.refreshStores.EXPECT().Execute(gomock.Any(), []uint64{item.LocationID}).Return(errRefresh)

	stockID, err := s.uc.Execute(s.ctx, item)
	s.Require().ErrorIs(err, errRefresh)
	s.Zero(stockID)
}
