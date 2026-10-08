package xtype

import (
	"strconv"
	"strings"

	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrtype"
)

type (
	// StoreCursor - позиция выборки мест хранения: записи, следующие за парой (TerritoryID, Code)
	// (нулевой TerritoryID - с начала), не более Limit.
	StoreCursor struct {
		TerritoryID uint64
		Code        string
		Limit       int
	}
)

// NewStoreCursor - создаёт объект StoreCursor из параметров курсорной пагинации: значение
// курсора - "territory_id|store_code" последней полученной записи; значение без разделителя,
// с неразбираемым, нулевым или выходящим за диапазон bigint territory_id означает первую страницу.
func NewStoreCursor(params mrtype.CursorParams) StoreCursor {
	params.Limit = mrstorage.PageLimit(params.Limit)

	territoryID, code, found := strings.Cut(params.Value, "|")
	if !found {
		return StoreCursor{
			Limit: params.Limit,
		}
	}

	// territory_id хранится в int8, поэтому значение ограничено 63 битами.
	parsedTerritoryID, err := strconv.ParseUint(territoryID, 10, 63)
	if err != nil || parsedTerritoryID == 0 {
		return StoreCursor{
			Limit: params.Limit,
		}
	}

	return StoreCursor{
		TerritoryID: parsedTerritoryID,
		Code:        code,
		Limit:       params.Limit,
	}
}
