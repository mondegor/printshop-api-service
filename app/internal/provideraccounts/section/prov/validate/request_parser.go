package validate

import (
	"github.com/mondegor/go-webcore/mrserver/request"
	"github.com/mondegor/go-webcore/mrserver/request/parser"

	"print-shop-back/pkg/transport/validate/prov"
)

type (
	// RequestProviderAccountsParser - агрегирующий интерфейс парсеров HTTP-запроса модуля в секции ProviderAPI:
	// парсер секции и изображение.
	RequestProviderAccountsParser interface {
		prov.RequestParser
		request.ParserImage
	}

	// Parser - реализация RequestProviderAccountsParser.
	Parser struct {
		*prov.Parser
		*parser.Image
	}
)

// NewParser - создаёт объект Parser.
func NewParser(sectionParser *prov.Parser, imageParser *parser.Image) *Parser {
	return &Parser{
		Parser: sectionParser,
		Image:  imageParser,
	}
}

// WithImage - возвращает копию парсера с заменённым парсером изображения.
func (p *Parser) WithImage(imageParser *parser.Image) *Parser {
	c := *p
	c.Image = imageParser

	return &c
}
