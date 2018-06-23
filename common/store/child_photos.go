package store

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

type ChildPhoto struct {
	PhotoId         sql.NullString
	ChildId         sql.NullString
	PublishedBy     sql.NullString
	ApprovedBy      sql.NullString
	ImageUri        sql.NullString
	Approved        bool
	PublicationDate time.Time
}

func (s *Store) AddChildPhoto(tx *gorm.DB, childPhoto ChildPhoto) error {
	db := s.dbOrTx(tx)

	childPhoto.PhotoId = s.newId()
	if err := db.Create(&childPhoto).Error; err != nil {
		return err
	}
	return nil
}

func (s *Store) ApprovePhoto(tx *gorm.DB, photoId, approvedBy string) error {
	db := s.dbOrTx(tx)
	return db.Table("child_photos").
		Where("photo_id = ?", photoId).
		Update(map[string]interface{}{"approved": true, "approvedBy": approvedBy}).Error
}
