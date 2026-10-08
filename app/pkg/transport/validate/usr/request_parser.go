package usr

import (
	"github.com/mondegor/go-webcore/mrserver/request"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
)

type (
	// RequestParser - агрегирующий интерфейс парсеров HTTP-запроса секции UserAPI:
	// базовые парсеры, парсеры контекста запроса и курсор списков.
	RequestParser interface {
		webvalidate.RequestParser
		webvalidate.RequestContextParser
		request.ParserListCursor
	}

	// Parser - реализация RequestParser, соединяющая парсеры go-webcore.
	Parser struct {
		*webvalidate.Parser
		*webvalidate.ContextParser
		*parser.ListCursor
	}
)

// NewParser - создаёт объект Parser.
func NewParser(
	baseParser *webvalidate.Parser,
	contextParser *webvalidate.ContextParser,
	listCursorParser *parser.ListCursor,
) *Parser {
	return &Parser{
		Parser:        baseParser,
		ContextParser: contextParser,
		ListCursor:    listCursorParser,
	}
}

// WithListCursor - возвращает копию парсера с заменённым парсером курсорной пагинации.
func (p *Parser) WithListCursor(listCursor *parser.ListCursor) *Parser {
	c := *p
	c.ListCursor = listCursor

	return &c
}
