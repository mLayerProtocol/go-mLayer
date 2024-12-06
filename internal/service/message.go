package service

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"gorm.io/gorm"
)

/*
Validate an agent authorization
*/
func ValidateMessageData(payload *entities.ClientPayload, topic *entities.Topic) (currentSubscription *models.SubscriptionState, err error) {
	// check fields of message
	// var currentState *models.MessageState

	// err = query.GetOne(models.MessageState{
	// 	Message: entities.Message{Subscriber: message.Subscriber, Topic: message.Topic},
	// }, &currentState)
	// if err != nil {
	// 	if err != gorm.ErrRecordNotFound {
	// 		//return nil, nil, apperror.Unauthorized("Not a subscriber")
	// 		// } else {
	// 		return nil, err
	// 	}
	// }
	message := payload.Data.(entities.Message)
	var subscription *entities.Subscription
	// err = query.GetOne(models.SubscriptionState{
	// 	Subscription: entities.Subscription{Subscriber: payload.Account, Topic: topicData.ID},
	// }, &subscription)
	if payload.Account != message.Sender {
		return nil,  apperror.BadRequest("Invalid message signer")
	}

	
	subsribers := []entities.DIDString{entities.DIDString(payload.Account.ToString()), entities.DIDString(payload.Agent)}
	// subscriptions, err := query.GetSubscriptionStateBySubscriber(payload.Subnet, message.Topic, subsribers, sql.SqlDb)
	subscriptions := []*entities.Subscription{}
	accountSubcribed, err := dsquery.GetSubscriptions(entities.Subscription{
		Subnet: payload.Subnet,
		Topic: message.Topic,
		Subscriber: subsribers[0],
	}, dsquery.DefaultQueryLimit, nil)

	if err != nil && !dsquery.IsErrorNotFound(err) {
		return nil, err
	}
	
		agentSubcribed, _err := dsquery.GetSubscriptions(entities.Subscription{
			Subnet: payload.Subnet,
			Topic: message.Topic,
			Subscriber: subsribers[1],
		}, dsquery.DefaultQueryLimit, nil)
		
		if _err != nil {
			return nil, _err
		}
		
		subscriptions=append(subscriptions, accountSubcribed...)
		subscriptions=append(subscriptions, agentSubcribed...)
	
	if len(subscriptions) > 0 {
		if  len(subscriptions) > 1 {
			// if string(payload.Account)  != "" && (*subscriptions)[0].Subscription.Subscriber.ToString() == string(payload.Account) {
			// 	subscription = (*subscriptions)[0]
			// } else {
			// 	subscription = (*subscriptions)[1]
			// }
			if  *((subscriptions)[0].Role) > *((subscriptions)[1].Role) {
				subscription = subscriptions[0]
			} else {
				subscription = (subscriptions)[1]
			}
		} else {
			subscription = (subscriptions)[0]
		}
		
		if *topic.ReadOnly && payload.Account != topic.Account && *subscription.Role < constants.TopicManagerRole {
			return nil, apperror.Unauthorized("Not allowed to post to this topic")
		}
		if payload.Account != topic.Account && *subscription.Role < constants.TopicWriterRole {
			return  nil, apperror.Unauthorized("Not allowed to post to this topic")
		}
		logger.Debugf("Found Subscribers: %v", subscription)
		return &models.SubscriptionState{Subscription: *subscription}, nil
	} else {
		// check if the sender is a subnet admin
		// subnet := models.SubnetState{}
		subnet, err := dsquery.GetSubnetStateById(payload.Subnet)
		
		if err != nil {
			return nil, apperror.BadRequest("Invalid subnet")
		}
		if payload.Account != subnet.Account {
			// check if its an admin
			// auth := models.AuthorizationState{}
			// err = query.GetOneState(entities.Authorization{Agent: payload.Agent, Account: payload.Account}, &auth)
			auth, err := dsquery.GetAccountAuthorizations(entities.Authorization{
				Agent: payload.Agent, Account: payload.Account, Subnet: payload.Subnet,
			}, nil, nil)
			if err != nil {
				return nil, apperror.Unauthorized("Invalid subnet")
			}
			
			if len(auth) == 0 || *auth[0].Priviledge  < constants.MemberPriviledge {
				return nil, apperror.Unauthorized("agent not authorized")
			}
			
		}
	}
	
	if subscription == nil {
		return nil, nil
	}
	return &models.SubscriptionState{Subscription: *subscription}, nil
}
func saveMessageEvent(where entities.Event, createData *entities.Event, updateData *entities.Event, txn *datastore.Txn, tx *gorm.DB ) (*entities.Event, error) {
	
	return SaveEvent(entities.MessageModel, where, createData, updateData, txn)

}
func HandleNewPubSubMessageEvent(event *entities.Event, ctx *context.Context) error {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to load config from context")
	}
	data := event.Payload.Data.(entities.Message)
	// var id = data.ID
	// if len(data.ID) == 0 {
	// 	id, _ = entities.GetId(data)
	// } else {
	// 	id = data.ID
	// }
	var topic =  models.TopicState{}
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.Event = *event.GetPath()
	data.EventSignature = event.Signature
	data.EventTimestamp = event.Timestamp
	hash, err := data.GetHash()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	data.Agent = event.Payload.Agent
	data.Sender = event.Payload.Account
	var subnet = event.Payload.Subnet
	

	eventData := PayloadData{Subnet: subnet, localDataState: nil, localDataStateEvent:  nil}
	tx := sql.SqlDb
	// defer func () {
	// 	if tx.Error != nil {
	// 		tx.Rollback()
	// 	} else {
	// 		tx.Commit()
	// 	}
	// }()
	stateTxn, err := stores.MessageStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	txn, err := stores.EventStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	defer stateTxn.Discard(context.Background())
	defer txn.Discard(context.Background())

	
	
	previousEventUptoDate,  authEventUpToDate, _, eventIsMoreRecent, err := ProcessEvent(event,  eventData, true, saveMessageEvent, &txn, nil, ctx)
	if err != nil {
		logger.Errorf("Processing Error...: %v", err)
		return err
	}
	logger.Debugf("Processing 2...: %v,  %v", previousEventUptoDate, authEventUpToDate)
	// get the topic, if not found retrieve it


	
	if previousEventUptoDate  && authEventUpToDate {
		_topic, err := dsquery.GetTopicById( data.Topic)

		if (err != nil && dsquery.IsErrorNotFound(err) || _topic == nil  ) {
			// get topic like we got subnet
			topicPath := entities.NewEntityPath(event.Validator, entities.TopicModel, data.Topic)
				pp, err := p2p.GetState(cfg, *topicPath, &event.Validator, &topic)
				if err != nil {
					logger.Error(err)
					return err
				}
				if len(pp.Event) < 2 {
					return  fmt.Errorf("invalid event data")
				}
				topicEvent, err := entities.UnpackEvent(pp.Event, entities.TopicModel)
				if err != nil {
					logger.Error(err)
					return  err
				}
				if topicEvent != nil  && *topicEvent.Synced && len(pp.States) > 0 {
					// return HandleNewPubSubTopicEvent(topicEvent, ctx)
					
					topic, err := entities.UnpackTopic(pp.States[0])
					if err != nil {
						return err
					}
					err = dsquery.CreateEvent(topicEvent, &txn)
					if err == nil {
						_, err = dsquery.CreateTopicState(&topic, nil)
					}
					if err != nil {
						return err
					}
					
				}
		}
		if _topic != nil {
			topic = models.TopicState{Topic: *_topic}
		}
		
		
		err = dsquery.IncrementCounters(event.Cycle, event.Validator, event.Subnet, &txn)
		if err != nil { 
			logger.Errorf("CounterError %v", err)
			return err
		}
		 _, err = ValidateMessageData(&event.Payload, _topic)
		if err != nil {
			// update error and mark as synced
			// notify validator of error
			logger.Errorf("MessageDataError: %v", err)
			saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, &txn, nil )
			
		} else {
			// TODO if event is older than our state, just save it and mark it as synced
			savedEvent, err := saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, tx );
			if eventIsMoreRecent && err == nil {
				// update state
				logger.Debugf("CreateMessageData: %+v", data)
				_, err = dsquery.CreateMessageState(&data, &stateTxn)
				if err != nil {
					stateTxn.Discard(context.Background())
					logger.Debugf("CreateMessageErrror: %+v", err)
					_, err = saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil)
	
					if err != nil {
						// tx.Rollback()
						stateTxn.Discard(context.Background())
						logger.Errorf("SaveStateError %v", err)
						return err
					}
				} else {
					err = stateTxn.Commit(context.Background())
				}
				
			}
			
			if err == nil {
				err =  txn.Commit(context.Background())
			}
			if err == nil {
				go func ()  {
					dsquery.IncrementStats(event, nil)
				 dsquery.UpdateAccountCounter(event.Payload.Account.ToString())
				OnFinishProcessingEvent(ctx, event,  &models.MessageState{
					Message: data,
				},  &savedEvent.Payload.Subnet)
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
