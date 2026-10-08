package adm_test

import (
	"testing"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"print-shop-back/pkg/transport/validate/adm"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ adm.RequestParser = (*adm.Parser)(nil)

// TestParserWithListPager - копия получает новый парсер постраничной навигации с прежним
// парсером сортировки, исходный парсер не изменяется.
func TestParserWithListPager(t *testing.T) {
	t.Parallel()

	basePager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{})
	customPager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{PageSizeDefault: 10, PageSizeMax: 100})
	sorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})
	contextParser := webvalidate.NewContextParser(nil, parser.NewUser(mrlog.NopLogger()), nil, nil)

	base := adm.NewParser(nil, contextParser, webvalidate.NewListParser(basePager, sorter), nil)
	got := base.WithListPager(customPager)

	require.NotSame(t, base, got)
	assert.Same(t, customPager, got.ListPager)
	assert.Same(t, sorter, got.ListSorter)
	assert.Same(t, contextParser, got.ContextParser)
	assert.Same(t, basePager, base.ListPager)
}

// TestParserWithListSorter - копия получает новый парсер сортировки с прежним
// парсером постраничной навигации, исходный парсер не изменяется.
func TestParserWithListSorter(t *testing.T) {
	t.Parallel()

	pager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{})
	baseSorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})
	customSorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})

	base := adm.NewParser(nil, nil, webvalidate.NewListParser(pager, baseSorter), nil)
	got := base.WithListSorter(customSorter)

	require.NotSame(t, base, got)
	assert.Same(t, customSorter, got.ListSorter)
	assert.Same(t, pager, got.ListPager)
	assert.Same(t, baseSorter, base.ListSorter)
}
