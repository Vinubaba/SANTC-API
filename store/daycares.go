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

func (s *Store) AddDaycare(tx *gorm.DB, daycare Daycare) (Daycare, error) {
	db := s.dbOrTx(tx)

	daycare.DaycareId = DbNullString(s.StringGenerator.GenerateUuid())

	if err := db.Create(&daycare).Error; err != nil {
		return Daycare{}, err
	}

	return daycare, nil
}

func (s *Store) DeleteDaycare(tx *gorm.DB, daycareId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.daycareExists(db, daycareId) {
		return ErrDaycareNotFound
	}

	if err := db.Where("daycare_id = ?", daycareId).Delete(&Daycare{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) daycareExists(tx *gorm.DB, daycareId string) bool {
	c := Daycare{DaycareId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(Daycare{}).Where("daycare_id = ?", daycareId).First(&c).RecordNotFound()
}

func (s *Store) GetDaycare(tx *gorm.DB, daycareId string, options SearchOptions) (Daycare, error) {
	db := s.dbOrTx(tx)
	query := db.Table("daycares").
		Select("daycares.daycare_id," +
			"daycares.name," +
			"daycares.address_1," +
			"daycares.address_2," +
			"daycares.city," +
			"daycares.state," +
			"daycares.zip")
	query = query.Where("daycares.daycare_id = ?", daycareId)

	rows, err := query.Rows()
	if err != nil {
		return Daycare{}, err
	}

	daycares, err := s.scanDaycareRows(rows)
	if err != nil {
		return Daycare{}, err
	}
	if len(daycares) == 0 {
		return Daycare{}, ErrDaycareNotFound
	}

	return daycares[0], nil
}

func (s *Store) ListDaycare(tx *gorm.DB, options SearchOptions) ([]Daycare, error) {
	db := s.dbOrTx(tx)

	query := db.Table("daycares").
		Select("daycares.daycare_id," +
			"daycares.name," +
			"daycares.address_1," +
			"daycares.address_2," +
			"daycares.city," +
			"daycares.state," +
			"daycares.zip")

	rows, err := query.Rows()
	if err != nil {
		return []Daycare{}, err
	}

	daycares, err := s.scanDaycareRows(rows)
	if err != nil {
		return []Daycare{}, err
	}

	return daycares, nil
}

func (s *Store) UpdateDaycare(tx *gorm.DB, daycare Daycare) (Daycare, error) {
	db := s.dbOrTx(tx)

	res := db.Where("daycare_id = ?", daycare.DaycareId).Model(&Daycare{}).Updates(daycare).First(&daycare)
	if res.RecordNotFound() {
		return Daycare{}, ErrDaycareNotFound
	}
	if err := res.Error; err != nil {
		return Daycare{}, err
	}

	return s.GetDaycare(db, daycare.DaycareId.String, SearchOptions{})
}
