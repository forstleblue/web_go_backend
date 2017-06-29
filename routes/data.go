package routes

import (
	"log"
	//"fmt"
	"encoding/json"

	"github.com/unirep/ur-local-web/app/db"
	"github.com/unirep/ur-local-web/app/render"
	"github.com/valyala/fasthttp"
)

func GetServiceCategories(ctx *fasthttp.RequestCtx) {
	type ServiceCategories struct {
		Id   int64  `json:"id"`
		Name string `json:"text"`
	}

	var serviceCategories = []ServiceCategories{}

	conn := db.Connect()
	defer conn.Close()
	rows, err := conn.Query("select service_name, service_id  from servicecategory")

	if err != nil {
		log.Printf("Error in routes/data.go GetServiceCategories(ctx *fasthttp.RequestCtx):Read servicecategory table failed %s\n", err.Error())
	} else {

		var id int64
		var name string

		for rows.Next() {
			err = rows.Scan(&name, &id)
			category := ServiceCategories{Name: name, Id: id}
			serviceCategories = append(serviceCategories, category)
			log.Println("**", name, id)
			if err != nil {
				log.Printf("Error in routes/data.go GetServiceCategories(ctx *fasthttp.RequestCtx):Read servicecategory table failed  %s\n", err.Error())
			}
		}
	}

	jsonb, err := json.Marshal(serviceCategories)
	json := string(jsonb)
	render.JSON(ctx, json, "ok", fasthttp.StatusOK)
}
