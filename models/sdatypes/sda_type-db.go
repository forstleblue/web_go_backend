package sdatypes

import "github.com/unirep/ur-local-web/app/db"

import "log"

const sdatypeTableName = "sda_types"

func GetSdaListWithID(refID int64, profileType string) []string {
	conn := db.Connect()
	defer conn.Close()

	row := conn.QueryRow("select sda_list from "+sdatypeTableName+" where ref_id=$1 AND profile_type=$2", refID, profileType)
	var sdaList []string
	err := row.Scan(&sdaList)
	if err != nil {
		log.Println("Error in sdaType/sda_types-db.go GetSdaList() failed to get sdaList:", err)
	}
	log.Println(sdaList)
	return sdaList
}
