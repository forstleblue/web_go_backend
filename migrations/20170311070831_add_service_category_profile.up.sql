ALTER TABLE ONLY profiles	
	ADD COLUMN service_category bigint,
	ADD CONSTRAINT profiles_service_category_foreign_key 
	FOREIGN KEY (service_category)  REFERENCES servicecategory (service_id);