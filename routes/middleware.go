package routes

import (
	"bytes"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/unirep/ur-local-web/app/render"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
)

func auth(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	//if no user in context redirect to login

	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		if ctx.UserValue("user") != nil {
			//user exists in session so pass control to the intended route handler
			h(ctx)
			return
		}
		isAjax := bytes.EqualFold(ctx.Request.Header.PeekBytes([]byte("X-Requested-With")), []byte("XMLHttpRequest"))
		if isAjax {
			render.JSON(ctx, nil, "", fasthttp.StatusUnauthorized)
			utils.Log(ctx, "routes/middleware.go auth()", fasthttp.StatusUnauthorized, len(ctx.Response.Body()), "No user in context, redirecting to /login-register via ajaxComplete() in client")
		} else {
			session := globalSessions.StartFasthttp(ctx)
			var redirectPage interface{} = string(ctx.Path())
			session.Set(RedirectPage, redirectPage)
			ctx.Redirect("/login-register", fasthttp.StatusSeeOther)
			utils.Log(ctx, "routes/middleware.go auth()", fasthttp.StatusUnauthorized, len(ctx.Response.Body()), "No user in context, redirecting to /login-register via server")
		}
	})

}

type middlewareParams struct {
	isAjax bool
	isHTML bool
	path   string
	host   string
}

//middleware is the first thing to run when a request is received.
//It first runs pre-request security code, then calls the intended route for the path that was requested and finally the post request handler code to clean up and log.
func middleware(ctx *fasthttp.RequestCtx) {
	//Recovery from panics:
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovery error: %s\n", err)
			log.Printf("Stack: %s\n", debug.Stack())
		}
	}()

	params := &middlewareParams{
		isAjax: bytes.EqualFold(ctx.Request.Header.PeekBytes([]byte("X-Requested-With")), []byte("XMLHttpRequest")),
		isHTML: bytes.Contains(ctx.Request.Header.PeekBytes([]byte("Accept")), []byte("text/html")),
		path:   string(ctx.Path()),
		host:   string(ctx.Host()),
	}

	//pre-request middleware code
	preHandler(ctx, params)

	if ctx.Response.StatusCode() >= 400 {
		return
	}

	//Call route handler for the path that was requested
	router.Handler(ctx)

	//post-request middleware code
	postHandler(ctx, params)

	//Must call sessions.Clear at the very end of each request lifetime. See "important Note" at bottom of https://github.com/go-gem/sessions. This is crashing!
	//sessions.Clear(ctx)

}

func preHandler(ctx *fasthttp.RequestCtx, params *middlewareParams) {
	//pre-request middleware code, you can use ctx.SetUserValue(key string, value interface{}) to share variables in the request pipeline if necessary

	// AllowedHosts is a list of fully qualified domain names that are allowed. Default is empty list, which allows any and all host names.
	isHostAllowed := false
	for _, allowedHost := range appConfig.Security.AllowedHosts {
		if strings.EqualFold(allowedHost, params.host) {
			isHostAllowed = true
		}
	}
	if isHostAllowed == false {
		utils.Log(ctx, "Middleware", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "Host not valid '"+params.host+"'")
		if params.isAjax {
			render.JSON(ctx, params.path, "Invalid Hostname", fasthttp.StatusBadRequest)
		} else {
			ctx.Error("Invalid Hostname", fasthttp.StatusBadRequest)
		}
	}

	//Cross Site Request Forgery (CSRF) check. All form posts must have a "csrftoken"" element
	csrfToken := ctx.Request.Header.Cookie("csrftoken")
	if (params.isAjax || ctx.IsPost()) && len(csrfToken) == 0 {
		utils.Log(ctx, "Middleware", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "Missing CSRF cookie - user was redirected to login")
		if params.isAjax {
			render.JSON(ctx, params.path, "Missing CSRF Token", fasthttp.StatusUnauthorized)
		} else {
			//ctx.Error("Missing CSRF Token", fasthttp.StatusUnauthorized)
			ctx.Redirect("/login-register", fasthttp.StatusUnauthorized)
		}
		return
	}

	if (params.isAjax || ctx.IsPost()) && bytes.EqualFold(ctx.FormValue("csrftoken"), csrfToken) == false {
		utils.Log(ctx, "Middleware", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "CSRF Failed from host '"+params.host+"'. Expecting csrftoken='"+string(csrfToken)+"' in request but got '"+string(ctx.PostBody())+"'")
		if params.isAjax {
			render.JSON(ctx, params.path, "Stale CSRF Token", fasthttp.StatusUnauthorized) //if cookie times out (20min) redirect to login
		} else {
			ctx.Error("Stale CSRF Token", fasthttp.StatusUnauthorized)
		}
		return
	}

	// //e.g. processing for only JS files (e.g. minification)
	// if bytes.Equal(ctx.Path()[:2], []byte("/js")) {

	// }

	//pre-processing for "text/html" content.
	if params.isHTML || params.isAjax {
		utils.Log(ctx, "Middleware", fasthttp.StatusOK, len(ctx.Response.Body()), "CSRF cookie ='"+string(csrfToken)+"' form='"+string(ctx.FormValue("csrftoken"))+"'")

		/*if params.isAjax && !ctx.IsPost() {
			utils.Log(ctx, "Middleware AJAX", fasthttp.StatusMethodNotAllowed, len(ctx.Response.Body()), "Ajax call not POST")
			if params.isAjax {
				render.JSON(ctx, path, "Invalid HTTP Method", fasthttp.StatusMethodNotAllowed)
			} else {
				ctx.Error("Invalid HTTP Method", fasthttp.StatusMethodNotAllowed)
			}
			return
		}*/

		AddSecurityHeaders(ctx)
		//TODO:
		//Redirect non WWW requests to WWW
		//Cross-Origin Resource Sharing

		getUserFromSession(ctx)
	}

}

func postHandler(ctx *fasthttp.RequestCtx, params *middlewareParams) {
	//post-request middleware code here, e.g. request timer

	//post-processing for "text/html" content.
	if params.isHTML || params.isAjax {
		log.Println("Middleware postHandler() completed route", params.path, ctx.Response.StatusCode(), len(ctx.Response.Body()), "\n\n")
	}

	//ToDo: Auto-log HTTP requests
}

//AddSecurityHeaders sets various security HTTP headers defined in config
func AddSecurityHeaders(ctx *fasthttp.RequestCtx) {

	// Determine if server is using HTTPS
	isHTTPS := bytes.EqualFold(ctx.Request.URI().Scheme(), []byte("https")) || ctx.IsTLS()

	//If we're not on HTTPS but using a reverse proxy like Nginx and that proxy server terminated HTTPS it will send us the header "X-Forwarded-Proto = https"
	if !isHTTPS && bytes.EqualFold(ctx.Request.Header.Peek("X-Forwarded-Proto"), []byte("https")) {
		isHTTPS = true
	}

	// If SSLRedirect=true, then force HTTPS. If SSLHost is set, then redirect all requests to that host. Default is empty string (same host as request).
	if appConfig.SSLRedirect && !isHTTPS {
		ctx.URI().SetSchemeBytes([]byte("https"))
		if len(appConfig.SSLHost) > 0 {
			ctx.URI().SetHost(appConfig.SSLHost)
		}
		ctx.RedirectBytes(ctx.URI().FullURI(), fasthttp.StatusMovedPermanently)
		utils.Log(ctx, "Middleware", fasthttp.StatusMovedPermanently, len(ctx.Response.Body()), "Redirecting to https://"+appConfig.SSLHost)
	}

	// STSSeconds is the max-age of the Strict-Transport-Security header. Default is 0, which would NOT include the header.
	// Only add header when we know it's an SSL connection.
	// See https://tools.ietf.org/html/rfc6797#section-7.2 for details.
	if appConfig.Security.STSSeconds != 0 && isHTTPS && appConfig.IsProduction {
		ctx.Response.Header.AddBytesK([]byte("Strict-Transport-Security"), fmt.Sprintf("max-age=%d; includeSubdomains; preload", appConfig.Security.STSSeconds))
	}

	// If CustomFrameOptions is set to `DENY` the browser will disallow iframed content. Default is empty string (omit from header and allow iframed content).
	if len(appConfig.Security.CustomFrameOptions) > 0 {
		ctx.Response.Header.AddBytesK([]byte("X-Frame-Options"), appConfig.Security.CustomFrameOptions)
	}

	// Content Type Options header.
	if appConfig.Security.ContentTypeNosniff {
		ctx.Response.Header.AddBytesKV([]byte("X-Content-Type-Options"), []byte("nosniff"))
	}

	//Cross-Site Request Forgery (XSS) Protection header.
	if appConfig.Security.BrowserXSSFilter {
		ctx.Response.Header.AddBytesKV([]byte("X-XSS-Protection"), []byte("1; mode=block"))
	}

	// HPKP header.
	if len(appConfig.Security.PublicKey) > 0 && isHTTPS {
		ctx.Response.Header.AddBytesK([]byte("Public-Key-Pins"), appConfig.Security.PublicKey)
	}

	// Content Security Policy header.
	if len(appConfig.Security.ContentSecurityPolicy) > 0 {
		ctx.Response.Header.AddBytesK([]byte("Content-Security-Policy"), appConfig.Security.ContentSecurityPolicy)
	}

}
