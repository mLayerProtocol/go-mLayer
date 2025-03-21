package migration

import (
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)



 func DropTopicIdColumnFromMessageState(db *gorm.DB) (err error) {
	if !db.Migrator().HasTable(&models.MessageState{}) {
		return nil
	}
	return  db.Migrator().DropColumn(&models.MessageState{}, "topic_id")
 }


 func DropAttachmentsColumnFromMessageState(db *gorm.DB) (err error) {
	if !db.Migrator().HasTable(&models.MessageState{}) {
		return nil
	}
	return  db.Migrator().DropColumn(&models.MessageState{}, "attachments")
 }