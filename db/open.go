package db

import (
	"log"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/config"
)

var connConfig pgx.ConnConfig

//Open connects to the database
func Open(dbConfig config.DatabaseConnectionParams) {

	if connConfig.Host == "" {

		connConfig = pgx.ConnConfig{
			User:              dbConfig.Username,
			Password:          dbConfig.Password,
			Host:              dbConfig.Host,
			Port:              5432,
			Database:          dbConfig.DatabaseName,
			TLSConfig:         nil,
			UseFallbackTLS:    false,
			FallbackTLSConfig: nil,
		}
	}
}

//Connect returns a connection to the database
func Connect() *pgx.Conn {
	conn, err := pgx.Connect(connConfig)
	if err != nil {
		log.Println("DB 500 0 Unable to establish connection: " + err.Error())
		//os.Exit(1)
	}
	return conn
}
