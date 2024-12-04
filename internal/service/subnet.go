package service

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"gorm.io/gorm"
	// query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
)

/*
Validate an agent authorization
*/
func ValidateSubnetData(clientPayload *entities.ClientPayload, chainID configs.ChainId) (currentSubnetState *models.SubnetState, err error) {
	// check fields of Subnet
	subnet := clientPayload.Data.(entities.Subnet)
	agent := entities.AddressFromString(string(subnet.Agent))
	account := entities.AddressFromString(string(subnet.Account))
	
	if len(subnet.Agent) > 0 && subnet.ID != "" {
		
		// TODO Check that this agent is an admin of subnet. Return error if not
		priv := constants.AdminPriviledge
		
		// err := query.GetOne(models.AuthorizationState{Authorization: entities.Authorization{
		// 	Agent: agent.ToDeviceString(),
		// 	Subnet: subnet.ID,
		// 	Priviledge: &priv,
		// 	Account: account.ToString(),
		// }}, &auth)
		authorizations, err := dsquery.GetAccountAuthorizations( entities.Authorization{
			Agent: agent.ToDeviceString(),
			Subnet: subnet.ID,
			Account: account.ToString(),
		}, dsquery.DefaultQueryLimit, nil)
		if err != nil  {
			if  dsquery.IsErrorNotFound(err) {
				return nil,  apperror.Unauthorized("agent not authorized")
			}
			return nil,  apperror.Internal("internal database error")
		}
		authorized := false
		// var auth models.AuthorizationState
		for _, _auth := range authorizations {
			if *(_auth.Priviledge) == priv {
				authorized = true
			}
			// auth = models.AuthorizationState{Authorization: *_auth}
		}
		if !authorized {
			return nil,  apperror.Unauthorized("agent not authorized")
		}
		
	}

	// TODO if agent is specified, ensure agent is allowed to sign on behalf of Owner

	if len(subnet.Ref) > 64 {
		return nil, apperror.BadRequest("Subnet ref cannont be more than 64 characters")
	}
	if len(subnet.Ref) > 0 && !utils.IsAlphaNumericDot(subnet.Ref) {
		return nil, apperror.BadRequest("Ref can only include alpha-numerics, and .")
	}
	var valid bool
	// b, _ := subnet.EncodeBytes()
	msg, err := clientPayload.GetHash()
	if err != nil {
		return nil, err
	}
	action :=  "Subnet"
	switch subnet.SignatureData.Type {
	case entities.EthereumPubKey:
		authMsg := fmt.Sprintf(constants.SignatureMessageString, action, chainID, subnet.Ref, encoder.ToBase64Padded(msg))
		logger.Debug("MSG:: ", authMsg)
		msgByte := crypto.EthMessage([]byte(authMsg))

		valid = crypto.VerifySignatureECC(entities.AddressFromString(string(subnet.Account)).Addr, &msgByte, subnet.SignatureData.Signature)

	case entities.TendermintsSecp256k1PubKey:
		
		decodedSig, err := base64.StdEncoding.DecodeString(subnet.SignatureData.Signature)
		if err != nil {
			return nil, err
		}
		// account := entities.AddressFromString(string(subnet.Account))
		publicKeyBytes, err := base64.RawStdEncoding.DecodeString(subnet.SignatureData.PublicKey)

		if err != nil {
			return nil, err
		}
		authMsg := fmt.Sprintf(constants.SignatureMessageString, action, chainID, subnet.Ref, encoder.ToBase64Padded(msg))
		logger.Debug("MSG:: ", authMsg)
		valid, err = crypto.VerifySignatureAmino(encoder.ToBase64Padded([]byte(authMsg)), decodedSig, account.Addr, publicKeyBytes)
		if err != nil {
			return nil, err
		}

	}
	if !valid {
		return nil, apperror.Unauthorized("Invalid subnet data signature")
	}
	
	if subnet.ID != "" {
		// var  curSt  models.SubnetState
		// query.GetOne(models.SubnetState{Subnet: entities.Subnet{ID: subnet.ID}}, &curSt)
		snetS, err := dsquery.GetSubnetStateById(subnet.ID)
		if err != nil {
			if !dsquery.IsErrorNotFound(err) {
				return nil, err
			} else {
				return nil, nil
			}
		}
		currentSubnetState = &models.SubnetState{Subnet: *snetS}
	}
	
	return currentSubnetState, nil
}

func saveSubnetEvent(where entities.Event, createData *entities.Event, updateData *entities.Event, txn *datastore.Txn, tx *gorm.DB) (*entities.Event, error) {
	// var createModel *models.SubnetEvent
	// if createData != nil {
	// 	createModel = &models.SubnetEvent{Event: *createData}
	// } else {
	// 	createModel = &models.SubnetEvent{}
	// }
	// var updateModel *models.SubnetEvent
	// if updateData != nil {
	// 	updateModel = &models.SubnetEvent{Event: *updateData}
	// }
	// model, _, err := query.SaveRecord(models.SubnetEvent{Event: where},  createModel, updateModel, tx)
	// if err != nil {
	// 	return nil, err
	// }
	// return &model.Event, err
	// get the key
	return SaveEvent(entities.SubnetModel, where, createData, updateData, txn)
// 	if createData != nil {
// 		// create
// 		id, err := createData.GetId()
// 		if err != nil {
// 			return nil, err
// 		}
// 		subnetEvent, err := dsquery.GetEventById(id, entities.SubnetModel)
// 		if err != nil && err != datastore.ErrNotFound {
// 			return nil, err
// 		}
// 		if subnetEvent != nil {
// 			return nil, fmt.Errorf("event exists")
// 		}
// 		// create the subnet event
// 		// event := createData.MsgPack()
// 		if err := dsquery.CreateEvent(createData); err != nil {
// 			return nil, err
// 		}
// 		return createData, nil
		
// 	} else {
// 		id, err := where.GetId()
// 		subnetEvent, err := dsquery.GetEventById(id,  entities.SubnetModel)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if subnetEvent == nil {
// 			return nil, fmt.Errorf("subnet event not found")
// 		}
// 		utils.UpdateStruct(updateData, subnetEvent)
// 		if err := dsquery.UpdateEvent(subnetEvent); err != nil {
// 			return  nil, err
// 		}
// 		return subnetEvent, nil
// 	}
 }


func HandleNewPubSubSubnetEvent(event *entities.Event, ctx *context.Context, ) error {

	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	
	if !ok {
		panic("Unable to load config from context")
	}
	
	
	data := event.Payload.Data.(entities.Subnet)
	data.Event = *event.GetPath()
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.EventSignature = event.Signature
	hash, err := data.GetHash()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	logger.Debugf("HandlingNewEvent: %s in subnet %s", data.ID, event.Payload.Subnet )
	var id string
	if len(data.ID) == 0 {
		id, _ = entities.GetId(data)
	} else {
		id = data.ID
	}
	
	var localState models.SubnetState
	// err := query.GetOne(&models.TopicState{Topic: entities.Topic{ID: id}}, &localTopicState)
	// err = sql.SqlDb.Where(&models.SubnetState{Subnet: entities.Subnet{ID: id}}).Take(&localState).Error

	 subnet, err := dsquery.GetSubnetStateById(id)
	 if err != nil && !dsquery.IsErrorNotFound(err){
		logger.Debugf("SubnetStateQueryError: %v", err)
		return err
	 }
	 if (subnet != nil ) {
	 	localState =  models.SubnetState{Subnet: *subnet}
	 }

	// if err != nil {
	// 	logger.Error(err)
	// }
	
	
	var localDataState *LocalDataState
	if localState.ID != "" {
		localDataState = &LocalDataState{
			ID: localState.ID,
			Hash: localState.Hash,
			Event: &localState.Event,
			Timestamp: localState.Timestamp,
		}
	}
	// localDataState := utils.IfThenElse(localTopicState != nil, &LocalDataState{
	// 	ID: localTopicState.ID,
	// 	Hash: localTopicState.Hash,
	// 	Event: &localTopicState.Event,
	// 	Timestamp: localTopicState.Timestamp,
	// }, nil)
	var stateEvent *entities.Event
	if localState.ID != "" {
		stateEvent, err = dsquery.GetEventFromPath(&localState.Event)
		if err != nil && err != query.ErrorNotFound && !dsquery.IsErrorNotFound(err) {
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

	eventData := PayloadData{Subnet: data.ID, localDataState: localDataState, localDataStateEvent:  localDataStateEvent}
	// tx := sql.SqlDb
	// defer func () {
	// 	if tx.Error != nil {
	// 		tx.Rollback()
	// 	} else {
	// 		tx.Commit()
	// 	}
	// }()
	txn, err := stores.EventStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	defer txn.Discard(context.Background())  
	previousEventUptoDate,  _, _, eventIsMoreRecent, err := ProcessEvent(event,  eventData, false, saveSubnetEvent, &txn, nil, ctx)
	if err != nil {
		logger.Debugf("Processing Error...: %v", err)
		return err
	}
	if id != "" {
		event.Subnet = id
		err = dsquery.IncrementCounters(event.Cycle, event.Validator, event.Subnet, &txn)
		if err != nil { 
			return err
		}
	}
	logger.Debugf("Processing 2...: %v", previousEventUptoDate)
	if previousEventUptoDate {
		_, err = ValidateSubnetData(&event.Payload, cfg.ChainId)
		
		
		if err != nil {
			// update error and mark as synced
			// notify validator of error
			logger.Infof("InvalidSubnetData: %v",  err)
			saveSubnetEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()}, &txn, nil )
			
		} else {
			// TODO if event is older than our state, just save it and mark it as synced
			
			savedEvent, err := saveSubnetEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid:  utils.TruePtr(), Subnet: event.Subnet, Synced:  utils.TruePtr()}, &txn, nil );
			if eventIsMoreRecent && err == nil {
				// update state
				if data.ID != "" {
					logger.Debug("DataID", data.ID)
					// _, _, err = query.SaveRecord(models.SubnetState{
					// 	Subnet: entities.Subnet{ID: data.ID},
					// }, &models.SubnetState{
					// 	Subnet: data,
					// }, utils.IfThenElse(event.EventType == uint16(constants.UpdateSubnetEvent), &models.SubnetState{
					// 	Subnet: data,
					// }, &models.SubnetState{}) , tx)
					_, err = dsquery.UpdateSubnetState(id, &data, nil)
					if err != nil {
						// TODO worker that will retry processing unSynced valid events with error
						_, err = saveSubnetEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil )
						
					}
				} else {
					//err = tx.Create(&models.SubnetState{Subnet: data}).Error
					logger.Infof("UpdateSubnetState: %s", data.ID)
					_, err = dsquery.CreateSubnetState(&data, nil)
					if err != nil {
						// TODO worker that will retry processing unSynced valid events with error
						_, err =  saveSubnetEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid:  utils.FalsePtr(), Synced:  utils.TruePtr()}, &txn, nil )
					}
				}
				if err != nil {
					// tx.Rollback()
					logger.Errorf("SaveStateError %v", err)
					return err
				} else {
					_, err = saveSubnetEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid: utils.TruePtr(), Synced:  utils.TruePtr()}, &txn, nil )
				}
				
			}
			if err == nil {
				if err = txn.Commit(context.Background()); err != nil {
					logger.Errorf("ErorrSavingEvent: %v", err)
					return err
				}
				go func ()  {
					dsquery.UpdateAccountCounter(data.Account.ToString())
					dsquery.IncrementStats(event, nil)

					OnFinishProcessingEvent(ctx, event, &models.SubnetState{
							Subnet: data,
						}, &savedEvent.ID)
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



	// logger.WithFields(logrus.Fields{"event": event}).Debug("New Subnet event from pubsub channel")
	// markAsSynced := false
	// updateState := false
	// var eventError string
	// // hash, _ := event.GetHash()
	// err := ValidateEvent(*event)

	// if err != nil {
	// 	logger.Error(err)
	// 	return
	// }

	// logger.Debugf("Event is a valid event %s", event.PayloadHash)
	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// // Extract and validate the Data of the paylaod which is an Events Payload Data,
	// data := event.Payload.Data.(entities.Subnet)
	// hash, _ := data.GetHash()
	// data.Hash = hex.EncodeToString(hash)
	// // authEventHash := event.AuthEvent
	// // authState, authError := query.GetOneAuthorizationState(entities.Authorization{Event: authEventHash})
	// logger.Debug("data.Meta Ref ", data.Meta, " ", data.Ref)
	// h, _ := data.GetHash()
	// logger.Debugf("data.Hash %v", h)

	// currentState, err := ValidateSubnetData(&data, cfg.ChainId)
	// if err != nil {
	// 	// penalize node for broadcasting invalid data
	// 	logger.Debugf("Invalid Subnet data %v. Node should be penalized", err)
	// 	return
	// }

	// // check if we are upto date on this event
	// prevEventUpToDate := query.EventExist(&event.PreviousEvent) || (currentState == nil && event.PreviousEvent.Hash == "") || (currentState != nil && currentState.Event.Hash == event.PreviousEvent.Hash)
	// // authEventUpToDate := query.EventExist(&event.AuthEvent) || (authState == nil && event.AuthEvent.Hash == "") || (authState != nil && authState.Event == authEventHash)

	// // Confirm if this is an older event coming after a newer event.
	// // If it is, then we only have to update our event history, else we need to also update our current state
	// isMoreRecent := false
	// if currentState != nil && currentState.Hash != data.Hash {
	// 	var currentStateEvent = &models.SubnetEvent{}
	// 	_ = query.GetOne(entities.Event{Hash: currentState.Event.Hash}, currentStateEvent)
	// 	isMoreRecent, markAsSynced = IsMoreRecent(
	// 		currentStateEvent.ID,
	// 		currentState.Event.Hash,
	// 		currentStateEvent.Payload.Timestamp,
	// 		event.Hash,
	// 		event.Payload.Timestamp,
	// 		markAsSynced,
	// 	)
	// 	// if uint64(currentStateEvent.Payload.Timestamp) < uint64(event.Payload.Timestamp) {
	// 	// 	isMoreRecent = true
	// 	// }
	// 	// if uint64(currentStateEvent.Payload.Timestamp) > uint64(event.Payload.Timestamp) {
	// 	// 	isMoreRecent = false
	// 	// }
	// 	// // if the authorization was created at exactly the same time but their hash is different
	// 	// // use the last 4 digits of their event hash
	// 	// if uint64(currentStateEvent.Payload.Timestamp) == uint64(event.Payload.Timestamp) {
	// 	// 	// get the event payload of the current state

	// 	// 	if err != nil && err != gorm.ErrRecordNotFound {
	// 	// 		logger.Error("DB error", err)
	// 	// 	}
	// 	// 	if currentStateEvent.ID == "" {
	// 	// 		markAsSynced = false
	// 	// 	} else {
	// 	// 		// if currentStateEvent.Payload.Timestamp < event.Payload.Timestamp {
	// 	// 		// 	isMoreRecent = true
	// 	// 		// }
	// 	// 		// if currentStateEvent.Payload.Timestamp == event.Payload.Timestamp {
	// 	// 			// logger.Debugf("Current state %v", currentStateEvent.Payload)
	// 	// 			csN := new(big.Int)
	// 	// 			csN.SetString(currentState.Event.Hash[56:], 16)
	// 	// 			nsN := new(big.Int)
	// 	// 			nsN.SetString(event.Hash[56:], 16)

	// 	// 			if csN.Cmp(nsN) < 1 {
	// 	// 				isMoreRecent = true
	// 	// 			}
	// 	// 		//}
	// 	// 	}
	// 	// }
	// }

	// // If no error, then we should act accordingly as well
	// // If are upto date, then we should update the state based on if its a recent or old event
	// if len(eventError) == 0 {
	// 	if prevEventUpToDate { // we are upto date
	// 		if currentState == nil || isMoreRecent {
	// 			updateState = true
	// 			markAsSynced = true
	// 		} else {
	// 			// Its an old event
	// 			markAsSynced = true
	// 			updateState = false
	// 		}
	// 	} else {
	// 		updateState = false
	// 		markAsSynced = false
	// 	}

	// }

	// // Save stuff permanently
	// tx := sql.SqlDb
	
	// // If the event was not signed by your node
	// if string(event.Validator) != (*cfg).PublicKey  {
	// 	// save the event
	// 	event.Error = eventError
	// 	event.IsValid = markAsSynced && len(eventError) == 0.
	// 	event.Synced = markAsSynced
	// 	event.Broadcasted = true
		
	// 	_, _, err = query.SaveRecord(&models.SubnetEvent{
	// 		Event: entities.Event{
	// 			PayloadHash: event.PayloadHash,
	// 		},
	// 	}, 
	// 	&models.SubnetEvent{
	// 		Event: *event,
	// 	}, nil,  tx)
	// 	if err != nil {
	// 		// tx.Rollback()
	// 		logger.Error("1000: Db Error", err)
	// 		return
	// 	}
	// } else {
	// 	if markAsSynced {
	// 		err = tx.Where(
	// 			&models.SubnetEvent{
	// 					Event: entities.Event{PayloadHash: event.PayloadHash},
	// 				}).Assign(
	// 			&models.SubnetEvent{
	// 			Event: entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0},
	// 		}).FirstOrCreate(&models.SubnetEvent{
	// 				Event: *event,
	// 			}).Error
	// 		// _, _, err = query.SaveRecord(&models.SubnetEvent{
	// 		// 	Event: entities.Event{PayloadHash: event.PayloadHash},
	// 		// },
	// 		// &models.SubnetEvent{
	// 		// 	Event: *event,
	// 		// },
	// 		// &models.SubnetEvent{
	// 		// 	Event: entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0},
	// 		// }, tx)
	// 		if err != nil {
	// 			logger.Error("DB error: ", err)
	// 		}
	// 	} else {
	// 		// mark as broadcasted
	// 		_, _, err = query.SaveRecord(&models.SubnetEvent{
	// 			Event: entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},
	// 		},
	// 			&models.SubnetEvent{
	// 				Event: entities.Event{Broadcasted: true},
	// 			},
	// 			&models.SubnetEvent{
	// 				Event: entities.Event{Broadcasted: true},
	// 			}, tx)
	// 		if err != nil {
	// 			logger.Error("DB error", err)
	// 		}
	// 	}
	// }

	
	// if err != nil {
	// 	logger.Errorf("Invalid event payload")
	// }
	// data.Event = *entities.NewEventPath(event.Validator, entities.SubnetModel, event.Hash)

	// data.Account = event.Payload.Account
	// // logger.Error("data.Public ", data.Public)
	// var newState *models.SubnetState
	// if updateState {
	// 	newState, _, err = query.SaveRecord(&models.SubnetState{
	// 		Subnet: entities.Subnet{ID: data.ID},
	// 	}, &models.SubnetState{
	// 		Subnet: data,
	// 	}, utils.IfThenElse(event.EventType == uint16(constants.UpdateSubnetEvent), &models.SubnetState{
	// 		Subnet: data,
	// 	}, nil), tx)
	// 	if err != nil {
	// 		// tx.Rollback()
	// 		logger.Error("7000: Db Error", err)
	// 		return
	// 	}
	// }
	// // tx.Commit()
	// if markAsSynced {
	// 	go OnFinishProcessingEvent(ctx, &data.Event, utils.IfThenElse(newState!=nil, &newState.ID, nil), utils.IfThenElse(event.Error!="", apperror.Internal(event.Error), nil))
	// }
	// if string(event.Validator) != (*cfg).PublicKey  {
	// 	dependent, err := query.GetDependentEvents(*event)
	// 	if err != nil {
	// 		logger.Debug("Unable to get dependent events", err)
	// 	}
	// 	for _, dep := range *dependent {
	// 		go HandleNewPubSubSubnetEvent(&dep, ctx)
	// 	}
	// }

	// TODO Broadcast the updated state
}
return nil
}

