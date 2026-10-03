package errcat

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:generate gotext -srclang=en-US update -out=../../../internal/localization/dict/errcat/catalog.go -lang=en-US,ru-RU print-shop-back/localization/dict/errcat
//go:generate gotext-catalog-fix -src=../../../internal/localization/dict/errcat/catalog.go -out=../../../internal/localization/dict/errcat/catalog.go

// Здесь приведены фразы используемы для локализации.
//
//nolint:unused
func list() {
	p := message.NewPrinter(language.MustParse("en-US"))

	p.Sprintf("internal error")            // 500
	p.Sprintf("system error")              // 503
	p.Sprintf("unexpected internal error") // 500, 599
	p.Sprintf("not implemented")           // 501

	p.Sprintf("401. client is unauthorized")                  // 401
	p.Sprintf("403. access forbidden")                        // 403
	p.Sprintf("access forbidden")                             // 403
	p.Sprintf("404. resource not found")                      // 404
	p.Sprintf("record not found")                             // 404
	p.Sprintf("record version conflict")                      // 409
	p.Sprintf("request body is not valid: '%[1]s'", "Reason") // 422
	p.Sprintf("too many requests")                            // 429

	p.Sprintf("the file with the specified key '%[1]s' was not uploaded", "Key")
	p.Sprintf("request param with key '%[1]s' of type '%[2]s' contains incorrect value '%[3]s'", "Key", "Type", "Value")
	p.Sprintf("request param with key '%[1]s' is empty", "Key")
	p.Sprintf("request param with key '%[1]s' contains value greater then max '%[2]s'", "Key", "Max")
	p.Sprintf("request param with key '%[1]s' has value length greater then max '%[2]s' characters", "Key", "MaxLength")
	p.Sprintf("input data is incorrect: '%[1]s'", "Reason")
	p.Sprintf("switching from '%[1]s' to '%[2]s' is rejected", "StatusFrom", "StatusTo")
	p.Sprintf("file is invalid")
	p.Sprintf("email is invalid")
	p.Sprintf("phone is invalid")
	p.Sprintf("email already exists")
	p.Sprintf("phone already exists")
	p.Sprintf("query %[1]s not found", "ShortLink")
	p.Sprintf("box with ID=%[1]s not found", "Id")
	p.Sprintf("box article '%[1]s' already exists", "Name")
	p.Sprintf("laminate with ID=%[1]s not found", "Id")
	p.Sprintf("laminate article '%[1]s' already exists", "Name")
	p.Sprintf("paper with ID=%[1]s not found", "Id")
	p.Sprintf("paper article '%[1]s' already exists", "Name")
	p.Sprintf("form ID is required")
	p.Sprintf("form with ID=%[1]s not found", "Id")
	p.Sprintf("rewrite name '%[1]s' already exists", "Name")
	p.Sprintf("param name '%[1]s' already exists", "Name")
	p.Sprintf("form with ID=%[1]s is disabled", "Id")
	p.Sprintf("form element with ID=%[1]s not found", "Id")
	p.Sprintf("param name '%[1]s' already exists", "Name")
	p.Sprintf("item detailing '%[1]s' not allowed for form detailing '%[2]s'", "Name1", "Name2")
	p.Sprintf("rewrite name '%[1]s' already exists", "Name")
	p.Sprintf("element template ID is required")
	p.Sprintf("element template with ID=%[1]s not found", "Id")
	p.Sprintf("element template with ID=%[1]s is disabled", "Id")
	p.Sprintf("laminate type ID is required")
	p.Sprintf("laminate type with ID=%[1]s is not available", "Id")
	p.Sprintf("laminate type with ID=%[1]s not found", "Id")
	p.Sprintf("paper color ID is required")
	p.Sprintf("paper color with ID=%[1]s is not available", "Id")
	p.Sprintf("paper color with ID=%[1]s not found", "Id")
	p.Sprintf("paper facture ID is required")
	p.Sprintf("paper facture with ID=%[1]s is not available", "Id")
	p.Sprintf("paper facture with ID=%[1]s not found", "Id")
	p.Sprintf("print format ID is required")
	p.Sprintf("print format with ID=%[1]s is not available", "Id")
	p.Sprintf("print format with ID=%[1]s not found", "Id")
	p.Sprintf("token not found or expired : %[1]s, %[2]s, %[1]s", "Value1", "Value2")
	p.Sprintf("token section %[1]s is invalid", "Key")
	p.Sprintf("confirm code is incorrect")
	p.Sprintf("confirm code is required")
	p.Sprintf("all attempts to confirm the operation have been spent")
	p.Sprintf("no attempts to resend confirm code")
	p.Sprintf("resend confirm code is not supported for the current action")
	p.Sprintf("sending new messages is temporarily restricted")
	p.Sprintf("operation is empty or invalid")
	p.Sprintf("operation already confirmed")
	p.Sprintf("operation already expired")
	p.Sprintf("operation is not confirmed")
	p.Sprintf("login not exists")
	p.Sprintf("password is too weak")
	p.Sprintf("password has recovery code format")
	p.Sprintf("auth token is empty or invalid")
	p.Sprintf("auth token not found or expired")
	p.Sprintf("session limit exceeded, try again later")
	p.Sprintf("signup already in progress, try again later")
	p.Sprintf("2fa is disabled")
	p.Sprintf("disable current 2fa before setting a new one")
	p.Sprintf("recovery code is not accepted at this step")
	p.Sprintf("totp code is incorrect")
	p.Sprintf("after node with ID=%[1]s not found", "Id")
	p.Sprintf("invalid file size, min size = %[1]sb", "Value")
	p.Sprintf("invalid file size, max size = %[1]sb", "Value")
	p.Sprintf("invalid file extension: %[1]s", "Value")
	p.Sprintf("invalid file total size, max total size = %[1]sb", "Value")
	p.Sprintf("the content type '%[1]s' does not match the detected type", "Value")
	p.Sprintf("unsupported file type '%[1]s'", "Value")
	p.Sprintf("invalid image width, max size = %[1]spx", "Value")
	p.Sprintf("invalid image height, max size = %[1]spx", "Value")

	// Ошибки (400) генерируемые go-playground валидатором.
	// ID формируется из "Validator" и имени валидатора, которое указано
	// в описании поля структуры (например: min, max, required).
	// Можно использовать следующие переменные:
	//   {Name} - название поля, где произошла ошибка;
	//   {Type} - тип поля (например: int32);
	//   {Value} - текущее значение поля;
	//   {Param} - параметр валидатора, если он указан (например: max=16);

	p.Sprintf("Validator_http_url: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_required: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")

	p.Sprintf("Validator_gte: %[1]s, %[2]s, %[3]s, %[4]s", "Name", "Type", "Value", "Param")
	p.Sprintf("Validator_lte: %[1]s, %[2]s, %[3]s, %[4]s", "Name", "Type", "Value", "Param")
	p.Sprintf("Validator_max: %[1]s, %[2]s, %[3]s, %[4]s", "Name", "Type", "Value", "Param")
	p.Sprintf("Validator_min: %[1]s, %[2]s, %[3]s, %[4]s", "Name", "Type", "Value", "Param")

	p.Sprintf("Validator_tag_article: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_email: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_email_phone: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_lang: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_name: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_password: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_phone: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_realm: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_rewrite_name: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_tz: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_variable: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_2d_size: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
	p.Sprintf("Validator_tag_3d_size: %[1]s, %[2]s, %[3]s", "Name", "Type", "Value")
}
