package user

import (
	"log"
	"time"
)

//PaymentHistory represents payment_request history
type PaymentHistory struct {
	PaymentHistoryID int64     `json:"payment_history_id"`
	PaymentID        int64     `json:"payment_id"`
	Amount           int32     `json:"amount"`
	Status           string    `json:"status"`
	Message          string    `json:"message"`
	UserID           int64     `json:"user_id"`
	Created          time.Time `json:"created,omitempty" db:"created"`
}

//GetBookingDuration returns Booking duration
func (prHistory *PaymentHistory) GetBookingDuration() int {
	log.Println("payment request history id = ", prHistory.PaymentHistoryID)
	hours := GetBookingDurationHoursWithPaymentHistory(prHistory.PaymentHistoryID)
	return hours
}
