package utils

import (
	"bytes"
	"time"

	"github.com/valyala/fasthttp"
)

// type Cookie struct {
// 	Name       string
// 	Value      string
// 	Path       string
// 	Domain     string
// 	Expires    time.Time
// 	RawExpires string

// 	// MaxAge=0 means no 'Max-Age' attribute specified.
// 	// MaxAge<0 means delete cookie now, equivalent to 'Max-Age: 0'
// 	// MaxAge>0 means Max-Age attribute present and given in seconds
// 	MaxAge   int
// 	Secure   bool
// 	HttpOnly bool
// 	Raw      string
// 	Unparsed []string // Raw text of unparsed attribute-value pairs
// }

//SetCookie sets a HTTP cookie
func SetCookie(ctx *fasthttp.RequestCtx, name string, value string, expires time.Time, httpOnly bool) {
	//expiration := time.Now().Add(365 * 24 * time.Hour)
	var c fasthttp.Cookie
	c.SetKey(name)
	c.SetValue(value)
	c.SetPathBytes([]byte("/"))
	c.SetExpire(expires)
	c.SetHTTPOnly(httpOnly)
	ctx.Response.Header.SetCookie(&c)
	//log.Println("/utils/http.go SetCookie:", string(c.Path())+string(c.Key())+"="+string(c.Value()))
}

//SplitQuerystring return an array containing an byte array of kev/value pairs in the query string
func SplitQuerystring(ctx *fasthttp.RequestCtx) [][][]byte {
	q := bytes.Split(ctx.QueryArgs().QueryString(), []byte("&"))
	qs := make([][][]byte, len(q))
	for i, l := range q {
		qs[i] = bytes.Split(l, []byte("="))
		//log.Println(string(qs[i][0]), string(qs[i][1]))
	}
	return qs
}
