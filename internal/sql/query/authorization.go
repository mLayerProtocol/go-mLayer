package query

import (
	"encoding/hex"

	"github.com/jinzhu/copier"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	db "github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"gorm.io/gorm"
)

func GetOneAuthorizationState(auth entities.Authorization) (*models.AuthorizationState, error) {

	data := models.AuthorizationState{}
	where := models.AuthorizationState{
		Authorization: auth,
	}
	if err := utils.CheckEmpty(where); err != nil {
		return nil, nil
	}
	err := db.SqlDb.Where(where).First(&data).Error
	if err != nil {

		return nil, err
	}
	return &data, nil
}

func GetOneAuthorizationEvent(event entities.Event) (*models.AuthorizationEvent, error) {
	where := models.AuthorizationEvent{Event: event}
	if err := utils.CheckEmpty(where); err != nil {
		return nil, nil
	}
	data := models.AuthorizationEvent{}
	err := db.SqlDb.Where(&where).First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func GetManyAuthorizationEvents(event entities.Event) (*models.AuthorizationEvent, error) {

	data := models.AuthorizationEvent{}
	err := db.SqlDb.Where(&models.AuthorizationEvent{Event: event}).Find(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}


// Save authorization only when it doesnt exist
func SaveAuthorizationState(auth *entities.Authorization, DB *gorm.DB) (*models.AuthorizationState, *gorm.DB) {

	data := models.AuthorizationState{
		// Privilege 	: auth.Priviledge,
		Authorization: *auth,
	}
	tx := DB
	if DB == nil {
		tx = db.SqlDb
	}
	result := tx.Where(utils.EnsureNotEmpty(models.AuthorizationState{
		Authorization: entities.Authorization{
			Authorized:  auth.Authorized,
			Application: auth.Application,
		},
	})).Assign(data).FirstOrCreate(&data)
	
	// if DB == nil {
	// 	tx.Commit()
	// }
	return &data, result
}

func SaveAuthorizationEvent(event *entities.Event, update bool, DB *gorm.DB) (model *models.AuthorizationEvent, created bool, err error) {
	tx := DB
	if DB == nil {
		tx = db.SqlDb
	}
	// dataByte, err := encoder.MsgPackStruct(event.Payload)
	if err != nil {
		return nil, false, err
	}
	authPayload := entities.ClientPayload{
		Data: entities.Authorization{},
	}
	copier.Copy(&authPayload, &event.Payload)
	data := models.AuthorizationEvent{
		Event: *event,
	}
	var result *gorm.DB
	if update {
		result = tx.Where(models.AuthorizationEvent{
			Event: entities.Event{Hash: event.Hash},
		}).Assign(data).FirstOrCreate(&data)
	} else {
		result = tx.Where(models.AuthorizationEvent{
			Event: entities.Event{PayloadHash: event.PayloadHash},
		}).FirstOrCreate(&data)
	}
	if result.Error != nil {
		// tx.Rollback()
		logger.Errorf("SQL: %v", result.Error)
		return nil, false, result.Error
	}
	// if DB == nil {
	// 	tx.Commit()
	// }
	return &data, result.RowsAffected > 0, nil
}

func UpdateAuthorizationEvent(where entities.Event, updateFields entities.Event, DB *gorm.DB) (model *models.AuthorizationEvent, err error) {
	tx := DB
	if DB == nil {
		tx = db.SqlDb
	}
	// dataByte, err := encoder.MsgPackStruct(event.Payload)
	if err != nil {
		return nil, err
	}

	result := tx.Where(models.AuthorizationEvent{
		Event: where,
	}).Updates(models.AuthorizationEvent{
		Event: updateFields,
	}).First(&model)

	if result.Error != nil {
		// tx.Rollback()
		logger.Errorf("SQL: %v", result.Error)
		return nil, result.Error
	}
	// if DB == nil {
	// 	tx.Commit()
	// }
	return model, nil
}

func SaveAuthorizationStateAndEvent(authEvent *entities.Event, tx *gorm.DB) (*models.AuthorizationState, *models.AuthorizationEvent, error) {
	if tx == nil {
		tx = db.SqlDb
	}

	auth := (*authEvent).Payload.Data.(entities.Authorization)

	hash, _ := auth.GetHash()
	auth.Hash = hex.EncodeToString(hash)
	id, _ := authEvent.GetId()
	auth.Event =  *entities.NewEventPath(authEvent.Validator, entities.AuthModel, id)
	event, created, err := SaveAuthorizationEvent(authEvent, false, tx)
	if err != nil {
		logger.Errorf("SQL: %v", err)
		return nil, nil, err
	}
	// auth.AuthorizationEventID = event.Event.ID
	if created {
		state, txrsl := SaveAuthorizationState(&auth, tx)
		if txrsl.Error != nil {
			logger.Errorf("SQL: %v", err)
			// tx.Rollback()
			return nil, nil, err
		}
		//tx.Commit()
		return state, event, nil
	} else {
		state, err := GetOneAuthorizationState(entities.Authorization{Account: auth.Account, Authorized: auth.Authorized})
		if err != nil {
			logger.Errorf("SQL: %v", err)
			// tx.Rollback()
			return nil, nil, err
		}
		// tx.Commit()
		return state, event, nil
	}

}
