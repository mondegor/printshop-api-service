package validate_test

import (
	"testing"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"print-shop-back/internal/provideraccounts/section/adm/validate"
	pkgvalidate "print-shop-back/pkg/provideraccounts/validate"
	"print-shop-back/pkg/transport/validate/adm"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ validate.RequestProviderAccountsParser = (*validate.Parser)(nil)

// TestParserWithListSorter - копия получает новый парсер сортировки,
// исходный парсер и собственные части модуля не изменяются.
func TestParserWithListSorter(t *testing.T) {
	t.Parallel()

	pager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{})
	baseSorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})
	customSorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})
	publicStatusParser := pkgvalidate.NewPublicStatusParser(mrlog.NopLogger())

	base := validate.NewParser(adm.NewParser(nil, nil, webvalidate.NewListParser(pager, baseSorter), nil), publicStatusParser)
	got := base.WithListSorter(customSorter)

	require.NotSame(t, base, got)
	assert.Same(t, customSorter, got.ListSorter)
	assert.Same(t, pager, got.ListPager)
	assert.Same(t, publicStatusParser, got.PublicStatusParser)
	assert.Same(t, baseSorter, base.ListSorter)
}
