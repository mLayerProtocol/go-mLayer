package migration

import (
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)


 func DropOwnerColumnFromApplicationState(db *gorm.DB) (err error) {
	if !db.Migrator().HasTable(&models.ApplicationState{}) {
		return nil
	}
	return db.Migrator().DropColumn(&models.ApplicationState{}, "owner")
 }

//  func DropAgentColumnFromApplicationState(db *gorm.DB) (err error) {
// 	// if db.Migrator().HasColumn(&models.ApplicationState{}, "Agent") {
// 	// 	err = db.Migrator().DropColumn(&models.ApplicationState{}, "Agent")
// 	// }
// 	return  db.Migrator().DropColumn(&models.ApplicationState{}, "Agent")
//  }
