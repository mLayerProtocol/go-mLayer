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
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"

	// query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
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
	// 	Subscription: entities.Subscription{Subscriber: subscription.Subscriber, Subnet: subscription.Subnet, Topic: subscription.Topic},
	// }, &currentState)
	_currentState, err := dsquery.GetSubscriptions(entities.Subscription{Subscriber: subscription.Subscriber, Subnet: subscription.Subnet, Topic: subscription.Topic}, dsquery.DefaultQueryLimit, nil)
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
	logger.Infof("ValidateSubscription...2: %v", (uint16(constants.SubscribeTopicEvent) == payload.EventType))
	if len(_currentState)  > 0 {
		currentState = &models.SubscriptionState{Subscription: *_currentState[0]}
	}
	if payload.EventType == uint16(constants.SubscribeTopicEvent) { 
		// someone inviting someone else
		logger.Infof("ValidateSubscription...3")
		if subscription.Subscriber != payload.Account && subscription.Agent != payload.Agent {
			if !slices.Contains([]constants.SubscriptionStatus{constants.InvitedSubscriptionStatus, constants.BannedSubscriptionStatus}, *subscription.Status) {
				return nil, apperror.Forbidden("Subscription status must be Invited or Banned")
			}
		}
	} else {
		// subscribing oneself
		// if the topic is not public, you have to have been invited
		if !(*topic.Public) && currentState == nil {
			return nil, apperror.Forbidden("Must be invited first")
		}
		if  currentState != nil && *currentState.Status == constants.BannedSubscriptionStatus {
			return nil, apperror.Forbidden("Banned subscriber")
		}
		if (currentState != nil && *currentState.Role != *subscription.Role) || (currentState == nil && *subscription.Role > *topic.DefaultSubscriberRole) {
			return nil, apperror.Forbidden("Invalid role selected")
		}
		if !slices.Contains([]constants.SubscriptionStatus{constants.UnsubscribedSubscriptionStatus, constants.SubscribedSubscriptionStatus}, *subscription.Status) {
			return nil, apperror.Forbidden("Subscription status must be Subscribed or Unsubscribed")
		}
	}

	return currentState, err
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
	// var id = data.ID
	// if len(data.ID) == 0 {
	// 	id, _ = entities.GetId(data)
	// } else {
	// 	id = data.ID
	// }
	var topic =  models.TopicState{}
	data.Event = *event.GetPath()
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.Agent = event.Payload.Agent
	data.EventSignature = event.Signature
	data.Timestamp = &event.Timestamp
	hash, err := data.GetHash()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	var subnet = data.Subnet

	localState := models.SubscriptionState{}
	// err := query.GetOne(&models.TopicState{Topic: entities.Topic{ID: id}}, &localTopicState)
	// err = sql.SqlDb.Where(&models.SubscriptionState{Subscription: entities.Subscription{Subnet: subnet, Topic: data.Topic, Subscriber: entities.AddressFromString(string(data.Subscriber)).ToDIDString()}}).Take(&localState).Error
	// if err != nil {
	// 	logger.Error(err)
	// }
	
	subs, err := dsquery.GetSubscriptions(entities.Subscription{Subnet: subnet, Topic: data.Topic, Subscriber: entities.AddressFromString(string(data.Subscriber)).ToDIDString()}, nil, nil)
	if err == nil && subs != nil && len(subs) > 0 {
		localState = models.SubscriptionState{
			Subscription: *subs[0],
		}
	}

	var localDataState *LocalDataState
	if localState.ID != "" {
		localDataState = &LocalDataState{
			ID: localState.ID,
			Hash: localState.Hash,
			Event: &localState.Event,
			Timestamp: *localState.Timestamp,
		}
	}

	stateTxn, err := stores.StateStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	txn, err := stores.EventStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	defer stateTxn.Discard(context.Background())
	defer txn.Discard(context.Background())
	// localDataState := utils.IfThenElse(localTopicState != nil, &LocalDataState{
	// 	ID: localTopicState.ID,
	// 	Hash: localTopicState.Hash,
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
			ID: stateEvent.ID,
			Hash: stateEvent.Hash,
			Timestamp: stateEvent.Timestamp,
		}
	}

	eventData := PayloadData{Subnet: subnet, localDataState: localDataState, localDataStateEvent:  localDataStateEvent}
	// tx := sql.SqlDb
	// defer func () {
	// 	if tx.Error != nil {
	// 		tx.Rollback()
	// 	} else {
	// 		tx.Commit()
	// 	}
	// }()
	
	
	previousEventUptoDate,  authEventUpToDate, _, eventIsMoreRecent, err := ProcessEvent(event,  eventData, true, saveSubscriptionEvent, &txn, nil, ctx)
	if err != nil {
		logger.Debugf("Processing Error...: %v", err)
		return err
	}
	// logger.Debugf("Processing 2...: %v,  %v", previousEventUptoDate, authEventUpToDate)
	// get the topic, if not found retrieve it
	
	if previousEventUptoDate  && authEventUpToDate {

		_topic, err := dsquery.GetTopicById(data.Topic)
		if _topic != nil {
			topic.Topic = *_topic
		}
		if topic.ID == "" ||  dsquery.IsErrorNotFound(err) {
			// get topic like we got subnet
			topicPath := entities.NewEntityPath(event.Validator, entities.TopicModel, data.Topic)
				pp, err := p2p.GetState(cfg, *topicPath, &event.Validator, &topic)
				if err != nil {
					logger.Error(err)
					
				}
				
				topicEvent, err := entities.UnpackEvent(pp.Event, entities.TopicModel)
				if err != nil {
					logger.Error(err)
					
				}
				if topicEvent != nil {
					err = HandleNewPubSubTopicEvent(topicEvent, ctx)
					if err != nil {
						return err
					}
				}

				
		}

		err = dsquery.IncrementCounters(event.Cycle, event.Validator, event.Subnet, &txn)
		if err != nil { 
			return err
		}
		_, err = ValidateSubscriptionData(&event.Payload, &topic.Topic)
		if err != nil {
			// update error and mark as synced
			// notify validator of error
			saveSubscriptionEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, &txn, nil )
		} else {
			// TODO if event is older than our state, just save it and mark it as synced
			savedEvent, err := saveSubscriptionEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil );
			if eventIsMoreRecent && err == nil {
				// update state
				
				_, err = dsquery.CreateSubscriptionState(&data, &stateTxn)
					if err != nil {
						stateTxn.Discard(context.Background())
						// TODO worker that will retry processing unSynced valid events with error
						_, err = saveSubscriptionEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil)
					} else {
						err = stateTxn.Commit(context.Background())
					}
				if err != nil {
					// tx.Rollback()
					logger.Errorf("SaveStateError %v", err)
					return err
				}
				// err = stateTxn.Commit(context.Background())
				if err == nil {
					err = txn.Commit(context.Background())
				}
				
			}
			if err == nil {
				go func ()  {
					dsquery.IncrementStats(event, nil)
					dsquery.UpdateAccountCounter(event.Payload.Account.ToString())
					OnFinishProcessingEvent(ctx, event, &models.SubscriptionState{
						Subscription: data,
					}, &savedEvent.Payload.Subnet)
				}()
			}
			if string(event.Validator) != cfg.PublicKeyEDDHex {
				go func () {
				dependent, err := dsquery.GetDependentEvents(event)
				if err != nil {
					logger.Debug("Unable to get dependent events", err)
				}
				for _, dep := range *dependent {
					HandleNewPubSubEvent(dep, ctx)
				}
				}()
			}
			
		}
	} 
		return nil
	
}
