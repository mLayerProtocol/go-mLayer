package query

import (
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)

// func GetManyWithEvent[T any, U any](filter T, event entities.Event, data *U) error {
// 	err := db.SqlDb.Where(&filter).Joins().Find(data).Error
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func GetEvent(grantor string, agent string) (*models.Config, error) {

// 	data := models.Config{}
// 	err := db.SqlDb.Where(&models.AuthorizationState{Grantor: grantor, Agent: agent}).First(&data).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &data, nil
// }

// func SaveEvent(event entities.Event, data models.Event) (*models.Event, error) {

// 	tx := db.SqlDb.Begin()
// 	err := tx.Where(models.Event{
// 			Hash: event.Hash,
// 			}).Assign(models.Event{
// 				Parents   : event.Parents,
// 				Synced : sql.NullBool{Valid: true, Bool: event.Synced},
// 				Hash: event.Hash,
// 				StateHash:  event.StateHash,
// 				EventType: int16(event.EventType),
// 				Payload: event.MsgPack(),
// 				Validator: event.Validator,
// 				Timestamp : event.Timestamp,
// 				Signature : event.Signature,
// 				// IsValid :  event.IsValid
// 				 }).FirstOrCreate(&data).Error
// 	if err != nil {
// 		tx.Rollback()
// 		return nil, err
// 	}
// 	tx.Commit()
// 	return &data, nil
// }

var ErrorNotFound = gorm.ErrRecordNotFound

func EventExist(ePath *entities.EventPath) bool {
	if ePath.ID == "" {
		return true
	}
	event, _ := GetEventFromPath(ePath)
	return event != nil
}

func GetEventFromPath(ePath *entities.EventPath) (*entities.Event, error) {
	if  ePath == nil || len(ePath.ID) == 0  {
		return nil, nil
	}
	if ePath.Model == entities.SubscriptionModel {
		var data *models.SubscriptionEvent
		err := GetOneWithOr(models.SubscriptionEvent{
			Event: entities.Event{Hash: ePath.ID},
		}, models.SubscriptionEvent{
			Event: entities.Event{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Event, nil
	}

	if ePath.Model == entities.TopicModel {
		var data *models.TopicEvent
		err := GetOneWithOr(models.TopicEvent{
			Event: entities.Event{Hash: ePath.ID},
		}, models.TopicEvent{
			Event: entities.Event{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Event, nil
	}

	if ePath.Model == entities.SubnetModel {
		var data *models.SubnetEvent
		err := GetOneWithOr(models.SubnetEvent{
			Event: entities.Event{Hash: ePath.ID},
		}, models.SubnetEvent{
			Event: entities.Event{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Event, nil
	}

	if ePath.Model == entities.AuthModel {
		var data *models.AuthorizationEvent
		err := GetOneWithOr(models.AuthorizationEvent{
			Event: entities.Event{Hash: ePath.ID},
		}, models.AuthorizationEvent{
			Event: entities.Event{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Event, nil
	}

	if ePath.Model == entities.MessageModel {
		var data *models.MessageEvent
		err := GetOneWithOr(models.MessageEvent{
			Event: entities.Event{Hash: ePath.ID},
		}, models.MessageEvent{
			Event: entities.Event{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Event, nil
	}

	return nil, nil
}

func GetStateFromPath(ePath *entities.EntityPath) (any, error) {
	if  ePath == nil || len(ePath.ID) == 0  {
		return nil, nil
	}
	if ePath.Model == entities.SubscriptionModel {
		var data *models.SubscriptionState
		err := GetOne(models.SubscriptionState{
			Subscription: entities.Subscription{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Subscription, nil
	}

	if ePath.Model == entities.TopicModel {
		var data *models.TopicState
		err := GetOne(models.TopicState{
			Topic: entities.Topic{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Topic, nil
	}

	if ePath.Model == entities.SubnetModel {
		var data *models.SubnetState
		err := GetOne(models.SubnetState{
			Subnet: entities.Subnet{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Subnet, nil
	}

	if ePath.Model == entities.AuthModel {
		var data *models.AuthorizationState
		err := GetOne(models.AuthorizationState{
			Authorization: entities.Authorization{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Authorization, nil
	}

	if ePath.Model == entities.MessageModel {
		var data *models.MessageState
		err := GetOne(models.MessageState{
			Message: entities.Message{ID: ePath.ID},
		}, &data)
		if err != nil {
			return nil, err
		}
		return &data.Message, nil
	}

	return nil, nil
}
