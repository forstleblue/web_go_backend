package ebay

import (
	"time"
)

//FeedbackComment is feedback history of ebay user
type FeedbackComment struct {
	FeedbackCommentID   int64     `json:"feedback_comment_id"`
	ProfileID           int64     `json:"profile_id"`
	CommentingUser      string    `json:"commenting_user"`
	CommentingUserScore string    `json:"commenting_user_score"`
	CommentText         string    `json:"comment_text"`
	CommentTime         time.Time `json:"comment_time"`
	CommentType         string    `json:"comment_type"`
	ItemID              string    `json:"item_id"`
	Role                string    `json:"role"`
	FeedbackID          string    `json:"feedback_id"`
	TransactionID       string    `json:"transaction_id"`
	ItemTitle           string    `json:"item_title"`
	ItemPrice           string    `json:"item_price"`
}
