package platform

import (
	"log"
	"net"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const platformTableName = "platforms"

func GetPlatform(id string) (*Platform, error) {

	conn := db.Connect()
	defer conn.Close()
	platform, err := scanSinglePlatform(conn.QueryRow("select "+platformFieldList(true)+" from "+platformTableName+" where platform_id = $1", id))

	if err != nil {
		log.Println("No platform found in DB by platform.GetPlatform() with value '" + id + "'")
		return nil, err
	}
	return platform, err
}

func GetPlatformByServerIP(ip net.IP) (*Platform, error) {
	conn := db.Connect()
	defer conn.Close()

	//platform, err := scanSinglePlatform(conn.QueryRow("select "+platformFieldList(true)+" from "+platformTableName+" where $1 = ANY (server_ips)", ip))
	platform, err := scanSinglePlatform(conn.QueryRow("select " + platformFieldList(true) + " from " + platformTableName + " where api_configuration @> '{\"authentication\":[{\"IP\":\"" + ip.String() + "\"}]}'::jsonb"))

	if err != nil {
		log.Println("No platform found in DB by platform.GetPlatformByServerIP() with value '" + ip.String() + "'")
		return nil, err
	}
	return platform, err
}

func GetPlatformByProfileType(profileType string) (*Platform, error) {

	conn := db.Connect()
	defer conn.Close()
	platform, err := scanSinglePlatform(conn.QueryRow("select "+platformFieldList(true)+" from "+platformTableName+" where profile_type = $1", profileType))

	if err != nil {
		log.Println("No platform found in DB by platform.GetPlatformByProfileType() with value '" + profileType + "'")
		return nil, err
	}
	return platform, err
}

// HELPER METHODS
func platformFieldList(withID bool) string {
	list := "name, profile_type, widget_access, api_configuration, created, updated"

	if withID {
		list = "platform_id, " + list
	}
	return list
}

func scanSinglePlatform(row *pgx.Row) (*Platform, error) {
	var platformID, name, profileType pgx.NullString
	var hasWidgetAccess pgx.NullBool
	var created, updated pgx.NullTime
	var apiConfig APIConfiguration
	err := row.Scan(&platformID, &name, &profileType, &hasWidgetAccess, &apiConfig, &created, &updated)

	if err != nil {
		log.Println("Error scanning single platform row: ", err.Error())
		return nil, err
	}

	platform := &Platform{}
	if platformID.Valid {
		platform.PlatformID, _ = gocql.ParseUUID(platformID.String)
	}
	if name.Valid {
		platform.Name = name.String
	}
	if profileType.Valid {
		platform.ProfileType = profileType.String
	}
	if hasWidgetAccess.Valid {
		platform.HasWidgetAccess = hasWidgetAccess.Bool
	}
	if created.Valid {
		platform.Created = created.Time
	}
	if updated.Valid {
		platform.Updated = updated.Time
	}
	platform.APIConfiguration = &apiConfig
	return platform, nil
}
