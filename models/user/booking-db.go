package user

import (
	"log"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const bookingTableName = "bookings"
const serviceCategoryTable = "servicecategory"
const serviceInputTable = "serviceinputtype"

//InsertBooking creates a new booking
func InsertBooking(b *Booking) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()

	err := conn.QueryRow("insert into "+bookingTableName+"(profile_id, user_id) values ($1, $2) Returning booking_id",
		b.Profile.ProfileID, b.User.UserID).Scan(&id)

	return id, err
}

func GetServiceInputType(categoryID int64) []bool {

	conn := db.Connect()
	defer conn.Close()

	categoryRow := conn.QueryRow("select service_style_key from "+serviceCategoryTable+" where service_id=$1", categoryID)

	var serviceStyleKey pgx.NullInt64

	err := categoryRow.Scan(&serviceStyleKey)
	if err != nil {
		log.Println("Error in /models/user/booking-db.go GetServiceInputType(): Error scanning table: " + err.Error())
	}

	serviceInputRow := conn.QueryRow("select from_date, to_date, from_time, to_time, frequency_unit, total_price from "+serviceInputTable+" where service_input_id=$1", serviceStyleKey.Int64)

	var fromDate, toDate, fromTime, toTime, frequencyUnit, totalPrice pgx.NullBool
	err = serviceInputRow.Scan(&fromDate, &toDate, &fromTime, &toTime, &frequencyUnit, &totalPrice)
	if err != nil {
		log.Println("Error scanning ServiceInput Table:", err.Error())
	}

	serviceInputValue := []bool{fromDate.Bool, toDate.Bool, fromTime.Bool, toTime.Bool, frequencyUnit.Bool, totalPrice.Bool}

	return serviceInputValue
}

func GetBookingWithBookingId(id int64) (*Booking, error) {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select "+bookingFieldList()+" from "+bookingTableName+" where booking_id = $1", id)

	booking, err := scanSingleBooking(row)
	if err != nil {
		log.Println("Error in user/booking-db.go GetBookingWithBookingId() No Booking found in DB by GetBookingWithBookingId()")
		return nil, err
	}
	return booking, err
}

func scanSingleBooking(row *pgx.Row) (*Booking, error) {
	var bookingItem *Booking
	bookingItem = &Booking{}
	var bookingID, profileID, userID pgx.NullInt64
	var bookingUUID string

	err := row.Scan(&bookingID, &profileID, &userID, &bookingUUID)

	if err != nil {
		log.Println("Can not find BookingData in app/models/user/booking-db.go", err)
	}

	if bookingID.Valid {
		bookingItem.BookingID = bookingID.Int64
	}
	if profileID.Valid {
		p, _ := GetProfile(profileID.Int64)
		bookingItem.Profile = *p
	}
	if userID.Valid {
		u, _ := GetUser(userID.Int64)
		bookingItem.User = *u
	}
	bookingItem.BookingUUID, _ = gocql.ParseUUID(bookingUUID)

	return bookingItem, err
}

func bookingFieldList() string {
	list := " booking_id, profile_id, user_id, booking_uuid"
	return list
}

func GetBookingByBookingUUID(bookingUUID string) (*Booking, error) {
	conn := db.Connect()
	defer conn.Close()

	rows := conn.QueryRow("select "+bookingFieldList()+" from "+bookingTableName+" where booking_uuid = $1", bookingUUID)

	booking, err := scanSingleBooking(rows)
	if err != nil {
		log.Println("No Booking found in DB by GetBookingByBookingUUID()")
		return nil, err
	}
	return booking, err
}
