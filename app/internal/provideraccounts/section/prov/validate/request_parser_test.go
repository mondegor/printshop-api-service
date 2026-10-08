package validate_test

import (
	"testing"

	"github.com/mondegor/go-webcore/mrserver/request/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"print-shop-back/internal/provideraccounts/section/prov/validate"
	"print-shop-back/pkg/transport/validate/prov"
)

// Проверка на этапе компиляции: при соединении парсеров методы не конфликтуют.
var _ validate.RequestProviderAccountsParser = (*validate.Parser)(nil)

// TestParserWithImage - копия получает новый парсер изображения, исходный парсер не изменяется.
func TestParserWithImage(t *testing.T) {
	t.Parallel()

	sectionParser := prov.NewParser(nil, nil)
	baseImage := &parser.Image{}
	customImage := &parser.Image{}

	base := validate.NewParser(sectionParser, baseImage)
	got := base.WithImage(customImage)

	require.NotSame(t, base, got)
	assert.Same(t, customImage, got.Image)
	assert.Same(t, sectionParser, got.Parser)
	assert.Same(t, baseImage, base.Image)
}
