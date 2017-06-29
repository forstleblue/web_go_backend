package notification

import (
	"github.com/unirep/ur-local-web/app/models/user"
)

const (
	//NotificationTypeBookingRequest constant for notification type of booking request, when user send booking request
	NotificationTypeBookingRequest = "BookingRequest"
	//NotificationTypeBookingResponse constant for nofication type booking response , when user accept or reject Booking request
	NotificationTypeBookingResponse = "BookingResponse"
	//NotificationTypeRating constant for notification type of rating
	NotificationTypeRating = "Rating"
	//NotificationTypePaymentRequest constant for notification type of a payment request
	NotificationTypePaymentRequest = "PaymentRequest"
	//NotificationTypePaymentRequestResponse const for notification type of payment request response
	NotificationTypePaymentRequestResponse = "PaymentRequestResponse"
	//NotificationTypePaymentRequestMessage constant for notification type a payment request message
	NotificationTypePaymentRequestMessage = "PaymentRequestMessage"
	//NotificationTypePaid notification sent to a user when a payment request has been paid
	NotificationTypePaid = "PaymentRequestPaid"
	//NotificationTypeDeclined notification sent to user if their payment request was declined
	NotificationTypeDeclined = "PaymentRequestDeclined"
	//NotificationTypeFeedbackDueSeven sent to customer or provier if they did not leave feedback for a week.
	NotificationTypeFeedbackDueSeven = "FeedbackDueSeven"
	//NotificationTypeFeedbackDueTwo sent to customer or provier to notify them only have 2 days to leave feedback.
	NotificationTypeFeedbackDueTwo = "FeedbackDueTwo"
	//NotificationTypeLeaveFeedback notification sent to users to request feedback after booking completion
	NotificationTypeLeaveFeedback = "LeaveFeedback"
	//NotificationTypeFeedbackReceived notifictaion sent to users to write feedback
	NotificationTypeFeedbackReceived = "FeedbackReceived"
)

//Notification represents notification data
type Notification struct {
	NotificationID   int64   `json:"notification_id"`
	NotificationType string  `json:"notification_type"`
	SenderID         int64   `json:"sender_id"`
	ReceiverID       int64   `json:"receiver_id"`
	EntityID         int64   `json:"entity_id"`
	EntityHistoryID  int64   `json:"entity_history_id"`
	NotificationText string  `json:"notification_string"`
	Unread           []int64 `json:"unread"`
	NotificationUUID string  `json:"notification_uuid"`
}

//Sender returns notification sender name
func (n Notification) Sender() *user.User {
	user, _ := user.GetUser(n.SenderID)
	return user
}

//Receiver returns notification receiver name
func (n Notification) Receiver() *user.User {
	user, _ := user.GetUser(n.ReceiverID)
	return user
}

//Booking returns Booking struct
func (n Notification) Booking() *user.Booking {
	var booking *user.Booking
	if n.NotificationType == NotificationTypeBookingRequest ||
		n.NotificationType == NotificationTypeBookingResponse ||
		n.NotificationType == NotificationTypeLeaveFeedback ||
		n.NotificationType == NotificationTypeFeedbackReceived {
		booking, _ = user.GetBookingWithBookingId(n.EntityID)
	} else if n.NotificationType == NotificationTypePaymentRequest ||
		n.NotificationType == NotificationTypePaymentRequestResponse ||
		n.NotificationType == NotificationTypePaymentRequestMessage ||
		n.NotificationType == NotificationTypePaid {
		pr, _ := user.GetPayment(n.EntityID)
		booking, _ = user.GetBookingWithBookingId(pr.Booking.BookingID)
	}
	return booking
}

//BookingHistory gets BookingHistory
func (n Notification) BookingHistory() *user.BookingHistory {
	var bookingHistory *user.BookingHistory
	if n.NotificationType == NotificationTypeBookingRequest ||
		n.NotificationType == NotificationTypeBookingResponse ||
		n.NotificationType == NotificationTypePaymentRequest ||
		n.NotificationType == NotificationTypePaymentRequestMessage ||
		n.NotificationType == NotificationTypePaymentRequestResponse ||
		n.NotificationType == NotificationTypeFeedbackReceived ||
		n.NotificationType == NotificationTypePaid ||
		n.NotificationType == NotificationTypeLeaveFeedback {
		bookingHistory, _ = user.GetBookingHistoryWithHistoryID(n.EntityHistoryID)
	}

	return bookingHistory
}

//Payment gets Payment
func (n Notification) Payment() *user.Payment {
	var pr *user.Payment
	if n.NotificationType == NotificationTypePaymentRequest ||
		n.NotificationType == NotificationTypePaid ||
		n.NotificationType == NotificationTypeDeclined ||
		n.NotificationType == NotificationTypePaymentRequestMessage ||
		n.NotificationType == NotificationTypePaymentRequestResponse {
		pr, _ = user.GetPayment(n.EntityID)
	} else if n.NotificationType == NotificationTypeLeaveFeedback {
		pr, _ = user.GetPaymentByBookingID(n.EntityID)
	}
	return pr
}

//PaymentHistory returns payment request history
func (n Notification) PaymentHistory() *user.PaymentHistory {
	var prHistory *user.PaymentHistory

	if n.NotificationType == NotificationTypePaymentRequest ||
		n.NotificationType == NotificationTypePaymentRequestMessage ||
		n.NotificationType == NotificationTypePaymentRequestResponse {
		prHistory, _ = user.GetPaymentHistory(n.EntityHistoryID)
	}
	return prHistory
}
