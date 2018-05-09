package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrChildNotFound = errors.New("child not found")
)

type Child struct {
	ChildId             sql.NullString
	DaycareId           sql.NullString
	FirstName           sql.NullString
	LastName            sql.NullString
	BirthDate           time.Time
	Gender              sql.NullString
	StartDate           time.Time
	ImageUri            sql.NullString
	Notes               sql.NullString
	SpecialInstructions SpecialInstructions `sql:"-"`
	Allergies           Allergies           `sql:"-"`
}

func (s *Store) AddChild(tx *gorm.DB, child Child) (Child, error) {
	db := s.dbOrTx(tx)

	child.ChildId = DbNullString(s.StringGenerator.GenerateUuid())

	if err := db.Create(&child).Error; err != nil {
		return Child{}, err
	}

	return child, nil
}

func (s *Store) DeleteChild(tx *gorm.DB, childId string) (err error) {
	db := s.dbOrTx(tx)

	if !s.childExists(db, childId) {
		return ErrChildNotFound
	}

	if err := db.Where("child_id = ?", childId).Delete(&Child{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) childExists(tx *gorm.DB, childId string) bool {
	c := Child{ChildId: sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}}
	return !tx.Model(Child{}).Where("child_id = ?", childId).First(&c).RecordNotFound()
}

func (s *Store) GetChild(tx *gorm.DB, childId string, options SearchOptions) (Child, error) {
	db := s.dbOrTx(tx)

	query := db.Table("children").
		Select("children.child_id," +
			"children.daycare_id," +
			"children.first_name," +
			"children.last_name," +
			"children.gender," +
			"children.birth_date," +
			"children.start_date," +
			"children.image_uri," +
			"children.notes," +
			"(SELECT string_agg(special_instructions.instruction, ',') FROM special_instructions WHERE special_instructions.child_id = children.child_id)," +
			"(SELECT string_agg(allergies.allergy, ',')  FROM allergies WHERE allergies.child_id = children.child_id)")
	if options.ResponsibleId != "" {
		query = query.Joins("left join responsible_of ON responsible_of.child_id = children.child_id").Where("responsible_of.responsible_id = ?", options.ResponsibleId)
	}
	if options.DaycareId != "" {
		query = query.Where("children.daycare_id = ?", options.DaycareId)
	}
	query = query.Where("children.child_id = ?", childId)

	rows, err := query.Rows()
	if err != nil {
		return Child{}, err
	}
	children, err := s.scanChildRows(rows)
	if err != nil {
		return Child{}, err
	}

	if len(children) == 0 {
		return Child{}, ErrChildNotFound
	}
	return children[0], nil
}

func (s *Store) scanChildRows(rows *sql.Rows) ([]Child, error) {
	children := []Child{}
	for rows.Next() {
		currentChild := Child{}
		if err := rows.Scan(&currentChild.ChildId,
			&currentChild.DaycareId,
			&currentChild.FirstName,
			&currentChild.LastName,
			&currentChild.Gender,
			&currentChild.BirthDate,
			&currentChild.StartDate,
			&currentChild.ImageUri,
			&currentChild.Notes,
			&currentChild.SpecialInstructions,
			&currentChild.Allergies); err != nil {
			return []Child{}, err
		}
		for i := range currentChild.SpecialInstructions {
			currentChild.SpecialInstructions[i].ChildId = currentChild.ChildId
		}
		children = append(children, currentChild)
	}

	return children, nil
}

func (s *Store) ListChildren(tx *gorm.DB, options SearchOptions) ([]Child, error) {
	db := s.dbOrTx(tx)

	query := db.Table("children").
		Select("children.child_id," +
			"children.daycare_id," +
			"children.first_name," +
			"children.last_name," +
			"children.gender," +
			"children.birth_date," +
			"children.start_date," +
			"children.image_uri," +
			"children.notes," +
			"(SELECT string_agg(special_instructions.instruction, ',') FROM special_instructions WHERE special_instructions.child_id = children.child_id)," +
			"(SELECT string_agg(allergies.allergy, ',') FROM allergies WHERE allergies.child_id = children.child_id)")
	if options.ResponsibleId != "" {
		query = query.Joins("left join responsible_of ON responsible_of.child_id = children.child_id").Where("responsible_of.responsible_id = ?", options.ResponsibleId)
	}
	if options.DaycareId != "" {
		query = query.Where("children.daycare_id = ?", options.DaycareId)
	}
	// TODO: when https://github.com/Vinubaba/SANTC-API/issues/19 is done
	/*if options.TeacherId != "" {
		query = query.Joins("left join roles ON roles.user_id = users.user_id")
	}*/

	query.LogMode(true)
	rows, err := query.Rows()
	if err != nil {
		return []Child{}, err
	}
	children, err := s.scanChildRows(rows)
	if err != nil {
		return []Child{}, err
	}

	return children, nil
}

func (s *Store) UpdateChild(tx *gorm.DB, child Child) (Child, error) {
	db := s.dbOrTx(tx)

	res := db.Where("child_id = ?", child.ChildId).Model(&Child{}).Updates(child).First(&child)
	if res.RecordNotFound() {
		return Child{}, ErrChildNotFound
	}
	if err := res.Error; err != nil {
		return Child{}, err
	}

	return s.GetChild(db, child.ChildId.String, SearchOptions{})
}
