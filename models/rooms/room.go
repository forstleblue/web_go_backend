package rooms

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/models/user"
)

type Room struct {
	RoomId    int64
	RoomOwner int64
	Users     []int64
	Created   time.Time
	RoomUUID  gocql.UUID
}

func (m Room) GetOtherProfilesByCurrentUserID(currentUserID int64) []int64 {
	var userIdIndex int64
	users := m.Users
	userIdIndex = -1
	profiles, err := user.GetProfileByUserID(currentUserID)
	if err != nil {
		log.Println("Error: in app/models/rooms/room.go/GetOtherProfilesByCurrentUserID(): can't get profile by userId ", err)
	}

	for i := 0; i < len(profiles); i++ {
		for p, v := range users {
			if v == profiles[i].ProfileID {
				userIdIndex = int64(p)
			}
		}
	}

	if userIdIndex == -1 {
		log.Println("Error: in app/models/rooms/room.go/GetOtherProfilesByCurrentUserID() finding profiles in Users: ")
		return nil
	}
	return append(users[:userIdIndex], users[userIdIndex+1:]...)
}

func (m Room) GetCurrentProfileIDByCurrentUserID(currentUserID int64) int64 {
	var profileId int64
	users := m.Users
	profileId = -1
	profiles, err := user.GetProfileByUserID(currentUserID)
	if err != nil {
		log.Println("Error in app/models/rooms/room.go/GetOtherProfileIDByCurrentUserID() get profiles By userId: ", err)
	}
	for i := 0; i < len(profiles); i++ {
		for p, v := range users {
			if v == profiles[i].ProfileID {
				profileId = users[p]
			}
		}
	}

	if profileId == -1 {
		log.Println("Error: in app/models/rooms/room.go/GetOtherProfileIDByCurrentUserID() finding profile in users: ")
		return profileId
	}
	return profileId
}
