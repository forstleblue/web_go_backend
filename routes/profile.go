package routes

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/franela/goreq"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/feedback"
	"github.com/unirep/ur-local-web/app/models/platform/ebay"
	"github.com/unirep/ur-local-web/app/models/service_category"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
)

const profileFormID = "profile-form"

//ProfileList gets a list of the logged in user's profiles
func ProfileList(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, false)
	if currUser == nil {
		return
	}

	//create data as struct variable for the view context
	var data = struct {
		Profiles         []*user.Profile
		UserID           int64
		ParentPage       string
		EbayProfileAdded bool
		FeedbackAverages []feedback.FeedbackAverage
		//Data to display user's reputattion widget
		User                *user.User
		Score               int16
		Count               int16
		FeedbackDescription string
		FullURL             string
	}{}

	profiles, err := user.GetProfileByUserID(currUser.UserID)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileList(): finding user {%s} profiles': %s\n", currUser.String(), err.Error())
		InternalServerErrorRedirect(ctx)
	}

	utils.Log(ctx, "ProfileList", 0, 0, "Found user profiles:"+string(len(profiles)))

	data.Profiles = profiles

	if currUser == nil {
		data.UserID = 0
	} else {
		data.UserID = currUser.UserID
	}

	for _, profileItem := range data.Profiles {
		FeedbackAverageItem := profileItem.UniversalReputationScore()
		data.FeedbackAverages = append(data.FeedbackAverages, *FeedbackAverageItem)
	}

	data.ParentPage = "profiles"
	data.EbayProfileAdded = user.EbayProfileAdded(currUser.UserID)

	//Data for show reputation widget of user
	data.Score, data.Count, _ = feedback.GetUniversalReputationScoreByUserID(currUser.UserID)
	data.FullURL = config.FullBaseURL()
	data.FeedbackDescription = ""
	data.User, _ = user.GetUser(currUser.UserID)

	if data.Score < 25 {
		data.FeedbackDescription = "Unacceptable"
	} else if data.Score < 40 {
		data.FeedbackDescription = "Needs Improvement"
	} else if data.Score < 55 {
		data.FeedbackDescription = "Acceptable"
	} else if data.Score < 70 {
		data.FeedbackDescription = "Met Expectations"
	} else if data.Score < 85 {
		data.FeedbackDescription = "Exceeds Expectations"
	} else if data.Score < 100 {
		data.FeedbackDescription = "Excellent"
	} else {
		// 100%
		data.FeedbackDescription = "Exceptional"
	}

	pg := &render.Page{Title: "Profiles", TemplateFileName: "authenticated/profiles.html", Data: data}
	pg.Render(ctx)
}

//ReputationSources gets a list of the logged in user's reputation sources
func ReputationSources(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, false)
	if currUser == nil {
		return
	}

	//create data as struct variable for the view context
	var data = struct {
		Profiles         []*user.Profile
		UserID           int64
		ParentPage       string
		EbayProfileAdded bool
		EbayProfileUUID  string
		//Data for display user's reputation widget
		User                *user.User
		Score               int16
		Count               int16
		FeedbackDescription string
		FullURL             string
	}{}

	// profiles, err := user.GetProfileByUserID(currUser.UserID)
	// if err != nil {
	// 	log.Printf("Error in routes/profile.go ProfileList(): finding user {%s} profiles': %s\n", currUser.String(), err.Error())
	// 	InternalServerErrorRedirect(ctx)
	// }

	//utils.Log(ctx, "ProfileList", 0, 0, "Found user profiles:"+string(len(profiles)))

	//data.Profiles = profiles

	if currUser == nil {
		data.UserID = 0
	} else {
		data.UserID = currUser.UserID
	}
	data.ParentPage = "profiles"
	data.EbayProfileAdded = user.EbayProfileAdded(currUser.UserID)
	if data.EbayProfileAdded == true {
		data.EbayProfileUUID = user.GetEbayProfileUUID(currUser.UserID)
	}

	//Data for show reputation widget of user
	data.Score, data.Count, _ = feedback.GetUniversalReputationScoreByUserID(currUser.UserID)
	data.FullURL = config.FullBaseURL()
	data.FeedbackDescription = ""
	data.User, _ = user.GetUser(currUser.UserID)

	if data.Score < 25 {
		data.FeedbackDescription = "Unacceptable"
	} else if data.Score < 40 {
		data.FeedbackDescription = "Needs Improvement"
	} else if data.Score < 55 {
		data.FeedbackDescription = "Acceptable"
	} else if data.Score < 70 {
		data.FeedbackDescription = "Met Expectations"
	} else if data.Score < 85 {
		data.FeedbackDescription = "Exceeds Expectations"
	} else if data.Score < 100 {
		data.FeedbackDescription = "Excellent"
	} else {
		// 100%
		data.FeedbackDescription = "Exceptional"
	}
	pg := &render.Page{Title: "Reputation Sources", TemplateFileName: "authenticated/reputation-sources.html", Data: data}
	pg.Render(ctx)
}

//ProfileAdd gets an empty profile to add
func ProfileAdd(ctx *fasthttp.RequestCtx) {

	//create data as struct variable for the view context
	var data = struct {
		serviceTypes []service_category.ServiceTypes
		Profile      *user.Profile
	}{
		Profile: &user.Profile{},
	}

	data.serviceTypes = service_category.GetServiceCategory()

	pg := &render.Page{Title: "Add Profile", TemplateFileName: "authenticated/profile-add.html", Data: data}
	pg.Render(ctx)
}

// AddEbayProfile imports profile from Ebay
func AddEbayProfile(ctx *fasthttp.RequestCtx) {
	CreateEbayProfile(ctx)
	ctx.Redirect("/profiles", fasthttp.StatusSeeOther)
}

//CreateEbayProfile create ebay profile
func CreateEbayProfile(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}
	timeLayout := "2006-01-02T15:04:05.000Z"
	profile := &user.Profile{}
	profile.User = *currUser
	profile.Title = "Ebay"
	profile.Heading = "Ebay Profile"
	profile.ProfileType = "e"
	profile.OauthToken = fmt.Sprintf("%s", ctx.QueryArgs().Peek("authToken"))
	test, err := time.Parse(timeLayout, string(ctx.QueryArgs().Peek("expiry")))
	profile.OauthExpiry = test
	if err != nil {
		log.Println("Error parsing time: ", err)
		return
	}
	log.Println("OauthToken: ", profile.OauthToken)
	log.Println("OauthExpiry: ", profile.OauthExpiry)

	requestXML := `"<?xml version='1.0' encoding='utf-8'?>
	<GetFeedbackRequest xmlns='urn:ebay:apis:eBLBaseComponents'> 
	<RequesterCredentials> 
	<eBayAuthToken>` + profile.OauthToken + `</eBayAuthToken> 
	</RequesterCredentials> 
	<UserID>neilkevenson</UserID>
	<DetailLevel>ReturnAll</DetailLevel>	
	<CommentType>Positive</CommentType>	
	</GetFeedbackRequest>"`

	requestBody := []byte(requestXML)

	req := goreq.Request{
		Method:      "POST",
		Uri:         appConfig.EbayAPI.URL,
		Body:        requestBody,
		ContentType: "application/xml; charset=utf-8",
		UserAgent:   "go-ebay-fetch-orders",
		ShowDebug:   false,
	}

	var appConfig = config.Config()
	req.AddHeader("X-EBAY-API-CALL-NAME", "GetFeedback")
	req.AddHeader("X-EBAY-API-DEV-NAME", appConfig.EbayAPI.DevID)
	req.AddHeader("X-EBAY-API-CERT-NAME", appConfig.EbayAPI.Secret)
	req.AddHeader("X-EBAY-API-APP-NAME", appConfig.EbayAPI.AppID)
	req.AddHeader("X-EBAY-API-COMPATIBILITY-LEVEL", "1003")
	req.AddHeader("X-EBAY-API-REQUEST-ENCODING", "XML")
	req.AddHeader("X-EBAY-API-RESPONSE-ENCODING", "XML")
	req.AddHeader("X-EBAY-API-SITEID", "15")

	res, err := req.Do()

	if err != nil {
		log.Println("ERROR in routes/profile.go CreateEbayProfile()  processUrl -> req.Do: " + err.Error())
		return
	}

	data, err := ioutil.ReadAll(res.Body)
	log.Println("ResponseData: ", string(data))
	if err != nil {
		log.Println("ERROR in routes/profile.go CreateEbayProfile() - ioutil.ReadAll : " + err.Error())
		return
	}

	type FeedbackDetail struct {
		CommentingUser      string `xml:"CommentingUser"`
		CommentingUserScore string `xml:"CommentingUserScore"`
		CommentText         string `xml:"CommentText"`
		CommentTime         string `xml:"CommentTime"`
		CommentType         string `xml:"CommentType"`
		ItemID              string `xml:"ItemID"`
		Role                string `xml:"Role"`
		FeedbackID          string `xml:"FeedbackID"`
		TransactionID       string `xml:"TransactionID"`
		ItemTitle           string `xml:"ItemTitle"`
		ItemPrice           string `xml:"ItemPrice"`
	}

	type FeedbackComment struct {
		XMLName             xml.Name         `xml:"GetFeedbackResponse"`
		Version             string           `xml:"Version"`
		FeedbackDetailArray []FeedbackDetail `xml:"FeedbackDetailArray>FeedbackDetail"`
	}

	v := FeedbackComment{Version: "none"}

	err = xml.Unmarshal([]byte(data), &v)
	log.Println("FeedbackComment: ", v.FeedbackDetailArray)
	if len(v.FeedbackDetailArray) == 0 {
		log.Println("User doen't have any valid information in ebay. Aborted!")
		return
	}
	if err != nil {
		log.Println("Error in routes/profile.go CreateEbayProfile()  unmarshal xml:", err)
		return
	} else {
		id, err := user.InsertProfile(profile)
		if err != nil {
			log.Printf("Error in app/routes/profile.go ProfileCreate(ctx *fasthttp.RequestCtx): inserting profile  {%s}: %s\n", profile.String(), err.Error())
			return
		}
		for _, item := range v.FeedbackDetailArray {
			feedbackComment := &ebay.FeedbackComment{}
			feedbackComment.ProfileID = id
			feedbackComment.CommentingUser = item.CommentingUser
			feedbackComment.CommentingUserScore = item.CommentingUserScore
			feedbackComment.CommentText = item.CommentText
			feedbackComment.CommentType = item.CommentType
			feedbackComment.ItemID = item.ItemID
			feedbackComment.Role = item.Role
			feedbackComment.FeedbackID = item.FeedbackID
			feedbackComment.TransactionID = item.TransactionID
			feedbackComment.ItemTitle = item.ItemTitle
			feedbackComment.ItemPrice = item.ItemPrice

			t, err := time.Parse(timeLayout, item.CommentTime)
			if err != nil {
				log.Println("Error in routes/profile.go CreateEbayProfile() convet time:", err)
			} else {
				feedbackComment.CommentTime = t
				_, err := ebay.InsertEbayFeedbackComment(feedbackComment)
				if err != nil {
					log.Println("Error in routes/profile.go CreateEbayProfile() fail to insert feedbackComment:", err)
				}
			}
		}
	}
}

// ProfileCreate create a profile
func ProfileCreate(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	profileType := string(ctx.FormValue("profileType"))
	heading := string(ctx.FormValue("heading"))
	serviceCategory := string(ctx.FormValue("serviceCategory"))
	title := string(ctx.FormValue("title"))
	tagsBytes := ctx.PostArgs().PeekMulti("tags[]")
	fee := string(ctx.FormValue("fee"))
	paymentNotes := string(ctx.FormValue("paymentNotes"))
	description := string(ctx.FormValue("description"))
	photoURL := string(ctx.FormValue("photoURL"))

	log.Println("PhotoURL:", photoURL)

	filePath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + photoURL
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file path:", err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file information:", err)
	}

	// Setup form error object in case needed
	var profileFormError = &JSONFormError{}
	profileFormError.Form = profileFormID

	if fileStat.Size() > MaxFileSize {
		log.Println("Image size should not be larger than 4MB.")
		profileFormError.Error = "Image size should not be larger than 4MB."
		render.JSON(ctx, profileFormError, "Image size should not be larger than 4MB.", fasthttp.StatusUnprocessableEntity)
		return
	}
	tags := []string{}
	for _, tag := range tagsBytes {
		tags = append(tags, string(tag))
	}

	var profile *user.Profile

	if title == "" || description == "" {
		profileFormError.Error = "Please fill in all required fields. Title and Description are all required."
		render.JSON(ctx, profileFormError, "Cannot process profile as not all required fields are filled", fasthttp.StatusUnprocessableEntity)
		return
	}

	profile = &user.Profile{}
	profile.User = *currUser

	// Tags can be multiple results, so need to use PostArgs

	profile.Title = title
	serviceCategoryID, err := strconv.ParseInt(serviceCategory, 10, 64)
	profile.ServiceCategory = serviceCategoryID

	//Set the profile ServiceCategory 0 for Seller
	if profileType == "s" {
		profile.ServiceCategory = 0
	}

	profile.Fee = fee
	profile.PaymentNotes = paymentNotes
	profile.Tags = tags
	profile.Description = description
	profile.PhotoURL = photoURL
	profile.Heading = heading
	profile.ProfileType = profileType

	_, err = user.InsertProfile(profile)
	if err != nil {
		log.Printf("Error in app/routes/profile.go ProfileCreate(ctx *fasthttp.RequestCtx): inserting profile  {%s}: %s\n", profile.String(), err.Error())

		profileFormError.Error = "There has been an internal error inserting the profile."
		profileFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, profileFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "ProfileSave", 0, 0, "Profile Saved:"+ctx.PostArgs().String())
	render.JSON(ctx, "/profiles", "Profile successfully saved, redirecting to profile page", fasthttp.StatusOK)

}

//ProfileEdit gets a profile to edit
func ProfileEdit(ctx *fasthttp.RequestCtx) {

	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	profileUUID := (ctx.UserValue("profileUUID")).(string)

	//create data as struct variable for the view context
	var data = struct {
		serviceTypes []service_category.ServiceTypes
		Profile      *user.Profile
	}{}

	profile, err := user.GetProfileByProfileUUID(profileUUID)

	if err != nil {
		log.Printf("Error in routes/profile.go ProfileEdit(ctx *fasthttp.RequestCtx): finding profiles with uuid '%s': %s\n", profileUUID, err.Error())
		NotFoundRedirect(ctx)
	}

	if profile.User.UserID != currUser.UserID {
		//ctx.Redirect("/profile/:profileID", fasthttp.StatusSeeOther)
		ForbiddenRedirect(ctx)
	}

	filepath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + profile.PhotoURL
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// profile photo does not exist so reset the photo
		profile.PhotoURL = ""
	}

	utils.Log(ctx, "ProfileEdit", 0, 0, "Found user profile:"+profile.String())

	data.Profile = profile

	data.serviceTypes = service_category.GetServiceCategory()
	pg := &render.Page{Title: "Edit Profile", TemplateFileName: "authenticated/profile-edit.html", Data: data}
	pg.Render(ctx)
}

//ProfileUpdate updates profile
func ProfileUpdate(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}
	var err error

	heading := string(ctx.FormValue("heading"))
	idString := string(ctx.FormValue("id"))
	profileType := string(ctx.FormValue("profileType"))
	title := string(ctx.FormValue("title"))
	serviceCategory := string(ctx.FormValue("serviceCategory"))
	fee := string(ctx.FormValue("fee"))
	paymentNotes := string(ctx.FormValue("paymentNotes"))
	description := string(ctx.FormValue("description"))
	photoURL := string(ctx.FormValue("photoURL"))

	filePath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + photoURL
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file path:", err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.Println("Error in routes/profile.go ProfileCreate() fail to get image file information:", err)
	}

	// Setup form error object in case needed
	var profileFormError = &JSONFormError{}
	profileFormError.Form = profileFormID

	if fileStat.Size() > MaxFileSize {
		log.Println("Image size should not be larger than 4MB.")
		profileFormError.Error = "Image size should not be larger than 4MB."
		render.JSON(ctx, profileFormError, "Image size should not be larger than 4MB.", fasthttp.StatusUnprocessableEntity)
		return
	}
	// Tags can be multiple results, so need to use PostArgs
	tagsBytes := ctx.PostArgs().PeekMulti("tags[]")
	tags := []string{}
	for _, tag := range tagsBytes {
		tags = append(tags, string(tag))
	}

	var profile *user.Profile

	if title == "" || description == "" {
		if profileType == "p" && serviceCategory == "" {
			profileFormError.Error = "Please fill in all required fields. Title and Description are all required."
			render.JSON(ctx, profileFormError, "Cannot process profile as not all required fields are filled", fasthttp.StatusUnprocessableEntity)
			return
		}
	}

	var ID int64
	ID, err = strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileUpdate(ctx *fasthttp.RequestCtx): parsing profile id '%s': %s\n", idString, err.Error())

		profileFormError.Error = "There has been an internal error processing the profile. Please refresh the page and try again."
		profileFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, profileFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	profile, err = user.GetProfile(ID)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileUpdate(ctx *fasthttp.RequestCtx): getting profile with id '%d': %s\n", ID, err.Error())

		profileFormError.Error = "The profile cannot be found, there may be an internal issue."
		profileFormError.Redirect = notFoundURL
		render.JSON(ctx, profileFormError, "Profile not found error", fasthttp.StatusNotFound)
		return
	}

	if profile.PhotoURL != "" && profile.PhotoURL != photoURL {
		// photo url has changed, new photo saved, delete the old one
		err = deleteProfilePhoto(currUser, profile.PhotoURL)
		if err != nil {
			// just log error for now, maybe we should have the idea of 'tasks' where certain things fail
			// that aren't super horrible, we log them to be manually fixed later
			log.Printf("Error in routes/profile.go ProfileUpdate(ctx *fasthttp.RequestCtx): deleting profile photo for user '%d' with name '%s': %s\n", ID, profile.PhotoURL, err.Error())
		}
	}

	serviceCategoryID, err := strconv.ParseInt(serviceCategory, 10, 64)
	profile.Heading = heading
	profile.Title = title
	profile.ProfileType = profileType

	//Set the profile ServiceCategory 1 for Seller and Customer
	if profileType == "p" {
		profile.ServiceCategory = serviceCategoryID
	} else {
		profile.ServiceCategory = 0
	}

	profile.Fee = fee
	profile.PaymentNotes = paymentNotes
	profile.Tags = tags
	profile.Description = description
	profile.PhotoURL = photoURL

	err = profile.Update()
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileUpdate(ctx *fasthttp.RequestCtx): updating profile {%s}: %s\n", profile.String(), err.Error())

		profileFormError.Error = "There has been an internal error updating the profile."
		profileFormError.Redirect = internalServerErrorURL
		render.JSON(ctx, profileFormError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "ProfileSave", 0, 0, "Profile Saved:"+ctx.PostArgs().String())
	render.JSON(ctx, "/profiles", "Profile successfully saved, redirecting to profile page", fasthttp.StatusOK)
}

//ProfileDelete Saves edited or added profile
func ProfileDelete(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	// Setup form error object in case needed
	var formError = &JSONFormError{}
	formError.Form = profileFormID

	idString := string(ctx.QueryArgs().Peek("id"))

	ID, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileDelete(ctx *fasthttp.RequestCtx): parsing profile id '%s': %s\n", idString, err.Error())

		formError.Error = "There has been an internal error processing the profile. Please refresh the page and try again."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	_, err = user.DeleteProfile(ID)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfileDelete(ctx *fasthttp.RequestCtx): deleting profile with id '%d': %s\n", ID, err.Error())

		formError.Error = "The profile cannot be deleted, there may be an internal issue."
		formError.Redirect = internalServerErrorURL
		render.JSON(ctx, formError, "Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	utils.Log(ctx, "ProfileDelete", 0, 0, "Profile Deleted:"+ctx.PostArgs().String())
	render.JSON(ctx, "/profiles", "Profile successfully deleted, redirecting to profile page", fasthttp.StatusOK)
}

//ProfilePhotoUpload Saves uploaded profile photo to disk
func ProfilePhotoUpload(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}

	uuid := string(ctx.FormValue("qquuid"))
	filename := string(ctx.FormValue("qqfilename"))
	filesizeString := string(ctx.FormValue("qqtotalfilesize"))
	fileHeader, err := ctx.FormFile("qqfile")
	if err != nil {
		log.Printf("Error in routes/profile.go ProfilePhotoUpload(ctx *fasthttp.RequestCtx): getting form file: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	fileSize, err := strconv.ParseInt(filesizeString, 10, 64)
	if err != nil {
		log.Println("Error in routes/user.go UserPhotoUpload() fail to get file size:", err)
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	if fileSize > MaxFileSize {
		log.Printf("Image size should be less than 4MB")
		ctx.SetBody([]byte(`{"success":false, "error": "Image size should be less than 4MB"}`))
		return
	}
	log.Printf("Upload photo details %s, %s, %s, %s\n", uuid, filename, filesizeString, fileHeader.Header.Get("Content-Type"))

	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		log.Printf("Error in routes/profile.go ProfilePhotoUpload(ctx *fasthttp.RequestCtx): opening form file: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	//create destination file making sure the path is writeable.
	err = os.MkdirAll("../static/images/profile-photos/"+strconv.FormatInt(currUser.UserID, 10), 0755)
	if err != nil {
		log.Printf("Error making directory path to save: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	dst, err := os.Create("../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + uuid + "-" + filename)
	defer dst.Close()
	if err != nil {
		log.Printf("Error opening destination to save: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}
	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying file to destination: %s", err.Error())
		ctx.SetBody([]byte(`{"success":false, "error": "Internal error"}`))
		return
	}

	utils.Log(ctx, "ProfilePhotoSave", 0, 0, "Profile Image Saved:"+ctx.PostArgs().String())
	//render.JSON(ctx, data, "Profile photo successfully saved", fasthttp.StatusOK)
	ctx.SetBody([]byte(`{"success":true, "filename": "` + uuid + "-" + filename + `"}`))
}

//ProfilePhotoDelete Deletes uploaded profile photo from disk
func ProfilePhotoDelete(ctx *fasthttp.RequestCtx) {
	currUser := getUserFromContext(ctx, true)
	if currUser == nil {
		return
	}
	uuid := (ctx.UserValue("uuid")).(string)
	filename := string(ctx.QueryArgs().Peek("filename"))

	err := deleteProfilePhoto(currUser, uuid+"-"+filename)
	if err != nil {
		log.Printf("Error in routes/profile.go ProfilePhotoUpload(ctx *fasthttp.RequestCtx): deleting profile photo file: %s", err.Error())
		render.JSON(ctx, nil, "Cannot delete profile photo", fasthttp.StatusInternalServerError)
		return
	}

	render.JSON(ctx, nil, "Photo delete successful", fasthttp.StatusOK)
}

//TagsSearch Allows for searching for tags
func TagsSearch(ctx *fasthttp.RequestCtx) {
	q := ctx.QueryArgs().Peek("q")
	log.Printf("q {%s}\n", q)

	tags := user.GetTagsBySearch(string(q))
	items := []*JSONSelect2Item{}

	for _, tag := range tags {
		items = append(items, &JSONSelect2Item{ID: tag, Text: tag})
	}

	// Currently hardcoded, but this needs to come from the database
	/*items := []*JSONSelect2Item{
		&JSONSelect2Item{ID: "Programmer", Text: "Programmer"},
		&JSONSelect2Item{ID: "Web Developer", Text: "Web Developer"},
		&JSONSelect2Item{ID: "Baby Sitter", Text: "Baby Sitter"},
		&JSONSelect2Item{ID: "House Cleaner", Text: "House Cleaner"},
	}*/

	var data = struct {
		Items []*JSONSelect2Item `json:"items"`
	}{
		items,
	}

	render.JSON(ctx, data, "Tag search successful", fasthttp.StatusOK)
}

// HELPER Functions
func deleteProfilePhoto(currUser *user.User, fileName string) error {
	filepath := "../static/images/profile-photos/" + strconv.FormatInt(currUser.UserID, 10) + "/" + fileName

	err := os.Remove(filepath)
	return err
}
