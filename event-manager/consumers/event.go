package consumers

type Event struct {
	Type     string `json:"type"`
	SenderId string `json:"senderId"`
	*ImageApproval
}

type ImageApproval struct {
	Image   string `json:"image"`
	ChildId string `json:"childId"`
}
