package platform

import (
	"time"

	"net"

	"github.com/gocql/gocql"
)

type APIConfiguration struct {
	Authentication []*APIAuthentication `json:"authentication,omitempty"`
}

type APIAuthentication struct {
	IP net.IP `json:"IP,omitempty"`
}

type Platform struct {
	PlatformID       gocql.UUID        `json:"platform_id"`
	Name             string            `json:"name"`
	ProfileType      string            `json:"profile_type"`
	HasWidgetAccess  bool              `json:"widget_access"`
	Created          time.Time         `json:"created,omitempty" db:"created"`
	Updated          time.Time         `json:"updated,omitempty" db:"updated"`
	APIConfiguration *APIConfiguration `json:"api_configuration,omitempty"`
}

func (p *Platform) GetLogoURL() string {
	return "/images/platforms/" + p.ProfileType + ".png"
}
