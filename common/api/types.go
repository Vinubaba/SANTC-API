package api

type PhotoRequestTransport struct {
	Filename *string `json:"filename"`
	ChildId  *string `json:"childId"`
	SenderId *string `json:"senderId"`
}

type ChildTransport struct {
	Id                  *string                       `json:"id"`
	DaycareId           *string                       `json:"daycareId"`
	ClassId             *string                       `json:"classId"`
	AddressSameAs       *string                       `json:"addressSameAs"`
	FirstName           *string                       `json:"firstName"`
	LastName            *string                       `json:"lastName"`
	BirthDate           *string                       `json:"birthDate"` // dd/mm/yyyy
	Gender              *string                       `json:"gender"`
	ImageUri            *string                       `json:"imageUri"`
	StartDate           *string                       `json:"startDate"` // dd/mm/yyyy
	Notes               *string                       `json:"notes"`
	Allergies           []AllergyTransport            `json:"allergies"`
	ResponsibleId       *string                       `json:"responsibleId"`
	Relationship        *string                       `json:"relationship"`
	SpecialInstructions []SpecialInstructionTransport `json:"specialInstructions"`
	Schedule            ScheduleTransport             `json:"schedule"`
}

type AllergyTransport struct {
	Id          *string `json:"id"`
	Allergy     *string `json:"allergy"`
	Instruction *string `json:"instruction"`
}

type SpecialInstructionTransport struct {
	Id          *string `json:"id"`
	ChildId     *string `json:"childId"`
	Instruction *string `json:"instruction"`
}

type ScheduleTransport struct {
	Id             *string `json:"id"`
	TeacherId      *string `json:"teacherId,omitempty"`
	ChildId        *string `json:"childId,omitempty"`
	WalkIn         *bool   `json:"walkIn"`
	MondayStart    *string `json:"mondayStart"`
	MondayEnd      *string `json:"mondayEnd"`
	TuesdayStart   *string `json:"tuesdayStart"`
	TuesdayEnd     *string `json:"tuesdayEnd"`
	WednesdayStart *string `json:"wednesdayStart"`
	WednesdayEnd   *string `json:"wednesdayEnd"`
	ThursdayStart  *string `json:"thursdayStart"`
	ThursdayEnd    *string `json:"thursdayEnd"`
	FridayStart    *string `json:"fridayStart"`
	FridayEnd      *string `json:"fridayEnd"`
	SaturdayStart  *string `json:"saturdayStart"`
	SaturdayEnd    *string `json:"saturdayEnd"`
	SundayStart    *string `json:"sundayStart"`
	SundayEnd      *string `json:"sundayEnd"`
}
