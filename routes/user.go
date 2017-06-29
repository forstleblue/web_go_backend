package routes

import (
	"encoding/json"
	"image"
	"image/jpeg"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"bytes"

	"fmt"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/feedback"
	"github.com/unirep/ur-local-web/app/models/message"
	"github.com/unirep/ur-local-web/app/models/notification"
	"github.com/unirep/ur-local-web/app/models/payments"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
)

const (
	// MaxAgeShort Session MaxAge for people who are logging in temporarily
	MaxAgeShort = 1800 // half hour
	// MaxAgePersistant Session MaxAge for when people's logins are saved
	MaxAgePersistant = 86400 * 365 // remember for a year

	userFormID       = "user-form"
	ccSaveFormID     = "cc-save-form"
	payoutSaveFormID = "payout-details-form"
)

//UserLogin is an AJAX call that logs a user in and sets a session cookie
func UserLogin(ctx *fasthttp.RequestCtx) {

	email := ctx.FormValue("email")
	password := ctx.FormValue("password")

	redirectData := ctx.FormValue("redirect")

	var redirectInfo map[string]string
	err := json.Unmarshal(redirectData, &redirectInfo)

	// Setup form error object in case needed
	var loginFormError = &JSONFormError{}
	loginFormError.Form = "login-form"

	u, err := user.GetUserByEmail(string(email))
	if err != nil {
		log.Printf("Error finding user by email: %s\n", err.Error())
		loginFormError.Error = "Email and password combination not found."
		render.JSON(ctx, loginFormError, "Email and password combination not found", fasthttp.StatusNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		loginFormError.Error = "Email and password combination not found."
		render.JSON(ctx, loginFormError, "Email and password combination not found", fasthttp.StatusNotFound)
		return
	}

	//Add user object to session
	// Get session

	session := globalSessions.StartFasthttp(ctx)
	session.Set("user", u)

	var redirectURL string
	redirectURL = "/dashboard"

	if redirectInfo["pageName"] != "" {
		redirectURL = "/" + redirectInfo["pageName"] + "/" + redirectInfo["profileId"]
		session.Set("redirect", redirectInfo)
	} else if session.Get(RedirectPage) != nil {
		redirectPage := session.Get(RedirectPage)
		redirectURL = redirectPage.(string)
	}

	utils.Log(ctx, "Login", 0, 0, "Logged in with:"+ctx.PostArgs().String())
	//ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
	render.JSON(ctx, redirectURL, "User successfully logged in, redirecting to dashboard", fasthttp.StatusOK)
}

// ResetPasswordSetToken shows  reset-password page
func ResetPasswordSetToken(ctx *fasthttp.RequestCtx) {
	emailText := ctx.FormValue("email")
	recaptcha := string(ctx.FormValue("recaptcha"))

	exists, _ := user.EmailExists(emailText)

	var formError = &JSONFormError{}
	formError.Form = "reset-password-form"

	if exists == false {
		formError.Error = "Email Does not exists:"
		render.JSON(ctx, formError, "Email Does not exists:", fasthttp.StatusInternalServerError)
		return
	}

	if len(recaptcha) == 0 {
		formError.Error = "Please tick the 'I'm not a robot' checkbox"
		render.JSON(ctx, formError, "Please tick the 'I'm not a robot' checkbox", fasthttp.StatusInternalServerError)
		return
	}

	uuid := gocql.TimeUUID()

	userInfo, err := user.GetUserByEmail(string(emailText))
	if err != nil {
		log.Println("Fail to get user in routes/user.go ResetPassword:", err)
	}
	userInfo.PasswordResetToken = uuid

	err = userInfo.Update()
	if err != nil {
		log.Println("Error in routes/user.go ResetPasswordSetToken()", err)
	}

	u := uuid.String()
	sU := u[:2] + u[5:7] + u[4:5] + u[2:4] + u[7:14] + "4" + u[15:]

	swappedUUID, err := gocql.ParseUUID(sU)
	if err != nil {
		log.Println("Error routes/user.go ResetPasswordSetToken() failed to get UUID from String:", err)
	} else {
		log.Println("Swapped UUID String:", swappedUUID.String())
	}

	defer emailing.SendNewPasswordResetEmail(userInfo, swappedUUID)
	render.JSON(ctx, "/login-register", "An Email has been sent!", fasthttp.StatusOK)

}

// ResetPassword function reset password
func ResetPassword(ctx *fasthttp.RequestCtx) {

	token := string(ctx.QueryArgs().Peek("token"))
	uuid, _ := gocql.ParseUUID(token)
	var appConfig = config.Config()
	var data = struct {
		Token     string
		AppConfig string
	}{}

	data.Token = uuid.String()
	data.AppConfig = appConfig.Recaptcha.SiteKey
	timeStamp := uuid.Time()
	log.Println("timestamp:", timeStamp)
	pg := &render.Page{Title: "Reset Password", TemplateFileName: "authenticated/reset-password.html", Data: data}
	pg.Render(ctx)
}

// UpdatePassword update password with new password
func UpdatePassword(ctx *fasthttp.RequestCtx) {

	email := ctx.FormValue("email")
	password := ctx.FormValue("password")
	confirmpassword := ctx.FormValue("confirmpassword")
	sU := string(ctx.FormValue("token"))
	recaptcha := string(ctx.FormValue("recaptcha"))

	u := sU[:2] + sU[5:7] + sU[4:5] + sU[2:4] + sU[7:14] + "1" + sU[15:]

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = "reset-password-form"

	if !bytes.EqualFold(confirmpassword, password) {
		formError.Error = "'Confirm Password' must match 'Password'"
		formError.FieldName = "confirmpassword"
		render.JSON(ctx, formError, "'Confirm Password' must match 'Password'", fasthttp.StatusUnprocessableEntity)
		return
	}

	if !isValidPassword(password) {
		formError.Error = "Password must be at least 6 characters long and contain at least one lowercase character, one uppercase character and one number."
		formError.FieldName = "password"
		render.JSON(ctx, formError, "Password must be at least 6 characters long and contain at least one lowercase character, one uppercase character and one number.", fasthttp.StatusUnprocessableEntity)
		return
	}
	if len(recaptcha) == 0 {
		formError.Error = "Please tick the 'I'm not a robot' checkbox"
		render.JSON(ctx, formError, "Please tick the 'I'm not a robot' checkbox", fasthttp.StatusInternalServerError)
		return
	}

	exists, err := user.EmailUUIDExists(email, u)

	if err != nil {
		log.Printf("Error checking for existing email and UUID: %s\n", err.Error())
		formError.Error = "Could not find  email and UUID"
		render.JSON(ctx, formError, "Could not find  email and UUID", fasthttp.StatusInternalServerError)
		return
	}
	if exists == false {
		formError.Error = "User doesn't own this email."
		formError.FieldName = "email"

		render.JSON(ctx, formError, "User requested password-reset  doesn't owns that email. Please retype email.", fasthttp.StatusNotFound)
		return
	}

	var hashedPassword []byte
	hashedPassword, err = bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		log.Printf("Error creating hash from password: %s\n", err.Error())
		formError.Error = "Cannot create hash from password"
		render.JSON(ctx, formError, "Cannot create hash from password", fasthttp.StatusInternalServerError)
		return
	}

	uuid, err := gocql.ParseUUID(u)
	timeStamp := uuid.Time()
	curTime := time.Now().UTC()
	elapsedHour := curTime.Sub(timeStamp).Hours()

	usr, _ := user.GetUserByEmail(string(email))

	if elapsedHour > 1.0 {
		usr.PasswordResetToken, _ = gocql.ParseUUID("00000000-0000-0000-0000-000000000000")
		usr.Update()
		log.Println("Password token has expired. Please  reset your password again within one hour")
		formError.Error = "Password token has expired. Please  reset your password again within one hour"
		formError.Redirect = "/reset-password-send-mail"
		render.JSON(ctx, formError, "Password token has expired. Please  reset your password again within one hour", fasthttp.StatusRequestTimeout)
		return
	}

	usr.Password = string(hashedPassword)
	usr.PasswordResetToken, _ = gocql.ParseUUID("00000000-0000-0000-0000-000000000000")
	usr.Update()

	// session := globalSessions.StartFasthttp(ctx)
	// session.Set("user", usr)

	render.JSON(ctx, "/login-register", "Password was updated, Please check your new password. Redirecting to login-register", fasthttp.StatusOK)
}

// ResetPasswordSendMail send mail to user
func ResetPasswordSendMail(ctx *fasthttp.RequestCtx) {
	var appConfig = config.Config()

	var data = struct {
		AppConfig string
	}{}
	data.AppConfig = appConfig.Recaptcha.SiteKey
	pg := &render.Page{Title: "Reset Password", TemplateFileName: "authenticated/reset-password-send-mail.html", Data: data}
	pg.Render(ctx)
}

//UserLoginRegister shows the resiter/login page
func UserLoginRegister(ctx *fasthttp.RequestCtx) {
	//If user is logged in, redirect to dashboard
	if ctx.UserValue("user") != nil {
		ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
	}
	pg := &render.Page{Title: "Login / Register", TemplateFileName: "register-login.html"}
	pg.Render(ctx)
}

//UserLogout removes the session cookie
func UserLogout(ctx *fasthttp.RequestCtx) {

	session := globalSessions.StartFasthttp(ctx)
	session.Set("user", nil)

	utils.SetCookie(ctx, "URSESSION", "", time.Now().Add(-365*24*time.Hour), true)
	ctx.Redirect("/login-register", fasthttp.StatusSeeOther)
}

func UpdateUserUnreadNotifications(ctx *fasthttp.RequestCtx) {
	userID, _ := strconv.ParseInt(string(ctx.FormValue("userId")), 10, 64)
	err := notification.UpdateUnreadNotifications(userID)
	if err != nil {
		return
	}

	render.JSON(ctx, "success", "update user unread notifications", fasthttp.StatusOK)
}

//UserDashboard shows a dashboard for the user
func UserDashboard(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)

	type FeedbackInfo struct {
		FeedbackDetail *feedback.Feedback
		SenderName     string
		SenderImageURL string
		BookingUUID    string
	}
	var data = struct {
		Notifications           []notification.Notification
		RoomPageNumbers         []int
		UserID                  int64
		PaypalRedirectURL       string
		UnreadNotificationCount int64
		FeedbackList            []FeedbackInfo
		//Data for display user's reputation widget
		User                *user.User
		Score               int16
		Count               int16
		FeedbackDescription string
		FullURL             string
	}{}

	notifications, _ := notification.GetNotifications(currUser.UserID)

	data.Notifications = notifications
	data.PaypalRedirectURL = fmt.Sprintf("https://%s?cmd=_ap-payment&paykey=", appConfig.PayPalAPI.RedirectURL)
	data.UnreadNotificationCount = notification.GetNotificationCount(currUser.UserID)
	messageCount := message.GetCountOfAllMessages(currUser.UserID)
	d := float64(messageCount) / float64(3)
	pageNumbers := make([]int, int(math.Ceil(d)))
	for i := range pageNumbers {
		pageNumbers[i] = 1 + i
	}
	data.RoomPageNumbers = pageNumbers

	data.UserID = currUser.UserID

	feedbackList, _ := feedback.GetAllFeedback(currUser.UserID)
	for _, feedbackItem := range feedbackList {
		senderProfile, _ := user.GetProfile(feedbackItem.SenderProfileID)
		sender, _ := user.GetUser(senderProfile.User.UserID)
		booking, _ := user.GetBookingWithBookingId(feedbackItem.BookingID)
		feedbackInfo := FeedbackInfo{FeedbackDetail: feedbackItem, SenderName: sender.FullName(), SenderImageURL: senderProfile.DisplayPhoto(), BookingUUID: booking.BookingUUID.String()}
		data.FeedbackList = append(data.FeedbackList, feedbackInfo)
	}

	//Data for show reputation widget of user
	data.Score, data.Count, _ = feedback.GetUniversalReputationScoreByUserID(currUser.UserID)
	data.FullURL = config.FullBaseURL()
	data.FeedbackDescription = ""
	data.User, _ = user.GetUser(currUser.UserID)

	if data.Score < 25 {
		data.FeedbackDescription = "Unacceptable"
	} else if data.Score < 40 {
		data.FeedbackDescription = "Needs Improvement"
	} else if data.Score < 55 {
		data.FeedbackDescription = "Acceptable"
	} else if data.Score < 70 {
		data.FeedbackDescription = "Met Expectations"
	} else if data.Score < 85 {
		data.FeedbackDescription = "Exceeds Expectations"
	} else if data.Score < 100 {
		data.FeedbackDescription = "Excellent"
	} else {
		// 100%
		data.FeedbackDescription = "Exceptional"
	}

	pg := &render.Page{Title: "Dashboard", TemplateFileName: "authenticated/dashboard.html", Data: data}

	pg.Render(ctx)
}

//UserRegister registers a new user to UR-Local and emails them
func UserRegister(ctx *fasthttp.RequestCtx) {
	firstname := ctx.FormValue("firstname")
	lastname := ctx.FormValue("lastname")
	email := ctx.FormValue("email")
	password := ctx.FormValue("password")
	confirmpassword := ctx.FormValue("confirmpassword")

	redirectData := ctx.FormValue("redirect")

	var redirectInfo map[string]string
	var err error //note: re-use error variable to minimize allocations

	err = json.Unmarshal(redirectData, &redirectInfo)

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = "register-form"

	if len(firstname) == 0 || len(lastname) == 0 || len(email) == 0 || len(password) == 0 {
		formError.Error = "'First Name', 'Last Name', 'Email' and 'Password' are all required fields"
		render.JSON(ctx, formError, "'First Name', 'Last Name', 'Email' and 'Password' must all be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	if !bytes.EqualFold(confirmpassword, password) {
		formError.Error = "'Confirm Password' must match 'Password'"
		formError.FieldName = "confirmpassword"
		render.JSON(ctx, formError, "'Confirm Password' must match 'Password'", fasthttp.StatusUnprocessableEntity)
		return
	}

	if !isValidPassword(password) {
		formError.Error = "Password must be at least 6 characters long and contain at least one lowercase character, one uppercase character and one number."
		formError.FieldName = "password"
		render.JSON(ctx, formError, "Password must be at least 6 characters long and contain at least one lowercase character, one uppercase character and one number.", fasthttp.StatusUnprocessableEntity)
		return
	}

	exists, err := user.EmailExists(email)
	if err != nil {
		log.Printf("Error checking for existing email: %s\n", err.Error())
		formError.Error = "Could not check for duplicate emails"
		render.JSON(ctx, formError, "Could not check for duplicate emails", fasthttp.StatusInternalServerError)
		return
	}
	if exists {
		formError.Error = "Account with that email already exists."
		formError.FieldName = "email"
		render.JSON(ctx, formError, "Email already exists", fasthttp.StatusConflict)
		return
	}

	var hashedPassword []byte
	hashedPassword, err = bcrypt.GenerateFromPassword(password, 10)
	if err != nil {
		log.Printf("Error creating hash from password: %s\n", err.Error())
		formError.Error = "Cannot create hash from password"
		render.JSON(ctx, formError, "Cannot create hash from password", fasthttp.StatusInternalServerError)
		return
	}

	//create user for DB
	usr := &user.User{
		FirstName: string(firstname),
		LastName:  string(lastname),
		Email:     string(email),
		Password:  string(hashedPassword),
		Roles:     []string{user.RoleUser},
	}

	//Save to the DB
	var id int64
	if id, err = usr.InitialRegister(); err != nil {
		log.Printf("Unable to insert user into database: %s\n", err.Error())
		formError.Error = "Unable to insert user into database"
		render.JSON(ctx, formError, "Unable to insert user into database.", fasthttp.StatusInternalServerError)
		return
	}
	usr.UserID = id

	// registration email
	defer emailing.SendRegistrationEmail(usr)

	// Add initial 'Customer' profile
	createDefaultProfile(usr)

	//Add user object to session
	// Get session from store.

	session := globalSessions.StartFasthttp(ctx)
	session.Set("user", usr)

	var redirectURL string
	redirectURL = "/dashboard"

	if redirectInfo["pageName"] != "" {
		redirectURL = "/" + redirectInfo["pageName"] + "/" + redirectInfo["profileId"]
		session.Set("redirect", redirectInfo)
	}
	//Set LastLogin field for user to time.Now()
	/*user.LastLogin = time.Now()
	_, err = rdb.Table("users").
		Get(user.UserID).
		Update(user).RunWrite(rdbSession)
	if err != nil {
		result["message"] = "Cannot update last login date"
		ren.JSON(w, http.StatusInternalServerError, result)
	}*/

	utils.Log(ctx, "Register", 0, 0, "Registered:"+ctx.PostArgs().String())
	render.JSON(ctx, redirectURL, "New user created, redirecting to dashboard", fasthttp.StatusOK)
}

//OauthCallback is called from the oauth provider with details of the user
func OauthCallback(ctx *fasthttp.RequestCtx) {
	u, err := user.CompleteUserAuth(ctx)
	if err != nil {
		utils.Log(ctx, "/models/user/oauth.go CompleteUserAuth()", fasthttp.StatusInternalServerError, 0, err.Error())
		return
	}
	var usr *user.User

	if u.Provider == "facebook" {

		usr = user.GetUserByEmailOrFacebookID(u.Email, u.UserID)

		if usr == nil {
			log.Printf("Facebook user NOT found in DB by email %s or UserID %s\n", u.Email, u.UserID)

			//this is a new user so save to DB
			photoURL := "http://graph.facebook.com/" + u.UserID + "/picture?width=270&height=270"
			usr = &user.User{
				FirstName:  u.FirstName,
				LastName:   u.LastName,
				Email:      u.Email,
				FacebookID: u.UserID,
				PhotoURL:   photoURL,
				Roles:      []string{"user"},
			}

			//Insert user into the DB
			var id int64
			if id, err = usr.InitialRegister(); err != nil {
				utils.Log(ctx, "Insert Facebook User", fasthttp.StatusInternalServerError, 0, "Unable to insert user into database:"+err.Error())
				return
			}
			usr.UserID = id
			log.Printf("Facebook new user inserted into DB as UserID %d\n", usr.UserID)

			//Reading Facebook image byte slice
			dst := []byte{}
			_, body, _ := fasthttp.Get(dst, photoURL)
			if err != nil {
				log.Println("Error with Connectting Facebook Image in app/routes/user.go")
			}

			img, _, _ := image.Decode(bytes.NewReader(body))

			//Define local path and image name
			fileName := "Picture.jpg"
			err = os.MkdirAll("../static/images/profile-photos/"+strconv.FormatInt(usr.UserID, 10), 0755)
			if err != nil {
				log.Printf("Error making directory path to save: %s", err.Error())
				ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
				return
			}
			out, err := os.Create("../static/images/profile-photos/" + strconv.FormatInt(usr.UserID, 10) + "/" + u.UserID + "-" + fileName)
			defer out.Close()
			if err != nil {
				log.Printf("Error routes/user.go OuathCallback() opening destination to save: %s", err.Error())
				ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
				return
			}

			//Convert byte array to JPEG image using Jpge Encode
			var opt jpeg.Options
			opt.Quality = 80
			err = jpeg.Encode(out, img, &opt)
			if err != nil {
				log.Println("Error with saving image local path in app/routes/user.go")
			} else {
				log.Println("Success writing facebook image to local Path in app/routes/user.go")
				log.Println("../static/images/profile-photos/" + strconv.FormatInt(usr.UserID, 10) + "/" + u.UserID + "-" + fileName)
			}

			imageLocalURL := u.UserID + "-" + fileName
			createDefaultProfileWPhoto(usr, imageLocalURL)

			usr.PhotoURL = imageLocalURL
			err = usr.Update()

			if err != nil {
				log.Println("Failed to save facebook user image url to databse in app/routes/user.go")
			} else {
				log.Println("Success to save facebook user image url to database in app/routes/user.go")
			}

		} else {

			log.Printf("Facebook user found in DB %s %s. Updating facebookID and photo if empty", u.FirstName, u.LastName)

			//user found in DB by email or facebook_id
			if len(usr.FacebookID) == 0 || len(usr.PhotoURL) == 0 {
				if len(usr.PhotoURL) == 0 {
					usr.PhotoURL = "http://graph.facebook.com/" + u.UserID + "/picture?width=270&height=270"
				}
				if len(usr.FacebookID) == 0 {
					//user was found by email so save the facebook ID
					usr.FacebookID = u.UserID

				}
				if err := usr.Update(); err != nil {
					utils.Log(ctx, "Facebook Login", fasthttp.StatusInternalServerError, 0, "Unable to update user in database after adding FacebookID:"+err.Error())
					return
				}
			}
		}

		session := globalSessions.StartFasthttp(ctx)
		session.Set("user", usr)
		log.Println("Oauth user session cookie saved, redirecting to dashboard")
		ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)

		// ctx.SetUserValue("user", usr)
		// pg := &render.Page{Title: "Dashboard", TemplateFileName: "authenticated/dashboard.html"}
		// pg.Render(ctx)

	}

}

//AccountSettings gets a user and details to edit
func AccountSettings(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	log.Printf("User ID: %s\n", currUser.UserID)

	//create data as struct variable for the view context
	var data = struct {
		User  *user.User
		Focus string
		//Data for display user's reputation widget
		Score               int16
		Count               int16
		FeedbackDescription string
		FullURL             string
	}{}

	user, err := user.GetUser(currUser.UserID)
	if err != nil {
		log.Printf("Error finding users with id '%d': %s\n", currUser.UserID, err.Error())
		NotFoundRedirect(ctx)
	}

	utils.Log(ctx, "UserEdit", 0, 0, "Found user:"+user.String())

	data.User = user
	data.Focus = string(ctx.QueryArgs().Peek("f"))
	//Data for show reputation widget of user
	data.Score, data.Count, _ = feedback.GetUniversalReputationScoreByUserID(currUser.UserID)
	data.FullURL = config.FullBaseURL()
	data.FeedbackDescription = ""

	if data.Score < 25 {
		data.FeedbackDescription = "Unacceptable"
	} else if data.Score < 40 {
		data.FeedbackDescription = "Needs Improvement"
	} else if data.Score < 55 {
		data.FeedbackDescription = "Acceptable"
	} else if data.Score < 70 {
		data.FeedbackDescription = "Met Expectations"
	} else if data.Score < 85 {
		data.FeedbackDescription = "Exceeds Expectations"
	} else if data.Score < 100 {
		data.FeedbackDescription = "Excellent"
	} else {
		// 100%
		data.FeedbackDescription = "Exceptional"
	}
	pg := &render.Page{Title: "Edit User", TemplateFileName: "authenticated/account-settings.html", Data: data}
	pg.Render(ctx)
}

//UserSave Saves edited user
func UserSave(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	idString := string(ctx.FormValue("id"))
	fname := string(ctx.FormValue("fname"))
	lname := string(ctx.FormValue("lname"))
	email := string(ctx.FormValue("email"))
	photoURL := string(ctx.FormValue("photoURL"))
	address := string(ctx.FormValue("address"))
	city := string(ctx.FormValue("city"))
	region := string(ctx.FormValue("region"))
	postcode := string(ctx.FormValue("postcode"))
	country := string(ctx.FormValue("country"))
	latStr := string(ctx.FormValue("lat"))
	lngStr := string(ctx.FormValue("lng"))

	filePath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + photoURL
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file path:", err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file information:", err)
	}

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = userFormID

	if fileStat.Size() > MaxFileSize {
		log.Println("Image size should not be larger than 4MB.")
		formError.Error = "Image size should not be larger than 4MB."
		render.JSON(ctx, formError, "Image size should not be larger than 4MB", fasthttp.StatusUnprocessableEntity)
		return
	}

	var u *user.User

	if fname == "" || lname == "" || email == "" || postcode == "" {
		formError.Error = "Please fill in all required fields. First name, last name, email and postcode are all required."
		render.JSON(ctx, formError, "Cannot process profile as not all required fields are filled", fasthttp.StatusUnprocessableEntity)
		return
	}

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error parsing user id '%s': %s\n", idString, err.Error())

		formError.Error = "There has been an internal error processing the user. Please refresh the page and try again."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	u, err = user.GetUser(ID)
	if err != nil {
		log.Printf("Error getting user with id '%d': %s\n", ID, err.Error())

		formError.Error = "The user cannot be found, there may be an internal issue."
		formError.Redirect = notFoundURL
		render.JSON(ctx, formError, "User not found error", fasthttp.StatusNotFound)
		return
	}

	/*if u.Postcode != postcode {
	lat, lng, err := utils.GetLatLng(postcode + "+AU")
	if err != nil {
		log.Printf("Error geocoding postcode: %s", err.Error())
		formError.Error = "There has been an error processing your postcode, please verify it is a valid Australian postcode."
		render.JSON(ctx, formError, "Cannot process profile as postcode cannot be geocoded: "+err.Error(), fasthttp.StatusUnprocessableEntity)
		return
	}*/
	u.Address = address
	u.City = city
	u.Region = region
	u.Postcode = postcode
	u.Country = country

	if latStr != "" && lngStr != "" {
		lat, _ := strconv.ParseFloat(latStr, 64)
		lng, _ := strconv.ParseFloat(lngStr, 64)
		u.Lat = lat
		u.Lng = lng
	}

	//}

	if u.PhotoURL != "" && u.PhotoURL != photoURL && !u.IsFacebookPhoto() {
		// photo url has changed, new photo saved, delete the old one
		err2 := deleteUserPhoto(currUser, u.PhotoURL)
		if err2 != nil {
			// just log error for now, maybe we should have the idea of 'tasks' where certain things fail
			// that aren't super horrible, we log them to be manually fixed later
			log.Printf("Error deleting user photo for user '%d' with name '%s': %s\n", ID, u.PhotoURL, err.Error())
		}
	}

	if u.Email != email {
		u.Email = email
		// plus setup for validation of new email, and in the future save previous email to history
	}

	u.FirstName = fname
	u.LastName = lname
	u.PhotoURL = photoURL

	err = u.Update()
	if err != nil {
		log.Printf("Error updating user {%s}: %s\n", u.String(), err.Error())

		formError.Error = "There has been an internal error updating the user."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	// update user in session and context
	updateUserInSessionAndContext(ctx, u)

	utils.Log(ctx, "UserSave", 0, 0, "User Saved:"+ctx.PostArgs().String())
	render.JSON(ctx, "/account-settings", "User successfully saved", fasthttp.StatusOK)
}

//UserPhotoUpload Saves uploaded user photo to disk
func UserPhotoUpload(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	uuid := string(ctx.FormValue("qquuid"))
	filename := string(ctx.FormValue("qqfilename"))
	filesizeString := string(ctx.FormValue("qqtotalfilesize"))
	fileHeader, err := ctx.FormFile("qqfile")
	if err != nil {
		log.Printf("Error getting form file: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	fileSize, err := strconv.ParseInt(filesizeString, 10, 64)
	if err != nil {
		log.Println("Error in routes/user.go UserPhotoUpload() fail to get file size:", err)
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	if fileSize > MaxFileSize {
		log.Printf("Image size should be less than 4MB")
		ctx.SetBody([]byte(`{"success":false, "error": "Image size should be less than 4MB"}`))
		return
	}
	log.Printf("Upload photo details %s, %s, %s, %s\n", uuid, filename, filesizeString, fileHeader.Header.Get("Content-Type"))

	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		log.Printf("Error opening form file: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	//create destination file making sure the path is writeable.
	err = os.MkdirAll("../static/images/profile-photos/"+strconv.FormatInt(currUser.UserID, 10), 0755)
	if err != nil {
		log.Printf("Error making directory path to save: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	dst, err := os.Create("../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + uuid + "-" + filename)
	defer dst.Close()
	if err != nil {
		log.Printf("Error opening destination to save: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file to destination: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	utils.Log(ctx, "UserPhotoSave", 0, 0, "User Image Saved:"+ctx.PostArgs().String())
	//render.JSON(ctx, data, "Profile photo successfully saved", fasthttp.StatusOK)
	ctx.SetBody([]byte(`{"success":true, "filename": "` + uuid + "-" + filename + `"}`))
}

//UserPhotoDelete Deletes uploaded profile photo from disk
func UserPhotoDelete(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}
	uuid := (ctx.UserValue("uuid")).(string)
	filename := string(ctx.QueryArgs().Peek("filename"))

	log.Printf("Delete photo details %s, %s\n", uuid, filename)

	err := deleteUserPhoto(currUser, uuid+"-"+filename)
	if err != nil {
		log.Printf("Error deleting file: %s", err.Error())
		render.JSON(ctx, nil, "Tag search successful", fasthttp.StatusInternalServerError)
		return
	}

	render.JSON(ctx, nil, "Photo delete successful", fasthttp.StatusOK)
}

//PaymentMethods gets a user's payment methods
func PaymentMethods(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	//create data as struct variable for the view context
	var data = struct {
		User *user.User
		//Data for display user's reputation widget
		Score               int16
		Count               int16
		FeedbackDescription string
		FullURL             string
	}{}

	user, err := user.GetUser(currUser.UserID)
	if err != nil {
		log.Printf("Error in routes/user.go PaymentMethods: User not found with id '%d': %s\n", currUser.UserID, err.Error())
		NotFoundRedirect(ctx)
	}

	utils.Log(ctx, "PaymentMethods", 0, 0, "Found user:"+user.String())

	data.User = user
	//Data for show reputation widget of user
	data.Score, data.Count, _ = feedback.GetUniversalReputationScoreByUserID(currUser.UserID)
	data.FullURL = config.FullBaseURL()
	data.FeedbackDescription = ""

	if data.Score < 25 {
		data.FeedbackDescription = "Unacceptable"
	} else if data.Score < 40 {
		data.FeedbackDescription = "Needs Improvement"
	} else if data.Score < 55 {
		data.FeedbackDescription = "Acceptable"
	} else if data.Score < 70 {
		data.FeedbackDescription = "Met Expectations"
	} else if data.Score < 85 {
		data.FeedbackDescription = "Exceeds Expectations"
	} else if data.Score < 100 {
		data.FeedbackDescription = "Excellent"
	} else {
		// 100%
		data.FeedbackDescription = "Exceptional"
	}
	pg := &render.Page{Title: "Payment Methods", TemplateFileName: "authenticated/payment-methods.html", Data: data}
	pg.Render(ctx)
}

//CreditCardSave Saves edited user
func CreditCardSave(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	idString := string(ctx.FormValue("userId"))
	creditCardName := string(ctx.FormValue("creditCardName"))
	creditCardNumber := string(ctx.FormValue("creditCardNumber"))
	expiryDate := string(ctx.FormValue("expiryDate"))
	CVV := string(ctx.FormValue("CVV"))

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = ccSaveFormID

	var u *user.User

	if creditCardName == "" || creditCardNumber == "" || expiryDate == "" || CVV == "" {
		formError.Error = "Please fill in all required fields. Name, credit card number, expiry date and CVV are all required."
		render.JSON(ctx, formError, "Cannot process credit card as not all required fields are filled", fasthttp.StatusUnprocessableEntity)
		return
	}

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardSave: Unable to parse user id '%s': %s\n", idString, err.Error())

		formError.Error = "There has been an internal error processing the user. Please refresh the page and try again."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	u, err = user.GetUser(ID)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardSave: User not found with id '%d': %s\n", ID, err.Error())

		formError.Error = "The user cannot be found, there may be an internal issue."
		formError.Redirect = notFoundURL
		render.JSON(ctx, formError, "User not found error", fasthttp.StatusNotFound)
		return
	}

	// save credit card to vault
	card := &payments.CreditCard{
		Number:     creditCardNumber,
		ExpiryDate: expiryDate,
		CVV:        CVV,
		Name:       creditCardName,
	}
	token, err := payments.StoreCreditCard(card)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardSave: Unable to save credit card to token system: %s\n", err.Error())

		formError.Error = "Unable to save credit card at this time. Please try again later."
		render.JSON(ctx, formError, "Unable to save credit card", fasthttp.StatusUnprocessableEntity)
		return
	}

	// Update user's details
	u.CreditCardID = token.ID
	u.CreditCardMask = token.Mask

	err = u.Update()
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardSave: Unable to update user's credit card {%s}: %s\n", u.String(), err.Error())

		formError.Error = "There has been an internal error updating credit card details."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "CreditCardSave", 0, 0, "Credit Card Saved:"+ctx.PostArgs().String())
	render.JSON(ctx, "/payment-methods", "User's credit card details successfully saved", fasthttp.StatusOK)
}

//CreditCardDelete Deletes saved credit card
func CreditCardDelete(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = ccSaveFormID

	var u *user.User

	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardDelete: Unable to parse user id '%s': %s\n", idString, err.Error())

		formError.Error = "There has been an internal error processing the user. Please refresh the page and try again."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	u, err = user.GetUser(ID)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardDelete: User not found with id '%d': %s\n", ID, err.Error())

		formError.Error = "The user cannot be found, there may be an internal issue."
		formError.Redirect = notFoundURL
		render.JSON(ctx, formError, "User not found error", fasthttp.StatusNotFound)
		return
	}

	// delete from vault, then remove details from user table
	// delete from vault
	err = payments.DeleteCreditCard(u.CreditCardID)
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardDelete: Unable to delete credit card from token system: %s\n", err.Error())

		formError.Error = "Unable to delete credit card at this time. Please try again later."
		render.JSON(ctx, formError, "Unable to delete credit card", fasthttp.StatusUnprocessableEntity)
		return
	}

	u.CreditCardID = ""
	u.CreditCardMask = ""

	err = u.Update()
	if err != nil {
		log.Printf("Error in routes/user.go CreditCardDelete: Unable to update user with credit card details removed {%s}: %s\n", u.String(), err.Error())

		formError.Error = "There has been an internal error removing credit card details."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "CreditCardDelete", 0, 0, "Credit Card Deleted:"+ctx.PostArgs().String())
	render.JSON(ctx, "/payment-methods", "User's credit card details successfully removed", fasthttp.StatusOK)
}

//PayoutMethodSave Saves users payout methods
func PayoutMethodSave(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	idString := string(ctx.FormValue("userId"))
	paypalPayoutMethodType := string(ctx.FormValue("paypalMethodType"))
	paypalAccount := string(ctx.FormValue("paypalAccount"))

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = payoutSaveFormID

	var u *user.User

	if paypalPayoutMethodType == "PPMOBILE" && paypalAccount == "" {
		formError.Error = "Please fill in all required fields. If Paypal Mobile is selected, you must fill in a mobile number."
		render.JSON(ctx, formError, "Save payout details as not all required fields are filled", fasthttp.StatusUnprocessableEntity)
		return
	}

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/user.go PayoutMethodSave: Unable to parse user id '%s': %s\n", idString, err.Error())

		formError.Error = "There has been an internal error processing the user. Please refresh the page and try again."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	u, err = user.GetUser(ID)
	if err != nil {
		log.Printf("Error in routes/user.go PayoutMethodSave: User not found with id '%d': %s\n", ID, err.Error())

		formError.Error = "The user cannot be found, there may be an internal issue."
		formError.Redirect = notFoundURL
		render.JSON(ctx, formError, "User not found error", fasthttp.StatusNotFound)
		return
	}

	// Update user's details
	u.PayoutType = paypalPayoutMethodType
	u.PayoutAccount = paypalAccount

	err = u.Update()
	if err != nil {
		log.Printf("Error in routes/user.go PayoutMethodSave: Unable to update user's credit card {%s}: %s\n", u.String(), err.Error())

		formError.Error = "There has been an internal error updating credit card details."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "PayoutMethodSave", 0, 0, "Payout Methods Saved:"+ctx.PostArgs().String())
	render.JSON(ctx, "/payment-methods#tab2", "User's payout details successfully saved", fasthttp.StatusOK)
}

//****** ONLY HELPER FUNCTIONS BELOW ******///

func deleteUserPhoto(currUser *user.User, fileName string) error {
	filepath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + fileName

	err := os.Remove(filepath)
	return err
}

//ValidatePassword writes "true" or "false" to the response depending on whether the password is validated
func ValidatePassword(ctx *fasthttp.RequestCtx) {
	if len(ctx.FormValue("password")) == 0 {
		ctx.Write([]byte("false"))
	} else {

		if isValidPassword(ctx.FormValue("password")) {
			ctx.Write([]byte("true"))
		} else {
			ctx.Write([]byte("false"))
		}
	}
}

func isValidPassword(s []byte) bool {
	sixOrMore, _, number, upper, lower, _ := verifyPassword(string(s))
	log.Println("Six or more: ", sixOrMore)
	log.Println("Number: ", number)
	log.Println("Upper: ", upper)
	log.Println("Lower: ", lower)
	if sixOrMore == true && number == true && upper == true && lower == true {
		return true
	}
	return false

}

func verifyPassword(s string) (sixOrMore, sixOrMoreLetters, number, upper, lower, special bool) {
	letters := 0
	for _, s := range s {
		switch {
		case unicode.IsNumber(s):
			number = true
		case unicode.IsUpper(s):
			upper = true
			letters++
		case unicode.IsLower(s):
			lower = true
			letters++
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			special = true
		case unicode.IsLetter(s) || s == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}
	sixOrMoreLetters = letters >= 6
	sixOrMore = len(s) >= 6
	return

}

func updateUserInSessionAndContext(ctx *fasthttp.RequestCtx, u *user.User) {
	session := globalSessions.StartFasthttp(ctx)
	session.Set("user", u)
	ctx.SetUserValue("user", u)
}

func getUserFromContext(ctx *fasthttp.RequestCtx, redirectToLogin bool) *user.User {

	//if no user in context redirect to login or return nil
	if ctx.UserValue("user") == nil {
		if redirectToLogin {
			isAjax := bytes.EqualFold(ctx.Request.Header.PeekBytes([]byte("X-Requested-With")), []byte("XMLHttpRequest"))
			if isAjax {
				render.JSON(ctx, nil, "", fasthttp.StatusUnauthorized)
				utils.Log(ctx, "routes/user.go getUserFromContext()", fasthttp.StatusUnauthorized, len(ctx.Response.Body()), "No user in context, redirecting to /login-register via ajaxComplete()")
			} else {
				ctx.Redirect("/login-register", fasthttp.StatusSeeOther)
				utils.Log(ctx, "routes/user.go getUserFromContext()", fasthttp.StatusUnauthorized, len(ctx.Response.Body()), "No user in context, redirecting to /login-register")
			}
		} else {
			utils.Log(ctx, "routes/user.go getUserFromContext()", fasthttp.StatusUnauthorized, len(ctx.Response.Body()), "No user in context, NO redirect was requested to /login-register")
		}
		return nil
	}
	currUser, ok := ctx.UserValue("user").(*user.User)
	if !ok {
		utils.Log(ctx, "routes/user.go getUserFromContext()", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "Could not type assert current user")
		return nil
	}
	return currUser

}

func getUserFromSession(ctx *fasthttp.RequestCtx) {
	// Check for user in session, if it exists, make available in ctx
	// This has to be run for both HTML and JSON requests
	session := globalSessions.StartFasthttp(ctx)
	// Retrieve our struct and type-assert it
	usr := session.Get("user")
	if usr == nil {
		utils.Log(ctx, "Middleware", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "getUserFromSession(): no user in session")

		// eventually, check if there needs to be a user AND ROLE(S) for a certain path, and redirect if necessary
		//ctx.Redirect("/login-register", fasthttp.StatusUnauthorized)
		//return
	} else {
		var u = &user.User{}
		var ok bool
		if u, ok = usr.(*user.User); !ok {
			utils.Log(ctx, "Middleware", fasthttp.StatusBadRequest, len(ctx.Response.Body()), "getUserFromSession(): Error getting user in session:"+u.Email)
			// Same as above, if we can't get the user properly, and the page is secure, redirect
			//ctx.Redirect("/login-register", fasthttp.StatusUnauthorized)
			//return
		} else {
			utils.Log(ctx, "Middleware", fasthttp.StatusOK, len(ctx.Response.Body()), "getUserFromSession(): Logged in user: "+u.Email)
			ctx.SetUserValue("user", u)
			log.Println("Set User Value in getUserFromSession", u.FullName())
		}
	}
}

func createDefaultProfile(usr *user.User) {
	createDefaultProfileWPhoto(usr, "")
}

func createDefaultProfileWPhoto(usr *user.User, photoURL string) {
	profile := &user.Profile{}
	profile.User = *usr
	profile.ServiceCategory = 0
	profile.ProfileType = "b"
	profile.Description = "Default customer profile for buying goods and services."
	profile.Heading = "Customer"
	if photoURL != "" {
		profile.PhotoURL = photoURL
	}

	_, err := user.InsertProfile(profile)
	if err != nil {
		log.Printf("Error inserting default profile {%s}: %s\n", profile.String(), err.Error())
		// we should have a way of logging tasks like these that aren't created automatically but should be.
		// this shouldn't fail the registration, as the user was created successfully, but need to decide what we should do
		// otherwise, we need to setup transactions so that if one fails, they all get rolled back.
	}
}
