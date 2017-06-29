package service_category

import (
	"fmt"
)

type ServiceCategories struct {	    
	Id int64 `json:"id"`
	Name string `json:"text"`
	ParentId int64 `json:"parent_id"`
}	

type ServiceTypes struct{
	Id 		int64				`json:"id"`
	Text     string 			`json:"text"`
	Children []ServiceCategories  `json:"children"`	
}

func (serviceType ServiceTypes) getServiceCategoryName(){
	fmt.Println(serviceType)
}