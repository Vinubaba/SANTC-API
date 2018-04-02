package store

import (
	"database/sql"
	"errors"
	"github.com/jinzhu/gorm"
)

var (
	ErrDaycareNotFound = errors.New("daycare not found")
)

type Daycare struct {
	DaycareId sql.NullString
	Name      sql.NullString
	Address_1 sql.NullString
	Address_2 sql.NullString
	City      sql.NullString
	State     sql.NullString
	Zip       sql.NullString
}

func (s *Store) AddDaycare(tx *gorm.DB, daycare Daycare) (Daycare, error) {
	db := s.dbOrTx(tx)

	daycare.DaycareId = sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}
	if err := db.Create(&daycare).Error; err != nil {
		return Daycare{}, err
	}

	return daycare, nil
}

func (s *Store) GetPublicDaycare(tx *gorm.DB) (Daycare, error) {
	db := s.dbOrTx(tx)
	rows, err := db.Table("daycares").
		Select("daycares.daycare_id,"+
			"daycares.name,"+
			"daycares.address_1,"+
			"daycares.address_2,"+
			"daycares.city,"+
			"daycares.state,"+
			"daycares.zip").
		Where("daycares.daycare_id = ?", "PUBLIC").
		Rows()
	if err != nil {
		return Daycare{}, err
	}
	daycares, err := s.scanDaycareRows(rows)
	if err != nil {
		return Daycare{}, err
	}

	if len(daycares) > 0 {
		return daycares[0], nil
	}
	return Daycare{}, ErrDaycareNotFound
}

func (s *Store) scanDaycareRows(rows *sql.Rows) ([]Daycare, error) {
	daycares := []Daycare{}
	for rows.Next() {
		currentDaycare := Daycare{}
		if err := rows.Scan(&currentDaycare.DaycareId,
			&currentDaycare.Name,
			&currentDaycare.Address_1,
			&currentDaycare.Address_2,
			&currentDaycare.City,
			&currentDaycare.State,
			&currentDaycare.Zip); err != nil {
			return []Daycare{}, err
		}
		daycares = append(daycares, currentDaycare)
	}

	return daycares, nil
}
