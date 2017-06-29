package utils

import (
	"log"

	"errors"

	geo "github.com/kellydunn/golang-geo"
	"github.com/unirep/ur-local-web/app/config"
)

var appConfig = config.Config()

//GetLatLng uses google geocoding to get the lat long from the address details
func GetLatLng(data string) (float64, float64, error) {

	if appConfig.GoogleAPIKey == "" {
		return 0, 0, errors.New("GoogleAPIKey must be set in the config")
	}

	geo.SetGoogleAPIKey(appConfig.GoogleAPIKey)
	g := new(geo.GoogleGeocoder)
	point, err := g.Geocode(data)
	if err != nil {
		return 0, 0, err
	}
	log.Printf("Point (lat, lng) '%f, %f'", point.Lat(), point.Lng())
	return point.Lat(), point.Lng(), nil
}
