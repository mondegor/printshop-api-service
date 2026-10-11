package usecase_test

//go:generate mockgen -destination=mock/usr.go -package=mock print-shop-back/internal/warehousing/actiongroup/usr ContainerStorage,StockStorage,StoreService
//go:generate mockgen -destination=mock/mrevent.go -package=mock github.com/mondegor/go-core/mrevent Emitter

// Общие ID локаций тестов юзкейсов: старшие биты ID задают вид локации.
const (
	testContainerID      uint64 = 2<<60 | 15 // вид локации: контейнер
	testGroupContainerID uint64 = 1<<60 | 3  // вид локации: групповой контейнер
	testStoreID          uint64 = 7          // вид локации: склад
)
