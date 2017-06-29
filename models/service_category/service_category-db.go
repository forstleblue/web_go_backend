package service_category

import (
	"log"

	"github.com/jackc/pgx"
	"github.com/unirep/ur-local-web/app/db"
)

func GetServiceCategory() []ServiceTypes {
	var serviceTypes = []ServiceTypes{}
	conn := db.Connect()
	defer conn.Close()
	rows, err := conn.Query("select service_id, service_name,  service_style_key, parent_id    from servicecategory where service_id > 0 order by parent_id asc")

	var id int64
	var name pgx.NullString
	var style_key pgx.NullInt64
	var parent_id pgx.NullInt64

	if err != nil {
		log.Printf("Read servicecategory table failed")
	} else {
		for rows.Next() {
			err = rows.Scan(&id, &name, &style_key, &parent_id)
			if parent_id.Int64 == 0 {
				serviceTypeItem := ServiceTypes{Id: id, Text: name.String, Children: []ServiceCategories{}}
				serviceTypes = append(serviceTypes, serviceTypeItem)
			} else {
				for i, item := range serviceTypes {
					if item.Id == parent_id.Int64 {
						serviceCategoryItem := ServiceCategories{id, name.String, parent_id.Int64}
						serviceTypes[i].Children = append(serviceTypes[i].Children, serviceCategoryItem)
					}
				}
			}
		}
	}

	return serviceTypes
}
