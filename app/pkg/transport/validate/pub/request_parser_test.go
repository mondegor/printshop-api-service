package pub_test

import (
	"print-shop-back/pkg/transport/validate/pub"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ pub.RequestParser = (*pub.Parser)(nil)
