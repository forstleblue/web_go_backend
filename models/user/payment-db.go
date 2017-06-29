package user

import (
	"encoding/json"
	"log"
	"time"

	"fmt"

	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"github.com/unirep/ur-local-web/app/db"
)

const paymentTableName = "payments"

//InsertPayment creates a new payment request
func InsertPayment(pr *Payment) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()

	pr.Status = PaymentRequestStatusNew
	pr.RequestDate = time.Now().UTC()

	jsonData, errValue := json.Marshal(pr.Message)
	if errValue != nil {
		log.Println("Error user/payment-db.go fail to marshal message data:", errValue)
		return 0, errValue
	}
	err := conn.QueryRow("insert into "+paymentTableName+"(booking_id, amount, request_date, status, message, tip) values ($1, $2, $3, $4, $5, $6) Returning payment_id",
		pr.Booking.BookingID, pr.Amount, pr.RequestDate, pr.Status, string(jsonData), pr.Tip).Scan(&id)

	return id, err
}

//Update saves the changes for the payment to the database
func (pr *Payment) Update() error {
	conn := db.Connect()
	defer conn.Close()
	jsonData, errValue := json.Marshal(pr.Message)
	if errValue != nil {
		log.Println("Error user/payment-db.go fail to marshal message data:", errValue)
		return errValue
	}
	_, err := conn.Exec("Update "+paymentTableName+" Set payment_date = $1, confirmed_date = $2, payment_method = $3, transaction_id = $4, payment_status = $5, acct_display = $6, paypal_token = $7, paypal_payer_id = $8, paypal_payer_status = $9, status = $10, declined_reason = $11, message = $12, tip = $13, amount=$14 where payment_id = $15",
		pr.PaymentDate, pr.ConfirmedDate, pr.PaymentMethod, pr.TransactionID, pr.PaymentStatus, pr.AcctDisplay, pr.PaypalToken, pr.PaypalPayerID, pr.PaypalPayerStatus, pr.Status, pr.DeclinedReason, string(jsonData), pr.Tip, pr.Amount, pr.PaymentID)

	return err
}

// UpadatedPaymentRequest returns true if service provider already sent payment requset.
func (pr *Payment) UpadatedPaymentRequest() bool {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("Select payment_id from "+paymentTableName+" where payment_id > $1 AND booking_id = $2", pr.PaymentID, pr.Booking.BookingID)
	var id pgx.NullInt64

	err := row.Scan(&id)
	if err != nil {
		log.Println("In user/payment-db.go UpadatedPaymentRequest() payment request did not updated:", err, "payment request id", pr.PaymentID, "booking id = ", pr.Booking.BookingID)
		return false
	}
	var updated bool
	if id.Valid {
		updated = true
	} else {
		updated = false
	}
	return updated
}

//GetMessage returns message content
func (pr *Payment) GetMessage(id int64) string {
	conn := db.Connect()
	defer conn.Close()
	idParam := fmt.Sprint(id)
	row := conn.QueryRow("Select message::json->"+idParam+" from "+paymentTableName+" where payment_id = $1", pr.PaymentID)
	var messageItem MessageData

	err := row.Scan(&messageItem)
	if err != nil {
		log.Println("Error in user/payment-db.go GetMessage() fail to get message:", err)
	}

	return messageItem.Message
}

//GetPayment gets a payment request by id
func GetPayment(id int64) (*Payment, error) {
	conn := db.Connect()
	defer conn.Close()

	rows := conn.QueryRow("select "+paymentFieldList()+" from "+paymentTableName+" where payment_id = $1", id)

	paymentRequest, err := scanSinglePayment(rows)
	if err != nil {
		log.Println("Error in models/user/payment-db.go GetPayment: No Payment found in DB:", err, "payment_id", id)
		return nil, err
	}
	return paymentRequest, err
}

//GetPaymentByBookingID gets a payment request by id
func GetPaymentByBookingID(id int64) (*Payment, error) {
	conn := db.Connect()
	defer conn.Close()

	rows := conn.QueryRow("select "+paymentFieldList()+" from "+paymentTableName+" where booking_id = $1", id)

	paymentRequest, err := scanSinglePayment(rows)
	if err != nil {
		log.Println("Error in models/user/payment-db.go GetPayment: No Payment found in DB:", err, "booking_id", id)
		return nil, err
	}
	return paymentRequest, err
}

//CheckPaymentByBookingID checks for service provider send payment request or not.
func CheckPaymentByBookingID(bookingID int64) bool {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select status "+" from "+paymentTableName+" where booking_id = $1", bookingID)

	var status pgx.NullString
	err := row.Scan(&status)
	if err != nil {
		// log.Println("Error in models/user/payment-db.go CheckPaymentRequestByBookingID(bookingID int64): scanning PaymentRequest Table with BookingID :", err)
	}

	return status.Valid
}

//GetPaymentStatusByBookingID returns booking status
func GetPaymentStatusByBookingID(bookingID int64) string {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select status, payment_id"+" from "+paymentTableName+" where  booking_id = $1 AND payment_id = (select max(payment_id) from "+paymentTableName+")", bookingID)

	var status pgx.NullString
	var id pgx.NullInt64
	err := row.Scan(&status, &id)
	if err != nil {
		log.Println("Error in models/user/payment-db.go CheckPaymentRequestByBookingID(bookingID int64): scanning PaymentRequest Table with BookingID :", err)
	}

	return status.String
}

func scanMultiplePaymentRequests(rows *pgx.Rows) ([]*Payment, error) {
	payments := []*Payment{}
	var err error
	for rows.Next() {
		var requestID, bookingID pgx.NullInt64
		var amount pgx.NullInt32
		var paymentMethod, transactionID, paymentStatus, acctDisplay, paypalToken, paypalPayerID, paypalPayerStatus, status, declinedReason pgx.NullString
		var message []MessageData
		var requestDate, confirmedDate, paymentDate pgx.NullTime
		var tip decimal.Decimal

		err = rows.Scan(&requestID, &bookingID, &amount, &requestDate, &confirmedDate, &paymentDate, &paymentMethod, &transactionID, &paymentStatus, &acctDisplay, &paypalToken, &paypalPayerID, &paypalPayerStatus, &status, &declinedReason, &message, &tip)

		if err != nil {
			log.Println("Error in models/user/payment-db.go scanMultiplePaymentRequests: scanning payment requests row in scanMultiplePaymentRequests: ", err.Error())
			return nil, err
		}

		paymentItem := fillPaymentRequest(&requestID, &bookingID, &amount, &requestDate, &confirmedDate, &paymentDate, &paymentMethod, &transactionID, &paymentStatus, &acctDisplay, &paypalToken, &paypalPayerID, &paypalPayerStatus, &status, &declinedReason, message, tip)
		payments = append(payments, paymentItem)
	}
	return payments, nil
}

func scanSinglePayment(row *pgx.Row) (*Payment, error) {
	var requestID, bookingID pgx.NullInt64
	var amount pgx.NullInt32
	var paymentMethod, transactionID, paymentStatus, acctDisplay, paypalToken, paypalPayerID, paypalPayerStatus, status, declinedReason pgx.NullString
	var requestDate, confirmedDate, paymentDate pgx.NullTime
	var message []MessageData
	var tip decimal.Decimal

	err := row.Scan(&requestID, &bookingID, &amount, &requestDate, &confirmedDate, &paymentDate, &paymentMethod, &transactionID, &paymentStatus, &acctDisplay, &paypalToken, &paypalPayerID, &paypalPayerStatus, &status, &declinedReason, &message, &tip)

	if err != nil {
		log.Println("Error in models/user/payment-db.go scanSinglePaymentRequest: issue scanning row", err)
	}

	paymentRequest := fillPaymentRequest(&requestID, &bookingID, &amount, &requestDate, &confirmedDate, &paymentDate, &paymentMethod, &transactionID, &paymentStatus, &acctDisplay, &paypalToken, &paypalPayerID, &paypalPayerStatus, &status, &declinedReason, message, tip)

	return paymentRequest, err
}

func paymentFieldList() string {
	list := " payment_id, booking_id, amount, request_date, confirmed_date, payment_date, payment_method, transaction_id, payment_status, acct_display, paypal_token, paypal_payer_id, paypal_payer_status, status, declined_reason, message, tip "
	return list
}

func fillPaymentRequest(requestID, bookingID *pgx.NullInt64, amount *pgx.NullInt32, requestDate, confirmedDate, paymentDate *pgx.NullTime, paymentMethod, transactionID, paymentStatus, acctDisplay, paypalToken, paypalPayerID, paypalPayerStatus, status, declinedReason *pgx.NullString, message []MessageData, tip decimal.Decimal) *Payment {
	payment := &Payment{}
	if requestID.Valid {
		payment.PaymentID = requestID.Int64
	}
	if bookingID.Valid {
		b, _ := GetBookingWithBookingId(bookingID.Int64)
		payment.Booking = *b
	}
	if amount.Valid {
		payment.Amount = amount.Int32
	}
	if requestDate.Valid {
		payment.RequestDate = requestDate.Time
	}
	if confirmedDate.Valid {
		payment.ConfirmedDate = confirmedDate.Time
	}
	if paymentDate.Valid {
		payment.PaymentDate = paymentDate.Time
	}
	if paymentMethod.Valid {
		payment.PaymentMethod = paymentMethod.String
	}
	if transactionID.Valid {
		payment.TransactionID = transactionID.String
	}
	if paymentStatus.Valid {
		payment.PaymentStatus = paymentStatus.String
	}
	if acctDisplay.Valid {
		payment.AcctDisplay = acctDisplay.String
	}
	if paypalToken.Valid {
		payment.PaypalToken = paypalToken.String
	}
	if paypalPayerID.Valid {
		payment.PaypalPayerID = paypalPayerID.String
	}
	if paypalPayerStatus.Valid {
		payment.PaypalPayerStatus = paypalPayerStatus.String
	}
	if status.Valid {
		payment.Status = status.String
	}
	if declinedReason.Valid {
		payment.DeclinedReason = declinedReason.String
	}
	payment.Message = message
	payment.Tip = tip
	return payment
}
