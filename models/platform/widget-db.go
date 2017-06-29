package platform

import (
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

const widgetTableName = "widgets"

//Update saves the profile to the database
func (w *Widget) Update() error {
	conn := db.Connect()
	defer conn.Close()

	w.Updated = time.Now()

	_, err := conn.Exec("Update "+widgetTableName+" Set configuration=$1, session_token=$2, updated=$3 where widget_id = $4",
		w.Configuration, w.SessionToken.String(), w.Updated, w.WidgetID.String())

	return err
}

func GetWidget(id string) (*Widget, error) {

	conn := db.Connect()
	defer conn.Close()
	widget, err := scanSingleWidget(conn.QueryRow("select "+fieldList(true)+" from "+widgetTableName+" where widget_id = $1", id))

	if err != nil {
		log.Println("No widget found in DB by widget.GetWidget() with value '" + id + "'")
		return nil, err
	}
	return widget, err
}

func GetWidgetBySessionToken(token gocql.UUID) (*Widget, error) {

	conn := db.Connect()
	defer conn.Close()
	widget, err := scanSingleWidget(conn.QueryRow("select "+fieldList(true)+" from "+widgetTableName+" where session_token = $1", token.String()))

	if err != nil {
		log.Println("No widget found in DB by widget.GetWidget() with session token '" + token.String() + "'")
		return nil, err
	}
	return widget, err
}

// HELPER METHODS
func fieldList(withID bool) string {
	list := "type, owner_id, owner_type, configuration, session_token, created, updated"

	if withID {
		list = "widget_id, " + list
	}
	return list
}

func scanSingleWidget(row *pgx.Row) (*Widget, error) {
	var widgetID, widgetType, ownerID, ownerType, sessionToken pgx.NullString
	var configuration interface{}
	var created, updated pgx.NullTime
	err := row.Scan(&widgetID, &widgetType, &ownerID, &ownerType, &configuration, &sessionToken, &created, &updated)

	if err != nil {
		log.Println("Error scanning single widget row: ", err.Error())
		return nil, err
	}

	widget := &Widget{}
	if widgetID.Valid {
		widget.WidgetID, _ = gocql.ParseUUID(widgetID.String)
	}
	if widgetType.Valid {
		widget.Type = widgetType.String
	}
	if ownerID.Valid {
		widget.OwnerID, _ = gocql.ParseUUID(ownerID.String)
	}
	if ownerType.Valid {
		widget.OwnerType = ownerType.String
	}
	if sessionToken.Valid {
		widget.SessionToken, _ = gocql.ParseUUID(sessionToken.String)
	}
	if created.Valid {
		widget.Created = created.Time
	}
	if updated.Valid {
		widget.Updated = updated.Time
	}
	widget.Configuration = configuration
	return widget, nil
}
