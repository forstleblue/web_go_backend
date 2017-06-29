package user

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/models/feedback"
)

// Profile is a
type Profile struct {
	ProfileID          int64                    `json:"profile_id"` //todo make this a type 1 GUID from the gocql lib
	User               User                     `json:"user"`
	Title              string                   `json:"title"`
	ServiceCategory    int64                    `json:"service_category"`
	Heading            string                   `json:"heading"`
	ProfileType        string                   `json:"profile_type,omitempty"`
	Tags               []string                 `json:"tags"`
	PhotoURL           string                   `json:"photo_url,omitempty"`
	Description        string                   `json:"description"`
	Fee                string                   `json:"fee"`
	PaymentNotes       string                   `json:"paymentNotes"`
	FeedbackRating     int32                    `json:"feedback_rating"`
	ReputationStatus   int32                    `json:"reputation_status"`
	Created            time.Time                `json:"created,omitempty" db:"created"`
	Updated            time.Time                `json:"updated,omitempty" db:"updated"`
	ProfileUUID        gocql.UUID               `json:"profile_uuid" db:"profile_uuid"`
	OauthToken         string                   `json:"ouath_token"`
	OauthExpiry        time.Time                `json:"oauth_expiry"`
	ExternalID         string                   `json:"external_id"`
	FeedbackAvergeInfo feedback.FeedbackAverage `json:"feedback_average"`
}

func (p Profile) String() string {
	return fmt.Sprintf("{ProfileID:%d, User:%s, Title: %s,  Tags:%s, PhotoURL:%s, Description:%s, Fee:%s, PaymentNotes: %s, ServiceCategory: %d Created:%s, Updated:%s, ProfileUUID:%s, ExternalID:%s}",
		p.ProfileID, p.User.String(), p.Title, p.Tags, p.PhotoURL, p.Description, p.Fee, p.PaymentNotes, p.ServiceCategory, p.Created, p.Updated, p.ProfileUUID, p.ExternalID)
}

//HeadingDisplay returns either the heading if set, or a default heading
func (p Profile) HeadingDisplay() string {
	if p.Heading != "" {
		return p.Heading
	}
	var heading string
	if p.ProfileType == "s" {
		heading = "Seller"
	} else if p.ProfileType == "p" {
		heading = "Provider"
	}
	return heading
}

//TitleDisplay returns either the title if set, or a default title
func (p Profile) TitleDisplay() string {
	if p.Title != "" {
		return p.Title
	}
	var title string
	if p.ProfileType == "s" {
		title = "Seller"
	} else if p.ProfileType == "p" {
		title = "Provider"
	}
	return title
}

//SinceDateFormatted returns the date that should show as since joined, formatted
func (p Profile) SinceDateFormatted() string {
	return p.User.Created.Format("2 Jan 2006")
}

//CreateDateFormatted returns the date that should show as since joined, formatted
func (p Profile) CreateDateFormatted() string {
	return p.Created.Format("2 Jan 2006")
}

//ShortDescription returns a shortened value for the description to fit in the preview area
func (p Profile) ShortDescription() string {
	var numRunes = 0
	for index := range p.Description {
		numRunes++
		if numRunes > 90 {
			return p.Description[:index] + "..."
		}
	}
	return p.Description
}

//DisplayPhoto The url based on if the PhotoURL is set, Facebook, or local
func (p Profile) DisplayPhoto() string {
	photoURL := ""
	if p.PhotoURL == "" {
		photoURL = p.User.DisplayPhoto()
	} else if strings.HasPrefix(p.PhotoURL, "http") || strings.HasPrefix(p.PhotoURL, "//") {
		photoURL = p.PhotoURL
	} else {
		photoURL = "/images/profile-photos/" + strconv.FormatInt(p.User.UserID, 10) + "/" + p.PhotoURL
	}
	return photoURL
}

func (p *Profile) DistanceFrom(lat, lng float64) int64 {
	if (p.User.Lat == 0 && p.User.Lng == 0) || (lat == 0 && lng == 0) {
		return -1
	}
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat * math.Pi / 180
	lo1 = lng * math.Pi / 180
	la2 = p.User.Lat * math.Pi / 180
	lo2 = p.User.Lng * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return int64((2 * r * math.Asin(math.Sqrt(h))) / 1000)
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

//FeedbackCount is
func (p Profile) FeedbackCount() int64 {
	return feedback.GetFeedbackCountByProfileID(p.ProfileID)
}

//FeedbackAverage returns average feedback score
func (p Profile) FeedbackAverage() int16 {
	return feedback.GetFeedbackAverageByProfileID(p.ProfileID)
}

//FeedbackDescription returns feedback description
func (p Profile) FeedbackDescription() string {
	average := p.FeedbackAverage()
	description := ""
	if average < 25 {
		description = "Unacceptable"
	} else if average < 40 {
		description = "Needs Improvement"
	} else if average < 55 {
		description = "Acceptable"
	} else if average < 70 {
		description = "Met Expectations"
	} else if average < 85 {
		description = "Exceeds Expectations"
	} else if average < 100 {
		description = "Excellent"
	} else {
		// 100%
		description = "Exceptional"
	}
	return description
}

func (p Profile) FeedbackTopSDA() []string {
	return feedback.GetFeedbackTopSDAByProfileID(p.ProfileID)
}

//GetProfileServiceName returns profile service name
func (p Profile) GetProfileServiceName(serviceCategory int64) string {
	serviceName := GetProfileServiceNameByID(serviceCategory)
	return serviceName
}

//UniversalReputationScore returns reputation information
func (p Profile) UniversalReputationScore() *feedback.FeedbackAverage {
	averageScore, count, sdaText := feedback.GetUniversalReputationScoreByProfileID(p.ProfileID)
	return &feedback.FeedbackAverage{
		Count:     count,
		Score:     averageScore / 20,
		SdaString: sdaText,
	}
}

//GetProfileType returns profile type
func GetProfileType(id int64) string {
	profileType := GetProfileTypeWithID(id)
	return profileType
}
