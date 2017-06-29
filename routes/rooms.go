package routes

import (
	"log"
	"math"
	"strconv"

	"time"

	"bytes"

	"github.com/CloudyKit/jet"
	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/message"
	"github.com/unirep/ur-local-web/app/models/rooms"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

var views = jet.NewHTMLSet("./views/partials")

type contentCache struct {
	Type string
	Body []byte
	Etag string
	Gzip bool
}

func CheckMessageRoom(ctx *fasthttp.RequestCtx) {

	profileIds := ctx.QueryArgs().PeekMulti("profileid")
	var profileIDs = []int64{}

	for _, v := range profileIds {
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			log.Println("Error in routes/rooms.go CheckMessageRoom(): convert string to int %s\n", err.Error())
		} else {
			profileIDs = append(profileIDs, i)
		}
	}

	var currentUserCustomProfileID = profileIDs[1]
	if currentUserCustomProfileID == 0 {
		ctx.Redirect("/login-register", fasthttp.StatusUnauthorized)
		return
	}
	var targetProfileID = profileIDs[0]
	var err error
	room, err := rooms.GetRoom(currentUserCustomProfileID, targetProfileID)

	if err != nil {
		var temp rooms.Room
		temp.Created = time.Now().UTC()
		profile, _ := user.GetProfile(currentUserCustomProfileID)
		temp.RoomOwner = profile.User.UserID
		temp.Users = append(temp.Users, currentUserCustomProfileID, targetProfileID)
		roomID, errR := rooms.InsertRoom(&temp)
		if errR != nil {
			log.Println("error in routes/rooms.go CheckMessageRoom(): insert roomtable", roomID)
			return
		}
		room, err = rooms.GetRoom(currentUserCustomProfileID, targetProfileID)
	}

	ctx.Redirect("/rooms/"+room.RoomUUID.String(), fasthttp.StatusSeeOther)
}

func DisplayRoom(ctx *fasthttp.RequestCtx) {

	//context to display
	var data = struct {
		room                 *rooms.Room
		Messages             []*message.Message
		Focus                string
		Profile              *user.Profile
		ParentPage           string
		UserID               int64
		currentUserProfileID int64
		ProfileType          string
		RoomUUID             string
	}{}
	roomUUID := (ctx.UserValue("roomUUID")).(string)

	//check currentUser is valid.
	if ctx.UserValue("user") == nil {
		data.UserID = 0
	} else {
		currentUser := getUserFromContext(ctx, true)
		data.UserID = currentUser.UserID
	}

	roomData, err := rooms.GetRoomByRoomUUIDAndCurrentUserID(roomUUID, data.UserID)
	if err != nil {
		log.Println("err: in app/routes/rooms.go DisplayRoom(): ", "currentUser is not valid!!")
		return
	}
	data.room = roomData
	data.Messages, _ = message.GetMessagesByRoomID(roomData.RoomId)
	data.Focus = string(ctx.QueryArgs().Peek("f"))
	data.RoomUUID = roomUUID

	//remove unreadmessgaes in currentRoom
	currentUserProfileIdIncurrentRoom := data.room.GetCurrentProfileIDByCurrentUserID(data.UserID)

	if currentUserProfileIdIncurrentRoom == -1 {
		return
	}
	data.currentUserProfileID = currentUserProfileIdIncurrentRoom
	err = message.RemoveCurrentUserFromUnreadMessagesInCurrentRoom(roomData.RoomId, currentUserProfileIdIncurrentRoom)
	//get userProfiles
	otherUserProfileID := data.room.GetOtherProfilesByCurrentUserID(data.UserID)[0]
	data.Profile, err = user.GetProfile(otherUserProfileID)
	data.ParentPage = "dashboard"

	data.ProfileType = user.GetProfileType(data.currentUserProfileID)
	if err != nil {
		log.Println("Error: in app/routes/rooms.go DisplayRoom(): can't get profile from profileID ", err)
		return
	}

	pg := &render.Page{Title: "Message Room", TemplateFileName: "authenticated/message-room.html", Data: data}
	pg.Render(ctx)
}

func ManageMessage(ctx *fasthttp.RequestCtx) {

	roomUUID := string(ctx.FormValue("roomUUId"))
	profileID, _ := strconv.ParseInt(string(ctx.FormValue("profileId")), 10, 64)
	userProfile, _ := user.GetProfile(profileID)
	userID := userProfile.User.UserID
	messageText := string(ctx.FormValue("messageText"))

	var messageContent message.Message
	messageContent.RoomUUID = roomUUID
	messageContent.MessageText = messageText
	messageContent.ProfileID = profileID
	messageContent.Created = time.Now().UTC()

	currentRoom, err := rooms.GetRoomByRoomUUIDAndCurrentUserID(roomUUID, userID)
	messageContent.Unread = message.GetUnreadUsers(currentRoom.Users, userID)
	messageContent.RoomID = currentRoom.RoomId
	log.Println("messageContent: ", messageContent)
	_, err = message.InsertMessage(&messageContent)
	if err != nil {
		log.Println("error: in app/routes/rooms.go ManageMessage():", err)
		render.JSON(ctx, "failed", "response to send-message", fasthttp.StatusOK)
	}
	receiveUserProfile, _ := user.GetProfile(messageContent.Unread[0])
	receiveUser := receiveUserProfile.User
	defer emailing.SendNewMessageEmail(&receiveUser, &userProfile.User, &messageContent)

	render.JSON(ctx, "success", "response to send-message", fasthttp.StatusOK)
}

func UpdateMessagesInDashboard(ctx *fasthttp.RequestCtx) {

	pageNum, _ := strconv.ParseInt(string(ctx.FormValue("currentPage")), 10, 64)
	userID, _ := strconv.ParseInt(string(ctx.FormValue("userId")), 10, 64)
	displayNum, _ := strconv.ParseInt(string(ctx.FormValue("displayCount")), 10, 64)

	messages, err := message.GetMessageIDsToDisplayInDashboard(userID, displayNum*(pageNum-1), displayNum)
	if err != nil {
		log.Println("Error in app/routes/rooms.go UpdateMessagesInDashboard(): to display messages: ", err)
	}
	var Data = struct {
		MessagesToDisplay []*message.Message
		UserID            int64
		RoomPageNumbers   []int
	}{}
	var AjaxData = struct {
		DashMessageContent   string
		DashPagination       string
		DashPaginationLength int
	}{}

	Data.MessagesToDisplay = messages
	Data.UserID = userID
	roomCount := message.GetCountOfAllMessages(Data.UserID)
	d := float64(roomCount) / float64(displayNum)
	pageNumbers := make([]int, int(math.Ceil(d)))
	for i := range pageNumbers {
		pageNumbers[i] = 1 + i
	}
	Data.RoomPageNumbers = pageNumbers

	vw, err := views.GetTemplate("_templ_messages.html")
	if err != nil {
		log.Println("Error: in app/routes/rooms.go UpdateMessagesInDashboard():", err)
		return
	}
	buf := bytes.Buffer{}
	err = vw.Execute(&buf, nil, Data)
	if err != nil {
		log.Println("Error: in app/routes/rooms.go UpdateMessagesInDashboard():", err)
		return
	}

	AjaxData.DashMessageContent = buf.String()

	vw, err = views.GetTemplate("_templ_pagination.html")
	if err != nil {
		log.Println("Error: in app/routes/rooms.go UpdateMessagesInDashboard():", err)
		return
	}
	buf = bytes.Buffer{}
	err = vw.Execute(&buf, nil, Data)
	if err != nil {
		log.Println("Error: in app/routes/rooms.go UpdateMessagesInDashboard():", err)
		return
	}

	AjaxData.DashPagination = buf.String()
	AjaxData.DashPaginationLength = len(pageNumbers)

	render.JSON(ctx, AjaxData, "Data to display", fasthttp.StatusOK)
}
