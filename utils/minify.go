package utils

import (
	"regexp"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

//Compress provides HTTP minification
type Compress struct {
	minifier *minify.M
}

//NewCompress initialises a new Compress minification struct
func NewCompress() *Compress {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	return &Compress{
		minifier: m,
	}
}

//MinifyBytes minifies given bytes
func (c *Compress) MinifyBytes(ct string, src []byte) []byte {
	dst, err := c.minifier.Bytes(ct, src)
	if err == nil {
		return dst
	}
	return src
}
