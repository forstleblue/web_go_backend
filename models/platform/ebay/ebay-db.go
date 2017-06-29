package ebay

import "github.com/unirep/ur-local-web/app/db"

const platformEbayTableName = "platform_ebay"

//InsertEbayFeedbackComment add EbayFeedbackComment
func InsertEbayFeedbackComment(e *FeedbackComment) (int64, error) {
	var id int64
	conn := db.Connect()
	defer conn.Close()

	err := conn.QueryRow("insert into "+platformEbayTableName+"(profile_id, commenting_user, commenting_user_score, comment_text, comment_time, comment_type, item_id, role, feedback_id, transaction_id, item_title, item_price) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) Returning feedback_comment_id",
		e.ProfileID, e.CommentingUser, e.CommentingUserScore, e.CommentText, e.CommentTime, e.CommentType, e.ItemID, e.Role, e.FeedbackID, e.TransactionID, e.ItemTitle, e.ItemPrice).Scan(&id)

	return id, err
}
