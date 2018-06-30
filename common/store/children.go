package store

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrChildNotFound  = errors.New("child not found")
	ErrSetResponsible = errors.New("failed to set responsible")
)

type Child struct {
	ChildId             sql.NullString
	DaycareId           sql.NullString
	ClassId             sql.NullString
	ScheduleId          sql.NullString
	AddressSameAs       sql.NullString
	FirstName           sql.NullString
	LastName            sql.NullString
	BirthDate           time.Time
	Gender              sql.NullString
	StartDate           time.Time
	ImageUri            sql.NullString
	Notes               sql.NullString
	SpecialInstructions SpecialInstructions `sql:"-"`
	Allergies           Allergies           `sql:"-"`
	ResponsibleId       sql.NullString      `sql:"-"`
	Relationship        sql.NullString      `sql:"-"`
	Schedule            Schedule            `sql:"-"`
}

func (s *Store) AddChild(tx *gorm.DB, child Child) (Child, error) {
	db := s.dbOrTx(tx)

	schedule, err := s.AddSchedule(db, child.Schedule)
	if err != nil {
		return Child{}, err
	}
	child.ScheduleId = schedule.ScheduleId
	child.Schedule = schedule

	child.ChildId = s.newId()
	if err := db.Create(&child).Error; err != nil {
		return Child{}, err
	}

	for i, specialInstruction := range child.SpecialInstructions {
		var err error
		specialInstruction.ChildId = child.ChildId
		child.SpecialInstructions[i], err = s.AddSpecialInstruction(tx, specialInstruction)
		if err != nil {
			return Child{}, errors.Wrap(err, "failed to set instruction")
		}
	}

	for i, allergy := range child.Allergies {
		var err error
		allergy.ChildId = child.ChildId
		child.Allergies[i], err = s.AddAllergy(tx, allergy)
		if err != nil {
			return Child{}, errors.Wrap(err, "failed to set allergy")
		}
	}

	if err := s.SetResponsible(tx, ResponsibleOf{
		ResponsibleId: child.ResponsibleId.String,
		ChildId:       child.ChildId.String,
		Relationship:  child.Relationship.String,
	}); err != nil {
		return Child{}, errors.Wrap(ErrSetResponsible, err.Error())
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

func (s *Store) baseChildQuery(tx *gorm.DB) *gorm.DB {
	db := s.dbOrTx(tx)
	query := db.Table("children").Select(
		"children.child_id," +
			"children.daycare_id," +
			"children.class_id," +
			"children.schedule_id," +
			"children.address_same_as," +
			"children.first_name," +
			"children.last_name," +
			"children.gender," +
			"children.birth_date," +
			"children.start_date," +
			"children.image_uri," +
			"children.notes," +
			"allergies.allergy_id," +
			"allergies.child_id," +
			"allergies.allergy," +
			"allergies.instruction," +
			"responsible_of.responsible_id," +
			"responsible_of.relationship," +
			"special_instructions.special_instruction_id," +
			"special_instructions.child_id," +
			"special_instructions.instruction," +
			"schedules.schedule_id," +
			"schedules.walk_in," +
			"schedules.monday_start," +
			"schedules.monday_end," +
			"schedules.tuesday_start," +
			"schedules.tuesday_end," +
			"schedules.wednesday_start," +
			"schedules.wednesday_end," +
			"schedules.thursday_start," +
			"schedules.thursday_end," +
			"schedules.friday_start," +
			"schedules.friday_end," +
			"schedules.saturday_start," +
			"schedules.saturday_end," +
			"schedules.sunday_start," +
			"schedules.sunday_end")
	query = query.Joins("left join allergies ON allergies.child_id = children.child_id")
	query = query.Joins("left join special_instructions ON special_instructions.child_id = children.child_id")
	query = query.Joins("left join responsible_of ON responsible_of.child_id = children.child_id")
	query = query.Joins("left join schedules ON schedules.schedule_id = children.schedule_id")
	return query
}

func (s *Store) GetChild(tx *gorm.DB, childId string, options SearchOptions) (Child, error) {
	db := s.dbOrTx(tx)

	query := s.baseChildQuery(db)
	if options.ResponsibleId != "" {
		query = query.Where("responsible_of.responsible_id = ?", options.ResponsibleId)
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
	children := make([]Child, 0)
	for rows.Next() {
		currentChild := Child{}
		allergy := Allergy{}
		specialInstruction := SpecialInstruction{}
		schedule := Schedule{}
		if err := rows.Scan(&currentChild.ChildId,
			&currentChild.DaycareId,
			&currentChild.ClassId,
			&currentChild.ScheduleId,
			&currentChild.AddressSameAs,
			&currentChild.FirstName,
			&currentChild.LastName,
			&currentChild.Gender,
			&currentChild.BirthDate,
			&currentChild.StartDate,
			&currentChild.ImageUri,
			&currentChild.Notes,
			&allergy.AllergyId,
			&allergy.ChildId,
			&allergy.Allergy,
			&allergy.Instruction,
			&currentChild.ResponsibleId,
			&currentChild.Relationship,
			&specialInstruction.SpecialInstructionId,
			&specialInstruction.ChildId,
			&specialInstruction.Instruction,
			&schedule.ScheduleId,
			&schedule.WalkIn,
			&schedule.MondayStart,
			&schedule.MondayEnd,
			&schedule.TuesdayStart,
			&schedule.TuesdayEnd,
			&schedule.WednesdayStart,
			&schedule.WednesdayEnd,
			&schedule.ThursdayStart,
			&schedule.ThursdayEnd,
			&schedule.FridayStart,
			&schedule.FridayEnd,
			&schedule.SaturdayStart,
			&schedule.SaturdayEnd,
			&schedule.SundayStart,
			&schedule.SundayEnd,
		); err != nil {
			return []Child{}, err
		}
		if s.childAlreadyScanned(children, currentChild) {
			for i, c := range children {
				if c.ChildId.String == currentChild.ChildId.String {
					children[i].Allergies.add(allergy)
					children[i].SpecialInstructions.add(specialInstruction)
				}
			}
		} else {
			currentChild.Allergies.add(allergy)
			currentChild.SpecialInstructions.add(specialInstruction)
			currentChild.Schedule = schedule
			children = append(children, currentChild)
		}
	}
	return children, nil
}

func (s *Store) childAlreadyScanned(children []Child, child Child) bool {
	for _, c := range children {
		if c.ChildId.String == child.ChildId.String {
			return true
		}
	}
	return false
}

func (s *Store) ListChildren(tx *gorm.DB, options SearchOptions) ([]Child, error) {
	db := s.dbOrTx(tx)

	query := s.baseChildQuery(db)
	if options.ResponsibleId != "" {
		query = query.Where("responsible_of.responsible_id = ?", options.ResponsibleId)
	}
	if options.DaycareId != "" {
		query = query.Where("children.daycare_id = ?", options.DaycareId)
	}
	// TODO: when https://github.com/Vinubaba/SANTC-API/issues/19 is done
	/*if options.TeacherId != "" {
		query = query.Joins("left join roles ON roles.user_id = users.user_id")
	}*/

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

func (s *Store) UpdateChild(tx *gorm.DB, child Child) error {
	var db *gorm.DB
	var mustCommitHere bool

	if tx != nil {
		// Caller is responsible for commiting the transaction
		mustCommitHere = false
		db = tx
	} else {
		// Transaction is fully handled here
		mustCommitHere = true
		db = s.Tx()
	}

	res := db.Where("child_id = ?", child.ChildId).Model(&Child{}).Updates(child).First(&child)
	if res.RecordNotFound() {
		db.Rollback()
		return ErrChildNotFound
	}
	if err := res.Error; err != nil {
		db.Rollback()
		return err
	}

	if len(child.SpecialInstructions) > 0 {
		if err := s.RemoveChildSpecialInstructions(db, child.ChildId.String); err != nil {
			db.Rollback()
			return err
		}
		for i, specialInstruction := range child.SpecialInstructions {
			var err error
			specialInstruction.ChildId = child.ChildId
			child.SpecialInstructions[i], err = s.AddSpecialInstruction(db, specialInstruction)
			if err != nil {
				db.Rollback()
				return errors.Wrap(err, "failed to set instruction")
			}
		}
	}

	if len(child.Allergies) > 0 {
		if err := s.RemoveAllergiesOfChild(db, child.ChildId.String); err != nil {
			db.Rollback()
			return err
		}
		for _, allergy := range child.Allergies {
			allergy.ChildId = child.ChildId
			_, err := s.AddAllergy(db, allergy)
			if err != nil {
				db.Rollback()
				return errors.Wrap(err, "failed to set allergy")
			}
		}
	}

	if child.ResponsibleId.String != "" {
		if err := s.RemoveChildResponsible(db, child.ChildId.String); err != nil {
			db.Rollback()
			return err
		}
		if err := s.SetResponsible(db, ResponsibleOf{
			ResponsibleId: child.ResponsibleId.String,
			ChildId:       child.ChildId.String,
			Relationship:  child.Relationship.String,
		}); err != nil {
			db.Rollback()
			return errors.Wrap(ErrSetResponsible, err.Error())
		}
	}

	if mustCommitHere {
		db.Commit()
	}
	return nil
}
