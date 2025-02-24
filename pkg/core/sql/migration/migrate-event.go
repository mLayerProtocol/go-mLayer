package migration

import (
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)


func RenameEventAndAuthHashColumns(db *gorm.DB) (err error) {
	models := []interface{}{
		&models.SubnetEvent{},
		&models.AuthorizationEvent{},
		&models.TopicEvent{},
		&models.MessageEvent{},
		&models.SubscriptionEvent{},
		&models.WalletEvent{},
	}
		
	for _, model := range models {
		if db.Migrator().HasTable(model) && db.Migrator().HasColumn(model, "previous_event_hash") {
			err = db.Migrator().RenameColumn(model, "previous_event_hash", "previous_event")
			if err != nil {
				return err
			}
		}
		if db.Migrator().HasTable(model) && db.Migrator().HasColumn(model, "auth_event_hash") {
			err = db.Migrator().RenameColumn(model, "auth_event_hash","auth_event")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
//  func RenameSubnetEventAndAuthHashColumns(db *gorm.DB) (err error) {
// 	model := &models.SubnetEvent{}
// 	if db.Migrator().HasTable(&models.SubnetEvent{}) && db.Migrator().HasColumn(&models.SubnetEvent{}, "previous_event_hash") {
// 		err = db.Migrator().RenameColumn(model, "previous_event_hash", "previous_event")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if db.Migrator().HasTable(&models.SubnetEvent{}) && db.Migrator().HasColumn(&models.SubnetEvent{}, "auth_event_hash") {
// 		err = db.Migrator().RenameColumn(model, "auth_event_hash","auth_event")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return err
//  }

//  func RenameAuthEventAndAuthHashColumns(db *gorm.DB) (err error) {
// 	model := &models.AuthorizationEvent{}
// 	if db.Migrator().HasTable(model) && db.Migrator().HasColumn(model, "previous_event_hash") {
// 		err = db.Migrator().RenameColumn(model, "previous_event_hash", "previous_event")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if db.Migrator().HasTable(&models.SubnetEvent{}) && db.Migrator().HasColumn(&models.SubnetEvent{}, "auth_event_hash") {
// 		err = db.Migrator().RenameColumn(model, "auth_event_hash","auth_event")
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return err
//  }

//  func DropAgentColumnFromSubnetState(db *gorm.DB) (err error) {
// 	// if db.Migrator().HasColumn(model, "Agent") {
// 	// 	err = db.Migrator().DropColumn(model, "Agent")
// 	// }
// 	return  db.Migrator().DropColumn(model, "Agent")
//  }
