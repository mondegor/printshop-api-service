package prov_test

import (
	"print-shop-back/pkg/transport/validate/prov"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ prov.RequestParser = (*prov.Parser)(nil)
