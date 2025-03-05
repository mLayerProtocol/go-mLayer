package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/global"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/smartlet"
	"gorm.io/gorm"
)

/*
Validate an agent authorization
*/
func ValidateMessageData(payload *entities.ClientPayload, topic *entities.Topic) (currentSubscription *entities.Subscription, err error) {
	defer utils.TrackExecutionTime(time.Now(), "ValidateMessageData::")

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
		return nil, apperror.BadRequest("Invalid message signer")
	}

	subsribers := []entities.AddressString{entities.AddressString(payload.Account), entities.AddressString(payload.Agent)}
	// subscriptions, err := query.GetSubscriptionStateBySubscriber(payload.Subnet, message.Topic, subsribers, sql.SqlDb)
	subscriptions := []*entities.Subscription{}
	accountSubcribed, err := dsquery.GetSubscriptions(entities.Subscription{
		Subnet:     payload.Subnet,
		Topic:      message.Topic,
		Subscriber: subsribers[0],
	}, dsquery.DefaultQueryLimit, nil)

	if err != nil {
		if !dsquery.IsErrorNotFound(err) {
			return nil, err
		}
	}

	agentSubcribed, _err := dsquery.GetSubscriptions(entities.Subscription{
		Subnet:     payload.Subnet,
		Topic:      message.Topic,
		Subscriber: subsribers[1],
	}, dsquery.DefaultQueryLimit, nil)

	if _err != nil {
		if dsquery.IsErrorNotFound(err) {

		}
		return nil, _err
	}

	subscriptions = append(subscriptions, accountSubcribed...)
	subscriptions = append(subscriptions, agentSubcribed...)

	if len(subscriptions) > 0 {
		if len(subscriptions) > 1 {
			// if string(payload.Account)  != "" && (*subscriptions)[0].Subscription.Subscriber.ToString() == string(payload.Account) {
			// 	subscription = (*subscriptions)[0]
			// } else {
			// 	subscription = (*subscriptions)[1]
			// }
			if *((subscriptions)[0].Role) > *((subscriptions)[1].Role) {
				subscription = subscriptions[0]
			} else {
				subscription = (subscriptions)[1]
			}
		} else {
			subscription = (subscriptions)[0]
		}



		if topic.ReadOnly != nil && *topic.ReadOnly &&
			payload.Account != topic.Account && (subscription.Role == nil ||
			*subscription.Role < constants.TopicManagerRole) {
			return nil, apperror.Unauthorized("Not allowed to post to this topic")
		}
		if payload.Account != topic.Account && *subscription.Role < constants.TopicWriterRole {
			return nil, apperror.Unauthorized("Not allowed to post to this topic")
		}

		return  subscription, nil
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
				Authorized: entities.AddressString(payload.Account), Account: subnet.Account, Subnet: payload.Subnet,
			}, nil, nil)
			if err != nil {
				// if !dsquery.IsErrorNotFound(err) {
				return nil, apperror.Unauthorized("Invalid authorization")
				// }
				// auth, err = dsquery.GetAccountAuthorizations(entities.Authorization{
				// 	Agent: entities.DeviceString(payload.Account), Account: subnet.Account, Subnet: payload.Subnet,
				// }, nil, nil)
			}

			if len(auth) == 0 || *auth[0].Priviledge < constants.ManagerPriviledge {
				return nil, apperror.Unauthorized("account not subscribed to topic")
			}

			// TODO only allow if authorized on all topics * or on specific topics

		}
	}

	if subscription == nil {
		return nil, nil
	}
	return subscription, nil
}
func saveMessageEvent(where entities.Event, createData *entities.Event, updateData *entities.Event, txn *datastore.Txn, tx *gorm.DB) (*entities.Event, error) {

	return SaveEvent(entities.MessageModel, where, createData, updateData, txn)

}
func HandleNewPubSubMessageEvent(event *entities.Event, ctx *context.Context) ( err error) {
	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to load config from context")
	}
	dataStates := dsquery.NewDataStates(event.ID, cfg)
	dataStates.AddEvent(*event)
	

	validator := utils.IfThenElse(event.IsLocal(cfg), "", string(event.Validator))

	data := event.Payload.Data.(entities.Message)
	// var topic =  models.TopicState{}
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.Event = *event.GetPath()
	data.EventSignature = event.Signature
	data.EventTimestamp = event.Timestamp
	var id = data.ID
	id, _ = entities.GetId(data, data.ID)
	hash, err := data.GetHash()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	data.Agent = event.Payload.Agent
	data.Sender = event.Payload.Account
	var subnet = event.Payload.Subnet
	var topic = &entities.Topic{}


	// if !event.IsLocal(cfg) {
	// 	syncDS := dsquery.NewDataStates(cfg)
	// 	for _, k := range event.StateEvents {
	// 		event, stateBytes , local, err := GetEventByPath(&k.Event, cfg, "")
	// 		if !local {
	// 			syncDS.AddEvent(*event)
	// 			item := entities.GetEventEntityFromModel(k.Event.Model)
	// 			err := encoder.MsgPackUnpackStruct(stateBytes, item)
	// 			st, err = dsquery.GetStateById(k.ID, k.Event.Model)
	// 			if err != nil {
	// 				syncDS.AddCurrentState(k.Event.Model, k.ID, item )
	// 			} else {
	// 				if st.Hash != item.Hash {

	// 				}
	// 			}

	// 			// save it 
	// 			// dsquery.UpdateEvent(event, nil, true)
	// 			// switch k.Event.Model {
	// 			// case entities.SubnetModel:
	// 			// 	state := entities.Subnet{}
	// 			// 	encoder.MsgPackUnpackStruct(stateBytes, state)
	// 			// 	dsqu

	// 			// }
	// 			// state := entities.GetStateModelFromEntityType(path.Model)
				
	// 			// id := state.GetId(state.ID)
	// 			// dsquery.SaveHistoricState(path.Model, )
	// 		}
	// 	}
	// }




	defer func() {
		if err != nil {
			logger.Errorf("NewMessageError: %v", err)
			dataStates.Commit(nil, nil, nil, event.ID, err)
			return
		}
		// for _, handler := range topic.Handlers{
			logger.Debugf("TOPICHANLDEREXISTS: %v", global.GlobalTopicHandlers)
			var app smartlet.App
		if f, ok := global.GlobalTopicHandlers[string(topic.Handler)]; ok {
			// The key exists, so call the function.
			
			keyStore := newKeyStore(&app, stores.GlobalHandlerStore.DB)
			stats := utils.RunWithConstraints(context.Background(), utils.DefaultResourceLimit, func(ctx context.Context) error {
				app = *smartlet.NewApplication(event, topic, logger, keyStore, &ctx)
				//ctx = context.WithValue(ctx, smartlet.Application, app)
				logger.Infof("NEWMESSAGEEEEE")
				app, err = f.PreCommit(app)
				
				// if stateUpdateError != nil {
				// 	logger.Error("HandleNewPubSubMessageEvent: ", stateUpdateError)
				// 	panic(stateUpdateError)
				// }
				// err = f.PostCommit(app)
				// logger.Error("PostCommitError: ", err)
				// if err != nil {
				// 	return err
				// }
				// keyStore.Commit()
				return err

			})
			if stats.Error != nil {
				logger.Error("PreCommitContraintError: ", stats.Error)
				err = stats.Error
				return
			}
			
			stateUpdateError := dataStates.Commit(nil, nil, nil, event.ID, stats.Error)
				if stateUpdateError != nil {
					logger.Error("HandleNewPubSubMessageEvent: ", stateUpdateError)
					panic(stateUpdateError)
				}

			state := utils.RunWithConstraints(context.Background(), utils.DefaultResourceLimit, func(ctx context.Context) error {
				//ctx = context.WithValue(ctx, smartlet.Application, app)
				app, err = f.PostCommit(app)
				if err != nil {
					logger.Error("PostCommitError: ", err)
					return err
				}
				app.KeyStore.(smartletKeyStore).Commit()
				
				return err

			})
			if state.Error != nil {
				logger.Errorf("PostCommitConstraintError: %v", state.Error)
				err = state.Error
			}
			
			// work with result
		}
		// }

		if err == nil {
			
			logger.Infof("POOOSIOIOSID")
			go OnFinishProcessingEvent(cfg, event, &data, app.GetResult())
			// go utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte("newMessage" + "\n"))
		}
	}()

	eventData := PayloadData{Subnet: subnet, localDataState: nil, localDataStateEvent: nil}
	// tx := sql.SqlDb
	// defer func () {
	// 	if tx.Error != nil {
	// 		tx.Rollback()
	// 	} else {
	// 		tx.Commit()
	// 	}
	// }()
	// stateTxn, err := stores.MessageStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	// if err != nil {
	// 	// either subnet does not exist or you are not uptodate
	// }
	// txn, err := stores.EventStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	// if err != nil {
	// 	// either subnet does not exist or you are not uptodate
	// }
	// defer stateTxn.Discard(context.Background())
	// defer txn.Discard(context.Background())

	logger.Debugf("Processing 1...: %s", event.ID)

	previousEventUptoDate, authEventUpToDate, _, _, err := ProcessEvent(event, eventData, true, saveMessageEvent, nil, nil, ctx, dataStates)
	if err != nil {
		logger.Errorf("Processing Error...: %v", err)
		return err
	}
	logger.Debugf("Processing 2...: %v,  %v, %s", previousEventUptoDate, authEventUpToDate, event.ID)
	// get the topic, if not found retrieve it

	if previousEventUptoDate && authEventUpToDate {
		
		_, err = SyncTypedStateById(data.Topic, topic, cfg, validator)
		if err != nil {
			return err
		}

		// err = validateState(event, data.Topic, &data.Event, entities.TopicModel, dataStates, cfg)
		// if err != nil {
		// 	logger.Debugf("validateStateError: %v", err)
		// 	return err
		// }
		// if _topic != nil {
		// 	topic = models.TopicState{Topic: *_topic}
		// }

		// errC := dsquery.IncrementCounters(event.Cycle, event.Validator, event.Subnet, nil)
		// if errC != nil {
		// 	logger.Errorf("CounterError %v", err)

		// 	// return err
		// }
		subscription := &entities.Subscription{}
		if !event.IsLocal(cfg) {
			subscription, err = ValidateMessageData(&event.Payload, topic)
		}
		if err != nil {
			// update error and mark as synced
			// notify validator of error
			logger.Errorf("MessageDataError: %v", err)
			// utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte("err:"))
			// utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte(event.ID))
			// utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte("\n"))
			//utils.UpdateStruct(entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, dataStates.Events[])
			dataStates.AddEvent(entities.Event{ID: event.ID, Error: err.Error(), IsValid: utils.FalsePtr(), Synced: utils.TruePtr()})
			// saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, &txn, nil )
			// return(err)
		} else {
			// TODO if event is older than our state, just save it and mark it as synced

			logger.Infof("VALIDATINGSTATE FOR SUBSCRIPTION %v", subscription)
			
			err = validateState(event, subscription.ID, &subscription.Event, entities.SubscriptionModel, dataStates, cfg)
			if err != nil {
				logger.Debugf("validateSubscriptionStateError: %v", err)
				return err
			}

			// savedEvent, err := saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, tx );
			dataStates.AddEvent(entities.Event{ID: event.ID, IsValid: utils.TruePtr(), Synced: utils.TruePtr()})
			dataStates.AddCurrentState(entities.MessageModel, id, data)
			//if  err == nil {
			// update state
			// logger.Debugf("CreateMessageData: %+v", data)
			// _, err = dsquery.CreateMessageState(&data, &stateTxn)

			// if err != nil {
			// 	stateTxn.Discard(context.Background())
			// 	logger.Debugf("CreateMessageErrror: %+v", err)
			// 	// _, err = saveMessageEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil)

			// 	if err != nil {
			// 		// tx.Rollback()
			// 		stateTxn.Discard(context.Background())
			// 		logger.Errorf("SaveStateError %v", err)
			// 		return err
			// 	}

			// } else {
			// 	err = stateTxn.Commit(context.Background())
			// }
			//}

			// if err == nil {
			// 	// err =  txn.Commit(context.Background())
			// } else {
			// 	panic(err)
			// 	// logger.Errorf("HandleNewPusSubMessageEventError: %v", err)
			// 	// return err
			// }
			// if err == nil {
			// 	go func ()  {
			// 		dsquery.IncrementStats(event, nil)
			// 	 dsquery.UpdateAccountCounter(event.Payload.Account.ToString())
			// 	OnFinishProcessingEvent(ctx, event,  &models.MessageState{
			// 		Message: data,
			// 	},  &event.Payload.Subnet)
			// 	}()
			// }  else {
			// 	panic(err)
			// 	// logger.Errorf("HandleNewPusSubMessageEventError: %v", err)
			// 	// return err
			// }

			// if string(event.Validator) != cfg.PublicKeyEDDHex {
			// 	go func () {
			// 	dependent, err := dsquery.GetDependentEvents(event)
			// 	if err != nil {
			// 		logger.Debug("Unable to get dependent events", err)
			// 	}
			// 	for _, dep := range *dependent {
			// 		HandleNewPubSubEvent(dep, ctx)
			// 	}
			// 	}()
			// }

		}
	}
	return nil
}


func validateState(event *entities.Event, stateId string, stateEvent  *entities.EventPath, stateModel entities.EntityModel, dataStates *dsquery.DataStates, cfg *configs.MainConfiguration, ) error {
	if event.IsLocal(cfg) {
		return nil
	}
		getEventStateId := slices.IndexFunc(event.StateEvents, func(b entities.StateEvents) bool {
			return b.ID == stateId
		})
		if getEventStateId < 0  || (getEventStateId > 0 && stateEvent.ID != (event.StateEvents)[getEventStateId].Event.ID)  {
			// getBoth Events
			currentStateEvent, _, _, err := GetEventByPath(stateEvent, cfg, "")
			if err != nil {
				logger.Debugf("GetEventByPathError1 %v", err)
				return err
			}
			logger.Infof("EVENTSTATES %v", event.StateEvents)
			eventStateEvent, eventState, isLocal, err := GetEventByPath(&event.StateEvents[getEventStateId].Event, cfg, "")
			if err != nil {
				logger.Debugf("GetEventByPathError2 %v", err)
				return err
			}
			if (!isLocal) {
				dataStates.AddEvent(
					*eventStateEvent,
				)
			}
			if IsMoreRecentEvent(
				eventStateEvent.Hash,
				int(eventStateEvent.Timestamp),
				currentStateEvent.Hash,
				int(currentStateEvent.Timestamp)) {
				if !isLocal {
					dataStates.AddHistoricState(stateModel, stateId, eventState)
				}
				// ensure that the event is not referrencing an old topic state, but in case of topics, it doesnt matter
				if !IsMoreRecentEvent(event.Hash, int(event.Timestamp), currentStateEvent.Hash, int(currentStateEvent.Timestamp)) {
					// event is using an invalid state
					err = fmt.Errorf("outdate event topic state")
					return err
				}
			} else {
				if !isLocal {
					dataStates.AddCurrentState(stateModel, stateId, eventStateEvent)
				}
			}
		}
		return nil
}