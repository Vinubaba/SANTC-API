package store

import (
	"time"
)

type User struct {
	UserId string `gorm:"primary_key:true"`
}

type Child struct {
	//User User `gorm:"ForeignKey:user_id;AssociationForeignKey:child_id"`
	ChildId   string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	FirstName string
	LastName  string
	BirthDate time.Time
	Gender    string
}

type AdultResponsible struct {
	//User User `gorm:"ForeignKey:user_id;AssociationForeignKey:responsible_id"`
	ResponsibleId string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	FirstName     string
	LastName      string
	Gender        string
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
