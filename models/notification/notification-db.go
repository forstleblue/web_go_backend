package notification

import (
	"fmt"
	"log"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const notificationTableName = "notifications"
const userTableName = "users"

//InsertNotification inserts notification
func InsertNotification(n *Notification) error {

	var id int64
	conn := db.Connect()
	defer conn.Close()

	err := conn.QueryRow("insert into "+notificationTableName+"(notification_type, sender_id, receiver_id, entity_id, entity_history_id, notification_text, unread) values ($1, $2, $3, $4, $5, $6, $7) Returning notification_id",
		n.NotificationType, n.SenderID, n.ReceiverID, n.EntityID, n.EntityHistoryID, n.NotificationText, n.Unread).Scan(&id)

	return err
}

//UpdateNotification updates notification
func UpdateNotification(n *Notification) error {
	conn := db.Connect()
	defer conn.Close()

	_, err := conn.Exec("Update "+notificationTableName+" Set notification_type=$1, sender_id=$2, receiver_id=$3, entity_id=$4, notification_string=$5 where notification_id=$6",
		n.NotificationType, n.SenderID, n.ReceiverID, n.EntityID, n.NotificationText, n.NotificationID)
	return err
}

//UpdateUnreadNotifications remove unread userid values
func UpdateUnreadNotifications(userID int64) error {

	conn := db.Connect()
	defer conn.Close()
	userIDString := fmt.Sprintf("%d", userID)
	_, err := conn.Query("UPDATE "+notificationTableName+" SET unread = array_remove(unread, CAST("+userIDString+" AS BIGINT)) where receiver_id = $1", userID)
	return err
}

//GetNotificationCount returns notificaion count
func GetNotificationCount(id int64) int64 {
	conn := db.Connect()
	defer conn.Close()
	var count int64
	rows := conn.QueryRow("select count(*) from "+notificationTableName+" where $1=any(unread)", id)
	rows.Scan(&count)
	return count
}

//GetNotificationByID returns notification
func GetNotificationByID(id int64) (Notification, error) {
	conn := db.Connect()
	defer conn.Close()
	log.Println("notification id:", id)
	row := conn.QueryRow("select "+notificationFieldList()+" from "+notificationTableName+" where notification_id = $1", id)
	var notificationID, senderID, receiverID, entityID, entityHistoryID pgx.NullInt64
	var notificationType, notificationText pgx.NullString
	var unreadArray []int64
	var uuid string

	err := row.Scan(&notificationID, &notificationType, &senderID, &receiverID, &entityID, &notificationText, &entityHistoryID, &unreadArray, &uuid)
	if err != nil {
		log.Println("Error in models/notification/notification-db.go GetNotificationByID() fail to get Notification:", err)
		return Notification{}, err
	}
	var notificationItem = Notification{}
	if notificationID.Valid {
		notificationItem.NotificationID = notificationID.Int64
	}
	if notificationType.Valid {
		notificationItem.NotificationType = notificationType.String
	}
	if senderID.Valid {
		notificationItem.SenderID = senderID.Int64
	}
	if receiverID.Valid {
		notificationItem.ReceiverID = receiverID.Int64
	}
	if entityID.Valid {
		notificationItem.EntityID = entityID.Int64
	}
	if entityHistoryID.Valid {
		notificationItem.EntityHistoryID = entityHistoryID.Int64
	}
	if notificationText.Valid {
		notificationItem.NotificationText = notificationText.String
	}
	if unreadArray != nil {
		notificationItem.Unread = unreadArray
	}
	notificationItem.NotificationUUID = uuid

	return notificationItem, nil
}

//GetNotifications return notification list
func GetNotifications(receiverID int64) ([]Notification, error) {
	conn := db.Connect()
	defer conn.Close()

	rows, err := conn.Query("select notification_id, notification_type, sender_id, receiver_id, entity_id, entity_history_id, notification_text, unread from "+notificationTableName+" where receiver_id=$1 order by notification_id DESC", receiverID)

	if err != nil {
		log.Println("Error in models/notification/notificaiton-db.go GetNotifications(receiverID int64): ", err)
	}

	var notifications = []Notification{}

	for rows.Next() {
		var notificationID, senderID, receiverID, entityID, entityHistoryID pgx.NullInt64
		var notificationType, notificationText pgx.NullString
		var users []int64
		err = rows.Scan(&notificationID, &notificationType, &senderID, &receiverID, &entityID, &entityHistoryID, &notificationText, &users)

		if err == nil {
			var notificationItem = Notification{}
			if notificationID.Valid {
				notificationItem.NotificationID = notificationID.Int64
			}
			if notificationType.Valid {
				notificationItem.NotificationType = notificationType.String
			}
			if senderID.Valid {
				notificationItem.SenderID = senderID.Int64
			}
			if receiverID.Valid {
				notificationItem.ReceiverID = receiverID.Int64
			}
			if entityID.Valid {
				notificationItem.EntityID = entityID.Int64
			}
			if entityHistoryID.Valid {
				notificationItem.EntityHistoryID = entityHistoryID.Int64
			}
			if notificationText.Valid {
				notificationItem.NotificationText = notificationText.String
			}
			if users != nil {
				notificationItem.Unread = users
			}
			notifications = append(notifications, notificationItem)
		}
	}

	return notifications, err
}

//CheckLeaveFeedback checks leave feedback or not
func (n *Notification) CheckLeaveFeedback(feedbackReceiverID int64) bool {
	conn := db.Connect()
	defer conn.Close()
	nType := "FeedbackReceived"
	row := conn.QueryRow("select notification_id "+" from "+notificationTableName+" where receiver_id = $1 AND  entity_history_id = $2 AND notification_type= $3", feedbackReceiverID, n.EntityHistoryID, nType)
	var nID pgx.NullInt64
	err := row.Scan(&nID)
	if err != nil {
		log.Println("Error in notification/notification-db.go fail to check notification:", err, "notification_id:", n.NotificationID)
	}
	var leaveFeedback bool
	if nID.Valid {
		leaveFeedback = true
	} else {
		leaveFeedback = false
	}

	return leaveFeedback
}

//CheckReceivedFeedback returns true if received feedback from provider or customer
func (n *Notification) CheckReceivedFeedback() bool {
	conn := db.Connect()
	defer conn.Close()
	nType := "FeedbackReceived"
	row := conn.QueryRow("select notification_id "+" from "+notificationTableName+" where receiver_id = $1 AND  entity_history_id = $2 AND notification_type= $3", n.ReceiverID, n.EntityHistoryID, nType)
	var nID pgx.NullInt64
	err := row.Scan(&nID)
	if err != nil {
		log.Println("Error in notification/notification-db.go fail to check notification:", err, "notification_id:", n.NotificationID)
	}
	var feedbackReceived bool
	if nID.Valid {
		feedbackReceived = true
	} else {
		feedbackReceived = false
	}

	return feedbackReceived
}

//TwoFeedbackCompleted checks two feedback completes or not.
func (n *Notification) TwoFeedbackCompleted() bool {
	conn := db.Connect()
	defer conn.Close()
	nType := "FeedbackReceived"
	var count int64
	rows := conn.QueryRow("select count(*) from "+notificationTableName+" where entity_id=$1 AND notification_type=$2", n.EntityID, nType)
	rows.Scan(&count)
	var feedbackCompleted bool
	if count == 2 {
		feedbackCompleted = true
	} else {
		feedbackCompleted = false
	}
	return feedbackCompleted
}

// HELPER METHODS
func notificationFieldList() string {
	list := "notification_id, notification_type, sender_id, receiver_id, entity_id, notification_text, entity_history_id, unread, notification_uuid"
	return list
}
