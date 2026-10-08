package xtype

import (
	"strconv"
	"strings"

	"github.com/mondegor/go-core/mrstorage"
	"github.com/mondegor/go-core/mrtype"
)

type (
	// ContainerCursor - позиция выборки контейнеров: записи, следующие за парой (Code, Marker)
	// (пустой Code - с начала), не более Limit.
	ContainerCursor struct {
		Code   string
		Marker uint16
		Limit  int
	}
)

// NewContainerCursor - создаёт объект ContainerCursor из параметров курсорной пагинации: значение
// курсора - "code|marker" последней полученной записи; значение без разделителя означает первую
// страницу, неразбираемый или выходящий за диапазон container_marker маркер - маркер 0.
func NewContainerCursor(params mrtype.CursorParams) ContainerCursor {
	params.Limit = mrstorage.PageLimit(params.Limit)

	code, marker, found := strings.Cut(params.Value, "|")
	if !found {
		return ContainerCursor{
			Limit: params.Limit,
		}
	}

	// container_marker хранится в int2, поэтому значение ограничено 15 битами.
	parsedMarker, err := strconv.ParseUint(marker, 10, 15)
	if err != nil {
		return ContainerCursor{
			Code:  code,
			Limit: params.Limit,
		}
	}

	return ContainerCursor{
		Code:   code,
		Marker: uint16(parsedMarker),
		Limit:  params.Limit,
	}
}
