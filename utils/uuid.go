package utils

import (
	"time"

	"github.com/gocql/gocql"
)

//HideTimeUUID alters time based UUIDs so it is obvious what their value is
func HideTimeUUID(uuid gocql.UUID) (gocql.UUID, error) {
	u := uuid.String()
	sU := u[:2] + u[5:7] + u[4:5] + u[2:4] + u[7:14] + "4" + u[15:]

	return gocql.ParseUUID(sU)
}

//UnhideTimeUUID reverses UUIDs formated with the HideTimeUUIDs func
func UnhideTimeUUID(uuid string) (gocql.UUID, error) {
	u := uuid[:2] + uuid[5:7] + uuid[4:5] + uuid[2:4] + uuid[7:14] + "1" + uuid[15:]
	return gocql.ParseUUID(u)
}

//IsHiddenUUIDValid allows passing in a string uuid hidden by HideTimeUUIDs and validates timing
func IsHiddenUUIDValid(uuidStr string, withinHours int64) (bool, error) {
	uuid, err := UnhideTimeUUID(uuidStr)
	if err != nil {
		return false, err
	}

	generateTime := uuid.Time().UTC()
	expiry := generateTime.Add(time.Duration(withinHours) * time.Hour)
	curTime := time.Now().UTC()

	valid := curTime.Before(expiry)
	return valid, nil
}
