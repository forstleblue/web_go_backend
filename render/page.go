package render

import (
	"log"
	"strings"

	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/message"
	"github.com/unirep/ur-local-web/app/models/notification"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/valyala/fasthttp"
)

//Page holds the model to be rendered for most handlers
type Page struct {
	TemplateFileName    string
	Path                string
	Title               string
	MetaTitle           string
	Locale              string
	User                *user.User
	HeaderStyles        string
	HeaderScripts       string
	FooterScripts       string
	IsAjax              bool
	IsInAdmin           bool
	IsLoggedIn          bool
	IsProduction        bool
	Data                interface{}
	CacheKey            string
	GoogleAPIKey        string
	UnreadMessages      []*message.Message
	UnreadNotifications int64
	Version             string
}

//Render renders a Jet template with content-type HTMLfrom cache if possible, otherwise renders and adds to the cache
func (pg *Page) Render(ctx *fasthttp.RequestCtx) {

	pg.Path = string(ctx.Path())
	pg.IsAjax = string(ctx.Request.Header.PeekBytes([]byte("X-Requested-With"))) == "XMLHttpRequest"

	if !pg.IsAjax && len(pg.Title) == 0 {
		renderError(ctx, "Page.Title not defined for path "+pg.Path)
		return
	}

	pg.Locale = GetUserLocale(ctx)
	pg.CacheKey = pg.Title + pg.Locale + string(ctx.RequestURI()) //ck: cache key includes the page title, locale and full url with querystring

	//if page has been sent from cache, simply exit
	if appConfig.PageCacheDurationMinutes > 0 && FromCache(ctx, string(pg.CacheKey)) {
		return
	}

	if len(pg.MetaTitle) == 0 {
		pg.MetaTitle = "Universal Reputations"
	}

	pg.GoogleAPIKey = appConfig.GoogleAPIKey

	//user, _ := GetUserFromSession(r)
	//pg.User = *user

	// if user.IsAdmin() {
	// 	pg.HeaderStyles += `<link href="/js/plugins/summernote/summernote.css" rel="stylesheet"><link href="/js/plugins/summernote/summernote-bs3.css" rel="stylesheet">`
	// 	pg.FooterScripts += `<script src="/js/plugins/summernote/summernote.min.js"></script>`
	// }

	pg.IsInAdmin = false
	pg.IsLoggedIn = false
	pg.IsProduction = appConfig.IsProduction

	pg.Version = config.Version()

	//check ctx for user (saved into ctx in /routes/middleware.go from session), and add to page
	val := ctx.UserValue("user")
	if val == nil {
		log.Println("/render/page.go Render() User not found in ctx.UserValue(\"user\")")
	} else {
		var u *user.User
		var ok bool
		if u, ok = val.(*user.User); ok {
			pg.User = u
			pg.IsLoggedIn = true
		}
		pg.UnreadNotifications = notification.GetNotificationCount(pg.User.UserID)

		//UnreadMessages
		unReadMessages, err := message.GetMessagesByUserIdInUnread(pg.User.UserID)
		if err != nil {
			log.Println("Error occuring in get messages: ", err)
		}
		pg.UnreadMessages = unReadMessages
		if len(pg.UnreadMessages) > 10 {
			pg.UnreadMessages = pg.UnreadMessages[:10]
		}
	}
	HTML(ctx, pg)
}

func (pg *Page) PageName() string {
	pageName := strings.Replace(pg.TemplateFileName, ".html", "", -1)
	lastIndex := strings.LastIndex(pageName, "/")
	if lastIndex != -1 {
		lastIndex++
		pageName = pageName[lastIndex:]
	}
	return pageName
}
