package routes

import (
	"fmt"
	"log"
	"strconv"

	"strings"

	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/feedback"
	"github.com/unirep/ur-local-web/app/models/notification"
	"github.com/unirep/ur-local-web/app/models/sdatypes"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

//WriteFeedback writes feedback
func WriteFeedback(ctx *fasthttp.RequestCtx) {
	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)

	} else {
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}
	bookingUUID := (ctx.UserValue("bookingUUID")).(string)

	bookingInfo, _ := user.GetBookingByBookingUUID(bookingUUID)
	bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
	bookingSenderProfile := bookingSenderProfiles[0]

	var data = struct {
		Profile             *user.Profile
		SdaCustomer         *sdatypes.SdaType
		ParentPage          string
		UserID              int64
		Booking             *user.Booking
		SdaProvider         *sdatypes.SdaType
		FeedbackProfileID   int64
		Feedback            *feedback.Feedback
		DisableFlag         bool
		FirstBookingHistory *user.BookingHistory
		LastBookingHistory  *user.BookingHistory
	}{}
	data.Profile = bookingSenderProfile
	data.Booking = bookingInfo
	data.UserID = currUser.UserID
	data.ParentPage = "Dashboard"
	data.SdaCustomer = &sdatypes.SdaType{RefID: data.Profile.ServiceCategory, ProfileType: data.Profile.ProfileType}
	data.SdaProvider = &sdatypes.SdaType{RefID: data.Booking.Profile.ServiceCategory, ProfileType: data.Booking.Profile.ProfileType}
	if data.UserID == data.Booking.User.UserID {
		data.FeedbackProfileID = data.Profile.ProfileID
	} else {
		data.FeedbackProfileID = data.Booking.Profile.ProfileID
	}
	var err error
	data.Feedback, err = feedback.GetFeedbackByBookingIDandCreatedProfileID(data.Booking.BookingID, data.FeedbackProfileID)
	if err != nil {
		log.Println("There is no feedback yet.")
	}
	if data.Feedback != nil {
		data.DisableFlag = true
	} else {
		data.DisableFlag = false
	}
	bookingHistories, err := user.GetBookingHistoriesWithBookingID(data.Booking.BookingID)
	data.LastBookingHistory = bookingHistories[0]
	data.FirstBookingHistory = bookingHistories[len(bookingHistories)-1]

	var pg *render.Page
	if currUser.UserID == bookingInfo.User.UserID {
		pg = &render.Page{Title: "Writing a Feedback", TemplateFileName: "to-provider-feedback.html", Data: data}
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		pg = &render.Page{Title: "Writing a Feedback", TemplateFileName: "to-customer-feedback.html", Data: data}
	}
	pg.Render(ctx)
}

//ViewFeedback shows feedback
func ViewFeedback(ctx *fasthttp.RequestCtx) {

	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)

	} else {
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}

	bookingUUID := (ctx.UserValue("bookingUUID")).(string)

	bookingInfo, _ := user.GetBookingByBookingUUID(bookingUUID)
	bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
	bookingSenderProfile := bookingSenderProfiles[0]
	var data = struct {
		Profile          *user.Profile
		ParentPage       string
		UserID           int64
		Booking          *user.Booking
		Feedback         *feedback.Feedback
		FeedbackComplete bool
	}{}

	if currUser.UserID == bookingInfo.User.UserID {
		data.Profile = &bookingInfo.Profile
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		data.Profile = bookingSenderProfile
	}
	data.UserID = currUser.UserID
	data.ParentPage = "WriteFeedback"
	var pg *render.Page

	checkTwoFeedback := feedback.CheckTwoFeedbackWrited(bookingInfo.BookingID)

	if checkTwoFeedback == false {
		//we can't show feedback because client or provider did not write feedback
		data.FeedbackComplete = false
	} else {
		data.FeedbackComplete = true
		//show feedback
		var err error
		data.Feedback, err = feedback.GetFeedbackByBookingIDandCreatedProfileID(bookingInfo.BookingID, data.Profile.ProfileID)
		if err != nil {
			log.Println("Error in routes/feedback.go ViewFeedback() fail to get feedback:", err)
			NotFoundRedirect(ctx)
		}
	}

	pg = &render.Page{Title: "Feedback Detail", TemplateFileName: "view-feedback.html", Data: data}
	pg.Render(ctx)
}

//CheckFeedback renders view feedback page if received feedback
func CheckFeedback(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)

	descriptionText := string(ctx.FormValue("descriptionText"))
	SDAtext := string(ctx.FormValue("SDAtext"))
	SDAtextArray := strings.Split(SDAtext, ",")
	score, _ := strconv.ParseInt(string(ctx.FormValue("score")), 10, 64)
	commentText := string(ctx.FormValue("commentText"))
	bookingID, _ := strconv.ParseInt(string(ctx.FormValue("bookingID")), 10, 64)
	senderProfileID, _ := strconv.ParseInt(string(ctx.FormValue("feedbackProfileID")), 10, 64)
	feedbackDataItem, err := feedback.GetFeedbackByBookingIDandCreatedProfileID(bookingID, senderProfileID)
	feedbackToCustomer := string(ctx.FormValue("feedbackToCustomer"))
	//block feedback which submitted again...

	if err == nil && feedbackDataItem != nil {
		log.Println("id :", feedbackDataItem.FeedbackID)
		return
	}

	bookingInfo, _ := user.GetBookingWithBookingId(bookingID)
	bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
	bookingSenderProfile := bookingSenderProfiles[0]

	var feedbackData *feedback.Feedback
	feedbackData = &feedback.Feedback{}
	feedbackData.BookingID = bookingID
	feedbackData.Comment = commentText
	feedbackData.SenderProfileID = senderProfileID
	// If the current user is customer, set the receive Feedback Profile set as provider profile ID
	// else if the current user is provider set the receiving Feedback Profile ID set as customer Profile ID
	if currUser.UserID == bookingInfo.User.UserID {
		feedbackData.ReceiverProfileID = bookingInfo.Profile.ProfileID
	} else {
		feedbackData.ReceiverProfileID = bookingSenderProfile.ProfileID
	}

	feedbackData.Description = descriptionText
	feedbackData.Score = int16(score)
	feedbackData.SdaText = SDAtextArray
	//check feedback to customer
	if feedbackToCustomer != "" {
		switch feedbackToCustomer {
		case "Positive":
			log.Println("FeedbackToCustomer: ", "positive")
			feedbackData.Positive = true
			feedbackData.Neutral = false
			feedbackData.Negative = false
			feedbackData.Score = 100
			break
		case "Neutral":
			log.Println("FeedbackToCustomer: ", "neutral")
			feedbackData.Positive = false
			feedbackData.Neutral = true
			feedbackData.Negative = false
			feedbackData.Score = 50
			break
		case "Negative":
			log.Println("FeedbackToCustomer: ", "negative")
			feedbackData.Positive = false
			feedbackData.Neutral = false
			feedbackData.Negative = true
			feedbackData.Score = 10
			break
		case "NotRated":
			log.Println("FeedbackToCustomer: Not Rated")
			feedbackData.Positive = false
			feedbackData.Neutral = false
			feedbackData.Negative = false
			feedbackData.Score = 0
			break
		default:

		}

	}

	feedbackData.FeedbackID, err = feedback.InsertFeedback(feedbackData)

	if err != nil {
		log.Println("Error app/routes/feedback.go/CheckFeedback() :", err)
	}

	var notificationItem *notification.Notification
	notificationItem = &notification.Notification{}
	notificationItem.NotificationType = notification.NotificationTypeFeedbackReceived
	notificationItem.EntityID = bookingID
	bookingHistories, err := user.GetBookingHistoriesWithBookingID(bookingID)
	if err != nil {
		log.Println("Error in routes/feedback.go fail to get booking histories:", err)
	}

	if currUser.UserID == bookingInfo.User.UserID {
		notificationItem.SenderID = currUser.UserID
		notificationItem.ReceiverID = bookingInfo.Profile.User.UserID
		notificationItem.Unread = append(notificationItem.Unread, bookingInfo.Profile.User.UserID)
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		notificationItem.SenderID = currUser.UserID
		notificationItem.ReceiverID = bookingInfo.User.UserID
		notificationItem.Unread = append(notificationItem.Unread, bookingInfo.User.UserID)
	}
	notificationItem.EntityHistoryID = bookingHistories[0].BookingHistoryID
	log.Println("notification data:", notificationItem)
	err = notification.InsertNotification(notificationItem)
	if err != nil {
		log.Println("Error in routes/feedback.go failed to insert feedback:", err)
		return
	}

	twoFeedbackWrited := feedback.CheckTwoFeedbackWrited(bookingInfo.BookingID)
	bookingIDstr := fmt.Sprintf("%v", bookingInfo.BookingID)
	var receiveUser user.User
	if currUser.UserID == bookingInfo.Profile.User.UserID {
		receiveUser = bookingInfo.User
	} else {
		receiveUser = bookingInfo.Profile.User
	}
	if twoFeedbackWrited == false {
		defer emailing.SendNewNotificationEmail(&receiveUser, currUser, notificationItem, "UR Local – You Have A New Feedback (ID "+bookingIDstr+")", "email_have_new_feedback.html", "email_have_new_feedback.txt")
	} else {
		defer emailing.SendNewNotificationEmail(&bookingInfo.Profile.User, currUser, notificationItem, "UR Local – You Have A New Feedback (ID "+bookingIDstr+")", "email_have_new_finished_feedback.html", "email_have_new_finished_feedback.txt")
	}
	var redirectURL string
	redirectURL = "/view-feedback/" + bookingInfo.BookingUUID.String()
	render.JSON(ctx, redirectURL, "redirecting feedback page", fasthttp.StatusOK)

}
