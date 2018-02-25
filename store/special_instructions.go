package store

import (
	"database/sql"
	"errors"
	"github.com/jinzhu/gorm"
	"strings"
)

type SpecialInstructions []SpecialInstruction

func (s *SpecialInstructions) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case string:
		instructions := strings.Split(v, ",")
		for _, instruction := range instructions {
			*s = append(*s, SpecialInstruction{Instruction: DbNullString(instruction)})
		}
	default:
		return errors.New("need string with roles separated by virgula")
	}
	return nil
}

func (s SpecialInstructions) ToList() []string {
	instructions := make([]string, 0)
	for _, instruction := range s {
		instructions = append(instructions, instruction.Instruction.String)
	}
	return instructions
}

type SpecialInstruction struct {
	ChildId     sql.NullString
	Instruction sql.NullString
}

func (SpecialInstruction) TableName() string {
	return "special_instructions"
}

func (s *Store) AddSpecialInstruction(tx *gorm.DB, specialInstruction SpecialInstruction) error {
	db := s.dbOrTx(tx)

	return db.Create(&specialInstruction).Error
}

func (s *Store) RemoveChildSpecialInstructions(tx *gorm.DB, childId string) error {
	db := s.dbOrTx(tx)
	return db.Where("child_id = ?", childId).Delete(&SpecialInstruction{}).Error
}
