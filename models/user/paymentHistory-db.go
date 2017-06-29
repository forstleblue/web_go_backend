package user

import (
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const paymentHistoryTableName = "payment_history"

//InsertPaymentHistory inserts payment request history
func InsertPaymentHistory(pH *PaymentHistory) (int64, error) {
	var id int64

	conn := db.Connect()
	defer conn.Close()
	err := conn.QueryRow("insert into "+paymentHistoryTableName+"(payment_id, amount, status, message, user_id) values ($1, $2, $3, $4, $5) Returning payment_history_id",
		pH.PaymentID, pH.Amount, pH.Status, pH.Message, pH.UserID).Scan(&id)

	return id, err
}

//GetPaymentHistory returns payment request history
func GetPaymentHistory(id int64) (*PaymentHistory, error) {
	conn := db.Connect()
	defer conn.Close()

	rows := conn.QueryRow("select "+paymentHistoryFieldList()+" from "+paymentHistoryTableName+" where payment_history_id = $1", id)

	paymentHistory, err := scanSinglePaymentHistory(rows)
	if err != nil {
		log.Println("Error in models/user/payment-db.go GetPaymentRequest: No PaymentRequest found in DB")
		return nil, err
	}
	return paymentHistory, err
}

//GetPaymentHistories returns payment request histories using payment requset id
func GetPaymentHistories(prID int64) ([]*PaymentHistory, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+paymentHistoryFieldList()+" from "+paymentHistoryTableName+" where payment_request_id = $1 order by payment_request_history_id DESC", prID)
	prHistories, err := scanMultiplePaymentHistory(rows)
	if err != nil {
		log.Println("Error in app/models/bookingHistory-db.go  GetBookingHistoriesWithBookingId(): scanning multiple BookingHistory", err)
	}
	return prHistories, err
}

//GetBookingDurationHoursWithPaymentHistory returns booking duration
func GetBookingDurationHoursWithPaymentHistory(id int64) int {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select from_time, to_time from "+paymentHistoryTableName+" where payment_history_id = $1", id)
	var fromTime, toTime pgx.NullString
	err := row.Scan(&fromTime, &toTime)
	if err != nil {
		log.Println("Error in user/paymentHistory-db.go fail to get from time and to time:", err, "payment requset history id = ", id)
		return 0
	}
	format := "Jan 2 2006 3:04 PM"
	from := "Jan 2 2006 " + fromTime.String
	to := "Jan 2 2006 " + toTime.String
	t1, err := time.Parse(format, from)
	if err != nil {
		log.Println("Error user/paymentHistory-db.go fail to convert time", err)
		return 0
	}

	t2, err := time.Parse(format, to)
	if err != nil {
		log.Println("Error user/paymentHistory-db.go fail to convert time", err)
		return 0
	}

	diff := t2.Sub(t1)

	hours := int(diff.Minutes() / 60)

	return hours
}

func paymentHistoryFieldList() string {

	list := " payment_history_id, payment_id, user_id, amount, status, message, created "
	return list

}

func scanSinglePaymentHistory(row *pgx.Row) (*PaymentHistory, error) {
	var paymentID, paymentHistoryID, userID pgx.NullInt64
	var amount pgx.NullInt32
	var status, message pgx.NullString
	var created pgx.NullTime
	err := row.Scan(&paymentHistoryID, &paymentID, &userID, &amount, &status, &message, &created)
	if err != nil {
		log.Println("Error in models/user/paymentHistory-db.go scanSinglePaymentHistory: issue scanning row", err)
	}

	paymentHistory := fillPaymentHistory(&paymentHistoryID, &paymentID, &userID, &amount, &status, &message, &created)

	return paymentHistory, err
}

func scanMultiplePaymentHistory(rows *pgx.Rows) ([]*PaymentHistory, error) {
	prHistories := []*PaymentHistory{}
	var err error

	for rows.Next() {
		var paymentID, paymentHistoryID, userID pgx.NullInt64
		var amount pgx.NullInt32
		var status, message, fromDate, toDate, fromTime, toTime pgx.NullString
		var created pgx.NullTime
		err := rows.Scan(&paymentID, &paymentHistoryID, &userID, &amount, &status, &message, &fromDate, &toDate, &fromTime, &toTime, &created)
		if err != nil {
			log.Println("Error in models/user/paymentHistory-db.go scanMultiplePaymentHistory: issue scanning row", err)
		}
		prHistory := fillPaymentHistory(&paymentHistoryID, &paymentID, &userID, &amount, &status, &message, &created)
		prHistories = append(prHistories, prHistory)
	}
	return prHistories, err
}

func fillPaymentHistory(paymentHistoryID, paymentID, userID *pgx.NullInt64, amount *pgx.NullInt32, status, message *pgx.NullString, created *pgx.NullTime) *PaymentHistory {
	paymentHistory := &PaymentHistory{}

	if paymentHistoryID.Valid {
		paymentHistory.PaymentHistoryID = paymentHistoryID.Int64
	}
	if paymentID.Valid {
		paymentHistory.PaymentID = paymentID.Int64
	}
	if userID.Valid {
		paymentHistory.UserID = userID.Int64
	}
	if amount.Valid {
		paymentHistory.Amount = amount.Int32
	}
	if status.Valid {
		paymentHistory.Status = status.String
	}
	if message.Valid {
		paymentHistory.Message = message.String
	}
	if created.Valid {
		paymentHistory.Created = created.Time
	}
	return paymentHistory

}
