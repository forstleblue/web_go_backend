package routes

import (
	
	"encoding/gob"
	"log"
	
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/buaazp/fasthttprouter"

	"github.com/kataras/go-sessions"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/db"
	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/service_category"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

const sessionName = "URSESSION"
const forbiddenURL = "/403"
const notFoundURL = "/404"
const internalServerErrorURL = "/500"

//RedirectPage is redirect session key
const RedirectPage = "REDIRECTPAGE"

//MaxFileSize is the max image file size for profile and user photo image
const MaxFileSize = 1024 * 1024 * 4

var appConfig = config.Config()
var router = fasthttprouter.New()

var SessionsConfig = sessions.Config{Cookie: sessionName,
	// see sessions_test.go on how to set encoder and decoder for cookie value(sessionid)
	Expires:                     time.Duration(100) * time.Hour,
	DisableSubdomainPersistence: false,
}
var globalSessions = sessions.New(SessionsConfig)

//GetRouter return the app level router object for use in other packages
func GetRouter() *fasthttprouter.Router {
	return router
}

func init() {
	// Register of custom structs to gob to enable session library to encode and decode
	gob.Register(&user.User{})

	// Create connection to db
	db.Open(appConfig.DB)

	// Setup email configuration
	emailing.Configure(&appConfig.Email)

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	router.RedirectTrailingSlash = true

	//Static File Routes
	router.ServeFiles("/js/*filepath", "../static/js") // i.e. /js/foo/bar.js will be served from ../static/js/foo/bar.js
	router.ServeFiles("/css/*filepath", "../static/css")
	router.ServeFiles("/images/*filepath", "../static/images")
	router.ServeFiles("/fonts/*filepath", "../static/fonts")
	router.ServeFiles("/audio/*filepath", "../static/audio")

	//Home page
	router.GET("/", func(ctx *fasthttp.RequestCtx) {
		var data = struct {
			serviceTypes []service_category.ServiceTypes
		}{}

		data.serviceTypes = service_category.GetServiceCategory()

		pg := &render.Page{Title: "Home Page", TemplateFileName: "index.html", Data: data}
		pg.Render(ctx)
	})

	//See /routes/user.go
	router.GET("/login-register", UserLoginRegister)
	router.POST("/user/login", UserLogin)
	router.GET("/user/logout", UserLogout)
	router.POST("/user/register", UserRegister)
	router.GET("/auth/:provider/callback", OauthCallback)
	router.GET("/auth/:provider", user.BeginAuthHandler)
	router.GET("/dashboard", auth(UserDashboard))
	router.GET("/account-settings", auth(AccountSettings))
	router.POST("/save-user", auth(UserSave))
	router.POST("/reset-password-set-token", ResetPasswordSetToken)
	router.GET("/reset-password-send-mail", ResetPasswordSendMail)
	router.GET("/reset-password", ResetPassword)
	router.POST("/update-password", UpdatePassword)
	router.GET("/platforms/ebay/:status", auth(user.GetUserToken))

	router.GET("/payment-methods", auth(PaymentMethods))
	router.POST("/save-cc", auth(CreditCardSave))
	router.POST("/save-payout-methods", auth(PayoutMethodSave))
	router.DELETE("/delete-cc", auth(CreditCardDelete))
	router.POST("/make-payment", auth(PayPaymentRequest))
	router.POST("/decline-payment", auth(DeclinePaymentRequest))
	router.GET("/process-paypal-payment", ProcessPaypalPayment)
	router.GET("/cancel-paypal-payment", CancelPaypalPayment)
	router.POST("/update-paypal-payment", UpdatePaypalPayment)
	router.GET("/refresh-paypal-payment", RefreshPaypalPayment)
	router.POST("/quick-payment-request", auth(QuickPaymentRequest))

	router.POST("/user/upload-photo", auth(UserPhotoUpload))
	router.DELETE("/user/delete-photo/:uuid", auth(UserPhotoDelete))

	//See /routes/profile.go

	router.GET("/profiles", auth(ProfileList))
	router.GET("/add-profile", auth(ProfileAdd))
	router.POST("/create-profile", auth(ProfileCreate))
	router.GET("/reputation-sources", auth(ReputationSources))
	router.GET("/add-ebay", auth(AddEbayProfile))

	router.GET("/edit-profile/:profileUUID", auth(ProfileEdit))
	router.POST("/update-profile", auth(ProfileUpdate))
	router.DELETE("/delete-profile", auth(ProfileDelete))

	router.POST("/profile/upload-photo", auth(ProfilePhotoUpload))
	router.DELETE("/profile/delete-photo/:uuid", auth(ProfilePhotoDelete))

	//See /routes/booking.go
	router.GET("/booking/:profileUUID", Booking)
	router.GET("/booking-response/:bookingUUID/:bookingHistoryUUID", auth(BookingResponse))
	router.GET("/send-payment-request/:bookingUUID/:bookingHistoryUUID", auth(SendPaymentRequest))
	router.POST("/booking/register", auth(BookingSave))
	router.POST("/booking/message-update", auth(BookingMessageUpdate))
	router.POST("/booking/update", auth(BookingUpdate))
	router.POST("/payment-request/:bookingID", auth(BookingPaymentRequest))
	router.POST("/pending-payment-message", auth(PendingPaymentMessage))
	router.POST("/add-tip", auth(AddTip))
	router.GET("/tags/search", TagsSearch)
	router.GET("/data/service_categories", GetServiceCategories)
	router.GET("/payment-request/:paymentRequestID", auth(GetPaymentRequest))
	router.POST("/quick-payment", QuickPaymentProcess)
	router.POST("/update-unread-notification", UpdateUserUnreadNotifications)

	//see /routes/feedback.go
	router.GET("/feedback/:bookingUUID", auth(WriteFeedback))
	router.GET("/view-feedback/:bookingUUID", auth(ViewFeedback))
	router.POST("/send-feedback", auth(CheckFeedback))
	//See /routes/find-person.go
	router.GET("/find-person", FindPerson)
	router.GET("/profile/:profileUUID", DisplayPublicProfile)

	//See /routes/rooms.go
	router.GET("/verify-rooms", CheckMessageRoom)
	router.GET("/rooms/:roomUUID", auth(DisplayRoom))
	router.POST("/send-message", auth(ManageMessage))
	router.GET("/update-message-notification", UpdateMessagesInDashboard)

	//See /routes/widgets.go
	router.GET("/reputation-widget/:widgetID", ReputationWidget)
	router.GET("/reputation-profile/:profileID", ReputationProfile)

	//There will be other AJAX routes here

	router.GET("/blog", func(ctx *fasthttp.RequestCtx) {
		pg := &render.Page{Title: "Blog", TemplateFileName: "blog.html"}
		pg.Render(ctx)
	})

	// Standard error pages. See below for common methods to call instead of using directly
	router.GET(forbiddenURL, func(ctx *fasthttp.RequestCtx) {
		pg := &render.Page{Title: "Forbidden", TemplateFileName: "errors/403.html"}
		pg.Render(ctx)
	})
	router.GET(notFoundURL, func(ctx *fasthttp.RequestCtx) {
		pg := &render.Page{Title: "Not Found", TemplateFileName: "errors/404.html"}
		pg.Render(ctx)
	})
	router.GET(internalServerErrorURL, func(ctx *fasthttp.RequestCtx) {
		pg := &render.Page{Title: "Internal Server Error", TemplateFileName: "errors/500.html"}
		pg.Render(ctx)
	})

	if len(appConfig.SSLHost) > 0 {
		/*
			// I can't seem to get this to work using another port
			certManager := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(appConfig.SSLHost), //your domain here
				Cache:      autocert.DirCache("certs"),                  //folder for storing certificates
			}

			tlsConfig := &tls.Config{
				GetCertificate: certManager.GetCertificate,
			}

			lnTLS, err := tls.Listen("tcp", ":3000", tlsConfig)
			if err != nil {
				log.Fatalf("cannot listen: %s", err)
			}

			s := &fasthttp.Server{
				Handler:      middleware,
				LogAllErrors: true,
			}

			//log.Printf("listening for https requests on %q", in)
			if err := s.Serve(lnTLS); err != nil {
				log.Fatalf("error in fasthttp server: %s", err)
			}
		*/

		// serve HTTP so we can redirect to HTTPS
		go func() {
			err_http := fasthttp.ListenAndServe(":80", middleware)
			if err_http != nil {
				log.Fatal("Web server (HTTP): ", err_http)
			}
		}()

		if err := fasthttp.Serve(autocert.NewListener(appConfig.SSLHost), middleware); err != nil {
			log.Fatalf("error in fasthttp server: %s", err)
		}
	} else {
		if err := fasthttp.ListenAndServe(appConfig.Port, middleware); err != nil {
			log.Fatalf("Error in ListenAndServe: %s", err)
		}
	}

}

// ForbiddenRedirect Common function to redirect for unauthorised requests
func ForbiddenRedirect(ctx *fasthttp.RequestCtx) {
	ctx.Redirect(forbiddenURL, fasthttp.StatusForbidden)
}

// NotFoundRedirect Common function to redirect for not found requests
func NotFoundRedirect(ctx *fasthttp.RequestCtx) {
	ctx.Redirect(notFoundURL, fasthttp.StatusNotFound)
}

// InternalServerErrorRedirect Common function to redirect for internal service error requests
func InternalServerErrorRedirect(ctx *fasthttp.RequestCtx) {
	ctx.Redirect(internalServerErrorURL, fasthttp.StatusInternalServerError)
}


