package user

import (
	"log"
	"strconv"

	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
	"github.com/unirep/ur-local-web/app/utils"
)

const profileTableName = "profiles"

//DATABASE FUNCTIONS

//InsertProfile saves the profile to the database
func InsertProfile(p *Profile) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()

	p.ProfileUUID = gocql.TimeUUID()
	p.Created = time.Now().UTC() //All dates must always be saved to DB in UTC
	p.Updated = p.Created        //All dates must always be saved to DB in UTC

	log.Println("Inserting profile to db: ", p.User.UserID, p.Title, p.PhotoURL, p.Description, p.FeedbackRating, p.ReputationStatus, p.Fee, p.PaymentNotes, p.ServiceCategory, p.ProfileType, p.Heading, p.Created, p.Updated)
	err := conn.QueryRow("insert into "+profileTableName+" (user_id, title, photo_url, description, feedback_rating, reputation_status, fee, payment_notes, service_category, profile_type, heading, created, updated, oauth_token, oauth_expiry, external_id, profile_uuid) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) Returning profile_id",
		p.User.UserID, p.Title, p.PhotoURL, p.Description, p.FeedbackRating, p.ReputationStatus, p.Fee, p.PaymentNotes, p.ServiceCategory, p.ProfileType, p.Heading, p.Created, p.Updated, p.OauthToken, p.OauthExpiry, p.ExternalID, p.ProfileUUID.String()).Scan(&id)

	p.ProfileID = id
	if err == nil {
		insertTags(id, p.Tags)
	} else {
		log.Printf("Error in models/user/profile-db.go InsertProfile(p *Profile): InsertProfile Error %s\n", err.Error())
	}
	return id, err
}

//DeleteProfile delete the profile from the database
func DeleteProfile(id int64) (int64, error) {
	conn := db.Connect()
	defer conn.Close()

	commandTag, err := conn.Exec("delete from "+profileTableName+" where profile_id=$1", id)
	if err != nil {
		return -1, err
	}
	deleteTags(id)

	return commandTag.RowsAffected(), err
}

//Update saves the profile to the database
func (p Profile) Update() error {
	conn := db.Connect()
	defer conn.Close()

	p.Updated = time.Now()

	_, err := conn.Exec("Update "+profileTableName+" Set title=$1, photo_url=$2, description=$3, feedback_rating=$4, reputation_status=$5, fee=$6, payment_notes=$7, service_category=$8, heading=$9, external_id=$10, created=$11, updated=$12 where profile_id = $13",
		p.Title, p.PhotoURL, p.Description, p.FeedbackRating, p.ReputationStatus, p.Fee, p.PaymentNotes, p.ServiceCategory, p.Heading, p.ExternalID, p.Created, p.Updated, p.ProfileID)

	if err == nil {
		updateTags(p.ProfileID, p.Tags)
	}
	return err
}

//GetProfile returns a Profile{} struct by ID and error
func GetProfile(id int64) (*Profile, error) {

	conn := db.Connect()
	defer conn.Close()

	profile, err := scanSingleProfile(conn.QueryRow("select "+profileFieldList(true)+" from "+profileTableName+" p where  profile_id = $1", id))

	if err != nil {
		log.Println("No profile found in DB by GetProfile() with value '" + string(id) + "'")
		return nil, err
	}
	return profile, err
}

// EbayProfileAdded return true if the user added Ebay Profile
func EbayProfileAdded(id int64) bool {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select profile_type from "+profileTableName+" where user_id=$1 AND profile_type='e'", id)
	var profileType pgx.NullString
	err := row.Scan(&profileType)
	if err != nil {
		return false
	}

	var profileAdded bool
	if profileType.Valid == false {
		profileAdded = false
	} else {
		profileAdded = true
	}
	return profileAdded
}

// GetEbayProfileUUID returns UUID of ebay profile
func GetEbayProfileUUID(id int64) string {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select profile_uuid from "+profileTableName+" where user_id=$1 AND profile_type='e'", id)
	var profileUUID pgx.NullString
	err := row.Scan(&profileUUID)
	if err != nil {
		log.Println("Error in models/user/profile-db.go EbayProfileAdded(): ", err)
		return ""
	}
	if profileUUID.Valid == false {
		return ""
	}
	return profileUUID.String
}

//GetProfileByUserID returns a list of Profile{} struct by user id and error
func GetProfileByUserID(userID int64) ([]*Profile, error) {

	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, "+userTableName+" u where p.user_id = u.user_id AND u.user_id = $1 order by profile_id", userID)
	if err != nil {
		log.Println("Error querying for profiles by user: ", err)
		return nil, err
	}
	profiles, err := scanMultipleProfiles(rows)
	if err != nil {
		log.Println("Error scanning multiple profiles for user: ", err)
		return nil, err
	}
	return profiles, nil
}

//GetProfileByUser returns a list of Profile{} struct by user and error
func GetProfileByUser(user *User) ([]*Profile, error) {
	return GetProfileByUserID(user.UserID)
}

//GetProfileByUserIDAndProfileType returns a list of Profile{} struct by user id and profile type and error
func GetProfileByUserIDAndProfileType(userID int64, profileType string) ([]*Profile, error) {

	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, "+userTableName+" u where p.user_id = u.user_id AND u.user_id = $1 AND p.profile_type=$2 order by profile_id", userID, profileType)
	if err != nil {
		log.Println("Error querying for profiles by user: ", err)
		return nil, err
	}
	profiles, err := scanMultipleProfiles(rows)
	if err != nil {
		log.Println("Error scanning multiple profiles for user: ", err)
		return nil, err
	}
	return profiles, nil
}

//GetLatestProfiles returns a list of Profile{} struct by number wanted ordered by create date and error
func GetLatestProfiles(count int64) ([]*Profile, error) {

	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, Users u where u.user_id = p.user_id AND u.postcode != '' AND profile_type != 'b' order by p.created desc LIMIT $1", count)
	if err != nil {
		log.Println("Error querying for latest profiles: ", err)
		return nil, err
	}
	profiles, err := scanMultipleProfiles(rows)
	if err != nil {
		log.Println("Error scanning returning profiles after querying for latest profiles: ", err)
		return nil, err
	}
	return profiles, nil
}

//SearchProfiles returns a list of Profile{} struct by several search params and error
func SearchProfiles(serviceCategory []int64, tags []string, loc string) ([]*Profile, error) {

	conn := db.Connect()
	defer conn.Close()

	var rows *pgx.Rows
	var err error
	var lat, lng float64
	var locQuery string

	if loc != "" {
		lat, lng, err = utils.GetLatLng(loc)
		locQuery = "ST_DWithin(u.location, ST_SetSRID(ST_MakePoint(" + strconv.FormatFloat(lng, 'f', 6, 64) + ", " + strconv.FormatFloat(lat, 'f', 6, 64) + "), 4326), 20000)"
	}

	if len(serviceCategory) != 0 && len(loc) != 0 {
		rows, err = conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, Users u where p.user_id = u.user_id AND profile_id in (select profile_id from "+profileTableName+" p, Users u where u.user_id = p.user_id AND u.postcode != '' AND service_category = any($1)  union select profile_id from tags where tag ilike any($2)) AND "+locQuery+" order by p.created desc", serviceCategory, tags)
	} else if len(serviceCategory) != 0 {
		rows, err = conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, Users u where p.user_id = u.user_id AND profile_id in (select profile_id from "+profileTableName+" p, Users u where u.user_id = p.user_id AND u.postcode != '' AND service_category = any($1)  union select profile_id from tags where tag ilike any($2)) order by p.created desc", serviceCategory, tags)
	} else if len(loc) != 0 {
		rows, err = conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, Users u where p.user_id = u.user_id AND profile_id in (select profile_id from "+profileTableName+" p, Users u where  u.user_id = p.user_id AND u.postcode != '' AND profile_type != 'b' union select profile_id from tags where tag ilike any($1)) AND "+locQuery+" order by p.created desc", tags)
	} else {
		rows, err = conn.Query("select "+profileFieldList(true)+" from "+profileTableName+" p, Users u where p.user_id = u.user_id AND profile_id in (select profile_id from tags where tag ilike any($1)) order by p.created desc", tags)
	}

	if err != nil {
		log.Println("Error searching for profiles: ", err)
		return nil, err
	}

	profiles, err := scanMultipleProfiles(rows)

	if err != nil {
		log.Println("Error scanning returning profiles after searching for profiles: ", err)
		return nil, err
	}
	return profiles, nil
}

//GetTagsBySearch returns a list of tags
func GetTagsBySearch(q string) []string {
	conn := db.Connect()
	defer conn.Close()

	var tag pgx.NullString
	tags := []string{}
	q = q + "%"

	rows, err := conn.Query("select distinct tag from tags where tag LIKE $1", q)
	if err != nil {
		log.Println("Error querying for available tags: ", err)
		return tags
	}
	for rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			log.Println("Error scanning tags in getTags: ", err.Error())
			return tags
		}
		if tag.Valid {
			tags = append(tags, tag.String)
		}
	}
	return tags
}

// HELPER METHODS
func profileFieldList(withID bool) string {
	list := "p.user_id, title, p.photo_url, description, feedback_rating, reputation_status, fee, payment_notes, service_category,  p.profile_type, p.heading, p.created, p.updated, p.profile_uuid, p.external_id"

	if withID {
		list = "profile_id, " + list
	}
	return list
}

func scanMultipleProfiles(rows *pgx.Rows) ([]*Profile, error) {
	profiles := []*Profile{}
	var err error
	for rows.Next() {
		var profileID, userID, serviceCategory pgx.NullInt64
		var title, description, photoURL, fee, paymentNotes, profileType, profileHeading, externalID pgx.NullString
		var feedbackRating, reputationStatus pgx.NullInt32
		var created, updated pgx.NullTime
		var profileUUID string
		err = rows.Scan(&profileID, &userID, &title, &photoURL, &description, &feedbackRating, &reputationStatus, &fee, &paymentNotes, &serviceCategory, &profileType, &profileHeading, &created, &updated, &profileUUID, &externalID)
		if err != nil {
			log.Println("Error scanning profile row in scanMultipleProfiles: ", err.Error())
			return nil, err
		}

		tags := getTags(profileID.Int64)

		profile := fillProfile(profileID, userID, title, description, photoURL, fee, paymentNotes, profileType, serviceCategory, profileHeading, feedbackRating, reputationStatus, created, updated, tags, profileUUID, externalID)
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func scanSingleProfile(row *pgx.Row) (*Profile, error) {
	var profileID, userID, serviceCategory pgx.NullInt64
	var title, description, photoURL, fee, paymentNotes, profileType, profileHeading, externalID pgx.NullString

	var feedbackRating, reputationStatus pgx.NullInt32
	var created, updated pgx.NullTime
	var profileUUID string
	//p.user_id, title, p.photo_url, description, feedback_rating, reputation_status, fee, payment_notes, service_category, p.created, p.updated
	err := row.Scan(&profileID, &userID, &title, &photoURL, &description, &feedbackRating, &reputationStatus, &fee, &paymentNotes, &serviceCategory, &profileType, &profileHeading, &created, &updated, &profileUUID, &externalID)

	if err != nil {
		log.Println("Error scanning profile row in scanSingleProfile: ", err.Error())
		return nil, err
	}

	tags := getTags(profileID.Int64)

	profile := fillProfile(profileID, userID, title, description, photoURL, fee, paymentNotes, profileType, serviceCategory, profileHeading, feedbackRating, reputationStatus, created, updated, tags, profileUUID, externalID)
	return profile, nil
}

func fillProfile(profileID, userID pgx.NullInt64, title, description, photoURL, fee, paymentNotes, profileType pgx.NullString, serviceCategory pgx.NullInt64, profileHeading pgx.NullString, feedbackRating, reputationStatus pgx.NullInt32, created, updated pgx.NullTime, tags []string, profileUUID string, externalID pgx.NullString) *Profile {
	profile := &Profile{}
	if profileID.Valid {
		profile.ProfileID = profileID.Int64
	}
	if userID.Valid {
		user, _ := GetUser(userID.Int64)
		profile.User = *user
	}
	if title.Valid {
		profile.Title = title.String
	}
	if photoURL.Valid {
		profile.PhotoURL = photoURL.String
	}
	if description.Valid {
		profile.Description = description.String
	}
	if feedbackRating.Valid {
		profile.FeedbackRating = feedbackRating.Int32
	}
	if reputationStatus.Valid {
		profile.ReputationStatus = reputationStatus.Int32
	}
	if fee.Valid {
		profile.Fee = fee.String
	}
	if paymentNotes.Valid {
		profile.PaymentNotes = paymentNotes.String
	}
	if profileType.Valid {
		profile.ProfileType = profileType.String
	}
	if serviceCategory.Valid {
		profile.ServiceCategory = serviceCategory.Int64
	}
	if created.Valid {
		profile.Created = created.Time
	}
	if updated.Valid {
		profile.Updated = updated.Time
	}
	if profileHeading.Valid {
		profile.Heading = profileHeading.String
	}
	if externalID.Valid {
		profile.ExternalID = externalID.String
	}
	profile.Tags = tags
	profile.ProfileUUID, _ = gocql.ParseUUID(profileUUID)
	return profile
}

func getTags(profileID int64) []string {
	conn := db.Connect()
	defer conn.Close()

	tags := []string{}
	var tag pgx.NullString
	rows, err := conn.Query("select tag from tags where profile_id = $1", profileID)
	if err != nil {
		log.Println("Error querying tags in getTags: ", err.Error())
		return tags
	}
	for rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			log.Println("Error scanning tags in getTags: ", err.Error())
			return tags
		}
		if tag.Valid {
			tags = append(tags, tag.String)
		}
	}
	return tags
}

func insertTags(profileID int64, tags []string) {
	conn := db.Connect()

	for _, tag := range tags {
		_, err := conn.Exec("insert into tags (profile_id, tag) values($1, $2)", profileID, tag)
		if err != nil {
			log.Println("Error inserting tag in insertTags: ", err.Error())
		}
	}
	conn.Close()
}

func deleteTags(profileID int64) {
	conn := db.Connect()

	_, err := conn.Exec("delete from tags where profile_id=$1", profileID)
	if err != nil {
		log.Println("Error deleting tags in deleteTags: ", err.Error())
	}
	conn.Close()
}

func updateTags(profileID int64, tags []string) {
	deleteTags(profileID)
	insertTags(profileID, tags)
}

func GetProfileIDsByUserID(userID int64) ([]int64, error) {
	conn := db.Connect()
	defer conn.Close()
	var profileIDs []int64
	rows, err := conn.Query("select profile_id from "+profileTableName+" p, "+userTableName+" u where p.user_id = u.user_id AND u.user_id = $1 order by profile_id", userID)
	if err != nil {
		log.Println("Error querying for profiles by user: ", err)
		return nil, err
	}

	for rows.Next() {
		var profileID pgx.NullInt64
		err = rows.Scan(&profileID)
		if err != nil {
			log.Println("Error scanning profile row in scanMultipleProfiles: ", err.Error())
			return nil, err
		}

		profileIDs = append(profileIDs, profileID.Int64)
	}
	return profileIDs, nil

}

func GetProfileTypeWithID(profileID int64) string {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select profile_type from "+profileTableName+" where profile_id=$1", profileID)
	var profileType pgx.NullString
	err := row.Scan(&profileType)
	if err != nil {
		log.Println("Error in models/user/profile-db.go GetProfileTypeWithID(): ", err)
		return ""
	}

	return profileType.String
}

//GetProfileServiceNameByID returns service name of profile
func GetProfileServiceNameByID(serviceCategory int64) string {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select service_name from "+serviceCategoryTable+" s, "+profileTableName+" p where s.service_id = p.service_category AND p.service_category = $1", serviceCategory)
	var serviceValue pgx.NullString
	err := row.Scan(&serviceValue)
	if err != nil {
		log.Println("Error in user/profile-db.go fail to get service category:", err)
	}
	var serviceName string
	if serviceValue.Valid == false {
		log.Println("Error in user/profile-db.go can't not get profile service name :", err)
		serviceName = ""
	} else {
		serviceName = serviceValue.String
	}
	return serviceName
}

func GetProfileByProfileUUID(profileUUID string) (*Profile, error) {
	conn := db.Connect()
	defer conn.Close()
	profile, err := scanSingleProfile(conn.QueryRow("select "+profileFieldList(true)+" from "+profileTableName+" p where  profile_uuid = $1", profileUUID))

	if err != nil {
		log.Println("No profile found in DB by GetProfile() with value '" + profileUUID + "'")
		return nil, err
	}
	return profile, err
}

func GetProfileByProfileUUIDAndProfileType(profileID, profileType string) (*Profile, error) {
	conn := db.Connect()
	defer conn.Close()
	profile, err := scanSingleProfile(conn.QueryRow("select "+profileFieldList(true)+" from "+profileTableName+" p where profile_uuid = $1 AND profile_type = $2", profileID, profileType))

	if err != nil {
		log.Println("No profile found in DB by GetProfileByUUIDAndProfileType() with value '" + profileID + "'")
		return nil, err
	}
	return profile, err
}

func GetProfileByExternalIDAndProfileType(profileID, profileType string) (*Profile, error) {
	conn := db.Connect()
	defer conn.Close()
	profile, err := scanSingleProfile(conn.QueryRow("select "+profileFieldList(true)+" from "+profileTableName+" p where external_id = $1 AND profile_type = $2", profileID, profileType))

	if err != nil {
		log.Println("No profile found in DB by GetProfileByExternalIDAndProfileType() with value '" + profileID + "'")
		return nil, err
	}
	return profile, err
}
