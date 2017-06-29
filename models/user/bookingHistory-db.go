package user

import (
	"log"

	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"github.com/unirep/ur-local-web/app/db"
)

const bookingHistoryTableName = "booking_history"

//InsertBookingHistory insert booking history to db
func InsertBookingHistory(bH *BookingHistory) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()

	bH.Created = time.Now().UTC()
	err := conn.QueryRow("insert into "+bookingHistoryTableName+" (booking_id, user_id, message, from_time, to_time, from_date, to_date, address, fee, total_price, booking_status, frequency_unit, frequency_value, created) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) Returning booking_history_id",
		bH.BookingID, bH.UserID, bH.Message, bH.FromTime, bH.ToTime, bH.FromDate, bH.ToDate, bH.Address, bH.Fee, bH.TotalPrice, bH.BookingStatus, bH.FrequencyUnit, bH.FrequencyValue, bH.Created).Scan(&id)

	return id, err
}

// UpdateBookingHistory updates booking history status to updated after customer accept or message to provider message
func UpdateBookingHistory(id int64, status string) error {
	conn := db.Connect()
	defer conn.Close()
	_, err := conn.Exec("Update "+bookingHistoryTableName+" Set booking_status=$1 where booking_history_id = $2", status, id)

	if err != nil {
		log.Println("Error in user/bookingHistory-db.go UpdateBookingHistory() fail to update booking_history:", err)
	}
	return err
}

//GetBookingHistoryWithHistoryID returns booking history
func GetBookingHistoryWithHistoryID(id int64) (*BookingHistory, error) {
	conn := db.Connect()
	defer conn.Close()
	rows := conn.QueryRow("select "+bookingHistoryFieldList()+" from "+bookingHistoryTableName+" where booking_history_id = $1", id)
	bookingHistory, err := scanSingleBookingHistory(rows)

	if err != nil {
		log.Println("No Booking History Data by GetBookingHistoryWithHistoryId():", err, "booking_history_id=", id)
		return nil, err
	}
	return bookingHistory, err
}

//CheckLastBookingHistory check this booking history is last
func CheckLastBookingHistory(bookingHistoryID int64, bookingID int64) bool {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select booking_history_id from "+bookingHistoryTableName+" where booking_history_id > $1 AND booking_id=$2", bookingHistoryID, bookingID)
	var historyID pgx.NullInt64
	err := row.Scan(&historyID)
	if err != nil {
		// log.Println("Error in modesl/bookingHistory-db.go failed to get next bookingHistory:", err)
	}
	return historyID.Valid
}

//GetBookingHistoriesWithBookingID returns booking histories using BookingID
func GetBookingHistoriesWithBookingID(id int64) ([]*BookingHistory, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+bookingHistoryFieldList()+" from "+bookingHistoryTableName+" where booking_id = $1 order by booking_history_id DESC", id)
	bookingHistories, err := scanMultipleBookingHistory(rows)
	if err != nil {
		log.Println("Error in app/models/bookingHistory-db.go  GetBookingHistoriesWithBookingId(): scanning multiple BookingHistory", err)
	}
	return bookingHistories, err
}

//GetBookingAccepted returns true if booking accepted or pending completion
func GetBookingAccepted(bookingID int64) bool {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select booking_status "+" from "+bookingHistoryTableName+" where booking_id = $1  AND (booking_status='Accepted' OR booking_status='Pending Completion' OR booking_status='Pending Payment')", bookingID)

	var status pgx.NullString
	err := row.Scan(&status)
	if err != nil {
		return false
	}

	return status.Valid
}

//GetBookingDeclineORCancel returns true if booking was declined or cancel
func GetBookingDeclineORCancel(id int64) bool {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select booking_status "+" from "+bookingHistoryTableName+" where booking_id = $1  AND (booking_status='Decline' OR booking_status='Cancel')", id)

	var status pgx.NullString
	err := row.Scan(&status)
	if err != nil {
		return false
	}

	return status.Valid
}

// GetBookingDurationHours returns booking duration
func GetBookingDurationHours(id int64) int {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select from_time, to_time from "+bookingHistoryTableName+" where booking_history_id = $1", id)
	var fromTime, toTime pgx.NullString
	err := row.Scan(&fromTime, &toTime)
	if err != nil {
		log.Println("Error in user/bookingHistory.go fail to get from time and to time:", err)
		return 0
	}
	format := "Jan 2 2006 3:04 PM"
	from := "Jan 2 2006 " + fromTime.String
	to := "Jan 2 2006 " + toTime.String
	t1, err := time.Parse(format, from)
	if err != nil {
		log.Println("Error user/bookingHistory.go fail to convert time", err)
		return 0
	}

	t2, err := time.Parse(format, to)
	if err != nil {
		log.Println("Error user/bookingHistory.go fail to convert time", err)
		return 0
	}

	diff := t2.Sub(t1)

	hours := int(diff.Minutes() / 60)

	return hours
}

//FeeChanged returns true if booking history fee changed
func (bH *BookingHistory) FeeChanged() bool {
	bookingHistories, err := GetBookingHistoriesWithBookingID(bH.BookingID)
	if err != nil {
		log.Println("Error in user/booking-history-db.go fail to get booking History data")
	}
	var FeeChanged bool
	if len(bookingHistories) < 2 {
		FeeChanged = false
		return FeeChanged
	}

	var oldBookinHistoryID int64
	for i, v := range bookingHistories {
		if bH.BookingHistoryID == v.BookingHistoryID {
			oldBookinHistoryID = int64(i) + 1
			break
		}
	}

	if bH.Fee.IntPart() == bookingHistories[oldBookinHistoryID].Fee.IntPart() {
		FeeChanged = false
	} else {
		FeeChanged = true
	}

	return FeeChanged
}

//DurationChanged returns true if booking duration changed
func (bH *BookingHistory) DurationChanged() bool {
	bookingHistories, err := GetBookingHistoriesWithBookingID(bH.BookingID)
	if err != nil {
		log.Println("Error in user/booking-history-db.go fail to get booking History data")
	}

	if len(bookingHistories) < 2 {
		return false
	}
	oldDuration := GetBookingDurationHours(bookingHistories[1].BookingHistoryID)
	lastDuration := GetBookingDurationHours(bH.BookingHistoryID)
	var DurationChanged bool
	if len(bookingHistories) == 1 {
		DurationChanged = false
	}
	if oldDuration == lastDuration {
		DurationChanged = false
	} else {
		DurationChanged = true
	}

	return DurationChanged
}

//DateChanged returns true if booking date changed
func (bH *BookingHistory) DateAndStartTimeChanged() (bool, bool) {
	bookingHistories, err := GetLastestBookingHistories(bH.BookingID)
	var IsDateChanged, IsStartTimeChanged bool
	IsDateChanged = false
	IsStartTimeChanged = false
	if err != nil {
		return false, false
	}
	if len(bookingHistories) == 2 {
		if bookingHistories[0].FromDate == bookingHistories[1].FromDate {
			IsDateChanged = false
		} else {
			IsDateChanged = true
		}
		if bookingHistories[0].FromTime == bookingHistories[1].FromTime {
			IsStartTimeChanged = false
		} else {
			IsStartTimeChanged = true
		}
	}
	return IsDateChanged, IsStartTimeChanged
}

//StartTimeChanged returns true if booking starttime changed
func (bH *BookingHistory) StartTimeChanged() bool {
	bookingHistories, err := GetLastestBookingHistories(bH.BookingID)
	if err != nil {
		return false
	}
	if len(bookingHistories) == 2 {
		if bookingHistories[0].FromTime == bookingHistories[1].FromTime {
			return false
		}
		return true
	}
	return false
}

//GetLastestBookingHistories returns lastest two bookigHistories with bookingID
func GetLastestBookingHistories(bookingID int64) ([]*BookingHistory, error) {
	conn := db.Connect()
	defer conn.Close()
	rows, err := conn.Query("select "+bookingHistoryFieldList()+" from "+bookingHistoryTableName+" where booking_id = $1 order by booking_history_id DESC limit 2", bookingID)
	if err != nil {
		log.Println("No Booking History Data by GetLastestBookingHistories(): ", err)
		return nil, err
	}
	bookingHistories, err := scanMultipleBookingHistory(rows)

	if err != nil {
		log.Println("No Booking History Data by GetLastestBookingHistories():", err)
		return nil, err
	}

	return bookingHistories, nil
}

func bookingHistoryFieldList() string {
	list := "booking_history_id, booking_id, user_id, message, from_time, to_time, from_date, to_date, address, fee, total_price, booking_status, frequency_unit, frequency_value, created, booking_history_uuid"
	return list
}

func scanSingleBookingHistory(row *pgx.Row) (*BookingHistory, error) {
	var bookingHistoryItem *BookingHistory
	bookingHistoryItem = &BookingHistory{}
	var bookingHistoryID, bookingID, userID, frequencyValue pgx.NullInt64
	var totalPrice, fee decimal.Decimal
	var created pgx.NullTime
	var message, fromTime, toTime, address, fromDate, toDate, bookingStatus, frequencyUnit pgx.NullString
	var bookingHistoryUUID string
	err := row.Scan(&bookingHistoryID, &bookingID, &userID, &message, &fromTime, &toTime, &fromDate, &toDate, &address, &fee, &totalPrice, &bookingStatus, &frequencyUnit, &frequencyValue, &created, &bookingHistoryUUID)

	if err != nil {
		log.Println("Can not find BookingHistoryData in app/models/user/bookingHistory-db.go", err)
	}

	if bookingHistoryID.Valid {
		bookingHistoryItem.BookingHistoryID = bookingHistoryID.Int64
	}
	if bookingID.Valid {
		bookingHistoryItem.BookingID = bookingID.Int64
	}
	if userID.Valid {
		bookingHistoryItem.UserID = userID.Int64
	}
	if frequencyValue.Valid {
		bookingHistoryItem.FrequencyValue = frequencyValue.Int64
	}
	if message.Valid {
		bookingHistoryItem.Message = message.String
	}
	if fromDate.Valid {
		bookingHistoryItem.FromDate = fromDate.String
	}
	if toDate.Valid {
		bookingHistoryItem.ToDate = toDate.String
	}
	if fromTime.Valid {
		bookingHistoryItem.FromTime = fromTime.String
	}
	if toTime.Valid {
		bookingHistoryItem.ToTime = toTime.String
	}
	if bookingStatus.Valid {
		bookingHistoryItem.BookingStatus = bookingStatus.String
	}
	if frequencyUnit.Valid {
		bookingHistoryItem.FrequencyUnit = frequencyUnit.String
	}
	if created.Valid {
		bookingHistoryItem.Created = created.Time
	}
	if address.Valid {
		bookingHistoryItem.Address = address.String
	}

	bookingHistoryItem.Fee = fee
	bookingHistoryItem.TotalPrice = totalPrice
	bookingHistoryItem.BookingHistoryUUID, _ = gocql.ParseUUID(bookingHistoryUUID)

	return bookingHistoryItem, err
}

func scanMultipleBookingHistory(rows *pgx.Rows) ([]*BookingHistory, error) {
	bookingHistories := []*BookingHistory{}
	var err error
	for rows.Next() {
		var bookingHistoryID, bookingID, userID, frequencyValue pgx.NullInt64
		var totalPrice, fee decimal.Decimal
		var created pgx.NullTime
		var message, fromTime, toTime, address, fromDate, toDate, bookingStatus, frequencyUnit pgx.NullString
		var bookingHistoryUUID string
		err = rows.Scan(&bookingHistoryID, &bookingID, &userID, &message, &fromTime, &toTime, &fromDate, &toDate, &address, &fee, &totalPrice, &bookingStatus, &frequencyUnit, &frequencyValue, &created, &bookingHistoryUUID)
		bookingHistory := fillBookingHistory(bookingHistoryID, bookingID, userID, frequencyValue, totalPrice, fee, message, fromTime, toTime, address, fromDate, toDate, bookingStatus, frequencyUnit, created, bookingHistoryUUID)
		bookingHistories = append(bookingHistories, bookingHistory)
	}
	return bookingHistories, err
}

func fillBookingHistory(bookingHistoryID, bookingID, userID, frequencyValue pgx.NullInt64, totalPrice, fee decimal.Decimal, message, fromTime, toTime, address, fromDate, toDate, bookingStatus, frequencyUnit pgx.NullString, created pgx.NullTime, bookingHistoryUUID string) *BookingHistory {
	bookingHistory := &BookingHistory{}
	if bookingHistoryID.Valid {
		bookingHistory.BookingHistoryID = bookingHistoryID.Int64
	}
	if bookingID.Valid {
		bookingHistory.BookingID = bookingID.Int64
	}
	if userID.Valid {
		bookingHistory.UserID = userID.Int64
	}
	if message.Valid {
		bookingHistory.Message = message.String
	}
	if fromTime.Valid {
		bookingHistory.FromTime = fromTime.String
	}
	if toTime.Valid {
		bookingHistory.ToTime = toTime.String
	}
	if fromDate.Valid {
		bookingHistory.FromDate = fromDate.String
	}
	if toDate.Valid {
		bookingHistory.ToDate = toDate.String
	}
	if frequencyUnit.Valid {
		bookingHistory.FrequencyUnit = frequencyUnit.String
	}
	if bookingStatus.Valid {
		bookingHistory.BookingStatus = bookingStatus.String
	}
	if created.Valid {
		bookingHistory.Created = created.Time
	}

	bookingHistory.Fee = fee
	bookingHistory.TotalPrice = totalPrice
	bookingHistory.BookingHistoryUUID, _ = gocql.ParseUUID(bookingHistoryUUID)

	return bookingHistory
}

func GetBookingHistoryByUUID(bookingHistoryUUID string) (*BookingHistory, error) {

	conn := db.Connect()
	defer conn.Close()
	rows := conn.QueryRow("select "+bookingHistoryFieldList()+" from "+bookingHistoryTableName+" where booking_history_uuid = $1", bookingHistoryUUID)
	bookingHistory, err := scanSingleBookingHistory(rows)

	if err != nil {
		log.Println("No Booking History Data by GetBookingHistoryWithHistoryUUID():", err)
		return nil, err
	}
	return bookingHistory, err
}
