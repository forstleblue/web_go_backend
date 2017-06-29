package routes

import (
	"log"
	"strconv"
	"strings"

	"github.com/unirep/ur-local-web/app/models/feedback"
	"github.com/unirep/ur-local-web/app/models/service_category"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

//FindPerson renders the Fin A Person page
func FindPerson(ctx *fasthttp.RequestCtx) {

	//create data as struct variable for the view context
	var data = struct {
		Loc              string
		Job              []string
		City             string
		Postcode         string
		Profiles         []*user.Profile
		LatestProfiles   []*user.Profile
		Heading          string
		serviceTypes     []service_category.ServiceTypes
		UserID           int64
		ParentPage       string
		FeedbackAverages []feedback.FeedbackAverage
	}{}

	jobs := ctx.QueryArgs().PeekMulti("job")
	var jobList = []int64{}
	var tags = []string{}

	for _, v := range jobs {
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			tags = append(tags, string(v))
		} else {
			jobList = append(jobList, i)
		}
	}

	loc := string(ctx.QueryArgs().Peek("loc"))
	if loc != "" {
		data.Loc = loc
	}
	log.Printf("User location:%s", loc)
	city := string(ctx.QueryArgs().Peek("city"))
	if city != "" {
		data.City = city
	}
	log.Printf("%s", city)
	postcode := string(ctx.QueryArgs().Peek("postcode"))
	if postcode != "" {
		data.Postcode = postcode
	}
	log.Printf("postcode:%s", postcode)

	latestProfiles, err := user.GetLatestProfiles(10)
	if err != nil {
		log.Printf("Error in routes/find-person.go FindPerson(): finding latest profiles': %s\n", err.Error())
		InternalServerErrorRedirect(ctx)
	}
	data.LatestProfiles = latestProfiles

	if len(jobList) != 0 || len(tags) != 0 || len(loc) != 0 {
		profiles, err := user.SearchProfiles(jobList, tags, data.Loc)
		if err != nil {
			log.Printf("Error in routes/find-person.go FindPerson(): searching profiles': %s\n", err.Error())
			InternalServerErrorRedirect(ctx)
		}
		data.Profiles = profiles
		data.Heading = strings.Join(data.Job, ", ")
	} else {
		data.Profiles = latestProfiles
		data.Heading = "Latest Profiles"
	}

	if ctx.UserValue("user") == nil {
		data.UserID = 0
	} else {
		currUser := getUserFromContext(ctx, true)
		data.UserID = currUser.UserID
	}
	for _, profileItem := range data.Profiles {
		FeedbackAverageItem := profileItem.UniversalReputationScore()
		data.FeedbackAverages = append(data.FeedbackAverages, *FeedbackAverageItem)
	}
	data.ParentPage = "find-person"
	data.serviceTypes = service_category.GetServiceCategory()
	pg := &render.Page{Title: "Find Person", TemplateFileName: "find-person.html", Data: data}
	pg.Render(ctx)
}

//DisplayPublicProfile renders the public profile page
func DisplayPublicProfile(ctx *fasthttp.RequestCtx) {

	profileUUID := (ctx.UserValue("profileUUID")).(string)

	//create data as struct variable for the view context
	var data = struct {
		Profile    *user.Profile
		UserID     int64
		ParentPage string
		ProfileID  int64
	}{}

	profile, err := user.GetProfileByProfileUUID(profileUUID)
	if err != nil {
		log.Printf("Error in routes/find-person.go DisplayPublicProfile(): finding profile with id x '%s': %s\n", profileUUID, err.Error())
		NotFoundRedirect(ctx)
	}

	//get current user's custom profile_id
	if ctx.UserValue("user") == nil {
		data.UserID = 0
	} else {
		currUser := getUserFromContext(ctx, true)
		data.UserID = currUser.UserID
		profiles, err := user.GetProfileByUserID(data.UserID)
		if err != nil {
			log.Println("Error: ", err)
		}
		data.ProfileID = profiles[0].ProfileID
	}
	data.Profile = profile

	data.ParentPage = "find-person"
	pg := &render.Page{Title: "Public Profile", TemplateFileName: "public-profile.html", Data: data}
	pg.Render(ctx)
}
