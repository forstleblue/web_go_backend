package sdatypes

type SdaType struct {
	SdaID       int64    `json:"sda_id"`
	RefID       int64    `json:"ref_id"`
	ProfileType string   `json:"profile_type"`
	SdaList     []string `json:"sda_list"`
}

func (sda *SdaType) GetSdaList() []string {
	list := GetSdaListWithID(sda.RefID, sda.ProfileType)
	return list
}
