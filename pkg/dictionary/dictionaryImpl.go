package dictionary

import (
	"context"
	"strings"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func (d *dictionary) get(lang, key string) (s string) {
	tag := language.Indonesian

	if strings.EqualFold(lang, EN) {
		tag = language.English
	}

	p := message.NewPrinter(tag)
	return p.Sprintf(key)
}

func (d *dictionary) getMessage(lang, key string) string {
	if strings.EqualFold(lang, EN) {
		return d.printerEn.Sprintf(key)
	}

	return d.printerID.Sprintf(key)
}

func (d *dictionary) getPrinter(lang string) (printer *message.Printer) {
	if strings.EqualFold(lang, EN) {
		return d.printerEn
	}

	return d.printerID
}

func (d *dictionary) getLanguage(ctx context.Context) string {
	if ctx.Value(pdkConst.CtxAcceptLanguageKey) != nil {
		return ctx.Value(pdkConst.CtxAcceptLanguageKey).(string)
	}
	return ""
}

func (d *dictionary) SetDictionaryMessage(
	ctx context.Context,
	dictCode string,
	args ...interface{}) string {
	lang := IN
	if d.getLanguage(ctx) != "" {
		lang = d.getLanguage(ctx)
	}

	if dictCode == "" {
		return ""
	}

	printer := d.getPrinter(lang)
	return printer.Sprintf(dictCode, args...)
}
