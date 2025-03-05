package service

// import (
// 	"context"
// 	"encoding/hex"
// 	"fmt"
// 	"reflect"
// 	"slices"
// 	"strings"

// 	"github.com/mlayerprotocol/go-mlayer/common/apperror"
// 	"github.com/mlayerprotocol/go-mlayer/common/constants"
// 	"github.com/mlayerprotocol/go-mlayer/common/utils"
// 	"github.com/mlayerprotocol/go-mlayer/configs"
// 	"github.com/mlayerprotocol/go-mlayer/entities"
// 	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
// 	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
// 	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
// 	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
// 	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/gorm"
// )

// /*
// Validate an agent authorization
// */
// // func dataValidator[Entt any, Auth any, M any](topic *Entt, authState *Auth) (currentTopicState *M, err error) {
// // 	subnet := models.SubnetState{}
// // 	rTopic := reflect.ValueOf(topic)
// // 	subnetId := rTopic.FieldByName("Subnet").Interface()

// // 	// TODO state might have changed befor receiving event, so we need to find state that is relevant to this event.
// // 	err = query.GetOne(models.SubnetState{Subnet: entities.Subnet{ID: fmt.Sprint(subnetId)}}, &subnet)
// // 	if err != nil {
// // 		if err == gorm.ErrRecordNotFound {
// // 			return nil, apperror.Forbidden("Invalid subnet id")
// // 		}
// // 		return nil, apperror.Internal(err.Error())
// // 	}
// // 	if  *authState.Priviledge < constants.MemberPriviledge {
// // 		return nil, apperror.Forbidden("Agent does not have enough permission to create topics")
// // 	}

// // 	if len(topic.Ref) > 40 {
// // 		return nil, apperror.BadRequest("Topic handle can not be more than 40 characters")
// // 	}
// // 	if !utils.IsAlphaNumericDot(topic.Ref) {
// // 		return nil, apperror.BadRequest("Handle must be alphanumeric, _ and . but cannot start with a number")
// // 	}
// // 	return nil, nil
// // }

// // func dataValidator[Event any, State any](event Event, state State, auth *models.AuthorizationState) (currState State, err error) {
// // 	rEvent := reflect.ValueOf(event)
// // 	return state, nil
// // }

// type CurrentState struct {
// 	EventHash string
// 	Hash      string
// 	Event     entities.EventPath
// }

// func HandleNewPubSubEventOld(eventPayloadDataType any, event *entities.Event, ctx *context.Context) {

// 	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
// 	logger.WithFields(logrus.Fields{"event": eventPayloadDataType}).Debug("New topic event from pubsub channel")
// 	// _, ok := (*ctx).Value(constants.EventCountStore).(*ds.Datastore)
// 	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

// 	if !ok {
// 		panic("Unable to get config")
// 	}
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	var tx *gorm.DB
// 	defer func () {
// 		if tx != nil {
// 			if tx.Error == nil {
// 				tx.Commit()
// 			} else {
// 				logger.Errorf("Transaction %v", tx.Error)
// 				tx.Rollback()
// 			}
// 		}
// 	}()
// 	var data any
// 	validAuth := false
// 	markAsSynced := false
// 	updateState := false
// 	var eventError string
// 	// hash, _ := event.GetHash()
// 	var authState *models.AuthorizationState
// 	// var localAuthState *models.AuthorizationState
// 	var account string
// 	var authError error
// 	var modelName entities.EntityModel
// 	//var entityState any
// 	var entityHash string
// 	var currentState *CurrentState
// 	var currentStateEvent entities.Event
// 	var where any
// 	var broadcastedWhere any
// 	var subnet string
// 	// var currentAgentAuthState *models.AuthorizationState

// 	previousEventHash := event.PreviousEvent
// 	authEventHash := event.AuthEvent
// 	// val := reflect.ValueOf(eventType)
// 	// previousEventHash := val.FieldByName("PreviousEvent").Interface().(entities.EventPath)
// 	// authEventHash := val.FieldByName("AuthEvent").Interface().(entities.EventPath)
// 	d, err := event.Payload.EncodeBytes()
// 	if err != nil {
// 		logger.Errorf("Invalid event payload")
// 	}
// 	agent, err := crypto.GetSignerECC(&d, &event.Payload.Signature)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = ValidateEvent(*event)
// 	if err != nil {
// 		logger.Error(err)
// 		return
// 	}
// 	// TODO Check if event signer is a validator
// 	// reject event if not

// 	// save event
// 	// if len(event.Payload.DeviceKey) > 0 {
// 	// 	currentAgentAuthState, err = query.GetOneAuthorizationState(entities.Authorization{Agent: event.Payload.DeviceKey, Subnet: event.Payload.Subnet})
// 	// 	if err != nil {
// 	// 		logger.Errorf("query: %v", err)
// 	// 	}
// 	// }
// 	switch evt := eventPayloadDataType.(type) {

// 		case entities.Topic:
// 			logger.Debugf("DataType %v", evt)
// 			modelName = entities.TopicModel
// 			topic := event.Payload.Data.(entities.Topic)

// 			hashByte, _ := topic.GetHash()
// 			entityHash = hex.EncodeToString(hashByte)
// 			topic.Hash = entityHash
// 			subnet = topic.Subnet
// 			account = string(event.Payload.Account)

// 			// check if the current local state is based on thise event
// 			authState, authError = query.GetOneAuthorizationState(entities.Authorization{Event: authEventHash})
// 			// validate from here
// 			// validateAuthState(authState)

// 			logger.Debugf("AUTHSTATE %s, %v", authEventHash, authState)
// 			curState, err := ValidateTopicData(&topic, authState)
// 			if err != nil {
// 				// penalize node for broadcasting invalid data
// 				logger.Debugf("Invalid topic data %v. Node should be penalized", err)
// 				return
// 			}
// 			if curState != nil {
// 				currentState = &CurrentState{EventHash: curState.Event.ID, Hash: curState.ID}
// 				var currStateEvent *models.TopicEvent
// 				query.GetOne(entities.Event{Hash: curState.Event.ID}, currStateEvent)
// 				if currStateEvent != nil {
// 					utils.CopyStructValues(currStateEvent.Event, currentStateEvent)
// 				}
// 			}
// 			where = models.TopicEvent{
// 				Event: entities.Event{Hash: event.Hash},
// 			}
// 			broadcastedWhere = models.TopicEvent{
// 				Event: entities.Event{Hash: event.Hash},
// 			}
// 			topic.Event = *entities.NewEventPath(event.Validator, modelName, event.Hash)
// 			topic.DeviceKey = entities.AddressFromString(agent).ToDeviceString()
// 			topic.Account = event.Payload.Account
// 			data = topic
// 		case entities.Authorization:

// 			modelName = entities.AuthModel
// 			authRequest := event.Payload.Data.(entities.Authorization)

// 			hashByte, _ := authRequest.GetHash()
// 			entityHash = hex.EncodeToString(hashByte)
// 			authRequest.Hash = entityHash
// 			subnet = authRequest.Subnet
// 			account = string(event.Payload.Account)
// 			agent = string(authRequest.DeviceKey)
// 			authRequest.DeviceKey = entities.AddressFromString(string(authRequest.DeviceKey)).ToDeviceString()
// 			// check if the current local state is based on thise event
// 			// authState, authError = query.GetOneAuthorizationState(entities.Authorization{Subnet: subnet, Agent: entities.DeviceString(agent)})
// 			// validate from here
// 			// validateAuthState(authState)

// 			logger.Debugf("AUTHSTATE %s, %v", authEventHash, authState)
// 			curState, grantorAuthState, subnetData, authError := ValidateAuthPayloadData(&authRequest, cfg.ChainId)

// 			if authError != nil {
// 				// penalize node for broadcasting invalid data

// 				// check if the subnet exists, if it doesnt, get it from the dht
// 				if strings.Contains(authError.Error(), "4004:")  && subnetData == nil {

// 					// get the subnet from the sending node

// 					subnetState := entities.Subnet{}
// 					_, err := p2p.GetState(cfg, *entities.NewEntityPath(event.Validator, entities.SubnetModel, subnet),nil, &subnetState)

// 					if err != nil {
// 						logger.Errorf("service/GetSubnet: %v", err)
// 						// return
// 					} else {
// 						logger.Debugf("SUBNETTTTT: %v", subnetState)
// 						if event.PreviousEvent.Model != entities.SubnetModel {

// 						}

// 					}
// 				}

// 				return
// 			} else {
// 				validAuth = true
// 			}
// 			if curState != nil {
// 				currentState = &CurrentState{EventHash: curState.Event.ID, Hash: curState.ID}
// 				var currStateEvent *models.AuthorizationEvent
// 				query.GetOne(entities.Event{Hash: curState.Event.ID}, currStateEvent)
// 				if currStateEvent != nil {
// 					utils.CopyStructValues(currStateEvent.Event, currentStateEvent)
// 				}
// 			}

// 			// TODO check if grantorAuthState is valid
// 			if grantorAuthState != nil {
// 				authState = grantorAuthState

// 			}
// 			where = models.AuthorizationEvent{
// 				Event: entities.Event{Hash: event.Hash},
// 			}
// 			broadcastedWhere = models.AuthorizationEvent{
// 				Event: entities.Event{Hash: event.Hash},
// 			}
// 			authRequest.Event = *entities.NewEventPath(event.Validator, modelName, event.Hash)
// 			authRequest.DeviceKey = entities.AddressFromString(agent).ToDeviceString()
// 			// authRequest.Account = event.Payload.Account
// 			data = authRequest

// 		//defer finalize(ctx, cfg, topic, event, &updateState, &markAsSynced, &tx)

// 	}
// 	tx = sql.SqlDb.Begin()
// 	tx.Where(where).FirstOrCreate(eventTypeFromWhere(where, *event))
// 		if tx.Error != nil {
// 			logger.Errorf("TXERROR %v", tx.Error)
// 			// tx.Rollback()
// 			return
// 		}

// 	if !slices.Contains([]string{string(entities.AuthModel), string(entities.SubnetModel)}, string(modelName)) {
// 		validAuth, err = validateAuthState(ctx, cfg, authState, authState, event, subnet, agent, account)
// 		if err != nil {
// 			logger.Errorf("validateAuthError: %v", err)
// 		}
// 	}

// 		logger.Debugf("PREVIOUSEVENTHASH: %s", previousEventHash.ToString())
// 	prevEventUpToDate := query.EventExist(&previousEventHash) || (currentState == nil && previousEventHash.Hash == "") || (currentState != nil && currentState.Event.ID == previousEventHash.Hash)
// 	authEventUpToDate := query.EventExist(&authEventHash) || (authState == nil && event.AuthEvent.ID == "") || (authState != nil && authState.Event == authEventHash)

// 	if validAuth {

// 		isMoreRecent := false
// 		if currentState != nil && currentState.ID != entityHash {

// 			isMoreRecent, markAsSynced = IsMoreRecent(
// 				currentStateEvent.ID,
// 				currentState.ID,
// 				currentStateEvent.Payload.Timestamp,
// 				event.Hash,
// 				event.Payload.Timestamp,
// 				markAsSynced,
// 			)
// 		}

// 		// get local auth state

// 		// logger.Debugf("Event is a valid event %s", iEvent.PayloadHash)

// 		if authError != nil {
// 			// check if we are upto date. If we are, then the error is an actual one
// 			// the error should be attached when saving the event
// 			// But if we are not upto date, then we might need to wait for more info from the network

// 			if prevEventUpToDate && authEventUpToDate {
// 				// we are upto date. This is an actual error. No need to expect an update from the network
// 				eventError = authError.Error()
// 				markAsSynced = true
// 			} else {
// 				if currentState == nil || (currentState != nil && isMoreRecent) { // it is a morer ecent event
// 					if strings.HasPrefix(authError.Error(), constants.ErrorForbidden) || strings.HasPrefix(authError.Error(), constants.ErrorUnauthorized) {
// 						markAsSynced = false
// 					} else {
// 						// entire event can be considered bad since the payload data is bad
// 						// this should have been sorted out before broadcasting to the network
// 						// TODO penalize the node that broadcasted this
// 						eventError = authError.Error()
// 						markAsSynced = true
// 					}

// 				} else {
// 					// we are upto date. We just need to store this event as well.
// 					// No need to update state
// 					markAsSynced = true
// 					eventError = authError.Error()
// 				}
// 			}

// 		}

// 		if len(eventError) == 0 {
// 			logger.Debugf("MARKASYNCED:: %v, %v", prevEventUpToDate, authEventUpToDate)
// 			if prevEventUpToDate && authEventUpToDate { // we are upto date
// 				if currentState == nil || isMoreRecent {
// 					updateState = true
// 					markAsSynced = true
// 				} else {
// 					// Its an old event
// 					markAsSynced = true
// 					updateState = false
// 				}
// 			} else {
// 				updateState = false
// 				markAsSynced = false
// 			}

// 		}
// 	} else {

// 		markAsSynced = false
// 	}

// 	// Save stuff permanently

// 	// If the event was not signed by your node
// 	if string(event.Validator) != (*cfg).PublicKey {
// 		// save the event
// 		event.Error = eventError
// 		event.IsValid = markAsSynced && len(eventError) == 0.
// 		event.Synced = markAsSynced
// 		event.Broadcasted = true

// 		rsl := tx.Where(where).Updates(eventTypeFromWhere(where, *event))
// 		if rsl.Error != nil {
// 			logger.Errorf("TXERROR %v", tx.Error)
// 			// tx.Rollback()
// 			return
// 		}

// 	} else {
// 		exists := eventTypeFromWhere(where, entities.Event{})
// 		tx.Where(broadcastedWhere).First(exists)
// 		rItem := reflect.ValueOf(exists).Elem()
// 		id := fmt.Sprint(rItem.FieldByName("ID").Interface())
// 		var tx2 *gorm.DB

// 		if markAsSynced {
// 			if len(id) > 0 {
// 				tx2 = tx.Where(broadcastedWhere).Updates(eventTypeFromWhere(where, entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0}))
// 			} else {
// 				event.Synced = true
// 				event.Broadcasted = true
// 				event.Error = eventError
// 				event.IsValid = len(eventError) == 0
// 				tx2 = tx.Where(broadcastedWhere).Create(eventTypeFromWhere(where, *event))
// 			}
// 		} else {
// 			// mark as broadcasted
// 			if len(id) > 0 {
// 				tx2 = tx.Where(broadcastedWhere).Updates(eventTypeFromWhere(where, entities.Event{Broadcasted: true}))
// 			} else {
// 				event.Broadcasted = true
// 				tx2 = tx.Where(broadcastedWhere).Create(eventTypeFromWhere(where, *event)).FirstOrCreate(eventTypeFromWhere(where, *event))
// 			}

// 		}
// 		if tx2.Error != nil {
// 			// tx2.Rollback()
// 			return
// 		}
// 	}

// 	if updateState {
// 		// var err error
// 		var newStateId string
// 		var newStateEvent *entities.EventPath
// 		var txResult *gorm.DB
// 		switch val := data.(type) {
// 			case entities.Topic:
// 				logger.Debugf("%vs", val.ID)
// 				topic := data.(entities.Topic)
// 				newState := models.TopicState{
// 					Topic: topic,
// 				}
// 				if event.EventType == uint16(constants.UpdateTopicEvent) {
// 					txResult = tx.Where(models.TopicState{Topic: entities.Topic{ID: topic.ID}}).Updates(&newState)
// 				} else {
// 					txResult = tx.Create(&newState)
// 				}
// 				if txResult.Error == nil {
// 					newStateId = newState.ID
// 					newStateEvent = &newState.Event
// 				} else {
// 					logger.Debug("TopicState is nil")
// 				}
// 			case entities.Authorization:
// 				auth := data.(entities.Authorization)

// 				// newStateTmp := models.AuthorizationState{
// 				// 	Authorization: auth,
// 				// }
// 				// query.GetOne(models.AuthorizationState{
// 				// 	Authorization: entities.Authorization{Agent: auth.DeviceKey, Subnet: auth.Subnet},
// 				// }, &auth)
// 				logger.Debug("AuthState", auth.DeviceKey, " ",  auth.Subnet, " ", auth.Account)
// 				// if len(auth.ID) > 0 {
// 				// 	newState, txResult = SaveState(models.AuthorizationState{
// 				// 		Authorization: entities.Authorization{Agent: auth.DeviceKey, Subnet: auth.Subnet},
// 				// 	}, models.AuthorizationState{Authorization: entities.Authorization{ID: auth.ID}}, true, tx)
// 				// } else {
// 				// 	txResult = tx.Model(&models.AuthorizationState{}).Create(&newState)
// 				// 	//newState, tx = SaveState(newState, models.AuthorizationState{Authorization: entities.Authorization{
// 				// 	//}}, false, tx)
// 				// }
// 				newState, txResult := query.SaveAuthorizationState(&auth, tx)
// 				if txResult.Error == nil {
// 					newStateId = newState.ID
// 					newStateEvent = &newState.Event
// 				} else {
// 					logger.Debug("AuthorizationState is nil")
// 				}

// 		}
// 		// if tx.Commit().Error != nil {
// 		// 	tx.Rollback()
// 		// 	logger.Error("7000: Db Error", err)
// 		// 	return
// 		// }
// 		if markAsSynced && newStateEvent != nil {
// 			go OnFinishProcessingEvent(ctx, newStateEvent, utils.IfThenElse(len(newStateId) > 0, &newStateId, nil), utils.IfThenElse(event.Error != "", apperror.Internal(event.Error), nil))
// 		}

// 		if string(event.Validator) != cfg.PublicKey {
// 			dependent, err := query.GetDependentEvents(event)
// 			if err != nil {
// 				logger.Debug("Unable to get dependent events", err)
// 			}
// 			for _, dep := range *dependent {
// 				go HandleNewPubSubTopicEvent(&dep, ctx)
// 			}
// 		}

// 	}

// 	// TODO Broadcast the updated state
// }

// func eventTypeFromWhere(where any, event entities.Event) (Model any) {
// 	switch val := where.(type) {
// 	case models.TopicEvent:
// 		logger.Debugf("DataType %s, %v", "Topic", val)
// 		return &models.TopicEvent{Event: event}
// 	case models.AuthorizationEvent:
// 		logger.Debugf("DataType %s, %v", "Topic", val)

// 		return &models.AuthorizationEvent{Event: event}

// 	}
// 	return nil
// }

// // func deriveStateTypeModel(data any) (Model any) {
// // 	switch val := data.(type) {
// // 	case entities.Topic:
// // 		logger.Debugf("DataType %v", val)
// // 		return &models.TopicState{Topic: data.(entities.Topic)}
// // 	}
// // 	return entities.Event{}
// // }
// // func finalize (ctx *context.Context, cfg *configs.MainConfiguration, data any, event *entities.Event, updateState *bool, markAsSynced *bool, tx *gorm.DB) {
// // 	if *updateState {
// // 	var	err error
// // 	var newStateId string
// // 	var newStateEvent *entities.EventPath

// // 	switch val := data.(type) {
// // 		case entities.Topic:
// // 			logger.Debug("VallContente: %v", val)
// // 			topic := data.(entities.Topic)
// // 			logger.Debugf("7000: %v", data)
// // 				// newState, _, _ := query.SaveRecord(models.TopicState{
// // 				// 	Topic: entities.Topic{ID: val.ID},
// // 				// }, &models.TopicState{
// // 				// 	Topic: val,
// // 				// }, utils.IfThenElse(event.EventType == uint16(constants.UpdateTopicEvent), &models.TopicState{
// // 				// 	Topic: val,
// // 				// }, &models.TopicState{}) , tx)
// // 				// newState, _, _ := query.SaveRecord(models.TopicState{
// // 				// 	Topic: entities.Topic{ID: val.ID},
// // 				// }, &models.TopicState{
// // 				// 	Topic: val,
// // 				// }, &models.TopicState{
// // 				// 	Topic: val,
// // 				// }, tx)
// // 				newState := models.TopicState{
// // 					Topic: topic,
// // 				}
// // 				copier.Copy(newState.Topic, topic)
// // 				// tx = tx.Table("topic_states").Where(models.TopicState{
// // 				// 	Topic: entities.Topic{ID: topic.ID},
// // 				// }).Assign( utils.IfThenElse(event.EventType == uint16(constants.UpdateTopicEvent), &models.TopicState{
// // 				// 	Topic: topic,
// // 				// }, &models.TopicState{})).Create(&newState)
// // 				tx = tx.Create(newState)

// // 				if tx.Error == nil {
// // 					newStateId = newState.ID
// // 					newStateEvent = &newState.Event
// // 				} else {
// // 					logger.Debug("TopicState is nil")
// // 				}

// // 	}
// // 	// if err != nil {
// // 	// 	tx.Rollback()
// // 	// 	logger.Error("7000: Db Error", err)
// // 	// 	return
// // 	// }

// // 	if tx !=nil {
// // 		tx.Commit()
// // 	}
// // 	if *markAsSynced {
// // 		go OnFinishProcessingEvent(ctx, newStateEvent, utils.IfThenElse(len(newStateId) > 0, &newStateId, nil), utils.IfThenElse(event.Error!="", apperror.Internal(event.Error), nil))
// // 	}

// // 	if string(event.Validator) != cfg.PublicKey  {
// // 		dependent, err := query.GetDependentEvents(*event)
// // 		if err != nil {
// // 			logger.Debug("Unable to get dependent events", err)
// // 		}
// // 		for _, dep := range *dependent {
// // 			go HandleNewPubSubTopicEvent(&dep, ctx)
// // 		}
// // 	}

// // 	}
// // }

// // The validateAuthState function checks if the authorization presented by an Agent is currently and valid.
// // It compares the presented authEvent with the local one. To do this, it must get the current state from the signing
// // node and checks if the event that resulted in it is older or more recent than its local
// func validateAuthState(ctx *context.Context, cfg *configs.MainConfiguration, eventAuthState *models.AuthorizationState, localAuthState *models.AuthorizationState, event *entities.Event, subnet string, agent string, account string) (valid bool, err error) {
// 	//1. I dont have the event auth, I dont have any auth for the agent
// 	//2. I dont have the event auth, but I have an auth for the agent. Get the event auth from the validator,save it if valid, check if its older than your local, if it is, check if the event depending on it is older than the event that modified my local authstate, if it is, then its valid, process the even if there is no other more recent event that has updated the state
// 	//3. I have the event auth, but its different from the current auth of the agent - check which is most recent, if mine, check if the event was initiated before my state, if it was, then its valid, else, check if the authEvent was older than mine, if it was, throw error else accept the event
// 	if eventAuthState == nil { // my local auth state is either different or I have none
// 		// get the auth event from the sending node
// 		payload := p2p.NewP2pPayload(cfg, p2p.P2pActionGetEvent, event.AuthEvent.MsgPack())

// 		response, err := payload.SendDataRequest(cfg.PrivateKeyEDD, string(event.AuthEvent.Validator))
// 		if response.ResponseCode != 0 {
// 			logger.Errorf("P2pPayload.SendRequest: %d-%s", response.ResponseCode, response.Error)
// 			return false, fmt.Errorf(response.Error)
// 		}
// 		if err != nil {
// 			logger.Errorf("P2pPayload.SendRequest: %v", err)
// 			return false, err
// 		}
// 		if !response.IsValid(cfg.ChainId) {
// 			return false, fmt.Errorf("service/validateAuthState: Unable to get authorization event")
// 		}
// 		p2pResp, err := p2p.UnpackP2pEventResponse([]byte(response.Data))
// 		if err != nil {
// 			return false, fmt.Errorf("service/: Invalid event data")
// 		}
// 		event, err := entities.UnpackEvent(p2pResp.Event, entities.AuthModel)
// 		if err != nil {
// 			return false, fmt.Errorf("service/validateAuthState: Invalid event data")
// 		}
// 		// now we have the event, push out for processing

// 		go HandleNewPubSubTopicEvent(event, ctx)
// 		valid = false
// 		// // lets get my local authstate and see if its the same
// 		// localAuthState, _ := query.GetOneAuthorizationState(entities.Authorization{Subnet: subnet, Agent: entities.DeviceString(agent), Account: entities.AccountString(account)})
// 		// if localAuthState == nil  { // I have no auth state record as well.
// 		// 	// all i can do is store this event and try to get the authstate from this node
// 		// 	// i will still have to validate it
// 		// 	return nil, nil
// 		// } else {
// 		// 	// I have a local authstate
// 		// 	// lets see if this was an older event than the one that updated mine. If it is, then I need to
// 		// 	localAuthStateEvent, _ := query.GetEventFromPath(&localAuthState.Event)
// 		// 	if localAuthStateEvent == nil { // this shouldnt happen oftne
// 		// 		// TODO find this event right away
// 		// 	}

// 		// 	// check if the current event we are processing is more recent than
// 		// 	// the event responsible for my local authstate
// 		// 	isMoreRecent, _ := IsMoreRecent(
// 		// 		localAuthStateEvent.ID,
// 		// 		localAuthState.ID,
// 		// 		localAuthStateEvent.Payload.Timestamp,
// 		// 		event.Hash,
// 		// 		event.Payload.Timestamp,
// 		// 		false,
// 		// 	)
// 		// 	if isMoreRecent { // this means that either the originating node is not upto date or I am.
// 		// 		// I must get the referred authstate event and update my authstate before I can process this event

// 		// 	}

// 		// 		// 1. get the event from the sending node
// 		// 		// 2. validate it
// 		// 		// 3. compute the possible authorization state
// 		// 		// 4. See if I should updated the state based on it

// 		// }
// 		return valid, nil
// 	} else {
// 		// it same as my local authstate, we can use it to authenticate this event
// 		return true, nil
// 	}

// }