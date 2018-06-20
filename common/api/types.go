package api

type PhotoRequestTransport struct {
	Filename string `json:"filename"`
	ChildId  string `json:"childId"`
	SenderId string `json:"senderId"`
}

type ChildTransport struct {
	Id                  string                        `json:"id"`
	DaycareId           string                        `json:"daycareId"`
	ClassId             string                        `json:"classId"`
	FirstName           string                        `json:"firstName"`
	LastName            string                        `json:"lastName"`
	BirthDate           string                        `json:"birthDate"` // dd/mm/yyyy
	Gender              string                        `json:"gender"`
	ImageUri            string                        `json:"imageUri"`
	StartDate           string                        `json:"startDate"` // dd/mm/yyyy
	Notes               string                        `json:"notes"`
	Allergies           []AllergyTransport            `json:"allergies"`
	ResponsibleId       string                        `json:"responsibleId"`
	Relationship        string                        `json:"relationship"`
	SpecialInstructions []SpecialInstructionTransport `json:"specialInstructions"`
}

type AllergyTransport struct {
	Id          string `json:"id"`
	Allergy     string `json:"allergy"`
	Instruction string `json:"instruction"`
}

type SpecialInstructionTransport struct {
	Id          string `json:"id"`
	ChildId     string `json:"childId"`
	Instruction string `json:"instruction"`
}
