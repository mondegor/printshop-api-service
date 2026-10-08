package adm

import (
	"github.com/mondegor/go-webcore/mrserver/request"
	"github.com/mondegor/go-webcore/mrserver/request/parser"
	webvalidate "github.com/mondegor/go-webcore/mrserver/request/validate"
)

type (
	// RequestParser - агрегирующий интерфейс парсеров HTTP-запроса секции AdminAPI: базовые парсеры,
	// парсеры контекста запроса, параметры списка (страница и сортировка) и фильтр по статусу элемента.
	RequestParser interface {
		webvalidate.RequestParser
		webvalidate.RequestContextParser
		webvalidate.RequestListParser
		request.ParserItemStatus
	}

	// Parser - реализация RequestParser, соединяющая парсеры go-webcore.
	Parser struct {
		*webvalidate.Parser
		*webvalidate.ContextParser
		*webvalidate.ListParser
		*parser.ItemStatus
	}
)

// NewParser - создаёт объект Parser.
func NewParser(
	baseParser *webvalidate.Parser,
	contextParser *webvalidate.ContextParser,
	listParser *webvalidate.ListParser,
	itemStatusParser *parser.ItemStatus,
) *Parser {
	return &Parser{
		Parser:        baseParser,
		ContextParser: contextParser,
		ListParser:    listParser,
		ItemStatus:    itemStatusParser,
	}
}

// WithListPager - возвращает копию парсера с заменённым парсером постраничной навигации.
func (p *Parser) WithListPager(listPager *parser.ListPager) *Parser {
	c := *p
	c.ListParser = webvalidate.NewListParser(listPager, p.ListSorter)

	return &c
}

// WithListSorter - возвращает копию парсера с заменённым парсером сортировки.
func (p *Parser) WithListSorter(listSorter *parser.ListSorter) *Parser {
	c := *p
	c.ListParser = webvalidate.NewListParser(p.ListPager, listSorter)

	return &c
}
