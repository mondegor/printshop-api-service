package msgcat

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:generate gotext -srclang=en-US update -out=../../../internal/localization/dict/msgcat/catalog.go -lang=en-US,ru-RU print-shop-back/localization/dict/msgcat
//go:generate gotext-catalog-fix -src=../../../internal/localization/dict/msgcat/catalog.go -out=../../../internal/localization/dict/msgcat/catalog.go

// Здесь приведены фразы используемы для локализации.
//
//nolint:unused
func list() {
	p := message.NewPrinter(language.MustParse("en-US"))

	p.Sprintf("Confirm the creation of the user by code")
	p.Sprintf("Confirm your identity to sign in by code")
	p.Sprintf("Confirm your identity to sign in by second factor")
	p.Sprintf("Confirm your identity to sign in by 2fa")
	p.Sprintf("Confirm your operation 'change email' by code")
	p.Sprintf("Confirm your operation 'change email' by second factor")
	p.Sprintf("Confirm your new email by code")
	p.Sprintf("Confirm your operation 'change phone' by code")
	p.Sprintf("Confirm your operation 'change password' by code")
	p.Sprintf("Confirm your operation 'change TOTP generator' by code")
	p.Sprintf("Confirm your operation 'regenerate recovery codes' by code")
	p.Sprintf("Confirm your operation 'disable 2fa' by code")
	p.Sprintf("Confirm your operation by 2fa")
	p.Sprintf("The confirmation code has been sent successfully")
}
