package rooms

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
	"github.com/unirep/ur-local-web/app/models/user"
)

const roomTableName = "rooms"

func GetRoom(roomMemberIdA int64, roomMemberIdB int64) (*Room, error) {

	conn := db.Connect()
	defer conn.Close()

	rows := conn.QueryRow("select "+"room_id, room_owner, users, created, room_uuid from "+roomTableName+" "+" where $1=any(users) AND $2=any(users)", roomMemberIdA, roomMemberIdB)

	var roomID, roomOwner pgx.NullInt64
	var users []int64
	var created pgx.NullTime
	var roomRow Room
	var roomUUID string

	err := rows.Scan(&roomID, &roomOwner, &users, &created, &roomUUID)
	if err != nil {
		return nil, err
	}
	if roomID.Valid {
		roomRow.RoomId = roomID.Int64
	}
	if roomOwner.Valid {
		roomRow.RoomOwner = roomOwner.Int64
	}
	roomRow.Users = users

	if created.Valid {
		roomRow.Created = created.Time
	}
	roomRow.RoomUUID, _ = gocql.ParseUUID(roomUUID)

	return &roomRow, err
}

func InsertRoom(room *Room) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()
	room.Created = time.Now().UTC()

	err := conn.QueryRow("insert into "+roomTableName+"(room_owner, users, created) values ($1, $2, $3) Returning room_id",
		room.RoomOwner, room.Users, room.Created).Scan(&id)
	if err != nil {
		log.Println("Error: in app/models/rooms/room-db.go/InsertRoom():", err)
	}
	return id, err
}

func RoomFieldList() string {

	list := "room_id, room_owner, users, created, room_uuid"

	return list

}

func scanSingleRoomID(row *pgx.Row) (*Room, error) {
	var roomID, roomOwner pgx.NullInt64
	var users []pgx.NullInt64
	var created pgx.NullTime
	var roomUUID string

	err := row.Scan(&roomID, &roomOwner, &users, &created, &roomUUID)

	if err != nil {
		return nil, err
	}
	roomRow := &Room{}

	if roomOwner.Valid {
		roomRow.RoomOwner = roomOwner.Int64
	}
	if users[0].Valid {
		roomRow.Users[0] = users[0].Int64
	}
	if created.Valid {
		roomRow.Created = created.Time
	}
	roomRow.RoomUUID, _ = gocql.ParseUUID(roomUUID)
	return roomRow, nil
}

func GetRoomByRoomUUIDAndCurrentUserID(roomUUId string, currentUserID int64) (*Room, error) {
	conn := db.Connect()
	defer conn.Close()
	profileIDs, _ := user.GetProfileIDsByUserID(currentUserID)
	rows := conn.QueryRow("select "+"room_id, room_owner, users, created, room_uuid from "+roomTableName+" "+" where room_uuid=$1 AND $2 && users", roomUUId, profileIDs)

	var roomID, roomOwner pgx.NullInt64
	var users []int64
	var created pgx.NullTime
	var roomRows Room
	var roomUUID string

	err := rows.Scan(&roomID, &roomOwner, &users, &created, &roomUUID)
	if err != nil {
		return nil, err
	}
	if roomID.Valid {
		roomRows.RoomId = roomID.Int64
	}
	if roomOwner.Valid {
		roomRows.RoomOwner = roomOwner.Int64
	}
	roomRows.Users = users
	roomRows.RoomUUID, _ = gocql.ParseUUID(roomUUID)

	if created.Valid {
		roomRows.Created = created.Time
	}

	return &roomRows, err
}

func GetRoomByRoomUUID(roomUUID string) (*Room, error) {
	conn := db.Connect()
	defer conn.Close()
	row := conn.QueryRow("select room_id, room_owner, users, created, room_uuid from "+roomTableName+" where room_uuid=$1", roomUUID)

	var roomID, roomOwner pgx.NullInt64
	var users []int64
	var created pgx.NullTime
	var roomRow Room

	err := row.Scan(&roomID, &roomOwner, &users, &created, &roomUUID)
	if err != nil {
		return nil, err
	}
	if roomID.Valid {
		roomRow.RoomId = roomID.Int64
	}
	if roomOwner.Valid {
		roomRow.RoomOwner = roomOwner.Int64
	}
	roomRow.Users = users
	roomRow.RoomUUID, _ = gocql.ParseUUID(roomUUID)

	if created.Valid {
		roomRow.Created = created.Time
	}

	return &roomRow, err
}
