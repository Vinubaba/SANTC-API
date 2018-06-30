package store

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

type SpecialInstruction struct {
	SpecialInstructionId sql.NullString
	ChildId              sql.NullString
	Instruction          sql.NullString
}

type SpecialInstructions []SpecialInstruction

func (s *SpecialInstructions) add(instruction SpecialInstruction) {
	if instruction.SpecialInstructionId.String == "" {
		return
	}
	for _, si := range *s {
		if si.SpecialInstructionId.String == instruction.SpecialInstructionId.String {
			return
		}
	}
	*s = append(*s, instruction)
}

func (SpecialInstruction) TableName() string {
	return "special_instructions"
}

func (s *Store) AddSpecialInstruction(tx *gorm.DB, specialInstruction SpecialInstruction) (SpecialInstruction, error) {
	db := s.dbOrTx(tx)
	specialInstruction.SpecialInstructionId = s.newId()
	err := db.Create(&specialInstruction).Error
	return specialInstruction, err
}

func (s *Store) RemoveChildSpecialInstructions(tx *gorm.DB, childId string) error {
	db := s.dbOrTx(tx)
	return db.Where("child_id = ?", childId).Delete(&SpecialInstruction{}).Error
}
