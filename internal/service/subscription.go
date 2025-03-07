package service

import (
	"context"
	"encoding/hex"
	"slices"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"

	// query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"

	"gorm.io/gorm"
)

/*
Validate an agent authorization
*/
func ValidateSubscriptionData(payload *entities.ClientPayload, topic *entities.Topic) (currentSubscriptionState *models.SubscriptionState, err error) {
	// check fields of subscription

	subscription := payload.Data.(entities.Subscription)
	var currentState *models.SubscriptionState

	logger.Info("ValidateSubscription...")
	// err = query.GetOne(models.SubscriptionState{
	// 	Subscription: entities.Subscription{Subscriber: subscription.Subscriber, Application: subscription.Application, Topic: subscription.Topic},
	// }, &currentState)
	_currentState, err := dsquery.GetSubscriptions(entities.Subscription{Subscriber: subscription.Subscriber, Application: subscription.Application, Topic: subscription.Topic}, dsquery.DefaultQueryLimit, nil)
	if err != nil {
		logger.Errorf("Invalid event payload %e ", err)

		if err != gorm.ErrRecordNotFound && !dsquery.IsErrorNotFound(err) {
			//return nil, nil, apperror.Unauthorized("Not a subscriber")
			// } else {
			return nil, err
		} else {
			// logger.Errorf("gorm.ErrRecordNotFound %e ", gorm.ErrRecordNotFound)
			// return nil, nil
		}
	}
	logger.Infof("ValidateSubscription...2: %v", (constants.SubscribeTopicEvent == payload.EventType))
	if len(_currentState) > 0 {
		currentState = &models.SubscriptionState{Subscription: *_currentState[0]}
	}
	
	// if payload.EventType == uint16(constants.SubscribeTopicEvent) {
	// 	// someone inviting someone else
	// 	logger.Infof("ValidateSubscription...3")
	// 	// subscribingSomeoneElse := subscription.Subscriber != payload.Account && subscription.AppKey != payload.AppKey 
	// 	subscribingSomeoneElse := subscription.Subscriber != payload.Account
	// 	if subscribingSomeoneElse {
	// 		if !slices.Contains([]constants.SubscriptionStatus{constants.InvitedSubscriptionStatus, constants.BannedSubscriptionStatus}, *subscription.Status) {
	// 			return nil, apperror.Forbidden("Subscription status must be Invited or Banned")
	// 		}
	// 	} else {
	// 		// subscribing oneself
	// 		// if the topic is not public, you have to have been invited
	// 		if !(*topic.Public) && currentState == nil {
	// 			return nil, apperror.Forbidden("Must be invited first")
	// 		}
	// 		if currentState != nil && *currentState.Status == constants.BannedSubscriptionStatus {
	// 			return nil, apperror.Forbidden("Banned subscriber")
	// 		}
	// 		if (currentState != nil && *currentState.Role != *subscription.Role) || (currentState == nil && *subscription.Role > *topic.DefaultSubscriberRole) {
	// 			return nil, apperror.Forbidden("Invalid role selected")
	// 		}
	// 		if !slices.Contains([]constants.SubscriptionStatus{constants.UnsubscribedSubscriptionStatus, constants.SubscribedSubscriptionStatus}, *subscription.Status) {
	// 			return nil, apperror.Forbidden("Subscription status must be Subscribed or Unsubscribed")
	// 		}
	// 	}
	// }
	if currentState == nil {
		err = ValidateSubscription(payload.Account, &subscription, topic, nil);
	} else {
		err = ValidateSubscription(payload.Account, &subscription, topic, &currentState.Subscription);
	}
	
	return currentState, err
}

func ValidateSubscription (account entities.AccountString, subscription *entities.Subscription,  topic *entities.Topic, currentState *entities.Subscription) error {
		// someone inviting someone else
		logger.Infof("ValidateSubscription...3")
		// subscribingSomeoneElse := subscription.Subscriber != payload.Account && subscription.AppKey != payload.AppKey 
		subscribingSomeoneElse := subscription.Subscriber != entities.AddressString(account)
		if len(subscription.Application) == 0 {
			return apperror.BadRequest("app must be specified")
		}
		if subscription.Application != topic.Application {
			return apperror.BadRequest("subscription and topic app does not match")
		}
		if subscribingSomeoneElse {
			if !slices.Contains([]constants.SubscriptionStatus{constants.InvitedSubscriptionStatus, constants.BannedSubscriptionStatus}, *subscription.Status) {
				return apperror.Forbidden("Subscription status must be Invited or Banned")
			}
		} else {
			// subscribing oneself
			// if the topic is not public, you have to have been invited
			if !(*topic.Public) && currentState == nil {
				return  apperror.Forbidden("Must be invited first")
			}
			if currentState != nil && *currentState.Status == constants.BannedSubscriptionStatus {
				return  apperror.Forbidden("Banned subscriber")
			}
			if (currentState != nil && *currentState.Role != *subscription.Role) || (currentState == nil && *subscription.Role > *topic.DefaultSubscriberRole) {
				return  apperror.Forbidden("Invalid role selected")
			}
			if !slices.Contains([]constants.SubscriptionStatus{constants.UnsubscribedSubscriptionStatus, constants.SubscribedSubscriptionStatus}, *subscription.Status) {
				return apperror.Forbidden("Subscription status must be Subscribed or Unsubscribed")
			}
		}
	return nil
}

func saveSubscriptionEvent(where entities.Event, createData *entities.Event, updateData *entities.Event, txn *datastore.Txn, tx *gorm.DB) (*entities.Event, error) {
	return SaveEvent(entities.SubscriptionModel, where, createData, updateData, txn)
}
func HandleNewPubSubSubscriptionEvent(event *entities.Event, ctx *context.Context) error {

	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to load config from context")
	}
	data := event.Payload.Data.(entities.Subscription)
	dataStates := dsquery.NewDataStates(event.ID, cfg)
	dataStates.AddEvent(*event)
	// var id = data.ID
	// if len(data.ID) == 0 {
	// 	id, _ = entities.GetId(data)
	// } else {
	// 	id = data.ID
	// }
	var topic = models.TopicState{}
	data.Event = *event.GetPath()
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.AppKey = event.Payload.AppKey
	data.EventSignature = event.Signature
	data.Timestamp = &event.Timestamp
	hash, err := data.GetHash()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	var app = data.Application
	var validator = ""
	if event.Validator != entities.PublicKeyString(cfg.PublicKeyEDDHex) {
		validator = string(event.Validator)
	}
	defer func() {
		// stateUpdateError := dataStates.Commit(nil, nil, nil, event.ID, err)
		stateUpdateError := dataStates.Save(event.ID)
		if stateUpdateError != nil {

			panic(stateUpdateError)
		} else {
			go OnFinishProcessingEvent(cfg, event, &data, nil)

			// go utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte("newMessage" + "\n"))
		}
	}()

	localState := models.SubscriptionState{}
	// err := query.GetOne(&models.TopicState{Topic: entities.Topic{ID: id}}, &localTopicState)
	// err = sql.SqlDb.Where(&models.SubscriptionState{Subscription: entities.Subscription{Application: app, Topic: data.Topic, Subscriber: entities.AddressFromString(string(data.Subscriber)).ToDIDString()}}).Take(&localState).Error
	// if err != nil {
	// 	logger.Error(err)
	// }

	subs, err := dsquery.GetSubscriptions(entities.Subscription{Application: app, Topic: data.Topic, Subscriber: data.Subscriber}, nil, nil)
	if err == nil && subs != nil && len(subs) > 0 {
		localState = models.SubscriptionState{
			Subscription: *subs[0],
		}
	}

	var localDataState *LocalDataState
	if localState.ID != "" {
		localDataState = &LocalDataState{
			ID:        localState.ID,
			Hash:      localState.ID,
			Event:     &localState.Event,
			Timestamp: *localState.Timestamp,
		}
	}

	// localDataState := utils.IfThenElse(localTopicState != nil, &LocalDataState{
	// 	ID: localTopicState.ID,
	// 	Hash: localTopicState.ID,
	// 	Event: &localTopicState.Event,
	// 	Timestamp: localTopicState.Timestamp,
	// }, nil)
	var stateEvent *entities.Event
	if localState.ID != "" {
		stateEvent, err = dsquery.GetEventFromPath(&localState.Event)
		if err != nil && !dsquery.IsErrorNotFound(err) {
			logger.Debug(err)
		}
	}
	var localDataStateEvent *LocalDataStateEvent
	if stateEvent != nil {
		localDataStateEvent = &LocalDataStateEvent{
			ID:        stateEvent.ID,
			Hash:      stateEvent.Hash,
			Timestamp: stateEvent.Timestamp,
		}
	}

	eventData := PayloadData{Application: app, localDataState: localDataState, localDataStateEvent: localDataStateEvent}
	// tx := sql.SqlDb
	// defer func () {
	// 	if tx.Error != nil {
	// 		tx.Rollback()
	// 	} else {
	// 		tx.Commit()
	// 	}
	// }()

	previousEventUptoDate, authEventUpToDate, _, eventIsMoreRecent, err := ProcessEvent(event, eventData, true, saveSubscriptionEvent, nil, nil, ctx, dataStates)
	if err != nil {
		logger.Debugf("Processing Error...: %v", err)
		return err
	}
	// logger.Debugf("Processing 2...: %v,  %v", previousEventUptoDate, authEventUpToDate)
	// get the topic, if not found retrieve it

	if previousEventUptoDate && authEventUpToDate {
		_, err := SyncTypedStateById(data.Topic, &entities.Topic{}, cfg, validator)
		if err != nil {
			logger.Error("HandleNewPubSubSubscriptionEvent/SyncTypedStateById", err)
			return err
		}

		if !event.IsLocal(cfg) {
			_, err = ValidateSubscriptionData(&event.Payload, &topic.Topic)
		}
		if err != nil {
			// update error and mark as synced
			// notify validator of error
			dataStates.AddEvent(entities.Event{ID: event.ID, Error: err.Error(), IsValid: utils.FalsePtr(), Synced: utils.TruePtr()})

		} else {
			// TODO if event is older than our state, just save it and mark it as synced
			logger.Infof("SAVINGSUBSCRIPITONS: %v, %v", event.ID, eventIsMoreRecent)
			dataStates.AddEvent(entities.Event{ID: event.ID, IsValid: utils.TruePtr(), Synced: utils.TruePtr()})
			data.ID, _ = entities.GetId(data, data.ID)
			if eventIsMoreRecent {
				// update state
				dataStates.AddCurrentState(entities.SubscriptionModel, data.DataKey(), data)
			} else {
				dataStates.AddHistoricState(entities.SubscriptionModel, data.DataKey(), data.MsgPack())
			}

		}
		go dsquery.UpdateAccountCounter(string(event.Payload.Account))
	}
	return nil

}
