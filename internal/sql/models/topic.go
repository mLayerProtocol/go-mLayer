package models

import (
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"gorm.io/gorm"
)

type TopicState struct {
	entities.Topic
	BaseModel
}

func (d *TopicState) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		hash, err := entities.GetId(*d, d.ID)
		if err != nil {
			panic(err)
		}
		d.ID = hash
	}
	return nil
}

func (d TopicState) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(&d.Topic)
	return b
}

type TopicEvent struct {
	entities.Event
	BaseModel
	// TopicID     uint64
	// Topic		TopicState
}

// func (d TopicEvent) ValidateData( cfg *configs.MainConfiguration) (currentAuthState any, err error) {
// 	authState := &models.AuthorizationState{}
// 	topic := d.Payload.Data.(*entities.Topic)
// 	err = db.SqlDb.Where(&AuthorizationState{
// 		Authorization: entities.Authorization{Event: d.Event.AuthEvent},
// 	}).Take(authState).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	app := ApplicationState{}
// 	// authState, authError := db.SqlDb.Where($AuthorizationState{Authorization: entities.Authorization{Event: authEventHash}}).Take(&app).Error
// 	// TODO state might have changed befor receiving event, so we need to find state that is relevant to this event.
// 	err = db.SqlDb.Where(&ApplicationState{Application: entities.Application{ID: d.Payload.Application}}).Take(&app).Error
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return authState,  apperror.Forbidden("Invalid app id")
// 		}
// 		return authState, apperror.Internal(err.Error())
// 	}
// 	if  authState.Priviledge < constants.MemberPriviledge {
// 		return authState, apperror.Forbidden("Agent does not have enough permission to create topics")
// 	}
	
// 	if len(topic.Ref) > 40 {
// 		return authState, apperror.BadRequest("Topic handle can not be more than 40 characters")
// 	}
// 	if !utils.IsAlphaNumericDot(topic.Ref) {
// 		return authState,  apperror.BadRequest("Handle must be alphanumeric, _ and . but cannot start with a number")
// 	}
// 	return authState, nil
// }

