package user

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/shopspring/decimal"
)

//BookingHistoryStatusNew is
const (
	BookingHistoryStatusNew                 = "New"
	BookingHistoryStatusUpdated             = "Updated"
	BookingHistoryStatusMessage             = "Message"
	BookingHistoryStatusPendingInCompletion = "Pending Completion"
	BookingHistoryStatusAccepted            = "Accepted"
	BookingHistoryStatusDecline             = "Decline"
	BookingHistoryStatusCancel              = "Cancel"
)

//BookingHistory is Booking History struct
type BookingHistory struct {
	BookingHistoryID   int64           `json:"booking_history_id"`
	BookingID          int64           `json:"booking_id"`
	UserID             int64           `json:"user_id"`
	Message            string          `json:"message"`
	FromTime           string          `json:"from_time"`
	ToTime             string          `json:"to_time"`
	FromDate           string          `json:"from_date"`
	ToDate             string          `json:"to_date"`
	Address            string          `json:"address"`
	Fee                decimal.Decimal `json:"fee"`
	TotalPrice         decimal.Decimal `json:"total_price"`
	BookingStatus      string          `json:"booking_status"`
	FrequencyUnit      string          `json:"frequency_unit"`
	FrequencyValue     int64           `json:"frequency_value"`
	Created            time.Time       `json:"created,omitempty" db:"created"`
	BookingHistoryUUID gocql.UUID      `json:"booking_history_uuid" db:"booking_history_uuid"`
}

//GetUserNameWithUserID returns user name
func (bH *BookingHistory) GetUserNameWithUserID(id int64) string {
	user, err := GetUser(id)
	if err != nil {
		log.Println("Error in models/bookingHistory.go GetUserNameWithUserID() failed to get user ", err)
		return ""
	}
	return user.FullName()
}

//CreatedTimeFormat returns formatted time string
func (bH *BookingHistory) CreatedTimeFormat() string {
	return bH.Created.Format("02 Jan 2006 15:04")
}

//FromTimeFormat returns from Date and time string
func (bH *BookingHistory) FromTimeFormat() string {
	return bH.FromDate + " " + bH.FromTime
}

//CheckBookingHistoryUpdate returns true if booking history updates
func (bH *BookingHistory) CheckBookingHistoryUpdate() bool {
	checkLastBookingHistory := CheckLastBookingHistory(bH.BookingHistoryID, bH.BookingID)
	return checkLastBookingHistory
}

//GetBookingDuration returns booking duration hour and minutes
func (bH *BookingHistory) GetBookingDuration() int {
	hours := GetBookingDurationHours(bH.BookingHistoryID)
	return hours
}

//CheckBookingAccept returns last booking history status to check booking accept
func (bH *BookingHistory) CheckBookingAccept() bool {
	bookingAccept := GetBookingAccepted(bH.BookingID)
	return bookingAccept
}

//BookingCancel returns true if booking cancel or decline
func (bH *BookingHistory) BookingCancel() bool {
	bookingCancel := GetBookingDeclineORCancel(bH.BookingID)
	return bookingCancel
}
