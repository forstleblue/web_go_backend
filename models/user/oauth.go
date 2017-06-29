/*
This is a port of package https://github.com/markbates/goth to work with https://github.com/valyala/fasthttp
*/

package user

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"net/url"

	"github.com/franela/goreq"
	"github.com/kataras/go-sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/facebook"
	"github.com/unirep/ur-local-web/app/config"
	"github.com/unirep/ur-local-web/app/utils"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

var appConfig = config.Config()

//Query is used to get SessionID
type SessionIDQuery struct {
	XMLName   xml.Name `xml:"GetSessionIDResponse"`
	TimeStamp string   `xml:"Timestamp"`
	Ack       string   `xml:"Ack"`
	Build     string   `xml:"Build"`
	SessionID string   `xml:"SessionID"`
}

//AuthTokenQuery is used to get AuthToken
type AuthTokenQuery struct {
	XMLName            xml.Name `xml:"FetchTokenResponse"`
	TimeStamp          string   `xml:"Timestamp"`
	Ack                string   `xml:"Ack"`
	Version            string   `xml:"Version"`
	Build              string   `xml:"Build"`
	EBayAuthToken      string   `xml:"eBayAuthToken"`
	HardExpirationTime string   `xml:"HardExpirationTime"`
}

//EbaySiteID = 15, SiteName = "eBay Australia", Global ID="EBAY-AU", get more information from http://developer.ebay.com/devzone/finding/Concepts/SiteIDToGlobalID.html
const EbaySiteID = "15"

//EbayCompatibilityLevel = 10003, Release Date = 2017-Feb-17, get more information from http://developer.ebay.com/DevZone/XML/docs/ReleaseNotes.html
const EbayCompatibilityLevel = "1003"

// GothSessionKey is the key used to access the session store.
const GothSessionKey = "oauth"

var GothSessionsConfig = sessions.Config{Cookie: GothSessionKey,
	// see sessions_test.go on how to set encoder and decoder for cookie value(sessionid)
	Expires:                     time.Duration(100) * time.Hour,
	DisableSubdomainPersistence: false,
}
var GothSessions = sessions.New(GothSessionsConfig)

// GothParams used to convert the ctx.QueryArgs() to goth's params
type GothParams map[string]string

// Get returns the value of a Goth param
func (g GothParams) Get(key string) string {
	return g[key]
}

var (
	_ goth.Params = GothParams{}
)

func init() {
	var appConfig = config.Config()

	goth.UseProviders(

		facebook.New(appConfig.FaceBookAuth.Key, appConfig.FaceBookAuth.Secret, config.FullBaseURL()+"/auth/facebook/callback"),
		//paypal.New(appConfig.PayPalAuth.Key, appConfig.PayPalAuth.Secret, config.FullBaseURL()+"/auth/paypal/callback"),
		//uber.New(os.Getenv("UBER_KEY"), os.Getenv("UBER_SECRET"), config.FullBaseURL()+"/auth/uber/callback"),
	)

	m := make(map[string]string)
	m["facebook"] = "Facebook"
	//m["paypal"] = "Paypal"
	//m["uber"] = "Uber"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

}

/*
BeginAuthHandler is a convienence handler for starting the authentication process.
It expects to be able to get the name of the provider from the named parameters
as either "provider" or url query parameter ":provider".
BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.
*/
func BeginAuthHandler(ctx *fasthttp.RequestCtx) {
	url, err := GetAuthURL(ctx)
	if err != nil {
		utils.Log(ctx, "BeginAuthHandler", fasthttp.StatusBadRequest, 0, err.Error())
		return
	}
	ctx.Redirect(url, fasthttp.StatusOK)
}

// SetState sets the state string associated with the given request.
// If no state string is associated with the request, one will be generated.
// This state is sent to the provider and can be retrieved during the
// callback.
var SetState = func(ctx *fasthttp.RequestCtx) string {
	state := ctx.QueryArgs().Peek("state")
	if len(state) > 0 {
		return string(state)
	}
	return "state"
}

// GetState gets the state returned by the provider during the callback.
// This is used to prevent CSRF attacks, see
// http://tools.ietf.org/html/rfc6749#section-10.12
var GetState = func(ctx *fasthttp.RequestCtx) string {
	return string(ctx.QueryArgs().Peek("state"))
}

/*
GetAuthURL starts the authentication process with the request provided.
It will return a URL that should be used to send users to.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or url query parameter ":provider".
I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(ctx *fasthttp.RequestCtx) (string, error) {

	providerName, err := GetProviderName(ctx)
	if err != nil {
		return "", err
	}
	//goth doesn't support Ebay oauth.
	if providerName == "ebay" {
		SessionXML := `"<?xml version="1.0" encoding="utf-8"?><GetSessionIDRequest xmlns="urn:ebay:apis:eBLBaseComponents">
				<RuName>` + appConfig.EbayAPI.RedirectURL + `</RuName>
				</GetSessionIDRequest>"`
		req := goreq.Request{
			Method:      "POST",
			Uri:         appConfig.EbayAPI.URL,
			Body:        strings.NewReader(SessionXML),
			ContentType: "application/xml; charset=utf-8",
			UserAgent:   "go-ebay-fetch-orders",
			ShowDebug:   false,
		}
		req.AddHeader("X-EBAY-API-DEV-NAME", appConfig.EbayAPI.DevID)
		req.AddHeader("X-EBAY-API-CERT-NAME", appConfig.EbayAPI.Secret)
		req.AddHeader("X-EBAY-API-APP-NAME", appConfig.EbayAPI.AppID)
		req.AddHeader("X-EBAY-API-REQUEST-ENCODING", "XML")
		req.AddHeader("X-EBAY-API-RESPONSE-ENCODING", "XML")
		req.AddHeader("X-EBAY-API-SITEID", EbaySiteID)
		req.AddHeader("X-EBAY-API-COMPATIBILITY-LEVEL", EbayCompatibilityLevel)
		req.AddHeader("X-EBAY-API-CALL-NAME", "GetSessionID")

		res, _ := req.Do()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Failed read data in response: ", err)
			return "", nil
		}
		var Response SessionIDQuery
		xml.Unmarshal(data, &Response)
		log.Println("SessionID: ", Response.SessionID)

		session := GothSessions.StartFasthttp(ctx)
		if session == nil {
			log.Printf("Failed to get oauth session")
			return "", errors.New("Could not get guth session")
		}

		session.Set("EbaySessionID", Response.SessionID)
		return "https://signin.sandbox.ebay.com/ws/eBayISAPI.dll?SignIn&Runame=" + appConfig.EbayAPI.RedirectURL + "&SessID=" + Response.SessionID, nil
		//return "https://signin.sandbox.ebay.com/authorize?client_id=" + appConfig.EbayAPI.AppID + "&response_type=code&redirect_uri=" + appConfig.EbayAPI.RedirectURL + "&scope=https://api.ebay.com/oauth/api_scope", nil
	}
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(SetState(ctx))
	if err != nil {
		return "", err
	}
	log.Println("Session: ", sess)
	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}
	log.Println("URL: ", url)
	// Get session from store.

	session := GothSessions.StartFasthttp(ctx)
	if session == nil {
		log.Printf("Failed to get oauth session")
		return "", errors.New("Could not get guth session")
	}

	session.Set(GothSessionKey, sess.Marshal())

	return url, nil
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all of the basic information about the user from the provider.
It expects to be able to get the name of the provider from the named parameters
as either "provider" or url query parameter ":provider".
*/
func CompleteUserAuth(ctx *fasthttp.RequestCtx) (goth.User, error) {
	// Get session from store.
	// session, err := store.Get(ctx, GothSessionKey)
	session := GothSessions.StartFasthttp(ctx)
	if session == nil {
		log.Printf("Failed to get oauth session")
		return goth.User{}, errors.New("Could not get guth session")
	}
	providerName, err := GetProviderName(ctx)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	if session.Get(GothSessionKey) == nil {
		//TODO: instead of the message below, delete all sessions and cookies and ask browser to redirect to "/login-register"
		ctx.Redirect("/login-register", fasthttp.StatusSeeOther)
		return goth.User{}, errors.New("could not find a matching session for this request, redirecting to /login-register")
	}

	sess, err := provider.UnmarshalSession(session.Get(GothSessionKey).(string))
	if err != nil {
		return goth.User{}, err
	}

	m := map[string]string{}

	ctx.QueryArgs().VisitAll(func(key []byte, value []byte) {
		m[string(key)] = string(value)
	})

	_, err = sess.Authorize(provider, GothParams(m))

	if err != nil {
		return goth.User{}, err
	}

	return provider.FetchUser(sess)
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
func GetProviderName(ctx *fasthttp.RequestCtx) (string, error) {

	provider := string(ctx.QueryArgs().Peek("provider"))
	if len(provider) == 0 {
		provider = ctx.UserValue("provider").(string)
	}
	if len(provider) == 0 {
		return "", errors.New("you must select a provider")
	}
	return provider, nil
}

//ProviderIndex is a list of oauth providers used by this application for the goth package to process
type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

func unmarshalUser(tsvUser string) *goth.User {

	fields := strings.Split(tsvUser, "\t")
	if len(fields) != 11 {
		log.Println("Error in /models/user/oauth.go func unmarshalUser(): Could not parse goth.User fields. Expecting 11 fields, got", len(fields), "Tab separated user was:", tsvUser)
		return &goth.User{}
	}

	tim, err := time.Parse(time.RFC3339, fields[9])
	if err != nil {
		log.Println("Error in /models/user/oauth.go func unmarshalUser(): Could not parse goth.User 'ExpiresAt' field into time.Time. ExpiresAt was", fields[9], err.Error())
		return &goth.User{}
	}

	u := &goth.User{
		Provider:     fields[0],
		Name:         fields[1],
		Email:        fields[2],
		NickName:     fields[3],
		Location:     fields[4],
		AvatarURL:    fields[5],
		Description:  fields[6],
		UserID:       fields[7],
		AccessToken:  fields[8],
		ExpiresAt:    tim,
		RefreshToken: fields[10],
	}

	return u
}

func marshalUser(user goth.User) []byte {
	//TODO: measure performance difference of storing separator \t indexof as an array header inside the object, like Protocol Buffers
	tpl := "<Provider>\t<Name>\t<Email>\t<NickName>\t<Location>\t<AvatarURL>\t<Description>\t<UserID>\t<AccessToken>\t<ExpiresAt>\t<RefreshToken>"
	Tpl, err := fasttemplate.NewTemplate(tpl, "<", ">")
	if err != nil {
		log.Println("Error in /models/user/oauth.go func marshalUser(): Could not create JSON template", err.Error())
		return []byte{}
	}

	var usr bytes.Buffer

	toString := func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "Provider":
			return w.Write([]byte(user.Provider))
		case "Name":
			return w.Write([]byte(user.Name))
		case "Email":
			return w.Write([]byte(user.Email))
		case "NickName":
			return w.Write([]byte(user.NickName))
		case "Location":
			return w.Write([]byte(user.Location))
		case "AvatarURL":
			return w.Write([]byte(user.AvatarURL))
		case "Description":
			return w.Write([]byte(user.Description))
		case "UserID":
			return w.Write([]byte(user.UserID))
		case "AccessToken":
			return w.Write([]byte(user.AccessToken))
		case "ExpiresAt":
			return w.Write([]byte(user.ExpiresAt.Format(time.RFC3339)))
		case "RefreshToken":
			return w.Write([]byte(user.RefreshToken))
		default:
			return w.Write([]byte(fmt.Sprintf("[unknown tag %q]", tag)))
		}
	}

	if _, err = Tpl.ExecuteFunc(&usr, toString); err != nil {
		log.Println("Error executing fasttemplate.ExecuteFunc() in /models/user/oauth.go func marshalUser()", err.Error())
	}

	return usr.Bytes()

}

//Ebay oauth redirect path.
func GetUserToken(ctx *fasthttp.RequestCtx) {
	session := GothSessions.StartFasthttp(ctx)
	SessionID := session.Get("EbaySessionID")

	log.Println("SessionID: ", SessionID)

	status := ctx.UserValue("status").(string)
	// applicationToken := string(ctx.QueryArgs().Peek("code"))
	// log.Println("ApplicationToken: ", applicationToken)
	log.Println("Status: ", status)
	if status == "declined" {
		fmt.Printf("User doesn't give permission. Go back to your dashboard.")
		ctx.Redirect("/dashboard", fasthttp.StatusSeeOther)
	}
	//var appConfig = config.Config()
	// client := &http.Client{}
	// applicationTokenURLEncoded, _ := url.Parse(applicationToken)
	// body := url.Values{
	// 	"grant_type":   {"authorization_code"},
	// 	"code":         {applicationTokenURLEncoded.String()},
	// 	"redirect_uri": {"Liu_Jin-LiuJin-urlocalw-pjrrsriwo"},
	// }
	// reqBody := bytes.NewBufferString(body.Encode())
	// log.Println("Reqbody: ", reqBody)
	// req, _ := http.NewRequest("POST", "https://api.sandbox.ebay.com/identity/v1/oauth2/token", reqBody)
	// //authorization := appConfig.EbayAPI.AppID + ":" + appConfig.EbayAPI.Secret
	// authorization := "LiuJin-urlocalw-SBX-269e2c47f-82f2fb6d" + ":" + "SBX-69e2c47f1565-c185-4379-8e62-e9f7"
	// authorizationBase64 := base64.StdEncoding.EncodeToString([]byte(authorization))
	// req.Header.Add("Authorization", "Basic "+authorizationBase64)
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// log.Println("Requset: ", req)
	// resp, _ := client.Do(req)
	// log.Println("resp: ", resp)
	// log.Println("ResBody: ", resp.Body)
	TokenXML := `"<?xml version="1.0" encoding="utf-8"?>
				<FetchTokenRequest xmlns="urn:ebay:apis:eBLBaseComponents">
				<SessionID>` + SessionID.(string) + `</SessionID>
				</FetchTokenRequest>"`
	req := goreq.Request{
		Method:      "POST",
		Uri:         appConfig.EbayAPI.URL,
		Body:        strings.NewReader(TokenXML),
		ContentType: "application/xml; charset=utf-8",
		UserAgent:   "go-ebay-fetch-orders",
		ShowDebug:   false,
	}
	req.AddHeader("X-EBAY-API-DEV-NAME", appConfig.EbayAPI.DevID)
	req.AddHeader("X-EBAY-API-CERT-NAME", appConfig.EbayAPI.Secret)
	req.AddHeader("X-EBAY-API-APP-NAME", appConfig.EbayAPI.AppID)
	req.AddHeader("X-EBAY-API-REQUEST-ENCODING", "XML")
	req.AddHeader("X-EBAY-API-RESPONSE-ENCODING", "XML")
	req.AddHeader("X-EBAY-API-SITEID", EbaySiteID)
	req.AddHeader("X-EBAY-API-COMPATIBILITY-LEVEL", EbayCompatibilityLevel)
	req.AddHeader("X-EBAY-API-CALL-NAME", "FetchToken")
	res, _ := req.Do()
	data, _ := ioutil.ReadAll(res.Body)
	var Response AuthTokenQuery
	xml.Unmarshal(data, &Response)
	log.Println("Ack: ", Response.Ack)
	log.Println("AuthToken:", Response.EBayAuthToken)

	redirectURL := url.QueryEscape(string(Response.EBayAuthToken))
	log.Println("RedirectURL: ", redirectURL)
	ctx.Redirect("/add-ebay?authToken="+redirectURL+"&expiry="+Response.HardExpirationTime, fasthttp.StatusSeeOther)
}
