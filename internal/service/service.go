package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"gorm.io/gorm"
)

type PayloadData struct {
	Subnet         string
	localDataState *LocalDataState

	localDataStateEvent *LocalDataStateEvent
}

type LocalDataState struct {
	Hash      string
	ID        string
	Event     *entities.EventPath
	Timestamp uint64
}
type LocalDataStateEvent struct {
	Hash      string
	ID        string
	Timestamp uint64
}

// deprecated
func SaveEvent(model entities.EntityModel, where entities.Event, createData *entities.Event, updateData *entities.Event, tx *datastore.Txn) (*entities.Event, error) {
	// logger.Debugf("SavingEvent: %+v, create: %+v, update: %+v", where.ID, createData, updateData)
	if createData != nil {
		// create
		id, err := createData.GetId()
		if err != nil {
			return nil, err
		}
		event, err := dsquery.GetEventById(id, model)
		if err != nil && err != datastore.ErrNotFound {
			return nil, err
		}
		if event != nil {
			return nil, dsquery.ErrorKeyExist
		}
		// create the subnet event
		// event := createData.MsgPack()

		if err := dsquery.UpdateEvent(createData, tx, nil, true); err != nil {
			return nil, err
		}
		return createData, nil

	} else {
		id := where.ID
		// if err != nil {
		// 	return nil, err
		// }
		// logger.Infof("UpdateEvent: Synced:  %+v, IsValid: %+v, %s",*updateData.Synced, *updateData.IsValid, updateData.Error)
		event, err := dsquery.GetEventByIdTxn(id, model, tx)
		if err != nil {
			logger.Errorf("GetEventErr: %s, %v", id, err)
			return nil, err
		}

		if event == nil {
			return nil, fmt.Errorf("subnet event not found")
		}

		utils.UpdateStruct(updateData, event)
		event.ID = id
		// logger.Infof("UpdatingEvent::: %s, %v, %v", where.ID, updateData.Synced, event.Synced)
		if err := dsquery.UpdateEvent(event, tx, nil, false); err != nil {
			logger.Errorf("UpdateError: %v, IsValid %v", err, *updateData.IsValid)
			return nil, err
		}
		return event, nil
	}
}

func ProcessEvent(
	event *entities.Event,
	data PayloadData,
	validAgentRequired bool,
	saveEvent func(entities.Event, *entities.Event, *entities.Event, *datastore.Txn, *gorm.DB) (*entities.Event, error),
	txn *datastore.Txn, tx *gorm.DB, ctx *context.Context, dataDataStates *dsquery.DataStates) (preEvUptoDate bool, authEvUptoDate bool, authSt *models.AuthorizationState, isRecent bool, err error) {
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	// logger.WithFields(logrus.Fields{"event": event}).Debug("New topic event from pubsub channel")
	// markAsSynced := false
	// updateState := false
	// var eventError string
	// // hash, _ := event.GetHash()

	if !event.IsLocal(cfg) {
		if len(event.AuthEvent.ID) > 0 {
			_, stateByte, _, err := GetEventByPath(&event.AuthEvent, cfg, string(event.Validator))
			if err != nil {
				return true, false, nil, true, err
			}
			state, err := entities.UnpackAuthorization(stateByte)
			if err != nil {
				return true, true, nil, true, nil
			}
			return true, true, &models.AuthorizationState{Authorization: state}, true, nil
		}
		return true, true, nil, true, nil
	}

	logger.Debugf("ProcessingEvent.. %s, %d", event.ID, event.Payload.Timestamp)
	if validAgentRequired && event.Payload.Timestamp > 0 && (uint64(event.Payload.Timestamp) > uint64(event.Timestamp)+15000 || uint64(event.Payload.Timestamp) < uint64(event.Timestamp)-15000) {

		return false, false, nil, false, errors.New("event timestamp exceeds payload timestamp")
	}

	logger.Debugf("ProcessEvent: Event is a valid event %s", event.PayloadHash)
	//cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// Extract and validate the Data of the paylaod which is an Events Payload Data,

	previousEventUptoDate := false
	authEventUptoDate := !validAgentRequired
	eventIsMoreRecent := true
	authMoreRecent := false
	var badEvent error
	// eventIsMoreRecent := true
	// authMoreRecent := false
	logger.Errorf("EVENTHASSSS RECEIVED: %d-%s %v", event.EventType, event.ID, event.Payload)
	err = ValidateEvent(*event)

	if err != nil {
		logger.Errorf("ValidateEventError: %v", err)
		return false, false, nil, eventIsMoreRecent, err
	}
	d, err := event.Payload.EncodeBytes()
	if err != nil || len(d) == 0 {
		logger.Debug("Invalid event payload")
		return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid event payload")
	}

	//subnet := models.SubnetState{}
	var _subnet *entities.Subnet
	if event.EventType != constants.CreateSubnetEvent && event.EventType != constants.UpdateSubnetEvent {
		// err = query.GetOne(models.SubnetState{Subnet: entities.Subnet{ID: data.Subnet}}, &subnet)
		_subnet, err = dsquery.GetSubnetStateById(data.Subnet)
		if _subnet == nil {
			if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
				logger.Debugf("EVENTINFO: %v %s", err, data.Subnet)
				// get the subnetstate from the sending node
				subPath := entities.NewEntityPath(event.Validator, entities.SubnetModel, data.Subnet)
				pp, err := p2p.GetState(cfg, *subPath, &event.Validator, &_subnet)
				if err != nil {
					logger.Error(err, " ", (*subPath).ToString(), " ", event.Validator)
					return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to get subnetdata")
				}
				logger.Infof("EVENTFOUND %s", pp.Event)
				logger.Infof("STATEFOUND %v", pp.States)
				if len(pp.Event) < 2 {
					return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to unpack subnetdata")
				}
				subnetEvent, err := entities.UnpackEvent(pp.Event, entities.SubnetModel)
				if err != nil {
					logger.Errorf("UnpackError: %v", err)
					return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to unpack subnetdata")
				}
				// err = dsquery.CreateEvent(subnetEvent, txn)
				// if err != nil {
				// 	return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to save subne event")
				// }
				dataDataStates.AddEvent(*subnetEvent)
				// dataDataStates.Events[subnetEvent.ID] = *subnetEvent
				for _, snetData := range pp.States {
					_subnet, err := entities.UnpackSubnet(snetData)
					if err == nil {
						// dataDataStates.CurrentStates[_subnet.ID] = _subnet
						dataDataStates.AddCurrentState(entities.SubnetModel, _subnet.ID, _subnet)
						// s, err := dsquery.CreateSubnetState(&_subnet, nil)
						// if err != nil {
						// 	return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to save subnet state")
						// }
						// _subnet = *s;
					}
				}
			} else {
				return false, false, nil, false, nil
			}
		}
	}

	 agent, _ := entities.AddressFromString(string(event.Payload.DeviceKey))
	if validAgentRequired || agent.Addr != "" {
		agentString, err := crypto.GetSignerECC(&d, &event.Payload.Signature)
		if err != nil {
			logger.Debug("Errpr", err)
			return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid agent signature")
		}

		if strings.Compare(agentString, agent.Addr) != 0 {
			logger.Debug("Invalid agent signer")
			return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid payload signer")
		}
	}
	// eventModel, _, err := query.SaveRecord(models.TopicEvent{Event: entities.Event{Hash: event.Hash}}, &models.TopicEvent{Event: *event}, nil, sql.SqlDb)
	// _, err = saveEvent( entities.Event{ID: event.ID},  event, nil, txn, tx)
	dataDataStates.AddEvent(*event)
	// if err != nil {

	// 	if err != dsquery.ErrorKeyExist {
	// 		logger.Errorf("saveEventError: %v", err)
	// 		return false, false,nil, eventIsMoreRecent, fmt.Errorf("event storage failed")
	// 	}
	// 	_, err = saveEvent( entities.Event{ID: event.ID},  nil, event, txn, tx)
	// 	if err != nil {
	// 		return false, false,nil, eventIsMoreRecent, fmt.Errorf("event storage failed")
	// 	}
	// }

	
	// get agent auth state

	var eventAuthState models.AuthorizationState
	currentLocaltAuthState := models.AuthorizationState{}
	// var agentAuthStateEvent models.AuthorizationEvent

	if validAgentRequired {
		// err = query.GetOne(models.AuthorizationState{Authorization: entities.Authorization{Agent: agent.ToDeviceString(), Subnet: event.Payload.Subnet}}, &agentAuthState)
		_agentAuthState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Authorized: agent.ToAddressString(), Account: event.Payload.Account, Subnet: event.Payload.Subnet}, dsquery.DefaultQueryLimit, nil)
		if err != nil && !dsquery.IsErrorNotFound(err) {
			return false, false, nil, eventIsMoreRecent, fmt.Errorf("db error: %s", err.Error())
		}
		// let check if the authstate authorizes this
		// var authorizationIndex int
		for _, _auth := range _agentAuthState {
			//if *_auth.Priviledge >= constants.MemberPriviledge {
			// authorizationIndex = i
			currentLocaltAuthState = models.AuthorizationState{Authorization: *(_auth)}
			//	}
		}
		// if len(_agentAuthState) > 0 && currentLocaltAuthState.ID == "" {
		// 	currentLocaltAuthState = models.AuthorizationState{Authorization: *(_agentAuthState[0])}
		// }
		if currentLocaltAuthState.Priviledge != nil && *currentLocaltAuthState.Priviledge < constants.MemberPriviledge && currentLocaltAuthState.ID == event.AuthEvent.ID {
			// authorizationIndex = i
			return false, false, nil, eventIsMoreRecent, fmt.Errorf("no write priviledge")
		}

	}
	// lets determine which authstate to use to validate this event

	// get all events and auth
	var previousEvent *entities.Event
	var authEvent *entities.Event

	// if agentAuthState == nil { // we dont have any info about the agent within this subnet
	// 	// we need to know if the agent has the right to process this event, else we cant do anything
	// 	// check the node that sent the event to see if it has the record

	// }
	logger.Debugf("PreviousEvent: %+v", event.PreviousEvent)
	if len(event.PreviousEvent.ID) > 0 {
		// previousEvent, err = query.GetEventFromPath(&event.PreviousEvent)

		previousEvent, err = dsquery.GetEventFromPath(&event.PreviousEvent)

		if err != nil && err != query.ErrorNotFound && !dsquery.IsErrorNotFound(err) {
			logger.Debug("GetLocalPreviousEventError: ", err)
			return false, false, nil, eventIsMoreRecent, fmt.Errorf("db err: %s", err.Error())
		}
		// check if we have the previous event locally, if we dont we can't proceed until we get it

		if previousEvent != nil {
			// logger.Debugf("FoundPreviousEvent: %s", previousEvent.ID)
			previousEventUptoDate = true
		} else {
			// get the previous event from the sending node and process it as well
			logger.Infof("GettingPreviousEvent: %s", event.PreviousEvent.ID)
			previousEvent, pl, err := getEventFromP2p(cfg, event.PreviousEvent, nil)
			if err != nil {
				logger.Debugf("ErrorRetrievingPreviousEvent: %v, %v", err, event.PreviousEvent)
				 
			}
			if  previousEvent != nil && previousEvent.Synced != nil && *previousEvent.Synced {
				// save event
				// err = dsquery.CreateEvent(previousEvent, txn)
				dataDataStates.Events[previousEvent.ID] = *previousEvent

				// if err != nil {
				// 	logger.Errorf("Create previous event error %v", err)
				// } else {
				for _, stData := range pl.States {
					dataDataStates.AddHistoricState(event.PreviousEvent.Model, event.PreviousEvent.ID, stData)
					// err = dsquery.SaveHistoricState(event.PreviousEvent.Model, event.PreviousEvent.ID, stData)
					// if err != nil {
					// 	logger.Errorf("Save previous historic state error %v", err)
					// }
				}
				previousEventUptoDate = true
				// }
			}

			// previousEvent, _, err = p2p.GetEvent(cfg, event.PreviousEvent, nil)
			// if err != nil {
			// 	logger.Error("GetPreviousEventError: ", err)
			// 	if event.Validator != event.PreviousEvent.Validator {
			// 		previousEvent, _, err = p2p.GetEvent(cfg, event.PreviousEvent, &event.Validator)
			// 		logger.Error("GetPReviousFromSameValidatorError %v", err)
			// 	}
			// }
			// if previousEvent != nil {
			// 	go HandleNewPubSubEvent(*previousEvent, ctx)
			// }

		}

	} else {
		previousEventUptoDate = true
	}

	updateAuthState := validAgentRequired
	syncAuthEvent := validAgentRequired
	authEventAuthState := &entities.Authorization{}
	if validAgentRequired && currentLocaltAuthState.ID == event.AuthEvent.ID && event.AuthEvent.ID  != "" {
		authEventAuthState = &currentLocaltAuthState.Authorization
		eventAuthState = models.AuthorizationState{Authorization: *authEventAuthState}
		authEvent, err = dsquery.GetEventFromPath(&event.AuthEvent)
		if err != nil {
			logger.Infof("AutheEventError: %v", err)
			syncAuthEvent = true
			authEventUptoDate = false
		} else {
			syncAuthEvent = false
			updateAuthState = true
		}
	}
	if syncAuthEvent || (validAgentRequired || len(event.AuthEvent.ID) > 0) && (currentLocaltAuthState.ID == "" || currentLocaltAuthState.Event.ID != event.AuthEvent.ID) {
		// check if we have the associated auth event locally, if we dont we can't proceed until we get it
		updateAuthState = false
		logger.Debugf("AuthEventValidation: %+v, %+v, %v", currentLocaltAuthState.ID, event.AuthEvent.ID, (validAgentRequired || len(event.AuthEvent.ID) > 0) && (currentLocaltAuthState.ID == "" || currentLocaltAuthState.Event.ID != event.AuthEvent.ID))
		if event.AuthEvent.ID == "" {
			return previousEventUptoDate, authEventUptoDate, nil, eventIsMoreRecent, fmt.Errorf("auth event not provided")
		}
		// authEvent, err = query.GetEventFromPath(&event.AuthEvent)

		authEvent, err = dsquery.GetEventFromPath(&event.AuthEvent)
		// if err != nil && !dsquery.IsErrorNotFound(err) {
		// 	logger.Errorf("GetEventFromPathError %v", err)
		// 	return previousEventUptoDate, authEventUptoDate, nil, eventIsMoreRecent, err
		// }
		

		if err != nil {
			if dsquery.IsErrorNotFound(err) {
				// get it from another node and broadcast it
				// authEv, payload, err := p2p.GetEvent(cfg, event.AuthEvent, &event.Validator)
				_, payload, err := getEventFromP2p(cfg, event.AuthEvent, &event.Validator)

				if err != nil {
					return previousEventUptoDate, authEventUptoDate, nil, eventIsMoreRecent, fmt.Errorf("auth event not found")
				}
				if len(payload.States) == 0 {
					return previousEventUptoDate, false, nil, eventIsMoreRecent, fmt.Errorf("authstate not found")
				}
				// dataDataStates.AddEvent(*authEv)
				// err = dsquery.CreateEvent(authEv, txn)
				// if err != nil {
				// 	return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to save auth event")
				//
				authEvent, err = entities.UnpackEvent(payload.Event, entities.AuthModel)
				if err != nil {
					return previousEventUptoDate, false, nil, eventIsMoreRecent, err
				}
				for _, authData := range payload.States {
					auth, err := entities.UnpackAuthorization(authData)
					if err != nil {
						logger.Errorf("AuthEventAuth %v", err)
						continue
					}
					authEventAuthState = &auth

				}
			} else {

				return previousEventUptoDate, false, nil, eventIsMoreRecent, err
			}

		} else {
			if authEvent.Synced != nil && *authEvent.Synced {
				authEventUptoDate = true
				// get the authstate since we have the event
				// err = query.GetOneState(entities.Authorization{Event: event.AuthEvent}, &eventAuthState)
				authEventAuthState, err = dsquery.GetAuthorizationByEvent(event.AuthEvent)
				if err != nil {
					if dsquery.IsErrorNotFound(err) {
						return previousEventUptoDate, false, nil, eventIsMoreRecent, fmt.Errorf("event auth state not found")
					}
					authEventUptoDate = false
				}
				eventAuthState = models.AuthorizationState{Authorization: *authEventAuthState}
			} else {
				// authEventUptoDate = false
				// event.Error = "authEvent not synced"
				// dataDataStates.AddEvent(*event)
				return previousEventUptoDate, false, nil, eventIsMoreRecent, fmt.Errorf("authEvent not synced")
			}
		}
		if authEvent != nil && currentLocaltAuthState.ID != authEventAuthState.ID {
			if authEvent.Timestamp >= event.Timestamp {
				return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid auth signature")
			}
			if *authEventAuthState.Priviledge < constants.MemberPriviledge {
				return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid event auth state")
			}
			hash, err := authEvent.Payload.GetHash()
			if err != nil {
				return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid event auth state")
			}
			if err = VerifyAuthDataSignature(*authEventAuthState, hash, event.Payload.ChainId); err != nil {
				return false, false, nil, eventIsMoreRecent, fmt.Errorf("invalid auth signature")
			}

			// localAuthState, err := dsquery.GetAccountAuthorizations(_auth, nil, nil)
			// if err != nil && !dsquery.IsErrorNotFound(err) {
			// 	return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to get local auth state")
			// }
			// if len(localAuthState) == 0 || IsMoreRecentEvent(localAuthState[0].Event.ID, int(*localAuthState[0].Timestamp), _auth.Event.ID, int(*_auth.Timestamp), ) {
			logger.Infof("CHECKINGIFAUTHISVALID: %s", currentLocaltAuthState.ID)
			updateAuthState = true
			

			// HandleNewPubSubAuthEvent(authEv, ctx)
		} else {
			return previousEventUptoDate, false, nil, eventIsMoreRecent, nil
		}

	}

	if updateAuthState {
		if (!validAgentRequired && currentLocaltAuthState.ID == "") || (currentLocaltAuthState.Event.ID ==  authEventAuthState.Event.ID) ||  IsMoreRecentEvent(currentLocaltAuthState.Event.ID, int(*currentLocaltAuthState.Timestamp), authEventAuthState.Event.ID, int(*authEventAuthState.Timestamp)) {
			// dataDataStates.CurrentStates[entities.EntityPath{Model: entities.AuthModel, Hash: authEventAuthState.ID}] = authEventAuthState
			logger.Info("SAVEAUTHEVENTSTATEASWELL")
			if authEventAuthState != nil {
				dataDataStates.AddCurrentState(entities.AuthModel, authEventAuthState.ID, *authEventAuthState)
			}

			dataDataStates.AddEvent(*authEvent)

			// _, err =	dsquery.CreateAuthorizationState(&_auth, nil)
		} else {
			// then make sure that the event is created before the auth was updated
			if event.Payload.Timestamp > *currentLocaltAuthState.Timestamp || event.Timestamp > *currentLocaltAuthState.Timestamp {
				return false, false, nil, eventIsMoreRecent, fmt.Errorf("auth event lag")
			}
			// err = dsquery.SaveHistoricState(entities.AuthModel, _auth.ID, authData)
			dataDataStates.AddEvent(*authEvent)
			dataDataStates.AddHistoricState(entities.AuthModel, authEventAuthState.ID, authEventAuthState.MsgPack())
		}
		// if err != nil {
		// 	return false, false, nil, eventIsMoreRecent, fmt.Errorf("unable to save auth state")
		// }

		authEventUptoDate = true
	}

	logger.Debugf("PreviousEventsUpdatoDate: %v, %v, %v", previousEventUptoDate, authEventUptoDate, authEventAuthState)
	if previousEventUptoDate && authEventUptoDate {
		// check a situation where we have either of current auth and event state locally, but the states events are not same as the events prev auth and event
		// get the topics state
		// var id = data.ID
		// var badEvent  error
		// if len(data.ID) == 0 {
		// 	id, _ = entities.GetId(data)
		// } else {
		// 	id = data.ID
		// }
		logger.Debugf("PREVIOUSANDAUTH %v, %v", previousEventUptoDate, authEventUptoDate)

		var entityState = data.localDataState
		var stateEvent = data.localDataStateEvent
		// err := query.GetOne(models.TopicState{Topic: entities.Topic{ID: id}}, &topic)
		// if err != nil && err != query.ErrorNotFound {
		// 	logger.Debug(err)
		// }

		if entityState != nil && len(entityState.ID) > 0 {
			// check if state.Event is same as events previous has
			if entityState.Event.ID != event.PreviousEvent.ID {
				// either we are not upto date, or the sender is not
				// get the event that resulted in current state
				// topicEvent, err = query.GetEventFromPath(&topicState.Event)
				// if err != nil && err != query.ErrorNotFound {
				// 	logger.Debug(err)
				// }
				logger.Debugf("STATEEVENT %v, %+v", stateEvent.Timestamp, event.Hash)
				if len(stateEvent.ID) > 0 {
					eventIsMoreRecent = IsMoreRecentEvent(stateEvent.Hash, int(stateEvent.Timestamp), event.Hash, int(event.Timestamp))

					logger.Debugf("STATEEVENT %v, %+v", stateEvent, previousEvent)
					// if this event is more recent, then it must referrence our local event or an event after it
					if previousEvent != nil && eventIsMoreRecent && stateEvent.Hash != event.PreviousEvent.ID {
						previousEventMoreRecent := IsMoreRecentEvent(stateEvent.Hash, int(stateEvent.Timestamp), previousEvent.ID, int(previousEvent.Timestamp))
						if !previousEventMoreRecent {
							badEvent = fmt.Errorf(constants.ErrorBadRequest)
						}
					}

				}

			}
		}

		// We need to determin which authstate is valid for this event
		// if its an old event, we can just save it since its not updating state
		// if its a new one, we have to confirm that it is referencing the true latest auth event

		// so lets get the referrenced authorization
		// if event.AuthEvent.ID != "" {
		// 	err = query.GetOne(models.AuthorizationState{Authorization: entities.Authorization{Event: event.AuthEvent}}, eventAuthState)
		// 	if err != nil && err != query.ErrorNotFound {
		// 		return false, false, nil, false,fmt.Errorf("db error: %s", err.Error())
		// 	}
		// 	// if we dont have it, get it from another node
		// }
		// if event is more recent that our local state, we have to check its validity since it updates state
	
		if eventIsMoreRecent && validAgentRequired && currentLocaltAuthState.ID != "" && currentLocaltAuthState.Event.ID != event.AuthEvent.ID && authEvent != nil {
			// get the event that is responsible for the current state
			// err := query.GetOne(models.AuthorizationEvent{Event: entities.Event{Hash: agentAuthState.Event.ID}}, &agentAuthStateEvent)
			localAuthEvent, err := dsquery.GetEventById(currentLocaltAuthState.Event.ID, entities.AuthModel)
			if err != nil && !dsquery.IsErrorNotFound(err) {
				logger.Debug(err)
				return previousEventUptoDate, authEventUptoDate, &eventAuthState, eventIsMoreRecent, err
			}
			// agentAuthStateEvent = models.AuthorizationEvent{Event: *authEvent}
			if localAuthEvent != nil && localAuthEvent.ID != "" {
				authMoreRecent = IsMoreRecentEvent(currentLocaltAuthState.ID, int(*currentLocaltAuthState.Timestamp), localAuthEvent.ID, int(localAuthEvent.Timestamp))
				// authMoreRecent = authMoreRecent &&
				if !authMoreRecent {
					// this is a bad event using an old auth state.
					// REJECT IT
					badEvent = fmt.Errorf(constants.ErrorUnauthorized)
				}
			}
		}

		if badEvent != nil {
			// update the event state with the error
			// _, _, err = query.SaveRecord(models.TopicEvent{Event: entities.Event{Hash: event.Hash}}, &models.TopicEvent{Event: *event}, &models.TopicEvent{Event: entities.Event{Error: badEvent.Error(), IsValid: false, Synced: true}}, nil)
			//utils.UpdateStruct(entities.Event{Error: badEvent.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, dataDataStates.Events[event.ID])
			//_, err = saveEvent(entities.Event{ID: event.ID}, event,  &entities.Event{Error: badEvent.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, txn, tx)
			dataDataStates.AddEvent(entities.Event{ID: event.ID, Error: badEvent.Error(), IsValid: utils.FalsePtr(), Synced: utils.TruePtr()})
			if err != nil {
				logger.Error("dataDataStates.AddEvent: ", err)
			}
			// notify the originator so it can correct it e.g. let it know that there is a new authorization

			// decide whether to report the node
			return false, false, &eventAuthState, eventIsMoreRecent, fmt.Errorf("db error: %s", badEvent)
		}

	}
	// we returned the events authstate and not the agents auth state because the agents authstate might
	// have been updated after this event was triggered
	return previousEventUptoDate, authEventUptoDate, &eventAuthState, eventIsMoreRecent, nil

}

func getEventFromP2p(cfg *configs.MainConfiguration, event entities.EventPath, validator *entities.PublicKeyString) (*entities.Event, *p2p.P2pEventResponse, error) {
	evt, payload, err := p2p.GetEvent(cfg, event, validator)

	if err != nil {
		if validator != nil && string(event.Validator) != string(*validator) {
			evt, payload, err = p2p.GetEvent(cfg, event, nil)
		}

	}
	return evt, payload, err
}

