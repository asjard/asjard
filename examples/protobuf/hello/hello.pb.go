package hello

// Say .
type Say struct {
	RegionID  string    `json:"region_id"`
	ProjectID string    `json:"project_id"`
	UserID    int64     `json:"user_id"`
	StrList   []string  `json:"str_list"`
	IntList   []int     `json:"int_list"`
	Obj       *SayObj   `json:"obj"`
	Objs      []*SayObj `json:"objs"`
}

// SayObj .
type SayObj struct {
	FieldInt int    `json:"field_int"`
	FieldStr string `json:"field_str"`
}
