package user

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	// PaymentRequestStatusNew newly created request
	PaymentRequestStatusNew = "NEW"
	// PaymentRequestStatusProcessing set before processing payment
	PaymentRequestStatusProcessing = "PROCESSING"
	// PaymentRequestStatusAmend amended payment response
	PaymentRequestStatusAmend = "AMEND"
	// PaymentRequestStatusPaid paid request
	PaymentRequestStatusPaid = "PAID"
	// PaymentRequestStatusDeclined declined the payment request
	PaymentRequestStatusDeclined = "DECLINED"

	// PaymentMethodCreditCard option for Credit card
	PaymentMethodCreditCard = "CC"
	// PaymentMethodPaypal option for Paypal
	PaymentMethodPaypal = "PAYPAL"

	//PaymentStatusAuthorised payment status of authorised
	PaymentStatusAuthorised = "AUTHORISED"
	//PaymentStatusCaptured payment status of captured
	PaymentStatusCaptured = "CAPTURED"
	//PaymentStatusApproved payment status of approved
	PaymentStatusApproved = "APPROVED"
	//PaymentStatusPending payment status of pending
	PaymentStatusPending = "PENDING"
	//PaymentStatusCleared payment status of cleared
	PaymentStatusCleared = "CLEARED"
)

//Payment is a request for payment for a booking
type Payment struct {
	PaymentID         int64           `json:"payment_request_id"` //todo make this a type 1 GUID from the gocql lib
	Booking           Booking         `json:"booking"`
	Amount            int32           `json:"amount"`
	RequestDate       time.Time       `json:"request_date"`   // the date the provider sends the request
	ConfirmedDate     time.Time       `json:"confirmed_date"` // the date the request is either accepted or denied
	PaymentDate       time.Time       `json:"payment_date"`   // the date the request is paid
	PaymentMethod     string          `json:"payment_method"`
	TransactionID     string          `json:"transaction_id"`
	PaymentStatus     string          `json:"payment_status"`
	AcctDisplay       string          `json:"acct_display"` // either paypal account or masked credit card
	PaypalToken       string          `json:"paypal_token"`
	PaypalPayerID     string          `json:"paypal_payer_id"`
	PaypalPayerStatus string          `json:"paypal_payer_status"`
	Status            string          `json:"status"`
	DeclinedReason    string          `json:"declined_reason"`
	Message           []MessageData   `json:"message_data"`
	Tip               decimal.Decimal `json:"tip"`
}

//MessageData is used when cutstomer and provider send message in payment_mode
type MessageData struct {
	Sender  string    `json:"sender"`
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

//RequestDateFormatted returns the date that should show as since joined, formatted
func (pr *Payment) RequestDateFormatted() string {
	return pr.RequestDate.Format("2, Jan 2006")
}

//ConfirmedDateFormatted returns the date that should show as since joined, formatted
func (pr *Payment) ConfirmedDateFormatted() string {
	return pr.ConfirmedDate.Format("2, Jan 2006")
}

//PaymentDateFormatted returns the date that should show as since joined, formatted
func (pr *Payment) PaymentDateFormatted() string {
	return pr.PaymentDate.Format("2, Jan 2006")
}
