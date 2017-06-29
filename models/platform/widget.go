package platform

import (
	"time"

	"github.com/gocql/gocql"
)

const (
	WidgetTypeProfile  = "REPUTATION"
	WidgetTypeFeedback = "FEEDBACK"

	OwnerTypePlatform = "PLATFORM"
)

type ReputationWidgetConfiguration struct {
}

type FeedbackWidgetConfiguration struct {
}

type Widget struct {
	WidgetID      gocql.UUID  `json:"widget_id"`
	Type          string      `json:"type"`
	OwnerID       gocql.UUID  `json:"owner_id"`
	OwnerType     string      `json:"owner_type"`
	Configuration interface{} `json:"configuration"`
	Created       time.Time   `json:"created,omitempty" db:"created"`
	Updated       time.Time   `json:"updated,omitempty" db:"updated"`
	SessionToken  gocql.UUID  `json:"session_token"`
}
