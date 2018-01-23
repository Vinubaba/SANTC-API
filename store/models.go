package store

import (
	"time"
)

type User struct {
	UserId   string `gorm:"primary_key:true"`
	Email    string
	Password string
}

type Child struct {
	ChildId   string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	FirstName string
	LastName  string
	BirthDate time.Time
	Gender    string
}

type AdultResponsible struct {
	ResponsibleId string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	Email         string `sql:"type:varchar(128) REFERENCES users(email)"`
	FirstName     string
	LastName      string
	Gender        string
	Phone         string
	Addres_1      string
	Addres_2      string
	City          string
	State         string
	Zip           string
}

type ResponsibleOf struct {
	ResponsibleId string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	ChildId       string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	Relationship  string
}

func (ResponsibleOf) TableName() string {
	return "responsible_of"
}

const (
	REL_MOTHER      = "mother"
	REL_FATHER      = "father"
	REL_GRANDMOTHER = "grandmother"
	REL_GRANDFATHER = "grandfather"
	REL_GUARDIAN    = "guardian"
)

var (
	allRelationships = []string{REL_FATHER, REL_MOTHER, REL_GRANDFATHER, REL_GRANDMOTHER, REL_GUARDIAN}
)
