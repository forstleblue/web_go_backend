package feedback

import (
	"log"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"github.com/unirep/ur-local-web/app/db"
)

const feedbackTableName = "feedbacks"

//InsertFeedback inserts feedback
func InsertFeedback(feedback *Feedback) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()
	err := conn.QueryRow("insert into "+feedbackTableName+" (booking_id, sender_profile_id, description, comment, score, sda_text, receiver_profile_id, positive, neutral, negative) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) Returning feedback_id",
		feedback.BookingID, feedback.SenderProfileID, feedback.Description, feedback.Comment, feedback.Score, feedback.SdaText, feedback.ReceiverProfileID, feedback.Positive, feedback.Neutral, feedback.Negative).Scan(&id)
	return id, err
}

// GetAllFeedback returns feedback list for user profileList
func GetAllFeedback(currUserID int64) ([]*Feedback, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select feedback_id, booking_id, sender_profile_id, description, comment, score, sda_text, receiver_profile_id, feedback_uuid, created, positive, neutral, negative from  "+feedbackTableName+" where receiver_profile_id in(select profile_id from profiles where user_id=$1)", currUserID)
	if err != nil {
		log.Println("Error routes/feedbacks.go GetFeedbacks() querying for Feedbacks:", err)
		return nil, err
	}
	feedbackList, err := scanMultipleFeedback(rows)
	if err != nil {
		log.Println("Error querying for feedback List:", err)
	}

	return feedbackList, err
}

//GetAllFeedbackByProfileID returns feedback using profileID
func GetAllFeedbackByProfileID(profileID int64) ([]*Feedback, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select feedback_id, booking_id, sender_profile_id, description, comment, score, sda_text, receiver_profile_id, feedback_uuid, created, positive, neutral, negative from "+feedbackTableName+" where receiver_profile_id=$1", profileID)
	if err != nil {
		log.Println("Error routes/feedbacks.go GetAllFeedbackByProfileID() querying for Feedbacks:", err)
		return nil, err
	}
	feedbackList, err := scanMultipleFeedback(rows)
	if err != nil {
		log.Println("Error querying for feedback List:", err)
	}

	return feedbackList, err
}

// scanMultipleFeedback returns feedback list
func scanMultipleFeedback(rows *pgx.Rows) ([]*Feedback, error) {
	feedbackList := []*Feedback{}
	var err error
	for rows.Next() {
		var feedbackID, bookingID, senderProfileID, receiverProfileID pgx.NullInt64
		var description, comment pgx.NullString
		var score pgx.NullInt16
		var sdaText []string
		var created pgx.NullTime
		var feedbackUUID string
		var positive, neutral, negative pgx.NullBool
		err = rows.Scan(&feedbackID, &bookingID, &senderProfileID, &description, &comment, &score, &sdaText, &receiverProfileID, &feedbackUUID, &created, &positive, &neutral, &negative)
		if err != nil {
			log.Println("Error scanning feedback row in scanMultipleFeedback: ", err.Error())
			return nil, err
		}
		feedback := fillFeedback(feedbackID, bookingID, senderProfileID, description, comment, score, sdaText, receiverProfileID, feedbackUUID, created, positive, neutral, negative)
		feedbackList = append(feedbackList, feedback)
	}
	return feedbackList, nil
}

// fillFeedback returns one feedback
func fillFeedback(feedbackID, bookingID, senderProfileID pgx.NullInt64, description, comment pgx.NullString, score pgx.NullInt16, sdaText []string, receiverProfileID pgx.NullInt64, feedbackUUID string, created pgx.NullTime, positive pgx.NullBool, neutral pgx.NullBool, negative pgx.NullBool) *Feedback {
	feedback := &Feedback{}
	if feedbackID.Valid {
		feedback.FeedbackID = feedbackID.Int64
	}
	if bookingID.Valid {
		feedback.BookingID = bookingID.Int64
	}
	if senderProfileID.Valid {
		feedback.SenderProfileID = senderProfileID.Int64
	}
	if description.Valid {
		feedback.Description = description.String
	}
	if comment.Valid {
		feedback.Comment = comment.String
	}
	if score.Valid {
		feedback.Score = score.Int16
	}
	if receiverProfileID.Valid {
		feedback.ReceiverProfileID = receiverProfileID.Int64
	}
	if created.Valid {
		feedback.Created = created.Time
	}
	if positive.Valid {
		feedback.Positive = positive.Bool
	}
	if neutral.Valid {
		feedback.Neutral = neutral.Bool
	}
	if negative.Valid {
		feedback.Negative = negative.Bool
	}
	feedback.FeedbackUUID, _ = gocql.ParseUUID(feedbackUUID)
	feedback.SdaText = sdaText

	return feedback
}

//GetFeedbackByBookingIDandCreatedProfileID gets feedback infor from booking and profileID
func GetFeedbackByBookingIDandCreatedProfileID(bookingIDparam int64, senderProfileIDparam int64) (*Feedback, error) {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select feedback_id, booking_id, sender_profile_id, description, comment, score, sda_text, positive, neutral, negative from "+feedbackTableName+" where booking_id=$1 AND sender_profile_id=$2", bookingIDparam, senderProfileIDparam)

	var feedbackID, bookingID, senderProfileID pgx.NullInt64
	var description, comment pgx.NullString
	var score pgx.NullInt16
	var sdaText []string
	var feedback Feedback
	var positive, neutral, negative pgx.NullBool

	err := row.Scan(&feedbackID, &bookingID, &senderProfileID, &description, &comment, &score, &sdaText, &positive, &neutral, &negative)
	if err != nil {
		return nil, err
	}
	if feedbackID.Valid {
		feedback.FeedbackID = feedbackID.Int64
	}
	if bookingID.Valid {
		feedback.BookingID = bookingID.Int64
	}
	if senderProfileID.Valid {
		feedback.SenderProfileID = senderProfileID.Int64
	}
	if description.Valid {
		feedback.Description = description.String
	}
	if comment.Valid {
		feedback.Comment = comment.String
	}
	if score.Valid {
		feedback.Score = score.Int16
	}
	if positive.Valid {
		feedback.Positive = positive.Bool
	}
	if neutral.Valid {
		feedback.Neutral = neutral.Bool
	}
	if negative.Valid {
		feedback.Negative = negative.Bool
	}
	feedback.SdaText = sdaText

	return &feedback, err
}

//CheckTwoFeedbackWrited returns true if two feedback completes
func CheckTwoFeedbackWrited(bookingID int64) bool {
	conn := db.Connect()
	defer conn.Close()
	var count int64
	rows := conn.QueryRow("select count(*) from "+feedbackTableName+" where booking_id = $1", bookingID)
	rows.Scan(&count)
	if count < 2 {
		return false
	}
	return true
}

func GetFeedbackAverageByProfileID(profileIDparam int64) int16 {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select avg(score) from "+feedbackTableName+" where receiver_profile_id  = $1", profileIDparam)

	if row == nil {
		log.Println("row nil ")
		return -1
	}

	var scoreAvg pgx.NullString

	err := row.Scan(&scoreAvg)
	if err != nil {
		return -1
	}
	if scoreAvg.Valid {
		dec, _ := decimal.NewFromString(scoreAvg.String)
		return int16(dec.IntPart())
	}

	return -1
}

func GetFeedbackAverageByUserID(userIDparam int64) int16 {

	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select avg(score) from "+feedbackTableName+" where receiver_profile_id in (select profile_id from profiles where user_id = $1)", userIDparam)

	if row == nil {
		log.Println("row nil ")
		return -1
	}

	var scoreAvg pgx.NullString

	err := row.Scan(&scoreAvg)
	if err != nil {
		return -1
	}
	if scoreAvg.Valid {
		dec, _ := decimal.NewFromString(scoreAvg.String)
		return int16(dec.IntPart())
	}

	return -1
}

func GetFeedbackCountByProfileID(profileIDparam int64) int64 {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select count(score) from "+feedbackTableName+" where receiver_profile_id  = $1", profileIDparam)

	if row == nil {
		return -1
	}

	var count pgx.NullInt64

	err := row.Scan(&count)
	if err != nil {
		return -1
	}
	if count.Valid {
		return count.Int64
	}

	return -1
}

func GetFeedbackCountByUserID(userIDparam int64) int64 {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select count(score) from "+feedbackTableName+" where receiver_profile_id in (select profile_id from profiles where user_id = $1)", userIDparam)

	if row == nil {
		return -1
	}

	var count pgx.NullInt64

	err := row.Scan(&count)
	if err != nil {
		return -1
	}
	if count.Valid {
		return count.Int64
	}

	return -1
}

func GetFeedbackTopSDAByProfileID(profileIDparam int64) []string {
	conn := db.Connect()
	defer conn.Close()

	returnCount := 3
	var returnVal []string

	rows, err := conn.Query(`select UNNEST(sda_text) as sda, COUNT(*) as sda_count, created 
		from (select * from `+feedbackTableName+` where receiver_id = $1) profile_feedback
		group by sda
		order by sda_count desc, created desc
		LIMIT $2`, profileIDparam, returnCount)

	if err != nil {
		log.Println("Error routes/feedbacks.go GetFeedbackTopSDAByProfileID querying for top SDA values:", err)
		return nil
	}

	var sda pgx.NullString
	var sdaCount pgx.NullInt64
	var created pgx.NullTime

	for rows.Next() {
		err := rows.Scan(&sda, &sdaCount, &created)
		if err != nil {
			log.Println("Error routes/feedbacks.go GetFeedbackTopSDAByProfileID scanning SDA value:", err)
			return nil
		}
		if sda.Valid {
			returnVal = append(returnVal, sda.String)
		}
	}

	return returnVal
}

func GetFeedbackTopSDAByUserID(userIDparam int64) []string {
	conn := db.Connect()
	defer conn.Close()

	returnCount := 3
	var returnVal []string

	rows, err := conn.Query(`select UNNEST(sda_text) as sda, COUNT(*) as sda_count, created 
		from (select * from `+feedbackTableName+` where receiver_profile_id in (select profile_id from profiles where user_id = $1)) profile_feedback
		group by sda
		order by sda_count desc, created desc
		LIMIT $2`, userIDparam, returnCount)

	if err != nil {
		log.Println("Error routes/feedbacks.go GetFeedbackTopSDAByUserID querying for top SDA values:", err)
		return nil
	}

	var sda pgx.NullString
	var sdaCount pgx.NullInt64
	var created pgx.NullTime

	for rows.Next() {
		err := rows.Scan(&sda, &sdaCount, &created)
		if err != nil {
			log.Println("Error routes/feedbacks.go GetFeedbackTopSDAByUserID scanning SDA value:", err)
			return nil
		}
		if sda.Valid {
			returnVal = append(returnVal, sda.String)
		}
	}

	return returnVal
}

func GetUniversalReputationScoreByUserID(userIDparam int64) (int16, int16, []string) {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select * from get_user_ur($1::BIGINT)", userIDparam)
	if row == nil {
		log.Println("Error in get reputation score  ")
		return -1, -1, nil
	}
	var sdaText []string
	var score int16
	var count int16
	err := row.Scan(&score, &count, &sdaText)
	if err != nil {
		log.Println("Error in scanning reputation score: ", err)
		return -1, 0, nil
	}

	return score, count, sdaText
}

func GetUniversalReputationScoreByProfileID(profileIDparam int64) (int16, int16, []string) {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select * from get_profile_ur($1::BIGINT)", profileIDparam)
	if row == nil {
		log.Println("Error in get reputation score  ")
		return -1, 0, nil
	}
	var sdaText []string
	var score int16
	var count int16
	err := row.Scan(&score, &count, &sdaText)
	if err != nil {
		log.Println("Error in scanning reputation score: ", err)
		return -1, 0, nil
	}

	return score, count, sdaText
}
