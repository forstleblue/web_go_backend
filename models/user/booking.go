package user

import "log"
import "github.com/gocql/gocql"

//Booking is
type Booking struct {
	BookingID   int64      `json:"booking_id"` //todo make this a type 1 GUID from the gocql lib
	User        User       `json:"user_id"`
	Profile     Profile    `json:"profile_id"`
	BookingUUID gocql.UUID `json:"booking_uuid"`
}

//CheckPaymentRequest returns true if service provider send payment request
func (b Booking) CheckPaymentRequest() bool {
	createdPayment := CheckPaymentByBookingID(b.BookingID)
	return createdPayment
}

//GetPaymentRequestStatus returns
func (b Booking) GetPaymentRequestStatus() string {
	status := GetPaymentStatusByBookingID(b.BookingID)
	return status
}

//GetBookingHistories returns booking history list with bookingID
func (b *Booking) GetBookingHistories(id int64) ([]*BookingHistory, error) {
	bookingHistories, err := GetBookingHistoriesWithBookingID(id)
	if err != nil {
		log.Println("Error in /models/booking.go GetBookingHistories() Failed to get BookingHistoires with BookingID:", err)
		return nil, err
	}
	return bookingHistories, err
}

//GetLatestBookingHistory returns last booking history
func (b *Booking) GetLatestBookingHistory() (*BookingHistory, error) {
	list, err := b.GetBookingHistories(b.BookingID)
	if err != nil {
		log.Println("Error in /models/booking.go GetBookingHistories() Failed to get BookingHistoires with BookingID:", err)
		return nil, err
	} else if len(list) == 0 {
		log.Println("Error in /models/booking.go GetBookingHistories() No booking histories exist for booking ID: ", b.BookingID)
		return nil, err
	}
	return list[0], nil
}

//CheckBookingAcceptedWithBookingID returns true if booking was accepted or pending completion.
func (b *Booking) CheckBookingAcceptedWithBookingID(id int64) bool {
	status := GetBookingAccepted(id)
	return status
}
