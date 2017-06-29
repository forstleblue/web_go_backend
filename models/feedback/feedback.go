package feedback

import (
	"time"

	"github.com/gocql/gocql"
)

//Feedback is feedback struct
type Feedback struct {
	FeedbackID        int64      `json:"feedback_id"`
	BookingID         int64      `json:"booking_id"`
	SenderProfileID   int64      `json:"sender_profile_id"`
	ReceiverProfileID int64      `json:"receiver_profile_id"`
	Description       string     `json:"description"`
	Comment           string     `json:"comment"`
	Score             int16      `json:"score"`
	SdaText           []string   `json:"sda_text"`
	FeedbackUUID      gocql.UUID `json:"feedback_uuid"`
	Created           time.Time  `json:"created,omitempty" db:"created"`
	Positive          bool       `json:"positive"`
	Neutral           bool       `json:"neutral"`
	Negative          bool       `json:"negative"`
}

//FeedbackAverage used showing feedback information in profile card
type FeedbackAverage struct {
	Score     int16    `json:"feedback_score"`
	Count     int16    `json:"feedback_count"`
	SdaString []string `json:"sda_string"`
}

//DateFormatted display formatted date information
func (f *Feedback) DateFormatted() string {
	return f.Created.Format("2, Jan 2006")
}
