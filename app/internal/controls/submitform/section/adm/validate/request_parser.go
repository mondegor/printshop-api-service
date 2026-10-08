package validate

import (
	"github.com/mondegor/go-webcore/mrserver/request"
	"github.com/mondegor/go-webcore/mrserver/request/parser"

	"print-shop-back/pkg/controls/validate"
	"print-shop-back/pkg/transport/validate/adm"
)

type (
	// RequestSubmitFormParser - агрегирующий интерфейс парсеров HTTP-запроса модуля в секции AdminAPI:
	// парсер секции, файл и фильтр детализации элементов.
	RequestSubmitFormParser interface {
		adm.RequestParser
		request.ParserFile
		validate.RequestDetailingParser
	}

	// Parser - реализация RequestSubmitFormParser.
	Parser struct {
		*adm.Parser
		*parser.File
		*validate.DetailingParser
	}
)

// NewParser - создаёт объект Parser.
func NewParser(
	sectionParser *adm.Parser,
	fileParser *parser.File,
	detailingParser *validate.DetailingParser,
) *Parser {
	return &Parser{
		Parser:          sectionParser,
		File:            fileParser,
		DetailingParser: detailingParser,
	}
}

// WithListPager - возвращает копию парсера с заменённым парсером постраничной навигации.
func (p *Parser) WithListPager(listPager *parser.ListPager) *Parser {
	c := *p
	c.Parser = p.Parser.WithListPager(listPager)

	return &c
}

// WithListSorter - возвращает копию парсера с заменённым парсером сортировки.
func (p *Parser) WithListSorter(listSorter *parser.ListSorter) *Parser {
	c := *p
	c.Parser = p.Parser.WithListSorter(listSorter)

	return &c
}

// WithFile - возвращает копию парсера с заменённым парсером файла.
func (p *Parser) WithFile(fileParser *parser.File) *Parser {
	c := *p
	c.File = fileParser

	return &c
}
