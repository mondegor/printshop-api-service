package validate_test

import (
	"testing"

	"github.com/mondegor/go-core/mrlog"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"print-shop-back/internal/controls/submitform/section/adm/validate"
	"print-shop-back/pkg/transport/validate/adm"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ validate.RequestSubmitFormParser = (*validate.Parser)(nil)

// TestParserWithListPager - копия получает новый парсер постраничной навигации,
// исходный парсер и собственные части модуля не изменяются.
func TestParserWithListPager(t *testing.T) {
	t.Parallel()

	basePager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{})
	customPager := parser.NewListPager(mrlog.NopLogger(), parser.ListPagerOptions{PageSizeDefault: 10, PageSizeMax: 100})
	sorter := parser.NewListSorter(mrlog.NopLogger(), parser.ListSorterOptions{})
	fileParser := &parser.File{}

	base := validate.NewParser(adm.NewParser(nil, nil, webvalidate.NewListParser(basePager, sorter), nil), fileParser, nil)
	got := base.WithListPager(customPager)

	require.NotSame(t, base, got)
	assert.Same(t, customPager, got.ListPager)
	assert.Same(t, sorter, got.ListSorter)
	assert.Same(t, fileParser, got.File)
	assert.Same(t, basePager, base.ListPager)
}

// TestParserWithFile - копия получает новый парсер файла, исходный парсер не изменяется.
func TestParserWithFile(t *testing.T) {
	t.Parallel()

	sectionParser := adm.NewParser(nil, nil, nil, nil)
	baseFile := &parser.File{}
	customFile := &parser.File{}

	base := validate.NewParser(sectionParser, baseFile, nil)
	got := base.WithFile(customFile)

	require.NotSame(t, base, got)
	assert.Same(t, customFile, got.File)
	assert.Same(t, sectionParser, got.Parser)
	assert.Same(t, baseFile, base.File)
}
