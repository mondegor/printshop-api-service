package validate

import (
	"github.com/mondegor/go-webcore/mrserver/request/parser"

	"print-shop-back/pkg/provideraccounts/validate"
	"print-shop-back/pkg/transport/validate/adm"
)

type (
	// RequestProviderAccountsParser - агрегирующий интерфейс парсеров HTTP-запроса модуля в секции AdminAPI:
	// парсер секции и фильтр по публичному статусу.
	RequestProviderAccountsParser interface {
		adm.RequestParser
		validate.RequestPublicStatusParser
	}

	// Parser - реализация RequestProviderAccountsParser.
	Parser struct {
		*adm.Parser
		*validate.PublicStatusParser
	}
)

// NewParser - создаёт объект Parser.
func NewParser(sectionParser *adm.Parser, publicStatusParser *validate.PublicStatusParser) *Parser {
	return &Parser{
		Parser:             sectionParser,
		PublicStatusParser: publicStatusParser,
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
