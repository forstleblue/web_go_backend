package user

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/models/feedback"
)

const (
	//RoleUser the const for role user
	RoleUser = "user"
	//RoleDeveloper the const for role developer
	RoleDeveloper = "developer"
	//RoleAdmin the const for role admin
	RoleAdmin = "admin"
)

//LoginDetails holds all important login information
type LoginDetails struct {
	UserAgent string
	Timestamp time.Time
}

//User is a CableStore website user
type User struct {
	UserID             int64                   `json:"user_id" db:"user_id"` //todo make this a type 1 GUID from the gocql lib
	FirstName          string                  `json:"first_name" db:"first_name"`
	MiddleName         string                  `json:"middle_name,omitempty" db:"middle_name"`
	LastName           string                  `json:"last_name,omitempty" db:"last_name"`
	Email              string                  `json:"email" db:"email"`
	DateOfBirth        time.Time               `json:"date_of_birth,omitempty" db:"date_of_birth"`
	FacebookID         string                  `json:"facebook_id,omitempty" db:"facebook_id"`
	PaypalID           string                  `json:"paypal_id,omitempty" db:"paypal_id"`
	PhotoURL           string                  `json:"photo_url,omitempty" db:"photo_url"`
	CreditCardID       string                  `json:"credit_card_id,omitempty" db:"credit_card_id"`
	CreditCardMask     string                  `json:"credit_card_mask,omitempty" db:"credit_card_mask"`
	PayoutType         string                  `json:"payout_type,omitempty" db:"payout_type"`
	PayoutAccount      string                  `json:"payout_account,omitempty" db:"payout_account"`
	Password           string                  `json:"password,omitempty" db:"password"`
	Phones             map[string]string       `json:"phones,omitempty" db:"phones"`
	Address            string                  `json:"address,omitempty" db:"address"`
	City               string                  `json:"city,omitempty" db:"city"`
	Region             string                  `json:"region,omitempty" db:"region"`
	Postcode           string                  `json:"postcode,omitempty" db:"postcode"`
	Country            string                  `json:"country,omitempty" db:"country"`
	Lat                float64                 `json:"lat,omitempty" db:"lat"`
	Lng                float64                 `json:"lng,omitempty" db:"lng"`
	LastLogin          map[string]LoginDetails `json:"last_login,omitempty" db:"last_login"`
	Roles              []string                `json:"roles,omitempty" db:"roles"`
	Created            time.Time               `json:"created,omitempty" db:"created"`
	Updated            time.Time               `json:"updated,omitempty" db:"updated"`
	PasswordResetToken gocql.UUID              `json:"passwordResetToken,omitempty" db:"passwordResetToken"`
	UUID               gocql.UUID              `json:"uuid,omitempty" db:"uuid"`
}

//SafeUser is a UR website user without private information
type SafeUser struct {
	UserID    int64
	Email     string
	FirstName string
	LastName  string
	Phones    map[string]string
	Postcode  string
	LastLogin map[string]LoginDetails
}

//GetSafeUser returns a UR website user without private information
func (u User) GetSafeUser() *SafeUser {
	safeUser := &SafeUser{
		UserID:    u.UserID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phones:    u.Phones,
		Postcode:  u.Postcode,
	}
	return safeUser
}

func (u User) hasRole(role string) bool {
	for _, b := range u.Roles {
		if b == role {
			return true
		}
	}
	return false
}

//IsCustomer returns whether the useris in the "customer" role
func (u User) IsCustomer() bool {
	return u.hasRole("customer")
}

//IsAdmin returns whether the useris in the "admin" role
func (u User) IsAdmin() bool {
	return u.hasRole("admin")
}

//IsDev returns whether the useris in the "developer" role
func (u User) IsDev() bool {
	return u.hasRole("developer")
}

//HasFacebook returns true if user has linked a Facebook account
func (u User) HasFacebook() bool {
	return u.FacebookID != ""
}

//IsFacebookPhoto returns true if user is using Facebook photo
func (u User) IsFacebookPhoto() bool {
	return strings.HasPrefix(u.PhotoURL, "http://graph.facebook.com")
}

//FullName gets user's full name
func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}

//FacebookPhotoURL returns full URL to Facebook photo if has connected to Facebook
func (u User) FacebookPhotoURL() string {
	if u.HasFacebook() {
		return "http://graph.facebook.com/" + u.FacebookID + "/picture?width=270&height=270"
	}
	return ""
}

//DisplayPhoto The url based on if the PhotoURL is set, Facebook, or local
func (u *User) DisplayPhoto() string {
	photoURL := ""
	if u.PhotoURL == "" {
		photoURL = "/images/default-avatar.png"
	} else if u.IsFacebookPhoto() {
		photoURL = u.PhotoURL
	} else {
		photoURL = "/images/profile-photos/" + strconv.FormatInt(u.UserID, 10) + "/" + u.PhotoURL
	}
	return photoURL
}

func (u *User) GetPayoutAccount() string {
	if u.PayoutType == "PPEMAIL" || u.PayoutType == "" {
		if u.PayoutAccount == "" {
			return u.Email
		}
	}
	return u.PayoutAccount
}

func (u *User) FormattedAddress() string {
	address := ""
	if u.Address != "" {
		address += u.Address + " "
	}
	address += u.City + " " + u.Region + " " + u.Country
	return address
}

func (u *User) FeedbackCount() int64 {
	return feedback.GetFeedbackCountByUserID(u.UserID)
}

func (u *User) FeedbackAverage() int16 {
	return feedback.GetFeedbackAverageByUserID(u.UserID)
}

func (u *User) FeedbackDescription() string {
	average := u.FeedbackAverage()
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

func (u *User) FeedbackTopSDA() []string {
	return feedback.GetFeedbackTopSDAByUserID(u.UserID)
}

func (u User) String() string {
	return fmt.Sprintf("{UserID:%d, Email:%s, Roles:%s}", u.UserID, u.Email, u.Roles)
}
