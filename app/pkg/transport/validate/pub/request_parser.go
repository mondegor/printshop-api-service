package pub

import (
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
)

type (
	// RequestParser - агрегирующий интерфейс парсеров HTTP-запроса секции PublicAPI:
	// базовые парсеры и парсеры контекста запроса.
	RequestParser interface {
		webvalidate.RequestParser
		webvalidate.RequestContextParser
	}

	// Parser - реализация RequestParser, соединяющая парсеры go-webcore.
	Parser struct {
		*webvalidate.Parser
		*webvalidate.ContextParser
	}
)

// NewParser - создаёт объект Parser.
func NewParser(baseParser *webvalidate.Parser, contextParser *webvalidate.ContextParser) *Parser {
	return &Parser{
		Parser:        baseParser,
		ContextParser: contextParser,
	}
}
