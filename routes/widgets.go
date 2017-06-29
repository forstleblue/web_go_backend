package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CloudyKit/jet"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/models/platform"
	"github.com/unirep/ur-local-web/app/models/user"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
)

var widgetviews = jet.NewHTMLSet("./views/widgets")

//ReputationWidget return formatted HTML for a profile card
func ReputationWidget(ctx *fasthttp.RequestCtx) {

	var data = struct {
		Profile *user.Profile
		Config  *platform.ReputationWidgetConfiguration
		FullURL string
	}{}

	data.FullURL = config.FullBaseURL()

	var profile *user.Profile
	var profileID string

	widgetTokenHidden := (ctx.UserValue("widgetID")).(string)
	log.Printf("Widget ID: %s\n", widgetTokenHidden)

	//Unhide token
	widgetToken, _ := utils.UnhideTimeUUID(widgetTokenHidden)
	isValid, _ := utils.IsHiddenUUIDValid(widgetTokenHidden, 24)

	if !isValid {
		log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Widget session token is not valid\n")
		return
	}

	w, err := platform.GetWidgetBySessionToken(widgetToken)
	if err != nil {
		log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Can't get widget from db: %s\n", err.Error())
		return
	}

	if w.Type != platform.WidgetTypeProfile {
		log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Wrong type of widget\n")
		return
	}
	if w.OwnerType == platform.OwnerTypePlatform {
		platformID := w.OwnerID
		p, err := platform.GetPlatform(platformID.String())
		if err != nil {
			log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Platform not found - %s \n", err.Error())
			return
		}
		if !p.HasWidgetAccess {
			log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Platform does not have access to profile widget \n")
			return
		}

		profileID = string(ctx.QueryArgs().Peek("pid"))
		profile, err = user.GetProfileByExternalIDAndProfileType(profileID, p.ProfileType)
		if err != nil {
			log.Printf("Error in routes/profile.go ReputationWidget(ctx *fasthttp.RequestCtx): finding profiles with external id '%s': %s\n", profileID, err.Error())
			return
		}

	} else {
		log.Printf("No other owner type is allowed at this stage. PLATFORM only. Owner Type: %s\n", w.OwnerType)
		return
	}

	config := platform.ReputationWidgetConfiguration{}
	if w.Configuration != nil {
		var ok bool
		config, ok = w.Configuration.(platform.ReputationWidgetConfiguration)
		if !ok {
			log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Could not properly cast config to reputation config\n")
			return
		}
	}
	data.Config = &config

	data.Profile = profile

	pg := &render.Page{Title: "Profile", TemplateFileName: "reputation-widget.html", Data: data}

	vw, err := widgetviews.GetTemplate("reputation-widget.html")
	if err != nil {
		log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Cant get template: %s\n", err.Error())
		return
	}

	buf := bytes.Buffer{}
	err = vw.Execute(&buf, nil, pg)
	if err != nil {
		log.Printf("Error in routes/widgets.go ReputationWidget(ctx *fasthttp.RequestCtx): Executing template: %s\n", err.Error())
		return
	}

	body := buf.Bytes()

	callback := ctx.FormValue("callback")
	jsonBytes, _ := json.Marshal(string(body))

	jsonBytes = []byte(fmt.Sprintf("%s(%s)", callback, jsonBytes))

	ctx.SetBody(jsonBytes)
	return
}

func ReputationProfile(ctx *fasthttp.RequestCtx) {

	//create data as struct variable for the view context
	var data = struct {
		Profile  *user.Profile
		Platform *platform.Platform
	}{}

	profileID := (ctx.UserValue("profileID")).(string)

	profile, err := user.GetProfileByProfileUUID(profileID)

	if err != nil {
		log.Printf("Error in routes/widgets.go ReputationProfile(ctx *fasthttp.RequestCtx): finding profile with uuid '%s': %s\n", profileID, err.Error())
		NotFoundRedirect(ctx)
	}

	platform, err := platform.GetPlatformByProfileType(profile.ProfileType)
	if err != nil {
		log.Printf("Error in routes/widgets.go ReputationProfile(ctx *fasthttp.RequestCtx): finding platform with profile type '%s': %s\n", profile.ProfileType, err.Error())
		NotFoundRedirect(ctx)
	}

	data.Profile = profile
	data.Platform = platform

	pg := &render.Page{Title: "Profile", TemplateFileName: "widgets/hybrid-profile.html", Data: data}
	pg.Render(ctx)
}
