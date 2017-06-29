package message

import (
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
	"github.com/unirep/ur-local-web/app/models/user"
)

const messageTableName = "messages"
const roomTableName = "rooms"

func InsertMessage(message *Message) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()
	message.Created = time.Now().UTC()
	err := conn.QueryRow("insert into "+messageTableName+"(room_id, message_text, profile_id, created, unread, room_uuid) values ($1, $2, $3, $4, $5, $6) Returning message_id",
		message.RoomID, message.MessageText, message.ProfileID, message.Created, message.Unread, message.RoomUUID).Scan(&id)
	return id, err
}

func GetMessagesByRoomID(roomID int64) ([]*Message, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select "+"message_id, room_id, message_text, profile_id, created, unread, message_uuid"+" from "+messageTableName+" where room_id=$1 ORDER BY created ASC", roomID)
	if err != nil {
		log.Println("Error in app/models/message/message-db.go GetMessageByRoomID() querying for messages by room_id: ", err)
		return nil, err
	}
	messages, err := scanMultipleMessages(rows)
	if err != nil {
		log.Println("Error in app/models/message/message-db.go GetMessageByRoomID() scanning multiple messages for room_id: ", err)
		return nil, err
	}
	return messages, nil
}

func scanMultipleMessages(rows *pgx.Rows) ([]*Message, error) {
	messages := []*Message{}
	var err error
	for rows.Next() {
		var messageID, roomID, profileID pgx.NullInt64
		var messageText pgx.NullString
		var created pgx.NullTime
		var users []int64
		var roomUUID string
		err = rows.Scan(&messageID, &roomID, &messageText, &profileID, &created, &users, &roomUUID)
		if err != nil {
			log.Println("Error  scanning message row in scanMultipleMessages: ", err)
			return nil, err
		}

		message := fillMessage(messageID, roomID, messageText, profileID, created, users, roomUUID)
		messages = append(messages, message)
	}
	return messages, nil
}

func fillMessage(messageID pgx.NullInt64, roomID pgx.NullInt64, messageText pgx.NullString, profileID pgx.NullInt64, created pgx.NullTime, users []int64, roomUUID string) *Message {
	message := &Message{}
	if messageID.Valid {
		message.MessageID = messageID.Int64
	}
	if roomID.Valid {
		message.RoomID = roomID.Int64
	}
	if messageText.Valid {
		message.MessageText = messageText.String
	}
	if profileID.Valid {
		message.ProfileID = profileID.Int64
	}
	if created.Valid {
		message.Created = created.Time
	}
	if users != nil {
		message.Unread = users
	}
	message.RoomUUID = roomUUID
	return message
}

func GetUnreadUsers(users []int64, userId int64) []int64 {
	var userIdIndex int64
	userIdIndex = -1
	profiles, err := user.GetProfileByUserID(userId)
	if err != nil {
		log.Println("Error:   ", err)
	}
	for i := 0; i < len(profiles); i++ {
		for p, v := range users {
			if v == profiles[i].ProfileID {
				userIdIndex = int64(p)
			}
		}
	}

	if userIdIndex == -1 {
		log.Println("Error: finding user in Unread:")
		return nil
	}
	return append(users[:userIdIndex], users[userIdIndex+1:]...)
}

func GetMessagesByUserIdInUnread(userID int64) ([]*Message, error) {
	conn := db.Connect()
	defer conn.Close()
	profileIDs, _ := user.GetProfileIDsByUserID(userID)
	rows, err := conn.Query("select "+"message_id, room_id, message_text, profile_id, created, unread, room_uuid "+"from "+messageTableName+" where $1 && unread ORDER BY created DESC", profileIDs)
	if err != nil {
		log.Println("Error querying for messages by UnreadUserId: ", err)
		return nil, err
	}
	messages, err := scanMultipleMessages(rows)
	if err != nil {
		log.Println("Error scanning multiple messages for UnreadUserId: ", err)
		return nil, err
	}
	return messages, nil
}

func RemoveCurrentUserFromUnreadMessagesInCurrentRoom(roomID int64, profileID int64) error {
	conn := db.Connect()
	defer conn.Close()
	profileIdString := fmt.Sprintf("%d", profileID)
	_, err := conn.Query("UPDATE "+messageTableName+" SET unread = array_remove(unread, CAST("+profileIdString+" AS BIGINT)) where room_id=$1 AND $2=any(unread)", roomID, profileID)
	if err != nil {
		log.Println("Error updating unread in messageTable: ", err)
	}
	return err
}

func GetMessageIDsToDisplayInDashboard(userID int64, pageNum int64, displayCount int64) ([]*Message, error) {
	conn := db.Connect()
	defer conn.Close()

	profileIDs, _ := user.GetProfileIDsByUserID(userID)

	//rows, err := conn.Query("select max(message_id), "+messageTableName+".room_id, max("+messageTableName+".created) from "+messageTableName+"inner join "+roomTableName+"on "+messageTableName+".room_id = "+roomTableName+".room_id where $1=any("+roomTableName+".users) GROUP BY "+messageTableName+".room_id ORDER BY MAX("+messageTableName+".created) DESC LIMIT $2 OFFSET $3", userID, displayCount, pageNum)
	rows, err := conn.Query("SELECT MAX(message_id), messages.room_id, MAX(messages.created) FROM messages  INNER JOIN rooms ON messages.room_id = rooms.room_id  WHERE $1 && rooms.users GROUP BY messages.room_id ORDER BY MAX(messages.created) DESC LIMIT $2 OFFSET $3", profileIDs, displayCount, pageNum)
	if err != nil {
		log.Println("Error querying for messages to display in dashboard: ", err)
		return nil, err
	}
	messages, err := scanMultipleMessagesToDisplayInDashboard(rows)
	if err != nil {
		log.Println("Error scanning multiple messages for display in dashboard: ", err)
		return nil, err
	}

	return messages, nil
}

func scanMultipleMessagesToDisplayInDashboard(rows *pgx.Rows) ([]*Message, error) {
	messages := []*Message{}
	var err error
	for rows.Next() {
		var messageID, roomID pgx.NullInt64
		var created pgx.NullTime
		err = rows.Scan(&messageID, &roomID, &created)
		if err != nil {
			log.Println("Error scanning message row in scanMultipleMessages: ", err)
			return nil, err
		}
		message := fillMessageToDisplayInDashboard(messageID, roomID, created)
		messages = append(messages, message)
	}
	return messages, nil
}

func fillMessageToDisplayInDashboard(messageID pgx.NullInt64, roomID pgx.NullInt64, created pgx.NullTime) *Message {
	message := &Message{}
	if messageID.Valid {
		message.MessageID = messageID.Int64
	}
	message, _ = GetMessageByMessageID(message.MessageID)

	return message
}

func GetMessageByMessageID(messageID int64) (*Message, error) {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select message_id, room_id, message_text, profile_id, created, room_uuid from "+messageTableName+" where message_id=$1", messageID)

	var messageId, roomID, profileID pgx.NullInt64
	var created pgx.NullTime
	var messageText pgx.NullString
	var roomUUID string

	err := row.Scan(&messageId, &roomID, &messageText, &profileID, &created, &roomUUID)
	if err != nil {
		log.Println("Error scanning message row in scanMultipleMessages: ", err)
		return nil, err
	}
	message := &Message{}
	if messageId.Valid {
		message.MessageID = messageId.Int64
	}
	if roomID.Valid {
		message.RoomID = roomID.Int64
	}
	if messageText.Valid {
		message.MessageText = messageText.String
	}
	if profileID.Valid {
		message.ProfileID = profileID.Int64
	}
	if created.Valid {
		message.Created = created.Time
	}
	message.RoomUUID = roomUUID
	return message, nil
}

func GetCountOfAllMessages(userID int64) int64 {
	conn := db.Connect()
	defer conn.Close()
	profileIDs, _ := user.GetProfileIDsByUserID(userID)
	var count int64
	//rows, err := conn.Query("select max(message_id), "+messageTableName+".room_id, max("+messageTableName+".created) from "+messageTableName+"inner join "+roomTableName+"on "+messageTableName+".room_id = "+roomTableName+".room_id where $1=any("+roomTableName+".users) GROUP BY "+messageTableName+".room_id ORDER BY MAX("+messageTableName+".created) DESC LIMIT $2 OFFSET $3", userID, displayCount, pageNum)
	rows := conn.QueryRow("select count(t1) from (SELECT MAX(message_id) as t1, messages.room_id as t2 FROM messages  INNER JOIN rooms ON messages.room_id = rooms.room_id  WHERE $1 && rooms.users GROUP BY messages.room_id ) k", profileIDs)
	rows.Scan(&count)
	return count

}
