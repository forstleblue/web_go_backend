package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/notification"
	"github.com/unirep/ur-local-web/app/models/payments"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

const bookingFormID = "booking-form"

// BookingSave save booking information
func BookingSave(ctx *fasthttp.RequestCtx) {
	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)

	} else {
		// Normally we should not reach here.
		// Just in case someone cracks the js file.
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}

	var err error

	profileIDParam := string(ctx.FormValue("profileId"))
	message := string(ctx.FormValue("message"))
	fromDate := string(ctx.FormValue("fromDate"))
	toDate := string(ctx.FormValue("toDate"))
	fromTime := string(ctx.FormValue("fromTime"))
	toTime := string(ctx.FormValue("toTime"))
	address := string(ctx.FormValue("address"))
	fee := string(ctx.FormValue("fee"))
	totalPrice := string(ctx.FormValue("totalPrice"))
	frequencyUnit := string(ctx.FormValue("frequencyUnit"))
	frequencyValue := string(ctx.FormValue("frequencyValue"))
	needMessage := string(ctx.FormValue("needMessage"))
	bookingUpdate := string(ctx.FormValue("bookingUpdate"))
	// Setup form error object in case needed
	var bookingFormError = &JSONFormError{}
	bookingFormError.Form = "booking-form"

	if needMessage == "YES" && len(message) == 0 {
		bookingFormError.Error = "'Message' is required field"
		render.JSON(ctx, bookingFormError, "'Message' must be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(address) == 0 {
		bookingFormError.Error = "'Address' is required field"
		render.JSON(ctx, bookingFormError, "'Address' must be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	if checkDateAndTimeValidaty(fromDate, toDate, fromTime, toTime) == false {
		log.Println("Error in app/routes/booking.go BookingSave() From Time should be earlier than To Time")
		bookingFormError.Error = "From Time should be earlier than To Time"
		render.JSON(ctx, bookingFormError, "From Time should be earlier than To Time", fasthttp.StatusUnprocessableEntity)
		return
	}
	var booking *user.Booking

	profileID, err := strconv.ParseInt(profileIDParam, 10, 64)
	profile, err := user.GetProfile(profileID)

	booking = &user.Booking{}

	booking.User = *currUser
	booking.Profile = *profile

	bookingHistory := &user.BookingHistory{}
	var errTotalPrice, errFee, errFrequency error
	var frequencyVal int64
	frequencyVal, errFrequency = strconv.ParseInt(frequencyValue, 10, 64)

	if errFrequency != nil && len(frequencyValue) != 0 {
		log.Println("Error routes/booking.go BookingSave() failed to convert frequency value to int:", errFrequency.Error())
		bookingFormError.Error = "Frequnecy should be number"
		render.JSON(ctx, bookingFormError, "Frequnecy should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	bookingHistory.FrequencyValue = frequencyVal

	var feeValue, totalPriceValue float64
	feeValue, errFee = strconv.ParseFloat(fee, 64)

	if errFee != nil && len(fee) != 0 {
		log.Println("Error routes/booking.go BookingSave() failed to convert fee to int:", errFee.Error())
		bookingFormError.Error = "Fee should be number"
		render.JSON(ctx, bookingFormError, "Fee should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	totalPriceValue, errTotalPrice = strconv.ParseFloat(totalPrice, 64)

	if errTotalPrice != nil && len(totalPrice) != 0 {
		log.Println("Error routes/booking.go BookingSave() failed to convert totalPrice to int:", errTotalPrice.Error())
		bookingFormError.Error = "Total Price should be number"
		render.JSON(ctx, bookingFormError, "Total Price should be number", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(fee) != 0 {
		bookingHistory.Fee = decimal.NewFromFloat(feeValue)
	}
	if len(totalPrice) != 0 {
		bookingHistory.TotalPrice = decimal.NewFromFloat(totalPriceValue)
	}

	if bookingUpdate == "Yes" {
		lastBookingHistoryIDParam := string(ctx.FormValue("bookingHistoryId"))
		lastBookingHistoryID, err := strconv.ParseInt(lastBookingHistoryIDParam, 10, 64)
		if err != nil {
			log.Println("Error in routes/booking.go BookingSave() fail to get bookingHistoryID:", err)
		}
		err = user.UpdateBookingHistory(lastBookingHistoryID, "Updated")
		if err != nil {
			log.Println("Error in routes/booking.go BookingSave() fail to update booking history")
		}
	}

	var bookingID int64
	bookingID, err = user.InsertBooking(booking)
	if err != nil {
		log.Printf("Error in app/routes/booking.go  InsertBooking Failed  %s\n", err.Error())
		bookingFormError.Error = "There has been an internal error inserting Booking Data."
		bookingFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingHistory.Message = message
	bookingHistory.BookingID = bookingID
	bookingHistory.FromDate = fromDate
	bookingHistory.ToDate = toDate
	bookingHistory.FromTime = fromTime
	bookingHistory.ToTime = toTime
	bookingHistory.Address = address

	bookingHistory.FrequencyUnit = frequencyUnit
	bookingHistory.FrequencyValue, err = strconv.ParseInt(frequencyValue, 10, 64)
	bookingHistory.BookingStatus = "New"
	bookingHistory.UserID = currUser.UserID

	var bookingHistoryID int64
	bookingHistoryID, err = user.InsertBookingHistory(bookingHistory)

	if err != nil {
		log.Printf("Error in app/routes/booking.go  InsertBooking History Failed  %s\n", err.Error())
		bookingFormError.Error = "There has been an internal error inserting Booking History Data."
		bookingFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	//remove redirect session information if booking profile id is same as redirection sesion information profileID
	session := globalSessions.StartFasthttp(ctx)
	redirectInfo := session.Get("redirect")
	if redirectInfo != nil && profileIDParam == redirectInfo.(map[string]string)["profileId"] {
		session.Set("redirect", nil)
		log.Println("remove redirect session information after booking save")
	}
	var notificationItem *notification.Notification
	notificationItem = &notification.Notification{}

	notificationItem.EntityID = bookingID
	notificationItem.EntityHistoryID = bookingHistoryID
	notificationItem.NotificationType = notification.NotificationTypeBookingRequest
	notificationItem.SenderID = currUser.UserID
	notificationItem.ReceiverID = profile.User.UserID
	notificationItem.NotificationText = ""
	notificationItem.Unread = append(notificationItem.Unread, notificationItem.ReceiverID)
	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go InsertNotification Failed %s\n", err.Error())
		return
	}
	bookingIDstr := fmt.Sprintf("%v", bookingID)
	defer emailing.SendNewNotificationEmail(&profile.User, currUser, notificationItem, "UR Local – You Have A New Booking Request (ID "+bookingIDstr+")", "email_booking_request.html", "email_booking_request.txt")
	render.JSON(ctx, "/dashboard", "Booking successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

func checkDateAndTimeValidaty(fromDate, toDate, fromTime, toTime string) bool {
	var fromValue string
	var toValue string
	emptyDateField := false
	if fromDate != "" && toDate != "" {
		if fromTime == "" && toTime == "" {
			fromValue = fromDate + " " + "3:04 PM"
			toValue = toDate + " " + "3:04 PM"
		} else if fromTime == "" || toTime == "" {
			if fromTime == "" {
				fromValue = fromDate + " " + toTime
				toValue = toDate + " " + toTime
			} else {
				fromValue = fromDate + " " + fromTime
				toValue = toDate + " " + fromDate
			}
		} else {
			fromValue = fromDate + " " + fromTime
			toValue = toDate + " " + toTime
		}

	} else if fromDate == "" && toDate == "" {
		if fromTime != "" && toTime != "" {
			fromValue = "2/01/2006" + " " + fromTime
			toValue = "2/01/2006" + " " + toTime
		} else {
			emptyDateField = true
		}
	} else {
		if fromDate == "" {
			if fromTime != "" && toTime != "" {
				fromValue = toDate + " " + fromTime
				toValue = toDate + " " + toTime
			} else {
				emptyDateField = true
			}

		} else {
			if fromTime != "" && toTime != "" {
				fromValue = fromDate + " " + fromTime
				toValue = fromDate + " " + toTime
			} else {
				emptyDateField = true
			}
		}
	}
	DateTimeLayOut := "2/01/2006 3:04 PM"

	timeStampFrom, _ := time.Parse(DateTimeLayOut, fromValue)
	timeStampTo, _ := time.Parse(DateTimeLayOut, toValue)

	if timeStampFrom.Unix() >= timeStampTo.Unix() && emptyDateField == false {
		return false
	}
	return true
}

//BookingMessageUpdate updates booking information in MESSAGE option
func BookingMessageUpdate(ctx *fasthttp.RequestCtx) {
	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)
	} else {
		// Normally we should not reach here.
		// Just in case someone cracks the js file.
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}
	var err error

	bookingIDParam := string(ctx.FormValue("bookingId"))
	message := string(ctx.FormValue("message"))
	fromDate := string(ctx.FormValue("fromDate"))
	toDate := string(ctx.FormValue("toDate"))
	fromTime := string(ctx.FormValue("fromTime"))
	toTime := string(ctx.FormValue("toTime"))
	address := string(ctx.FormValue("address"))
	fee := string(ctx.FormValue("fee"))
	totalPrice := string(ctx.FormValue("totalPrice"))
	frequencyUnit := string(ctx.FormValue("frequencyUnit"))
	frequencyValue := string(ctx.FormValue("frequencyValue"))

	bookingID, err := strconv.ParseInt(bookingIDParam, 10, 64)
	// Setup form error object in case needed
	var bookingResponseFormError = &JSONFormError{}
	bookingResponseFormError.Form = "booking-response-form"
	if len(message) == 0 {
		bookingResponseFormError.Error = "'Message' is required field"
		render.JSON(ctx, bookingResponseFormError, "'Message' must be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(address) == 0 {
		bookingResponseFormError.Error = "'Address' is required field"
		render.JSON(ctx, bookingResponseFormError, "'Address' must be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	if checkDateAndTimeValidaty(fromDate, toDate, fromTime, toTime) == false {
		log.Println("Error in app/routes/booking.go BookingMessageUpdate() From Time should be earlier than To Time")
		bookingResponseFormError.Error = "From Time should be earlier than To Time"
		render.JSON(ctx, bookingResponseFormError, "From Time should be earlier than To Time", fasthttp.StatusUnprocessableEntity)
		return
	}

	bookingHistory := &user.BookingHistory{}
	bookingHistory.Message = message
	bookingHistory.BookingID = bookingID
	bookingHistory.FromDate = fromDate
	bookingHistory.ToDate = toDate
	bookingHistory.FromTime = fromTime
	bookingHistory.ToTime = toTime
	bookingHistory.Address = address
	bookingHistory.BookingStatus = "Message"
	bookingHistory.UserID = currUser.UserID

	bookingInfo, _ := user.GetBookingWithBookingId(bookingID)
	var errTotalPrice, errFee, errFrequency error
	var frequencyVal int64
	frequencyVal, errFrequency = strconv.ParseInt(frequencyValue, 10, 64)

	if errFrequency != nil && len(frequencyValue) != 0 {
		log.Println("Error routes/booking.go BookingMessageUpdate() failed to convert frequency value to int:", errFrequency.Error())
		bookingResponseFormError.Error = "Frequnecy should be number"
		render.JSON(ctx, bookingResponseFormError, "Frequnecy should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	bookingHistory.FrequencyValue = frequencyVal

	var feeValue, totalPriceValue float64
	feeValue, errFee = strconv.ParseFloat(fee, 64)
	if errFee != nil && len(fee) != 0 {
		log.Println("Error routes/booking.go BookingMessageUpdate() failed to convert fee to int:", errFee.Error())
		bookingResponseFormError.Error = "Fee should be number"
		render.JSON(ctx, bookingResponseFormError, "Fee should be number", fasthttp.StatusUnprocessableEntity)
		return
	}

	totalPriceValue, errTotalPrice = strconv.ParseFloat(totalPrice, 64)
	if errTotalPrice != nil && totalPrice != "" {
		log.Println("Error routes/booking.go BookingMessageUpdate() failed to convert fee and totalPrice to int:", errTotalPrice)
		bookingResponseFormError.Error = "Total Price should be number"
		render.JSON(ctx, bookingResponseFormError, "Total Price should be number", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(fee) != 0 {
		bookingHistory.Fee = decimal.NewFromFloat(feeValue)
	}
	if len(totalPrice) != 0 {
		bookingHistory.TotalPrice = decimal.NewFromFloat(totalPriceValue)
	}

	bookingHistory.FrequencyUnit = frequencyUnit
	bookingHistory.FrequencyValue, err = strconv.ParseInt(frequencyValue, 10, 64)

	var bookingHistoryID int64
	bookingHistoryID, err = user.InsertBookingHistory(bookingHistory)

	if err != nil {
		log.Printf("Error in app/routes/booking.go  BookingMessageUpdate() Failed  %s\n", err.Error())
		bookingResponseFormError.Error = "There has been an internal update Booking Data."
		bookingResponseFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	var notificationItem *notification.Notification
	notificationItem = &notification.Notification{}
	notificationItem.SenderID = currUser.UserID
	notificationItem.EntityID = bookingID

	if currUser.UserID == bookingInfo.User.UserID {
		notificationItem.ReceiverID = bookingInfo.Profile.User.UserID
	} else {
		notificationItem.ReceiverID = bookingInfo.User.UserID
	}
	notificationItem.NotificationType = notification.NotificationTypeBookingResponse
	notificationItem.EntityHistoryID = bookingHistoryID
	notificationItem.Unread = append(notificationItem.Unread, notificationItem.ReceiverID)
	notificationItem.NotificationText = ""
	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go BookingMessageUpdate() Failed to insert notification %s\n", err.Error())
		return
	}

	bookingIDstr := fmt.Sprintf("%v", bookingInfo.BookingID)
	if currUser.UserID == bookingInfo.Profile.User.UserID {
		defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local – You Have a Message About a Booking (ID "+bookingIDstr+")", "email_booking_request_message_to_customer.html", "email_booking_request_message_to_customer.txt")
	} else {
		defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local – Booking Message From Customer (ID "+bookingIDstr+")", "email_booking_response_message_from_customer.html", "email_booking_response_message_from_customer.txt")
	}
	render.JSON(ctx, "/dashboard", "Booking successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

// BookingUpdate updates booking information
func BookingUpdate(ctx *fasthttp.RequestCtx) {
	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)
	} else {
		// Normally we should not reach here.
		// Just in case someone cracks the js file.
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}

	var err error

	notificationIDParam := string(ctx.FormValue("notificationID"))
	message := string(ctx.FormValue("message"))
	bookingStatus := string(ctx.FormValue("bookingStatus"))

	notificationID, err := strconv.ParseInt(notificationIDParam, 10, 64)

	// Setup form error object in case needed
	var bookingResponseFormError = &JSONFormError{}
	if bookingStatus == user.BookingHistoryStatusPendingInCompletion {
		bookingResponseFormError.Form = "booking-accept-form"
	} else if bookingStatus == user.BookingHistoryStatusDecline {
		bookingResponseFormError.Form = "booking-decline-form"
	} else if bookingStatus == user.BookingHistoryStatusCancel {
		bookingResponseFormError.Form = "booking-cancel-form"
	}

	if bookingStatus != user.BookingHistoryStatusPendingInCompletion && len(message) == 0 {
		log.Println("'Message' field is required.", err)
		bookingResponseFormError.Error = "'Message' field is required."
		render.JSON(ctx, bookingResponseFormError, "'Message' field is required.", fasthttp.StatusUnprocessableEntity)
		return
	}

	notificationItem, err := notification.GetNotificationByID(notificationID)
	if err != nil {
		log.Println("Error in routes/booking.go BookingUpdate() fail to get notification:", err)
		bookingResponseFormError.Error = "There has been an internal error getting notification Data."
		bookingResponseFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingHistory, _ := user.GetBookingHistoryWithHistoryID(notificationItem.EntityHistoryID)
	bookingInfo, _ := user.GetBookingWithBookingId(notificationItem.EntityID)
	bookingHistory.Message = message
	bookingHistory.BookingStatus = bookingStatus

	bookingHistoryID, err := user.InsertBookingHistory(bookingHistory)
	if err != nil {
		log.Println("Error in routes/booking.go BookingUpdate() fail to insert booking history:", err)
		bookingResponseFormError.Error = "There has been an internal error inserting booking history data."
		bookingResponseFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	var notificationResponse *notification.Notification
	notificationResponse = &notification.Notification{}
	notificationResponse.SenderID = currUser.UserID
	notificationResponse.EntityID = bookingInfo.BookingID
	notificationResponse.NotificationType = notification.NotificationTypeBookingResponse

	if currUser.UserID == bookingInfo.User.UserID {
		notificationResponse.ReceiverID = bookingInfo.Profile.User.UserID
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		notificationResponse.ReceiverID = bookingInfo.User.UserID
	}

	notificationResponse.EntityHistoryID = bookingHistoryID
	notificationResponse.Unread = append(notificationItem.Unread, notificationResponse.ReceiverID)
	notificationResponse.NotificationText = ""
	err = notification.InsertNotification(notificationResponse)

	if err != nil {
		log.Printf("Error in app/routes/booking.go InsertNotification Failed %s\n", err.Error())
		bookingResponseFormError.Error = "There has been an internal error inserting notification Data."
		bookingResponseFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	if currUser.UserID == bookingInfo.Profile.User.UserID && bookingStatus == user.BookingHistoryStatusPendingInCompletion {
		var notificationResponse *notification.Notification
		notificationResponse = &notification.Notification{}
		notificationResponse.SenderID = currUser.UserID
		notificationResponse.ReceiverID = currUser.UserID
		notificationResponse.EntityID = bookingInfo.BookingID
		notificationResponse.NotificationType = notification.NotificationTypeBookingResponse
		notificationResponse.EntityHistoryID = bookingHistoryID
		notificationResponse.Unread = append(notificationItem.Unread, notificationResponse.ReceiverID)
		notificationResponse.NotificationText = ""
		err = notification.InsertNotification(notificationResponse)
		if err != nil {
			log.Printf("Error in app/routes/booking.go InsertNotification Failed %s\n", err.Error())
			bookingResponseFormError.Error = "There has been an internal error inserting notification Data."
			bookingResponseFormError.Redirect = internalServerErrorURL
			render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
			return
		}
	} else if currUser.UserID == bookingInfo.User.UserID && bookingStatus == user.BookingHistoryStatusPendingInCompletion {
		var notificationResponse *notification.Notification
		notificationResponse = &notification.Notification{}
		notificationResponse.SenderID = currUser.UserID
		notificationResponse.ReceiverID = currUser.UserID
		notificationResponse.EntityID = bookingInfo.BookingID
		notificationResponse.NotificationType = notification.NotificationTypeBookingResponse
		notificationResponse.EntityHistoryID = bookingHistoryID
		notificationResponse.Unread = append(notificationItem.Unread, notificationResponse.ReceiverID)
		notificationResponse.NotificationText = ""
		err = notification.InsertNotification(notificationResponse)
		if err != nil {
			log.Printf("Error in app/routes/booking.go InsertNotification Failed %s\n", err.Error())
			bookingResponseFormError.Error = "There has been an internal error inserting notification Data."
			bookingResponseFormError.Redirect = internalServerErrorURL
			render.JSON(ctx, bookingResponseFormError, "Internal server error", fasthttp.StatusInternalServerError)
			return
		}
	}

	bookingIDstr := fmt.Sprintf("%v", bookingInfo.BookingID)
	if bookingStatus == "Pending Completion" {
		if currUser.UserID == bookingInfo.Profile.User.UserID {
			defer emailing.SendNewNotificationEmail(notificationResponse.Receiver(), notificationResponse.Sender(), notificationResponse, "UR Local – Your Booking Request is Accepted (ID "+bookingIDstr+")", "email_booking_request_accepted_to_customer.html", "email_booking_request_accepted_to_customer.txt")
		} else {
			defer emailing.SendNewNotificationEmail(notificationResponse.Receiver(), notificationResponse.Sender(), notificationResponse, "UR Local – Your Booking Request is Accepted (ID "+bookingIDstr+")", "email_booking_response_accepted_by_customer.html", "email_booking_response_accepted_by_customer.txt")
		}
	}
	if bookingStatus == "Decline" {
		defer emailing.SendNewNotificationEmail(notificationResponse.Receiver(), notificationResponse.Sender(), notificationResponse, "UR Local – Your Booking Request was Unsuccessful (ID "+bookingIDstr+")", "email_booking_request_declined_to_customer.html", "email_booking_request_declined_to_customer.txt")
	}
	if bookingStatus == "Cancel" {
		if currUser.UserID == bookingInfo.Profile.User.UserID {
			defer emailing.SendNewNotificationEmail(notificationResponse.Receiver(), notificationResponse.Sender(), notificationResponse, "UR Local – Booking Cancelled by Service Provider (ID "+bookingIDstr+")", "email_booking_cancelled_by_serviceprovider.html", "email_booking_cancelled_by_serviceprovider.txt")
		} else {
			defer emailing.SendNewNotificationEmail(notificationResponse.Receiver(), notificationResponse.Sender(), notificationResponse, "UR Local – Booking Cancelled by Customer (ID "+bookingIDstr+")", "email_booking_cancelled_by_customer.html", "email_booking_cancelled_by_customer.txt")
		}
	}
	render.JSON(ctx, "/dashboard", "Booking successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

//BookingPaymentRequest sends payment request to customer
func BookingPaymentRequest(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	bookingIDParam := string(ctx.FormValue("bookingId"))

	paymentMessage := string(ctx.FormValue("paymentMessage"))
	fromDate := string(ctx.FormValue("fromDate"))
	toDate := string(ctx.FormValue("toDate"))
	fromTime := string(ctx.FormValue("fromTime"))
	toTime := string(ctx.FormValue("toTime"))
	address := string(ctx.FormValue("address"))
	fee := string(ctx.FormValue("fee"))
	totalPriceParam := string(ctx.FormValue("totalPrice"))
	frequencyUnit := string(ctx.FormValue("frequencyUnit"))
	frequencyValue := string(ctx.FormValue("frequencyValue"))

	var paymenRequestFormError = &JSONFormError{}
	paymenRequestFormError.Form = "payment-request-form"

	if len(address) == 0 {
		paymenRequestFormError.Error = "'Address' is required field"
		render.JSON(ctx, paymenRequestFormError, "'Address' must be specified", fasthttp.StatusUnprocessableEntity)
		return
	}

	bookingID, err := strconv.ParseInt(bookingIDParam, 10, 64)
	if err != nil {
		log.Println("Error in routes/booking.go BookingPaymentRequest() fail to get Booking ID:", err)
		paymenRequestFormError.Error = "There has been an internal error getting Booking ID."
		paymenRequestFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, paymenRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingHistory := &user.BookingHistory{}

	var errTotalPrice, errFee, errFrequency error
	var frequencyVal int64
	frequencyVal, errFrequency = strconv.ParseInt(frequencyValue, 10, 64)

	if errFrequency != nil && len(frequencyValue) != 0 {
		log.Println("Error routes/booking.go BookingPaymentRequest() failed to convert frequency value to int:", errFrequency.Error())
		paymenRequestFormError.Error = "Frequnecy should be number"
		render.JSON(ctx, paymenRequestFormError, "Frequnecy should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	bookingHistory.FrequencyValue = frequencyVal

	var feeValue, totalPriceValue float64
	feeValue, errFee = strconv.ParseFloat(fee, 64)

	if errFee != nil && len(fee) != 0 {
		log.Println("Error routes/booking.go BookingPaymentRequest() failed to convert fee to int:", errFee.Error())
		paymenRequestFormError.Error = "Fee should be number"
		render.JSON(ctx, paymenRequestFormError, "Fee should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	totalPriceValue, errTotalPrice = strconv.ParseFloat(totalPriceParam, 64)

	if errTotalPrice != nil && len(totalPriceParam) != 0 {
		log.Println("Error routes/booking.go BookiBookingPaymentRequestngSave() failed to convert totalPrice to int:", errTotalPrice.Error())
		paymenRequestFormError.Error = "Total Price should be number"
		render.JSON(ctx, paymenRequestFormError, "Total Price should be number", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(fee) != 0 {
		bookingHistory.Fee = decimal.NewFromFloat(feeValue)
	} else if len(totalPriceParam) != 0 {
		bookingHistory.TotalPrice = decimal.NewFromFloat(totalPriceValue)
	}

	bookingHistory.Message = paymentMessage
	bookingHistory.BookingID = bookingID
	bookingHistory.FromDate = fromDate
	bookingHistory.ToDate = toDate
	bookingHistory.FromTime = fromTime
	bookingHistory.ToTime = toTime
	bookingHistory.Address = address

	bookingHistory.FrequencyUnit = frequencyUnit
	bookingHistory.FrequencyValue, err = strconv.ParseInt(frequencyValue, 10, 64)
	bookingHistory.BookingStatus = "Pending Payment"
	bookingHistory.UserID = currUser.UserID

	var bookingHistoryID int64
	bookingHistoryID, err = user.InsertBookingHistory(bookingHistory)

	var totalPrice int32
	if bookingHistory.Fee.IntPart() != 0 {
		totalPrice = int32(bookingHistory.Fee.IntPart())
	} else if bookingHistory.TotalPrice.IntPart() == 0 {
		totalPrice = int32(bookingHistory.TotalPrice.IntPart())
	}

	bookingInfo, err := user.GetBookingWithBookingId(bookingID)

	if err != nil {
		log.Println("Error in routes/booking.go BookingPaymentRequest() fail to get Booking Data:", err)
		paymenRequestFormError.Error = "There has been an internal error getting Booking Data."
		paymenRequestFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, paymenRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	createdTime := time.Now().UTC()
	messageItem := user.MessageData{Sender: currUser.FullName(), Time: createdTime, Message: paymentMessage}
	var messages []user.MessageData
	messages = append(messages, messageItem)
	var paymentID int64
	prCreated := bookingInfo.CheckPaymentRequest()
	var prStatus, notiType string
	if prCreated == false {
		prStatus = user.PaymentRequestStatusNew
		notiType = notification.NotificationTypePaymentRequest
		pr := &user.Payment{Booking: *bookingInfo, Amount: totalPrice, Message: messages, Status: prStatus}
		paymentID, err = user.InsertPayment(pr)
		if err != nil {
			log.Println("Error in routes/booking.go BookingPaymentRequest(ctx *fasthttp.RequestCtx) Fail to insert payment request ", err)
			paymenRequestFormError.Error = "There has been an internal error inserting payment request."
			paymenRequestFormError.Redirect = internalServerErrorURL
			render.JSON(ctx, paymenRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
			return
		}
	} else {
		notiType = notification.NotificationTypePaymentRequestResponse
		payment, err := user.GetPaymentByBookingID(bookingInfo.BookingID)
		if err != nil {
			log.Println("Error in routes/booking.go BookingPaymentRequest(ctx *fasthttp.RequestCtx) Fail to get payment.", err)
			paymenRequestFormError.Error = "There has been an internal error getting payment."
			paymenRequestFormError.Redirect = internalServerErrorURL
			render.JSON(ctx, paymenRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
			return
		}
		if payment.Amount != totalPrice {
			payment.Amount = totalPrice
			err := payment.Update()
			if err != nil {
				log.Println("Error in routes/booking.go BookingPaymentRequest(ctx *fasthttp.RequestCtx) Fail to updating payment.", err)
				paymenRequestFormError.Error = "There has been an internal error updateing payment."
				paymenRequestFormError.Redirect = internalServerErrorURL
				render.JSON(ctx, paymenRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
				return
			}
		}
		paymentID = payment.PaymentID
	}

	var receiverID int64
	if currUser.UserID == bookingInfo.User.UserID {
		receiverID = bookingInfo.Profile.User.UserID
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		receiverID = bookingInfo.User.UserID
	}

	notificationItem := &notification.Notification{
		EntityID:         paymentID,
		NotificationType: notiType,
		SenderID:         currUser.UserID,
		ReceiverID:       receiverID,
		EntityHistoryID:  bookingHistoryID,
		Unread:           []int64{receiverID},
	}

	err = notification.InsertNotification(notificationItem)
	if err != nil {
		log.Println("Error in /routes/booking.go PaymentRequest(): Failed to insert payment request notificaiton", err)
	}

	bookingIDstr := fmt.Sprintf("%v", bookingInfo.BookingID)
	//
	if currUser.UserID == bookingInfo.Profile.User.UserID {
		defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local – Payment Request for Booking (ID "+bookingIDstr+")", "email_payment_request_for_booking.html", "email_payment_request_for_booking.txt")
	} else {
		defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local – Message from Customer about Payment (ID "+bookingIDstr+")", "email_payment_request_amend.html", "email_payment_request_amend.txt")
	}
	render.JSON(ctx, "/dashboard", "PaymentRequest successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

// Booking render booking page.
func Booking(ctx *fasthttp.RequestCtx) {
	profileUUID := (ctx.UserValue("profileUUID")).(string)

	var data = struct {
		Profile           *user.Profile
		IsSelf            bool
		IsCustomer        bool
		IsLoggedIn        bool
		UserID            int64
		ParentPage        string
		serviceInputArray []bool
		inputData         map[string]string
	}{}

	profile, err := user.GetProfileByProfileUUID(profileUUID)
	if err != nil {
		log.Printf("Error in routes/booking.go Booking(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go Booking(): Cannot find profile with id '%s': %s\n", profileUUID, err.Error())
		NotFoundRedirect(ctx)
	}

	serviceCategoryID := profile.ServiceCategory
	data.Profile = profile
	data.ParentPage = "booking"
	data.UserID = 0
	data.serviceInputArray = user.GetServiceInputType(serviceCategoryID)
	if ctx.UserValue("user") != nil {
		currUser := getUserFromContext(ctx, true)
		data.IsSelf = profile.User.UserID == currUser.UserID

		if currUser != nil {
			data.IsLoggedIn = true
		} else {
			data.IsLoggedIn = false
		}
	} else {
		data.IsLoggedIn = false
		data.IsSelf = false
	}

	data.IsCustomer = profile.ProfileType == "b"
	session := globalSessions.StartFasthttp(ctx)
	redirectInfo := session.Get("redirect")
	if redirectInfo != nil && string(profile.ProfileID) == redirectInfo.(map[string]string)["profileId"] {
		data.inputData = redirectInfo.(map[string]string)
	}

	pg := &render.Page{Title: "Public Profile", TemplateFileName: "booking.html", Data: data}
	pg.Render(ctx)
}

// BookingResponse accept or cancel booking
func BookingResponse(ctx *fasthttp.RequestCtx) {
	bookingUUID := (ctx.UserValue("bookingUUID")).(string)
	bookingHistoryUUID := (ctx.UserValue("bookingHistoryUUID")).(string)

	currUser := getUserFromContext(ctx, true)

	bookingInfo, _ := user.GetBookingByBookingUUID(bookingUUID)
	bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
	bookingSenderProfile := bookingSenderProfiles[0]

	var data = struct {
		Profile        *user.Profile
		ParentPage     string
		UserID         int64
		Booking        *user.Booking
		BookingHistory *user.BookingHistory
	}{}
	data.Profile = bookingSenderProfile
	data.UserID = currUser.UserID
	data.ParentPage = "Dashboard"
	data.Booking = bookingInfo

	data.BookingHistory, _ = user.GetBookingHistoryByUUID(bookingHistoryUUID)
	pg := &render.Page{Title: "Booking Process", TemplateFileName: "booking-response.html", Data: data}

	pg.Render(ctx)
}

//SendPaymentRequest renders payment request page
func SendPaymentRequest(ctx *fasthttp.RequestCtx) {
	bookingUUID := (ctx.UserValue("bookingUUID")).(string)
	bookingHistoryUUID := (ctx.UserValue("bookingHistoryUUID")).(string)

	currUser := getUserFromContext(ctx, true)

	bookingInfo, _ := user.GetBookingByBookingUUID(bookingUUID)
	bookingSenderProfiles, _ := user.GetProfileByUserID(bookingInfo.User.UserID)
	bookingSenderProfile := bookingSenderProfiles[0]

	var data = struct {
		Profile        *user.Profile
		ParentPage     string
		UserID         int64
		Booking        *user.Booking
		BookingHistory *user.BookingHistory
	}{}
	data.Profile = bookingSenderProfile
	data.UserID = currUser.UserID
	data.ParentPage = "Dashboard"
	data.Booking = bookingInfo

	data.BookingHistory, _ = user.GetBookingHistoryByUUID(bookingHistoryUUID)

	pg := &render.Page{Title: "Payment Request", TemplateFileName: "authenticated/payment-request.html", Data: data}

	pg.Render(ctx)
}

//PendingPaymentMessage sends message in pending payment mode
func PendingPaymentMessage(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)

	message := string(ctx.FormValue("paymentMessage"))

	var paymentRequetMessageForm = &JSONFormError{}
	paymentRequetMessageForm.Form = "pay-request-message"

	if len(message) == 0 {
		paymentRequetMessageForm.Error = "'Message' is required filed."
		render.JSON(ctx, paymentRequetMessageForm, "'Message' is required filed.", fasthttp.StatusInternalServerError)
		return
	}
	paymentRequestIDParam := string(ctx.FormValue("paymentRequestId"))

	paymentRequestID, err := strconv.ParseInt(paymentRequestIDParam, 10, 64)

	if err != nil {
		log.Println("Error in routes/booking.go PendingPaymentMessage() fail to get PaymentRequestID:", err)
		paymentRequetMessageForm.Error = "There has been an internal error getting PaymentRequestID."
		paymentRequetMessageForm.Redirect = internalServerErrorURL
		render.JSON(ctx, paymentRequetMessageForm, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	pr, err := user.GetPayment(paymentRequestID)
	if err != nil {
		log.Println("Error in routes/booking.go PendingPaymentMessage() fail to get Payment:", err)
		paymentRequetMessageForm.Error = "There has been an internal error getting Payment."
		paymentRequetMessageForm.Redirect = internalServerErrorURL
		render.JSON(ctx, paymentRequetMessageForm, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingInfo, err := user.GetBookingWithBookingId(pr.Booking.BookingID)
	if err != nil {
		log.Println("Error in routes/booking.go PendingPaymentMessage() fail to get Booking Data:", err)
		paymentRequetMessageForm.Error = "There has been an internal error getting Booking Data."
		paymentRequetMessageForm.Redirect = internalServerErrorURL
		render.JSON(ctx, paymentRequetMessageForm, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingHistories, _ := user.GetBookingHistoriesWithBookingID(bookingInfo.BookingID)
	lastBookingHistory := bookingHistories[0]
	bookingHistory := &user.BookingHistory{}
	bookingHistory.BookingID = lastBookingHistory.BookingID
	bookingHistory.UserID = currUser.UserID
	bookingHistory.Message = message
	bookingHistory.FromDate = lastBookingHistory.FromDate
	bookingHistory.ToDate = lastBookingHistory.ToDate
	bookingHistory.FromTime = lastBookingHistory.FromTime
	bookingHistory.ToTime = lastBookingHistory.ToTime
	bookingHistory.Address = lastBookingHistory.Address
	bookingHistory.Fee = lastBookingHistory.Fee
	bookingHistory.TotalPrice = lastBookingHistory.TotalPrice
	bookingHistory.BookingStatus = lastBookingHistory.BookingStatus
	bookingHistory.FrequencyUnit = lastBookingHistory.FrequencyUnit
	bookingHistory.FrequencyValue = lastBookingHistory.FrequencyValue

	bookingHistoryID, err := user.InsertBookingHistory(bookingHistory)
	if err != nil {
		log.Println("Error in routes/booking.go PendingPaymentMessage() fail to insert Booking History Data:", err)
		paymentRequetMessageForm.Error = "There has been an internal error getting Booking History Data."
		paymentRequetMessageForm.Redirect = internalServerErrorURL
		render.JSON(ctx, paymentRequetMessageForm, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	var receiverID int64
	if currUser.UserID == bookingInfo.User.UserID {
		receiverID = bookingInfo.Profile.User.UserID
	} else if currUser.UserID == bookingInfo.Profile.User.UserID {
		receiverID = bookingInfo.User.UserID
	}
	notificationItem := &notification.Notification{
		EntityID:         paymentRequestID,
		NotificationType: notification.NotificationTypePaymentRequestMessage,
		SenderID:         currUser.UserID,
		ReceiverID:       receiverID,
		EntityHistoryID:  bookingHistoryID,
		Unread:           []int64{receiverID},
	}

	err = notification.InsertNotification(notificationItem)
	if err != nil {
		log.Println("Error in /routes/booking.go PaymentRequest(): Failed to insert payment request notificaiton", err)
		paymentRequetMessageForm.Error = "There has been an internal error update payment request data."
		paymentRequetMessageForm.Redirect = internalServerErrorURL
		render.JSON(ctx, paymentRequetMessageForm, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	render.JSON(ctx, "/dashboard", "PaymentRequest successfully update, redirecting to dashboard page", fasthttp.StatusOK)
}

// QuickPaymentRequest sends payment request without booking
func QuickPaymentRequest(ctx *fasthttp.RequestCtx) {
	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)

	} else {
		log.Println("user did not log in")
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}

	customerProfileIDParam := string(ctx.FormValue("customerProfileID"))
	customerProfileID, err := strconv.ParseInt(customerProfileIDParam, 10, 64)
	providerProfileIDParam := string(ctx.FormValue("providerProfileID"))
	providerProfileID, err := strconv.ParseInt(providerProfileIDParam, 10, 64)
	amountParam := string(ctx.FormValue("amount"))
	amountValue, err := strconv.ParseInt(amountParam, 10, 64)
	var amount int32
	if err != nil {
		log.Println("Error routes/booking.go QuickPaymentRequest() Failed to convert amount :", err)
		return
	}
	amount = int32(amountValue)
	log.Println("amountParam:", amountParam, "amount:", amount, "amountValue:", amountValue)
	message := string(ctx.FormValue("message"))

	var quickPaymentRequestFormError = &JSONFormError{}
	quickPaymentRequestFormError.Form = "quick-payment-request-form"

	if err != nil && len(amountParam) != 0 {
		log.Println("Error routes/booking.go QuickPaymentRequest() failed to convert amount value to int:", err.Error())
		quickPaymentRequestFormError.Error = "Amount should be integer"
		render.JSON(ctx, quickPaymentRequestFormError, "Amount should be integer", fasthttp.StatusUnprocessableEntity)
		return
	}

	booking := &user.Booking{}

	customerProfile, err := user.GetProfile(customerProfileID)
	if err != nil {
		log.Println("Error routes/booking.go QuickPaymentRequest() failed to get Profile:", err)
		quickPaymentRequestFormError.Error = "Failed to get Profile."
		render.JSON(ctx, quickPaymentRequestFormError, "Failed to get Profile.", fasthttp.StatusUnprocessableEntity)
		return
	}
	booking.User = customerProfile.User
	bookingProfile, err := user.GetProfile(providerProfileID)
	booking.Profile = *bookingProfile
	bookingID, err := user.InsertBooking(booking)
	if err != nil {
		log.Printf("Error in app/routes/booking.go QuickPaymentRequest() InsertBooking Failed  %s\n", err.Error())
		quickPaymentRequestFormError.Error = "There has been an internal error inserting Booking Data."
		quickPaymentRequestFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, quickPaymentRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	bookingHistory := &user.BookingHistory{}
	bookingHistory.UserID = currUser.UserID
	bookingHistory.BookingID = bookingID
	bookingHistory.BookingStatus = "Accepted"
	bookingHistory.Message = message
	bookingHistory.TotalPrice, err = decimal.NewFromString(amountParam)
	bookingHistoryID, err := user.InsertBookingHistory(bookingHistory)
	if err != nil {
		log.Printf("Error in app/routes/booking.go QuickPaymentRequest() Insert BookingHistory Failed  %s\n", err.Error())
		quickPaymentRequestFormError.Error = "There has been an internal error inserting Booking Data."
		quickPaymentRequestFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, quickPaymentRequestFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	bookingInfo, _ := user.GetBookingWithBookingId(bookingID)

	pr := &user.Payment{Booking: *bookingInfo, Amount: amount}

	id, err := user.InsertPayment(pr)
	if err != nil {
		log.Println("Error in routes/booking.go QuickPaymentRequest(ctx *fasthttp.RequestCtx) Failed to insert payment request ", err)
		return
	}

	notificationItem := &notification.Notification{
		EntityID:         id,
		NotificationType: notification.NotificationTypePaymentRequest,
		SenderID:         currUser.UserID,
		ReceiverID:       bookingInfo.User.UserID,
		EntityHistoryID:  bookingHistoryID,
		Unread:           []int64{bookingInfo.User.UserID},
	}

	err = notification.InsertNotification(notificationItem)
	if err != nil {
		log.Println("Error in /routes/booking.go PaymentRequest(): Failed to insert payment request notificaiton", err)
	}

	render.JSON(ctx, "/dashboard", "PaymentRequest successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

// QuickPaymentProcess process payment request without booking
func QuickPaymentProcess(ctx *fasthttp.RequestCtx) {

	var currUser *user.User

	if ctx.UserValue("user") != nil {
		currUser = getUserFromContext(ctx, true)

	} else {
		log.Println("user did not log in")
		render.JSON(ctx, "/login-register", "You should login first.", fasthttp.StatusOK)
		return
	}

	profileIDParam := string(ctx.FormValue("profileId"))
	profileID, err := strconv.ParseInt(profileIDParam, 10, 64)
	userIDParam := string(ctx.FormValue("userId"))
	_, err = strconv.ParseInt(userIDParam, 10, 64)
	message := string(ctx.FormValue("message"))
	amountParam := string(ctx.FormValue("amount"))
	amount, err := strconv.ParseInt(amountParam, 10, 64)
	paymentType := string(ctx.FormValue("paymentType"))

	var quickPaymentFormError = &JSONFormError{}
	quickPaymentFormError.Form = "quick-payment-form"

	if err != nil && len(amountParam) != 0 {
		log.Println("Error routes/booking.go QuickPaymentProcess() failed to convert amount value to int:", err.Error())
		quickPaymentFormError.Error = "Amount should be integer"
		render.JSON(ctx, quickPaymentFormError, "Amoun should be integer", fasthttp.StatusUnprocessableEntity)
		return
	}

	booking := &user.Booking{}
	booking.User = *currUser
	profile, err := user.GetProfile(profileID)
	booking.Profile = *profile

	bookingID, err := user.InsertBooking(booking)

	if err != nil {
		log.Printf("Error in app/routes/booking.go QuickPayment() InsertBooking Failed  %s\n", err.Error())
		quickPaymentFormError.Error = "There has been an internal error inserting Booking Data."
		quickPaymentFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, quickPaymentFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingHistory := &user.BookingHistory{}
	bookingHistory.BookingID = bookingID
	bookingHistory.BookingStatus = "Accepted"
	bookingHistory.Message = message

	_, err = user.InsertBookingHistory(bookingHistory)
	if err != nil {
		log.Printf("Error in app/routes/booking.go QuickPayment() Insert BookingHistory Failed  %s\n", err.Error())
		quickPaymentFormError.Error = "There has been an internal error inserting Booking Data."
		quickPaymentFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, quickPaymentFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	bookingInfo, _ := user.GetBookingWithBookingId(bookingID)

	pr := &user.Payment{Booking: *bookingInfo, Amount: int32(amount)}

	id, err := user.InsertPayment(pr)
	if err != nil {
		log.Println("Error in routes/booking.go PaymentRequest(ctx *fasthttp.RequestCtx) Failed to insert payment request ", err)
	}

	ctx.SetUserValue("quickPayment", "true")
	ctx.SetUserValue("paymentRequestId", id)
	ctx.SetUserValue("paymentType", paymentType)

	PayPaymentRequest(ctx)

}

//AddTip adds tip to payment request table.
func AddTip(ctx *fasthttp.RequestCtx) {
	tipParam := string(ctx.FormValue("tip"))
	var tipValue float64

	var addTipFormError = &JSONFormError{}
	addTipFormError.Form = "add-tip-form"

	tipValue, err := strconv.ParseFloat(tipParam, 64)
	if err != nil {
		log.Println("Error routes/booking.go AddTip() fail to convert tip to float:", err.Error())
		addTipFormError.Error = "Tip should be number"
		render.JSON(ctx, addTipFormError, "Tip should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	paymentRequestIDParam := string(ctx.FormValue("paymentRequestId"))
	paymnetRequestID, err := strconv.ParseInt(paymentRequestIDParam, 10, 64)
	if err != nil {
		log.Println("Error routes/booking.go AddTip() fail to convert tip to float:", err.Error())
		addTipFormError.Error = "Tip should be number"
		render.JSON(ctx, addTipFormError, "Tip should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	pr, err := user.GetPayment(paymnetRequestID)
	if err != nil {
		log.Println("Error routes/booking.go AddTip() fail to convert tip to float", err.Error())
		addTipFormError.Error = "Tip should be number"
		render.JSON(ctx, addTipFormError, "Tip should be number", fasthttp.StatusUnprocessableEntity)
		return
	}

	if len(tipParam) != 0 {
		pr.Tip = decimal.NewFromFloat(tipValue)
	}

	err = pr.Update()
	if err != nil {
		log.Println("Error routes/booking.go AddTip() fail to convert tip to float:", err.Error())
		addTipFormError.Error = "Tip should be number"
		render.JSON(ctx, addTipFormError, "Tip should be number", fasthttp.StatusUnprocessableEntity)
		return
	}
	render.JSON(ctx, "/dashboard", "Booking successfully saved, redirecting to dashboard page", fasthttp.StatusOK)
}

func GetPaymentRequest(ctx *fasthttp.RequestCtx) {
	paymentRequestIDParam := (ctx.UserValue("paymentRequestID")).(string)
	paymentRequestID, err := strconv.ParseInt(paymentRequestIDParam, 10, 64)

	if err != nil {
		log.Printf("Error in routes/booking.go PaymentRequest(ctx *fasthttp.RequestCtx): Payment Request ID Invalid: %s\n", paymentRequestIDParam)
		NotFoundRedirect(ctx)
	}

	isAjax := bytes.EqualFold(ctx.Request.Header.PeekBytes([]byte("X-Requested-With")), []byte("XMLHttpRequest"))

	var data = struct {
		PaymentRequest *user.Payment
	}{}

	request, err := user.GetPayment(paymentRequestID)
	if err != nil {
		log.Printf("Error in routes/booking.go PaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): Cannot find payment request with id '%d': %s\n", paymentRequestID, err.Error())
		NotFoundRedirect(ctx)
	}

	data.PaymentRequest = request

	if isAjax {
		jsonb, _ := json.Marshal(data)
		json := string(jsonb)
		render.JSON(ctx, json, "ok", fasthttp.StatusOK)
	} else {
		pg := &render.Page{Title: "Payment Request", TemplateFileName: "payment.html", Data: data}
		pg.Render(ctx)
	}

}

//PayPaymentRequest make a credit card against the payment request
func PayPaymentRequest(ctx *fasthttp.RequestCtx) {

	data := struct {
		Success     bool
		Error       string
		RedirectURL string
	}{
		Success:     false,
		Error:       "",
		RedirectURL: "",
	}
	var paymentRequestIDParam string
	var paymentRequestID int64
	var err error
	quickPay := ctx.UserValue("quickPayment")

	var paymentType string
	if quickPay == nil {

		paymentRequestIDParam = string(ctx.FormValue("id"))
		paymentType = string(ctx.FormValue("type"))
		paymentRequestID, err = strconv.ParseInt(paymentRequestIDParam, 10, 64)
		if err != nil {
			log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Payment Request ID Invalid: %s\n", paymentRequestIDParam)
			render.JSON(ctx, data, "Error", fasthttp.StatusNotFound)
			return
		}
	} else if quickPay.(string) == "true" {
		log.Println("pymentRequestID:", (ctx.UserValue("paymentRequestId")).(int64))
		log.Println("paymentType:", (ctx.UserValue("paymentType")).(string))
		log.Println("Payment method is quickpay")
		paymentRequestID = (ctx.UserValue("paymentRequestId")).(int64)
		paymentType = (ctx.UserValue("paymentType")).(string)
	}

	//_, err = user.GetPaymentRequest(paymentRequestID)
	payRequest, err := user.GetPayment(paymentRequestID)
	if err != nil {
		log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): Cannot find payment request with id '%d': %s\n", paymentRequestID, err.Error())
		render.JSON(ctx, data, "Error", fasthttp.StatusNotFound)
		return
	}

	if paymentType == user.PaymentMethodCreditCard || paymentType == user.PaymentMethodPaypal {
		// set status to processing so we know if something has gone wrong if the update fails after the payment
		payRequest.PaymentMethod = paymentType
		payRequest.Status = user.PaymentRequestStatusProcessing
		err = payRequest.Update()
		if err != nil {
			log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): error updating payment request to status Processing: %s\n", err.Error())
			data.Error = "INTERNAL_ERROR"
			render.JSON(ctx, data, "Error", fasthttp.StatusBadRequest)
			return
		}
	}

	if paymentType == user.PaymentMethodCreditCard {

		err = payments.MakeCreditCardPayment(payRequest)
		if err != nil {
			// reset back to new status
			resetPayment(payRequest)

			errorResponse := err.(*payments.CreditCardPaymentError)
			data.Error = errorResponse.ErrorCode
			log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): payment failed: %s\n", err.Error())
			render.JSON(ctx, data, "Error", fasthttp.StatusBadRequest)
			return
		}
	} else if paymentType == user.PaymentMethodPaypal {

		amountDec := decimal.NewFromFloat(float64(payRequest.Amount))
		feeRate, _ := decimal.NewFromString(".05")

		urAmountDec := amountDec.Mul(feeRate)
		fmt.Println("DEBUG: UR Amount (5% fee): ", urAmountDec)
		userAmountDec := amountDec.Sub(urAmountDec)
		fmt.Println("DEBUG: User Amount (less 5% fee): ", userAmountDec)

		urAmount, _ := amountDec.Float64()
		userAmount, _ := userAmountDec.Float64()

		urEmail := appConfig.PaypalEmail
		userEmail := payRequest.Booking.Profile.User.GetPayoutAccount()

		userReceiver := &payments.APReceiver{
			Amount:  userAmount,
			Email:   &userEmail,
			Primary: false,
		}
		urReceiver := &payments.APReceiver{
			Amount:  urAmount,
			Email:   &urEmail,
			Primary: true,
		}

		receivers := []*payments.APReceiver{urReceiver, userReceiver}

		receiverList := &payments.APReceiverList{
			Receiver: receivers,
		}

		ipnNotificationURL := fmt.Sprintf("%s/update-paypal-payment?id=%d", config.FullBaseURL(), payRequest.PaymentID)
		paypalRequest := &payments.APPaymentRequest{
			ActionType:         payments.AdaptivePaymentsActionTypePay,
			CancelURL:          fmt.Sprintf("%s/cancel-paypal-payment?id=%d", config.FullBaseURL(), payRequest.PaymentID),
			ReturnURL:          fmt.Sprintf("%s/process-paypal-payment?id=%d", config.FullBaseURL(), payRequest.PaymentID),
			IPNNotificationURL: &ipnNotificationURL,
			CurrencyCode:       "AUD",
			ReceiverList:       receiverList,
		}

		response, err := payments.AdaptivePaymentsPay(paypalRequest)
		if err != nil {
			// reset back to new status
			resetPayment(payRequest)

			//errorResponse := err.(*payments.CreditCardPaymentError)
			//data.Error = errorResponse.ErrorCode
			data.Error = "PAYPAL_FAILED"
			log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): payment with adaptive payments failed: %s\n", err.Error())
			render.JSON(ctx, data, "Error", fasthttp.StatusBadRequest)
			return
		}
		redirectURL := fmt.Sprintf("https://%s?cmd=_ap-payment&paykey=%s", appConfig.PayPalAPI.RedirectURL, response.PayKey)
		data.RedirectURL = redirectURL

		payRequest.PaypalToken = response.PayKey
		err = payRequest.Update()
		if err != nil {
			// not sure what should do here, as the payment has returned successfully, but pay request not updated. It shouldn't kill everything.
			log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): error updating payment request with paypal paykey: %s\n", err.Error())
		}

		// Payment successful
		data.Success = true

		render.JSON(ctx, data, "ok", fasthttp.StatusOK)
		return
	} else {
		// don't have any other payment type for the moment
		data.Error = "INVALID_PAYMENT_TYPE"
		log.Printf("Error in routes/booking.go PayPaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): invalid payment type, only CC available at the moment \n")
		render.JSON(ctx, data, "Error", fasthttp.StatusBadRequest)
		return
	}

	// Payment successful
	data.Success = true
	log.Println("Complate Payment Process call 1")
	// update payment request with data set in MakePayment methods and send payment notification
	completePayment(payRequest)

	render.JSON(ctx, data, "ok", fasthttp.StatusOK)
}

//DeclinePaymentRequest decline a payment request with or without reason
func DeclinePaymentRequest(ctx *fasthttp.RequestCtx) {
	data := struct {
		Success bool
		Error   string
	}{
		Success: false,
		Error:   "",
	}

	paymentRequestIDParam := string(ctx.FormValue("id"))
	paymentRequestID, err := strconv.ParseInt(paymentRequestIDParam, 10, 64)

	if err != nil {
		log.Printf("Error in routes/booking.go DeclinePaymentRequest(ctx *fasthttp.RequestCtx): Payment Request ID Invalid: %s\n", paymentRequestIDParam)
		render.JSON(ctx, data, "Error", fasthttp.StatusNotFound)
		return
	}

	reason := string(ctx.FormValue("reason"))

	payRequest, err := user.GetPayment(paymentRequestID)
	if err != nil {
		log.Printf("Error in routes/booking.go DeclinePaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): Cannot find payment request with id '%d': %s\n", paymentRequestID, err.Error())
		render.JSON(ctx, data, "Error", fasthttp.StatusNotFound)
		return
	}

	payRequest.Status = user.PaymentRequestStatusDeclined
	payRequest.ConfirmedDate = time.Now().UTC()
	payRequest.DeclinedReason = reason

	err = payRequest.Update()
	if err != nil {
		log.Printf("Error in routes/booking.go DeclinePaymentRequest(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): error updating payment request to be declined: %s\n", err.Error())
		data.Error = "INTERNAL_ERROR"
		render.JSON(ctx, data, "Error", fasthttp.StatusInternalServerError)
		return
	}

	notificationItem := &notification.Notification{
		EntityID:         payRequest.PaymentID,
		NotificationType: notification.NotificationTypeDeclined,
		SenderID:         payRequest.Booking.User.UserID,
		ReceiverID:       payRequest.Booking.Profile.User.UserID,
		NotificationText: "Payment declined for booking " + string(payRequest.Booking.BookingID),
		Unread:           []int64{payRequest.Booking.Profile.User.UserID},
	}

	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go DeclinePaymentRequest(ctx *fasthttp.RequestCtx): InsertNotification Failed %s\n", err.Error())
	}

	//defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem)
	render.JSON(ctx, data, "ok", fasthttp.StatusOK)
}

//ProcessPaypalPayment the return URL after user is redirected to payal
func ProcessPaypalPayment(ctx *fasthttp.RequestCtx) {

	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/booking.go ProcessPaypalPayment(ctx *fasthttp.RequestCtx): parsing payment request id '%s': %s\n", idString, err.Error())
	} else {

		request, err := user.GetPayment(ID)
		if err != nil {
			log.Printf("Error in routes/booking.go ProcessPaypalPayment(ctx *fasthttp.RequestCtx): Cannot find payment request with id '%d': %s\n", ID, err.Error())

		}
		log.Println("Complate Payment Process call 2")
		completePayment(request)
	}

	ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
}

//CancelPaypalPayment the cancel URL after user is redirected to payal
func CancelPaypalPayment(ctx *fasthttp.RequestCtx) {
	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/booking.go CancelPaypalPayment(ctx *fasthttp.RequestCtx): parsing payment request id '%s': %s\n", idString, err.Error())
	} else {

		request, err := user.GetPayment(ID)
		if err != nil {
			log.Printf("Error in routes/booking.go CancelPaypalPayment(ctx *fasthttp.RequestCtx): Cannot find payment request with id '%d': %s\n", ID, err.Error())

		} else {
			request.Status = user.PaymentRequestStatusNew
			request.PaypalToken = ""

			err = request.Update()
			if err != nil {
				log.Printf("Error in routes/booking.go CancelPaypalPayment(ctx *fasthttp.RequestCtx): Error in app/routes/booking.go PaymentRequest(): error updating payment request after paypal cancelled: %s\n", err.Error())
			}
		}
	}

	ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
}

//RefreshPaypalPayment check status of payment and update details
func RefreshPaypalPayment(ctx *fasthttp.RequestCtx) {
	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/booking.go RefreshPaypalPayment(ctx *fasthttp.RequestCtx): parsing payment request id '%s': %s\n", idString, err.Error())
	} else {

		request, err := user.GetPayment(ID)
		if err != nil {
			log.Printf("Error in routes/booking.go RefreshaypalPayment(ctx *fasthttp.RequestCtx): Cannot find payment request with id '%d': %s\n", ID, err.Error())

		}

		// first check to see if the payment has been updated by IPN while the user was making this request
		// make sure still is PROCESSING otherwise fall through and refresh
		if request.Status == user.PaymentRequestStatusProcessing {
			detailsRequest := &payments.APPaymentDetailsRequest{
				PayKey: request.PaypalToken,
			}

			response, err := payments.AdaptivePaymentsGetPaymentDetails(detailsRequest)
			if err != nil {
				log.Printf("Error in routes/booking.go RefreshaypalPayment(ctx *fasthttp.RequestCtx): Error contacting Paypal for details: %s\n", err.Error())

			} else {
				log.Printf("Response received: %s", response.ResponseEnvelope.Ack)
				if response.Status == "COMPLETED" {
					log.Println("Complate Payment Process call 3")
					completePayment(request)
				} else if response.Status == "ERROR" || response.Status == "EXPIRED" {
					resetPayment(request)
				}
				// otherwise do nothing as we aren't clear what is going on yet, it is processing
				// CREATED means they should click 'Continue' and others mean it is still processing
			}
		}
	}

	ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
}

//UpdatePaypalPayment the IPN notification URL after user is redirected to payal
func UpdatePaypalPayment(ctx *fasthttp.RequestCtx) {

	paypalURL := "https://" + appConfig.PayPalAPI.RedirectURL

	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/booking.go UpdatePaypalPayment(ctx *fasthttp.RequestCtx): parsing payment request id '%s': %s\n", idString, err.Error())
	}

	fmt.Printf("Payment request to update: %d", ID)

	// IPN HANDSHAKE BEGIN

	// *********************************************************
	// HANDSHAKE STEP 1 -- Write back an empty HTTP 200 response
	// *********************************************************
	fmt.Printf("Write Status 200")
	ctx.SetStatusCode(200)

	// *********************************************************
	// HANDSHAKE STEP 2 -- Send POST data (IPN message) back as verification
	// *********************************************************
	// Get Content-Type of request to be parroted back to paypal
	contentType := ctx.Request.Header.ContentType()
	// Read the raw POST body
	body := ctx.PostBody()
	//body, _ := ioutil.ReadAll(ctx.PostBody())
	// Prepend POST body with required field
	body = append([]byte("cmd=_notify-validate&"), body...)
	// Make POST request to paypal
	resp, _ := http.Post(paypalURL, string(contentType), bytes.NewBuffer(body))

	// *********************************************************
	// HANDSHAKE STEP 3 -- Read response for VERIFIED or INVALID
	// *********************************************************
	verifyStatus, _ := ioutil.ReadAll(resp.Body)

	// *********************************************************
	// Test for VERIFIED
	// *********************************************************
	if string(verifyStatus) != "VERIFIED" {
		log.Printf("Response: %v", string(verifyStatus))
		log.Println("This indicates that an attempt was made to spoof this interface, or we have a bug.")
		return
	}
	// We can now assume that the POSTed information in `body` is VERIFIED to be from Paypal.
	log.Printf("Response: %v", string(verifyStatus))

	values, _ := url.ParseQuery(string(body))
	for i, v := range values {
		fmt.Println(i, v)
	}
}

func resetPayment(payRequest *user.Payment) {
	// set common paid fields
	payRequest.Status = user.PaymentRequestStatusNew

	// update payment request
	err := payRequest.Update()
	if err != nil {
		// need to think about best way to handle this
		log.Printf("Error in routes/booking.go resetPayment(payRequest *user.Payment): error resetting payment request: %s\n", err.Error())
	}
}

func completePayment(payRequest *user.Payment) {
	currDate := time.Now().UTC()

	// set common paid fields
	payRequest.Status = user.PaymentRequestStatusPaid
	payRequest.PaymentDate = currDate
	payRequest.ConfirmedDate = currDate

	// update payment request
	err := payRequest.Update()
	if err != nil {
		// need to think about best way to handle this, the payment was successful, but the request wasn't updated properly
		log.Printf("Error in routes/booking.go completePayment(payRequest *user.Payment): error updating payment request: %s\n", err.Error())
	}

	bookingHistories, err := user.GetBookingHistoriesWithBookingID(payRequest.Booking.BookingID)
	if err != nil {
		log.Println("Error in routes/booking.go completePayment():", err)
	}
	log.Println("***Payment Paid Finish Notification Checking***")
	// payment notification
	notificationItem := &notification.Notification{
		EntityID:         payRequest.PaymentID,
		NotificationType: notification.NotificationTypePaid,
		SenderID:         payRequest.Booking.User.UserID,
		ReceiverID:       payRequest.Booking.Profile.User.UserID,
		NotificationText: "Payment made for booking " + string(payRequest.Booking.BookingID),
		EntityHistoryID:  bookingHistories[0].BookingHistoryID,
		Unread:           []int64{payRequest.Booking.Profile.User.UserID},
	}

	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go completePayment(payRequest *user.Payment): InsertNotification for payment paid notification Failed %s\n", err.Error())
	}
	bookingIDstr := fmt.Sprintf("%v", payRequest.Booking.BookingID)
	defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local – Payment Received for Booking (ID "+bookingIDstr+")", "email_payment_received.html", "email_payment_received.txt")

	// feedback notification to provider to leave feedback
	notificationItem = &notification.Notification{
		EntityID:         payRequest.Booking.BookingID,
		NotificationType: notification.NotificationTypeLeaveFeedback,
		SenderID:         payRequest.Booking.User.UserID,
		ReceiverID:       payRequest.Booking.Profile.User.UserID,
		EntityHistoryID:  bookingHistories[0].BookingHistoryID,
		Unread:           []int64{payRequest.Booking.Profile.User.UserID},
	}

	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go completePayment(payRequest *user.Payment): InsertNotification for feedback to service provider Failed %s\n", err.Error())
	}
	//send email to service provider to leave feedback.
	defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local - Please Leave Feedback (ID "+bookingIDstr+")", "email_leave_feedback.html", "email_leave_feedback.txt")

	// feedback notification to customer to leave feedback
	notificationItem = &notification.Notification{
		EntityID:         payRequest.Booking.BookingID,
		NotificationType: notification.NotificationTypeLeaveFeedback,
		SenderID:         payRequest.Booking.Profile.User.UserID,
		ReceiverID:       payRequest.Booking.User.UserID,
		EntityHistoryID:  bookingHistories[0].BookingHistoryID,
		Unread:           []int64{payRequest.Booking.User.UserID},
	}

	err = notification.InsertNotification(notificationItem)

	if err != nil {
		log.Printf("Error in app/routes/booking.go completePayment(payRequest *user.Payment): InsertNotification for feedback to customer Failed %s\n", err.Error())
	}

	//send email to customer to leave feedback
	defer emailing.SendNewNotificationEmail(notificationItem.Receiver(), notificationItem.Sender(), notificationItem, "UR Local - Please Leave Feedback (ID "+bookingIDstr+")", "email_leave_feedback.html", "email_leave_feedback.txt")

}
