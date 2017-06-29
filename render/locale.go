package render

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

//GetUserLocale returns the client browser's "accept-language" or first (default) locale in config file
func GetUserLocale(ctx *fasthttp.RequestCtx) string {
	langs := string(ctx.Request.Header.PeekBytes([]byte("Accept-Language")))
	if langs != "" {
		return getAgentLocale(langs) //return the browser's best accepted locale
	}
	return appConfig.Locales[0][0] //return default locale
}

func getAgentLocale(AcceptLangs string) string {
	lqs := parseAcceptLanguage(AcceptLangs)
	q := 0.0
	l := appConfig.Locales[0][0] //default locale is first locale in config file
	for _, lq := range lqs {
		for _, lc := range appConfig.Locales {
			if lq.Q > q && strings.HasPrefix(lq.Lang, lc[0]) {
				q = lq.Q
				l = lc[0]
			}
		}
	}
	return l
}

type langQ struct {
	Lang string
	Q    float64
}

func parseAcceptLanguage(AcceptLangs string) []langQ {
	var lqs []langQ
	arrLang := strings.Split(AcceptLangs, ",")
	for _, lang := range arrLang {
		arrLang := strings.Split(strings.TrimSpace(lang), ";")
		ln := len(arrLang)
		if ln == 0 {
			continue
		}
		if ln == 1 {
			lq := langQ{arrLang[0], 1}
			lqs = append(lqs, lq)
		} else {
			qp := strings.Split(arrLang[1], "=")
			if len(qp) > 1 {
				q, err := strconv.ParseFloat(qp[1], 64)
				if err != nil {
					continue
				}
				lq := langQ{arrLang[0], q}
				lqs = append(lqs, lq)
			}
		}
	}
	return lqs
}
