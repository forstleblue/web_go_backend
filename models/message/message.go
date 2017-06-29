package message

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/unirep/ur-local-web/app/models/rooms"
	"github.com/unirep/ur-local-web/app/models/user"
)

type Message struct {
	MessageID   int64      `json:"message_id"`
	RoomUUID    string     `json:"room_uuid"`
	MessageText string     `json:"message_text"`
	ProfileID   int64      `json:"user_id"`
	Created     time.Time  `json:"created"`
	Unread      []int64    `json:"unread"`
	RoomID      int64      `json:"room_id"`
	MessageUUID gocql.UUID `json:"message_uuid"`
}

func (m Message) GetUser(ProfileID int64) *user.User {
	profile, _ := user.GetProfile(ProfileID)
	user := profile.User
	return &user
}

func (m Message) ReduceLengthOfMessageText(i int) string {
	runes := []rune(m.MessageText)
	if len(runes) > i {
		result := string(runes[:i])
		result = result + "..."
		return result
	}
	return m.MessageText
}

func (m Message) GetRoom() *rooms.Room {
	room, _ := rooms.GetRoomByRoomUUID(m.RoomUUID)
	return room
}

func (m Message) GetUserByUserID(UserID int64) *user.User {
	user, _ := user.GetUser(UserID)
	return user
}

func (m Message) GetProfileByProfileID(ProfileID int64) *user.Profile {
	profile, _ := user.GetProfile(ProfileID)
	return profile
}
