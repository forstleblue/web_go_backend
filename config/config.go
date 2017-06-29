package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

//DatabaseConnectionParams contains all settings required to connect to a database
type DatabaseConnectionParams struct {
	Host         string
	DatabaseName string
	Username     string
	Password     string
	LoadData     bool
	LoadFile     string
}

//PaypalAPI contains all settings required to connect to the PayPal API
type PaypalAPI struct {
	URL         string
	RedirectURL string
	Username    string
	Password    string
	Signature   string
	ClientID    string
	Secret      string
	AppID       string
}

//Recaptcha for Reset Password
type Recaptcha struct {
	SiteKey   string
	SecretKey string
}

//EbayAPI contains all settings required to connect to the Ebay API
type EbayAPI struct {
	URL         string
	RedirectURL string
	AppID       string
	DevID       string
	Secret      string
	Version     string
}

//OAuthSettings is a generic oath container
type OAuthSettings struct {
	Key    string
	Secret string
	Name   string
}

//SecuritySettings contains all website security settings
type SecuritySettings struct {
	AllowedHosts          []string
	STSSeconds            int
	CustomFrameOptions    string
	ContentTypeNosniff    bool
	ContentSecurityPolicy string
	BrowserXSSFilter      bool
	PublicKey             string
}

//EmailSettings contains all email settings
type EmailSettings struct {
	AWSRegion          string
	CredentialsSet     string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	FromEmail          string
}

//Configuration stores global site parameters
type Configuration struct {
	IsProduction             bool
	LogRequests              bool
	Locales                  [][]string
	Host                     string
	SSLHost                  string
	SSLRedirect              bool
	Port                     string
	Protocol                 string
	URL                      string
	PageCacheDurationMinutes int
	DB                       DatabaseConnectionParams
	Security                 SecuritySettings
	PayPalAPI                PaypalAPI
	Recaptcha                Recaptcha
	PayPalAuth               OAuthSettings
	EbayAPI                  EbayAPI
	FaceBookAuth             OAuthSettings
	Email                    EmailSettings
	GoogleAPIKey             string
	PaymentSandbox           bool
	PaypalEmail              string
	AdminEmail               string
	SalesEmail               string
	SupportEmail             string
}

//ReCAPTCHA V2 settings (see https://developers.google.com/recaptcha/intro)
type ReCAPTCHA struct {
	SiteKey    string
	SiteSecret string
}

var config *Configuration
var version string

//Config loads and caches configuration for dev or production
func Config() *Configuration {

	if config != nil {
		return config
	}

	mode := "DEV"

	//TODO: Need to test production mode locally before deploying on server
	if os.Getenv("ENV") == "PROD" {
		mode = "PROD"
		config = loadJSON("./config/config-prod.json")
		config.IsProduction = true
	} else if os.Getenv("ENV") == "STAGE" {
		mode = "STAGE"
		config = loadJSON("./config/config-stage.json")
		config.IsProduction = false
	} else if os.Getenv("ENV") == "TEST" {
		mode = "TEST"
		config = loadJSON("./config/config-test.json")
		config.IsProduction = false
	} else {
		config = loadJSON("./config/config-dev.json")
		config.IsProduction = false
	}
	if len(config.SSLHost) > 0 {
		log.Println("Running as "+mode+" on", "https://"+config.SSLHost)
	} else {
		log.Println("Running as "+mode+" on", config.Protocol+config.Host+config.Port)
	}
	return config
}

//FullBaseURL returns the full base URL
func FullBaseURL() string {
	Config()
	if len(config.SSLHost) > 0 {
		return "https://" + config.SSLHost
	}
	return config.Protocol + config.Host + config.Port
}

func loadJSON(path string) *Configuration {
	//http://blog.2c-why.com/posts/2013-10-23-Part-009-Template-Function-With-Multiple-Arguments.html
	var config Configuration
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Cannot load config file at /config/config.go loadJSON(): " + err.Error())
	}
	err = json.Unmarshal(file, &config)

	if err != nil {
		log.Fatal("Cannot unmarshal config file at /config/config.go loadJSON(): " + err.Error())
	}

	if len(config.SSLHost) == 0 && config.Port == "" {
		config.Port = ":3000"
	}

	return &config
}

func Version() string {
	return version
}
