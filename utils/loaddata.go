package utils

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/unirep/ur-local-web/app/db"
)

func init() {

	if appConfig.DB.LoadData {
		loadfile := "./db/data-dev.sql"
		if appConfig.DB.LoadFile != "" {
			loadfile = appConfig.DB.LoadFile
		}
		log.Println("Loading data from file: " + loadfile)
		success := LoadData(loadfile)

		if success {
			log.Println("Finished loading data, consider altering the config file so the DB.LoadData property is false.")
		}
	}
}

// LoadData loads data from sql file
func LoadData(filename string) bool {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		// handle error
		log.Println("Error reading file: "+filename, err.Error())
		return false
	}

	conn := db.Connect()
	defer conn.Close()

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		request = strings.Trim(request, " ")
		if request != "" {
			result, err := conn.Exec(request)
			// do whatever you need with result and error
			if err != nil {
				log.Println("Error executing request: "+request, err.Error())
				log.Println("Stopping load, please fix SQL file before continuing")
				return false
			}
			log.Println("Result: " + result)
		}
	}
	return true
}
