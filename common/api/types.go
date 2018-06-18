package api

type PhotoRequestTransport struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
	ChildId  string `json:"childId"`
	SenderId string `json:"senderId"`
}
