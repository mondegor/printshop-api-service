package usr_test

import (
	"testing"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"print-shop-back/pkg/transport/validate/usr"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ usr.RequestParser = (*usr.Parser)(nil)

// TestParserWithListCursor - копия получает новый парсер курсора, исходный парсер
// не изменяется, парсер контекста запроса у копии общий с исходным.
func TestParserWithListCursor(t *testing.T) {
	t.Parallel()

	baseCursor := parser.NewListCursor(mrlog.NopLogger(), parser.ListCursorOptions{})
	customCursor := parser.NewListCursor(mrlog.NopLogger(), parser.ListCursorOptions{LimitDefault: 10, LimitMax: 100})
	contextParser := webvalidate.NewContextParser(nil, parser.NewUser(mrlog.NopLogger()), nil, nil)

	base := usr.NewParser(nil, contextParser, baseCursor)
	got := base.WithListCursor(customCursor)

	require.NotSame(t, base, got)
	assert.Same(t, customCursor, got.ListCursor)
	assert.Same(t, contextParser, got.ContextParser)
	assert.Same(t, baseCursor, base.ListCursor)
}
