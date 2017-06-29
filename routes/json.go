package routes

//JSONFormError used for returning form errors in a JSON call
type JSONFormError struct {
	Form      string `json:"form"`
	FieldName string `json:"fieldName"`
	Redirect  string `json:"redirect"`
	Error     string `json:"error"` // Once some sort of localisation is implemented, this should probably be a key instead of actual text
}

//JSONFormSuccess used for a successful form call
type JSONFormSuccess struct {
	Redirect     string `json:"redirect"`
	NewCSRFToken string `json:"newCSRFToken"`
}

//JSONSelect2Item used for select2
type JSONSelect2Item struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}
