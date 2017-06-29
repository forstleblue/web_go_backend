package user

import (
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const userTableName = "users"

//DATABASE FUNCTIONS

//InitialRegister inserts basic user information (name, email, password) to the database
func (u *User) InitialRegister() (int64, error) {
	var id int64

	conn := db.Connect()
	defer conn.Close()

	u.Created = time.Now().UTC() //All dates must always be saved to DB in UTC
	u.Updated = time.Now().UTC() //All dates must always be saved to DB in UTC

	phones := pgx.Hstore{}
	if u.Phones != nil {
		phones = map[string]string(u.Phones)
	}

	// TODO setup last login
	lastLogin := pgx.NullHstore{}

	geoField := ""
	geoText := ""

	if u.Lat != 0 && u.Lng != 0 {
		geoField = ",location"
		geoText = ",ST_GeographyFromText('SRID=4326;POINT(" + strconv.FormatFloat(u.Lng, 'f', 6, 64) + " " + strconv.FormatFloat(u.Lat, 'f', 6, 64) + ")')"
	}

	err := conn.QueryRow("Insert Into Users ("+fieldList(false, false)+geoField+") Values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21"+geoText+") Returning user_id",
		u.FirstName, u.MiddleName, u.LastName, u.Email, u.FacebookID, u.CreditCardID, u.CreditCardMask, u.Password, phones, u.PhotoURL, lastLogin, u.Created, u.Updated, u.Roles, u.Address, u.City, u.Region, u.Postcode, u.Country, u.PayoutType, u.PayoutAccount).Scan(&id)
	//_, err := conn.Exec("Insert Into Users (first_name, last_name, email, password, created, updated) Values ($1,$2,$3,$4,$5,$6)", u.FirstName, u.LastName, u.Email, u.Password, u.Created, u.Updated)
	u.UserID = id
	return id, err
}

//Update saves the user to the database
func (u User) Update() error {
	conn := db.Connect()
	defer conn.Close()

	u.Updated = time.Now().UTC() //All dates must always be saved to DB in UTC

	geoText := ""

	if u.Lat != 0 && u.Lng != 0 {
		geoText = ", location=ST_GeographyFromText('SRID=4326;POINT(" + strconv.FormatFloat(u.Lng, 'f', 6, 64) + " " + strconv.FormatFloat(u.Lat, 'f', 6, 64) + ")')"
	}

	_, err := conn.Exec("Update Users Set first_name=$1, middle_name=$2, last_name=$3, email=$4, facebook_id=$5, credit_card_id=$6, credit_card_mask=$7, password=$8, photo_url=$9, address=$10, city=$11, region=$12, postcode=$13, country=$14, payout_type=$15, payout_account=$16, created=$17, updated=$18, password_reset_token=$19"+geoText+" where user_id = $20",
		u.FirstName, u.MiddleName, u.LastName, u.Email, u.FacebookID, u.CreditCardID, u.CreditCardMask, u.Password, u.PhotoURL, u.Address, u.City, u.Region, u.Postcode, u.Country, u.PayoutType, u.PayoutAccount, u.Created, u.Updated, u.PasswordResetToken.String(), u.UserID)
	return err
}

//GetUser returns a User{} struct by ID and error
func GetUser(id int64) (*User, error) {

	conn := db.Connect()
	defer conn.Close()
	user, err := scanSingleUser(conn.QueryRow("select "+fieldList(true, true)+" from "+userTableName+" where user_id = $1", id))

	if err != nil {
		log.Println("No user found in DB by GetUser() with value '" + string(id) + "'")
		return nil, err
	}
	return user, err
}

//GetUserByEmail returns a User{} struct by Email and error
func GetUserByEmail(email string) (*User, error) {

	conn := db.Connect()
	defer conn.Close()
	user, err := scanSingleUser(conn.QueryRow("select "+fieldList(true, true)+" from "+userTableName+" where email = $1", email))

	if err != nil {
		log.Println("No user found in DB by GetUserByEmail() with value '" + email + "'")
		return nil, err
	}
	return user, err
}

//GetUserByEmailOrFacebookID returns a User{} struct by (Email or FacebookID) and error
func GetUserByEmailOrFacebookID(email, facebookid string) *User {

	conn := db.Connect()
	defer conn.Close()

	user, err := scanSingleUser(conn.QueryRow("select "+fieldList(true, true)+" from "+userTableName+" where email = $1 or facebook_id = $2", email, facebookid))

	if err != nil {
		log.Println("No user found in DB by GetUserByEmailOrFacebookID() with values '" + email + "' and '" + facebookid + "'")
		return nil
	}
	return user
}

//EmailExists returns true if email already exists
func EmailExists(email []byte) (bool, error) {

	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select exists(select 1 from users where email = $1 limit 1)", email)
	if err != nil {
		return false, err
	}

	exists := false
	rows.Next()
	err = rows.Scan(&exists)
	return exists, err
}

//EmailUUIDExists returns true if email, uuid already exists
func EmailUUIDExists(email []byte, token string) (bool, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select exists(select 1 from users where email = $1 AND password_reset_token=$2 limit 1)", email, token)
	if err != nil {
		return false, err
	}

	exists := false
	rows.Next()
	err = rows.Scan(&exists)
	return exists, err
}

func GetUserFullName(id int64) string {

	conn := db.Connect()
	defer conn.Close()

	user, err := scanSingleUser(conn.QueryRow("select "+fieldList(true, true)+" from "+userTableName+" where user_id = $1", id))
	var fullName string
	if err != nil {
		log.Println("Can not find User Name in Database")
	}

	fullName = user.FirstName + " " + user.LastName

	return fullName
}

// HELPER METHODS
func fieldList(withID bool, withLocation bool) string {
	list := "first_name, middle_name, last_name, email, facebook_id, credit_card_id, credit_card_mask, password, phones, photo_url, last_login, created, updated, roles, address, city, region, postcode, country, payout_type, payout_account"

	if withID {
		list = "user_id, " + list
	}
	if withLocation {
		list = list + ", ST_X(location::geometry) as lng, ST_Y(location::geometry) as lat"
	}
	return list
}

func scanSingleUser(row *pgx.Row) (*User, error) {
	var userID pgx.NullInt64
	var firstName, middleName, lastName, email, facebookID, creditCardID, creditCardMask, password, photoURL, address, city, region, postcode, country, payoutType, payoutAccount pgx.NullString
	var lat, lng pgx.NullFloat64
	var phones pgx.NullHstore
	var lastLogin pgx.NullHstore
	var created, updated pgx.NullTime
	var roles []string
	err := row.Scan(&userID, &firstName, &middleName, &lastName, &email, &facebookID, &creditCardID, &creditCardMask, &password, &phones, &photoURL, &lastLogin, &created, &updated, &roles, &address, &city, &region, &postcode, &country, &payoutType, &payoutAccount, &lng, &lat)

	if err != nil {
		log.Println("Error scanning single user row: ", err.Error())
		return nil, err
	}

	user := &User{}
	if userID.Valid {
		user.UserID = userID.Int64
	}
	if firstName.Valid {
		user.FirstName = firstName.String
	}
	if middleName.Valid {
		user.MiddleName = middleName.String
	}
	if lastName.Valid {
		user.LastName = lastName.String
	}
	if email.Valid {
		user.Email = email.String
	}
	if facebookID.Valid {
		user.FacebookID = facebookID.String
	}
	if creditCardID.Valid {
		user.CreditCardID = creditCardID.String
	}
	if creditCardMask.Valid {
		user.CreditCardMask = creditCardMask.String
	}
	if password.Valid {
		user.Password = password.String
	}
	if phones.Valid {
		phonesTmp := make(map[string]string)
		for key, value := range phones.Hstore {
			if value.Valid {
				phonesTmp[key] = value.String
			} else {
				phonesTmp[key] = ""
			}
		}
		user.Phones = phonesTmp
	}
	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}
	if address.Valid {
		user.Address = address.String
	}
	if city.Valid {
		user.City = city.String
	}
	if region.Valid {
		user.Region = region.String
	}
	if postcode.Valid {
		user.Postcode = postcode.String
	}
	if country.Valid {
		user.Country = country.String
	}
	if lat.Valid {
		user.Lat = lat.Float64
	}
	if lng.Valid {
		user.Lng = lng.Float64
	}
	if payoutType.Valid {
		user.PayoutType = payoutType.String
	}
	if payoutAccount.Valid {
		user.PayoutAccount = payoutAccount.String
	}
	if lastLogin.Valid {
		lastLoginTmp := make(map[string]LoginDetails)
		for key, value := range lastLogin.Hstore {
			ld := LoginDetails{}
			if value.Valid {
				// I'm not sure how this will come out of the db yet to know how to set the fields
				// TODO set this data
				lastLoginTmp[key] = ld
			} else {
				lastLoginTmp[key] = ld
			}
		}
		user.LastLogin = lastLoginTmp
	}
	if created.Valid {
		user.Created = created.Time
	}
	if updated.Valid {
		user.Updated = updated.Time
	}
	user.Roles = roles
	return user, nil
}
