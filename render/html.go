package render

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet"
	"github.com/patrickmn/go-cache"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
)

var appConfig = config.Config()
var appCache = cache.New(time.Duration(appConfig.PageCacheDurationMinutes)*time.Minute, time.Minute)
var myCompress = utils.NewCompress()
var views = jet.NewHTMLSet("./views")

func init() {
	views.SetDevelopmentMode(true) //TODO: read from config and remove in production
}

//HTML renders a Jet template with content-type HTML, minifies it if appConfig.IsProduction == true, then gzips and finally adds result to cache if appConfig.PageCacheDurationMinutes > 0
func HTML(ctx *fasthttp.RequestCtx, page *Page) {

	//On initial request, create a random CSRF token in context variable (later in form) and also cookie.
	csrf := utils.RandomString(11)
	utils.SetCookie(ctx, "csrftoken", csrf, time.Now().Add(20*time.Minute), true)
	views.AddGlobal("CSRF", csrf)

	ctx.SetContentType("text/html")
	cache := contentCache{Type: "text/html"}

	vw, err := views.GetTemplate(page.TemplateFileName)
	if err != nil {
		renderError(ctx, "/render/HTML.go: Jet GetTemplate() error:"+err.Error())
		return
	}

	buf := bytes.Buffer{}
	err = vw.Execute(&buf, nil, page)
	if err != nil {
		renderError(ctx, "/render/HTML.go: Jet Execute() error:"+err.Error())
		return
	}

	var body []byte
	body = buf.Bytes()
	// if appConfig.IsProduction == false {
	// 	body = buf.Bytes()
	// } else {
	// 	body = myCompress.MinifyBytes("text/html", buf.Bytes())
	// }

	//TODO: Consider calling pre-gzip content modifying middleware functions here, e.g. ChangeHTML(body)

	gzip := strings.Contains(string(ctx.Request.Header.Peek("Accept-Encoding")), "gzip")
	gzipped := false
	if gzip && len(body) > 512 {
		zbuf := utils.Gzip(body)
		n := len(zbuf)
		if n > 0 && n < len(body) {
			gzipped = true
			ctx.Response.Header.Set("Content-Encoding", "gzip")
			ctx.SetBody(zbuf)
			if page.CacheKey != "" {
				cache.Gzip = true
				cache.Body = make([]byte, n)
				copy(cache.Body, zbuf)
			}
		}
	}

	if !gzipped {
		ctx.SetBody(body)
		if page.CacheKey != "" && len(body) > 0 {
			cache.Body = make([]byte, len(body))
			copy(cache.Body, body)
		}
	}

	//Cache the page
	if appConfig.PageCacheDurationMinutes > 0 && page.CacheKey != "" && len(cache.Body) > 0 {
		cache.Etag = utils.HashStr(cache.Body)
		ctx.Response.Header.Set("ETag", cache.Etag)
		appCache.Add(page.CacheKey, cache, time.Duration(appConfig.PageCacheDurationMinutes)*time.Minute)
	}

	utils.Log(ctx, "Render HTML", fasthttp.StatusOK, len(ctx.Response.Body()), "")
}

//RenderError renders an error
func renderError(ctx *fasthttp.RequestCtx, msg string) {
	utils.Log(ctx, "Error", fasthttp.StatusInternalServerError, 0, msg)
	if appConfig.IsProduction == false {
		ctx.SetBodyString(msg) //TODO: create user friendly error page (for HTML content type only, not JSON)
	} else {
		ctx.Error("Server Error", fasthttp.StatusInternalServerError)
	}
}

type contentCache struct {
	Type string
	Body []byte
	Etag string
	Gzip bool
}

//FromCache finds a cached HTTP content item by given key and retirns whether it was found and added to the Response Body
func FromCache(ctx *fasthttp.RequestCtx, ck string) bool {

	if ck == "" || appConfig.IsProduction == false {
		return false
	}

	cd, ok := appCache.Get(ck)
	if !ok {
		return false
	}
	ce := cd.(contentCache)

	//if etag sent from browser matches the cached page, then return 304 status (not modified)
	etag := string(ctx.Request.Header.Peek("If-None-Match"))
	if etag == ce.Etag {
		ctx.NotModified()
		if ce.Type == "text/html" {
			utils.Log(ctx, "Render Cache", fasthttp.StatusNotModified, 0, "etag="+ce.Etag)
		}

	} else {

		ctx.SetContentType(ce.Type)

		//if browser does not support gzip but cached page is gzipped, then unzip and send
		gzip := strings.Contains(string(ctx.Request.Header.Peek("Accept-Encoding")), "gzip")

		if !gzip && ce.Gzip {
			buf, err := utils.Gunzip(ce.Body)
			if err != nil {
				renderError(ctx, "Gunzip error: "+err.Error())
				return false
			}
			ctx.SetBody(buf)
			ctx.Response.Header.Set("Content-Length", strconv.Itoa(len(buf)))

		} else {
			//otherwise set content body
			ctx.SetBody(ce.Body)
			ctx.Response.Header.Set("Content-Length", strconv.Itoa(len(ce.Body)))
			if ce.Gzip {
				ctx.Response.Header.Set("Content-Encoding", "gzip")
			}
		}
		ctx.Response.Header.Set("ETag", ce.Etag)
		if ce.Type == "text/html" {
			utils.Log(ctx, "Render Cache", fasthttp.StatusOK, len(ce.Body), "")
		}
	}
	return true
}
