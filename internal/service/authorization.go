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

	//query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"

	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"gorm.io/gorm"
)

func ValidateAuthPayloadData(clientPayload *entities.ClientPayload, cfg *configs.MainConfiguration, validator string) (prevAuthState *models.AuthorizationState, grantorAuthState *models.AuthorizationState, subnet *models.SubnetState, err error) {
	auth := clientPayload.Data.(entities.Authorization)
	// if err != nil {
	// 	return nil, nil, nil, err
	// }

	logger.Debug("auth.SignatureData.Signature:: ", auth.SignatureData.Signature)

	if auth.Subnet == "" {
		return nil, nil, nil, apperror.BadRequest("Subnet is required")
	}

	// TODO find subnets state prior to the current state
	// err = query.GetOne(models.SubnetState{Subnet: entities.Subnet{ID: auth.Subnet}}, &subnet)
	// _subnet, err := dsquery.GetSubnetStateById(auth.Subnet)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
	// 		snet, err := SyncStateFromPeer(auth.Subnet, entities.SubnetModel, cfg, validator)
	// 		if err != nil {
	// 			return nil, nil, nil, err
	// 		}
	// 		_subnet = snet.(*entities.Subnet)
			
	// 	} else {
	// 		return nil, nil, nil, err
	// 	}
	// }
	 _subnet := entities.Subnet{}
	_, err = SyncTypedStateById(auth.Subnet, &_subnet,  cfg, validator )
	 if err != nil {
		return  nil, nil, subnet, err
	 }
	//  _subnet := snet.(entities.Subnet)
	 if *_subnet.Status ==  0 {
		return nil, nil, subnet, apperror.Forbidden("Subnet is disabled")
	}
	subnet = &models.SubnetState{Subnet: _subnet}

	if auth.Account != subnet.Account && *auth.Priviledge > *subnet.DefaultAuthPrivilege {
		return nil, nil, subnet, apperror.Internal("invalid auth priviledge. Cannot be higher than subnets default")
	}
	account, err := entities.AddressFromString(string(auth.Account))
	if err != nil {
		return nil, nil, subnet, apperror.BadRequest("account: " + err.Error())
	}
	if !account.IsAccount(){
		return nil, nil, subnet, apperror.BadRequest("account: " + err.Error())
	}
	grantor, err := entities.AddressFromString(string(auth.Grantor))
	if err != nil {
		return nil, nil, subnet, apperror.BadRequest("grantor: " + err.Error())
	}
	agent, err := entities.AddressFromString(string(auth.Authorized))
	if err != nil {
		return nil, nil, subnet, apperror.BadRequest("authorized: " + err.Error())
	}
	if account.Addr == agent.Addr {
		return nil, nil, subnet, apperror.Internal("cannot reassign subnet owner role")
	}
	if account.Addr == agent.Addr {
		return nil, nil, subnet, apperror.Internal("cannot reassign subnet owner role")
	}

	msg, err := clientPayload.GetHash()
	if err != nil {
		return nil, nil, subnet, err
	}
	/////

	if err = VerifyAuthDataSignature(auth, msg, cfg.ChainId); err != nil {
		return nil, nil, subnet, apperror.Unauthorized("Invalid authorization data signature")
	}
	if auth.Grantor != auth.Account  {
		// grantorAuthState, err = query.GetOneAuthorizationState(entities.Authorization{Account: entities.AccountString(string(account.ToDeviceString())), Subnet: auth.Subnet, Agent: grantor.ToDeviceString()})
		_grantorAuthState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Account: entities.AccountString(account.ToString()), Subnet: auth.Subnet, Authorized: grantor.ToAddressString()}, dsquery.DefaultQueryLimit, nil)

		// if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
		// 	accAuth := entities.Authorization{Account: entities.AccountString(string(account.ToDeviceString())), Subnet: auth.Subnet, Agent: grantor.ToDeviceString()}
		// 	auth:= entities.Authorization{}
		// 	pp, err := p2p.GetState(cfg, entities.EntityPath{Model: entities.AuthModel, Hash:  accAuth.ToAccountAuthKey(), }, nil, &auth)
		// 	if err != nil {
		// 		return nil, nil, subnet, apperror.Unauthorized("Grantor not authorized agent")
		// 	}
		// 	if len(pp.Event) < 2 {
		// 		return nil, nil, subnet, fmt.Errorf("invalid event data")
		// 	}
		// 	authEvent, err := entities.UnpackEvent(pp.Event, entities.AuthModel)
		// 	if err != nil {
		// 		logger.Error(err)
		// 		return  nil, nil, subnet, err
		// 	}
			
		// 	if authEvent != nil  && *authEvent.Synced && len(pp.States) > 0 {
		// 		// return HandleNewPubSubTopicEvent(topicEvent, ctx)
		// 		// dataStates.Events[topicEvent.ID] = *topicEvent
		// 		dataStates.AddEvent(*authEvent)
		// 		authState, err := entities.UnpackAuthorization(pp.States[0])
		// 		if err != nil {
		// 			return nil, nil, subnet, err
		// 		}
		// 		// err = dsquery.CreateEvent(topicEvent, &txn)
				
		// 		// dataStates.CurrentStates[*entities.NewEntityPath("", entities.TopicModel, topic.ID)] = topic
		// 		dataStates.AddCurrentState(entities.AuthModel, authState.ID, &authState)
		// 		// if err == nil {
		// 		// 	_, err = dsquery.CreateTopicState(&topic, nil)
		// 		// }
		// 		if err != nil {
		// 			return nil, nil, subnet, err
		// 		}
				
		// 	}
			
		// }
		if err != nil ||! dsquery.IsErrorNotFound(err) {
			return nil, grantorAuthState, subnet, apperror.Forbidden(" Grantor does not have enough permission")
		}
		if err == nil && len(_grantorAuthState) > 0 {
			grantorAuthState = &models.AuthorizationState{Authorization: *_grantorAuthState[0]}
			if *grantorAuthState.Authorization.Priviledge != constants.AdminPriviledge {
				return nil, grantorAuthState, subnet, apperror.Forbidden(" Grantor does not have enough permission")
			}
		}
	}
	// prevAuthState, err = query.GetOneAuthorizationState(entities.Authorization{Agent:  agent.ToDeviceString(), Subnet: auth.Subnet})
	_prevAuthState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Account: entities.AccountString(string(account.ToDeviceString())), Subnet: auth.Subnet, Authorized: agent.ToAddressString()}, dsquery.DefaultQueryLimit, nil)
	if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
		return nil, nil, subnet, err
	}
	if len(_prevAuthState) > 0 {
		prevAuthState = &models.AuthorizationState{Authorization: *_prevAuthState[0]}
	}
	// if !valid {
	// 	return prevAuthState, grantorAuthState, subnet, errors.New("4000: Invalid authorization data signature")
	// }

	return prevAuthState, grantorAuthState, subnet, nil

}

func VerifyAuthDataSignature(auth entities.Authorization, msg []byte, chainId configs.ChainId) error {

	account, err := entities.AddressFromString(string(auth.Account))
	if err != nil {
		return  apperror.BadRequest("account: " + err.Error())
	}
	if !account.IsAccount(){
		return  apperror.BadRequest("account: " + fmt.Sprint("invalid account"))
	}
	grantor, err := entities.AddressFromString(string(auth.Grantor))
	if err != nil {
		return  apperror.BadRequest("grantor: " + err.Error())
	}
	agent, err := entities.AddressFromString(string(auth.Authorized))
	if err != nil {
		return  apperror.BadRequest("authorized: " + err.Error())
	}
	valid := false

	action := "write_authorization"
	switch auth.SignatureData.Type {
	case entities.EthereumPubKey:
		authMsg := fmt.Sprintf(constants.SignatureMessageString, action, agent.ToAddressString(), chainId, encoder.ToBase64Padded(msg))
		logger.Debug("MSG:: ", authMsg)

		msgByte := crypto.EthMessage([]byte(authMsg))
		signer := utils.IfThenElse(len(string(auth.Grantor)) == 0, account.Addr, grantor.Addr)
		// if len(agent.Addr) > 0 {
		// 	signer = agent.Addr
		// }
		// logger.Debug("Signer:: ", signer)
		valid = crypto.VerifySignatureECC(signer, &msgByte, string(auth.SignatureData.Signature))

		if !valid {
			// check if the grantor is authorized
			return apperror.Unauthorized("invalid auth signature")
			// check if agent is authorized by grantor
			// if agent.Addr != "" {
			// 	agentAuthState, err := query.GetOneAuthorizationState(entities.Authorization{Account: entities.AccountString(grantor.ToDeviceString()), Subnet: auth.Subnet, Agent: agent.ToDeviceString()})
			// 	if err == gorm.ErrRecordNotFound {
			// 		return nil, nil, subnet, apperror.Unauthorized("Agent not authorized to act on behalf of grantor")
			// 	}
			// 	if *agentAuthState.Authorization.Priviledge != constants.AdminPriviledge {
			// 		return nil, grantorAuthState,  subnet, apperror.Forbidden("Agent does not have enough permission")
			// 	}
			// }
		}

	case entities.TendermintsSecp256k1PubKey:

		decodedSig, err := base64.StdEncoding.DecodeString(string(auth.SignatureData.Signature))
		if err != nil {
			return err
		}
		publicKeyBytes, err := base64.RawStdEncoding.DecodeString(string(auth.SignatureData.PublicKey))

		if err != nil {
			return err
		}
		// grantor, err := entities.AddressFromString(auth.Grantor)

		// if err != nil {
		// 	return nil, nil, err
		// }

		// decoded, err := hex.DecodeString(grantor.Addr)
		// if err == nil {
		// 	address = crypto.ToBech32Address(decoded, "cosmos")
		// }
		authMsg := fmt.Sprintf(constants.SignatureMessageString, action, chainId, agent.Addr, encoder.ToBase64Padded(msg))

		valid, err = crypto.VerifySignatureAmino(encoder.ToBase64Padded([]byte(authMsg)), decodedSig, grantor.Addr, publicKeyBytes)
		if err != nil {
			return err
		}

	}
	if !valid {
		return apperror.BadRequest("not valid signature")
	}
	return nil
}

func saveAuthorizationEvent(where entities.Event, createData *entities.Event, updateData *entities.Event, txn *datastore.Txn, tx *gorm.DB) (*entities.Event, error) {
	return SaveEvent(entities.AuthModel, where, createData, updateData, txn)
}

func HandleNewPubSubAuthEvent(event *entities.Event, ctx *context.Context) error {

	cfg, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to load config from context")
	}
	dataStates := dsquery.NewDataStates(event.ID, cfg)
	dataStates.AddEvent(*event)

	data := event.Payload.Data.(entities.Authorization)
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.Event = *event.GetPath()
	hash, err := data.GetHash()
	data.EventSignature = event.Signature
	// data.Authorized = entities.AddressString(data.Authorized)
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	var subnet = data.Subnet

	defer func () {
	
		
		// stateUpdateError := dataStates.Commit(nil, nil, nil, event.ID, err)
		stateUpdateError := dataStates.Save( event.ID)
		if stateUpdateError != nil {
			logger.Error("HandleNewPubSubAuthEvent/Commit", err)
			panic(stateUpdateError)
		} else {
			go OnFinishProcessingEvent(cfg, event,  data, nil)
			
			// go utils.WriteBytesToFile(filepath.Join(cfg.DataDir, "log.txt"), []byte("newMessage" + "\n"))
		}
	
	}()

	var localState *models.AuthorizationState
	// err := query.GetOne(&models.TopicState{Topic: entities.Topic{ID: id}}, &localTopicState)
	// err = sql.SqlDb.Where(&models.AuthorizationState{Authorization: entities.Authorization{Subnet: subnet, Agent: entities.AddressFromString(string(data.DeviceKey)).ToDeviceString()}}).Take(&localState).Error
	stateTxn, err := stores.StateStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	defer stateTxn.Discard(context.Background())
	tx := sql.SqlDb
	txn, err := stores.EventStore.NewTransaction(context.Background(), false) // true for read-write, false for read-only
	if err != nil {
		// either subnet does not exist or you are not uptodate
	}
	defer txn.Discard(context.Background())
	_localState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Subnet: subnet, Account: data.Account, Authorized: data.Authorized}, dsquery.DefaultQueryLimit, &stateTxn)
	if err != nil {
		logger.Error("GetAccountAuthorizations:", err)
	}
	if len(_localState) > 0 {
		localState = &models.AuthorizationState{Authorization: *_localState[0]}
	} else {
		localState = &models.AuthorizationState{}
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
		stateEvent, err = dsquery.GetEventFromPath(&(localState.Event))
		if err != nil && dsquery.IsErrorNotFound(err) {
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

	eventData := PayloadData{Subnet: subnet, localDataState: localDataState, localDataStateEvent: localDataStateEvent}

	previousEventUptoDate, authEventUpToDate, _, eventIsMoreRecent, err := ProcessEvent(event, eventData, false, saveAuthorizationEvent, &txn, tx, ctx, dataStates)
	if err != nil {
		logger.Warnf("Processing Error: %v", err)
		return err
	}
	logger.Debugf("Processing 2 auth...: %v,  %v", previousEventUptoDate, authEventUpToDate)
	if previousEventUptoDate && authEventUpToDate {
		err = dsquery.IncrementCounters(event.Cycle, event.Validator, event.Subnet, &txn)
		if err != nil {
			logger.Debugf("IncrementError: %+v", err)
			return err
		}
		if !event.IsLocal(cfg) {
			_, _, _, err = ValidateAuthPayloadData(&event.Payload, cfg, string(event.Validator))
		}
		if err != nil {
			logger.Infof("ErrorValidatingAuth %v", err)
			// saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced: utils.TruePtr()}, &txn, nil)
			dataStates.AddEvent(entities.Event{ID: event.ID, Error: err.Error(), IsValid: utils.FalsePtr(), Synced:  utils.TruePtr()})

		} else {
			// TODO if event is older than our state, just save it and mark it as synced

			// savedEvent, err := saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)
			dataStates.AddEvent(entities.Event{ID: event.ID, IsValid:  utils.TruePtr(), Synced:  utils.TruePtr()})
		
			if eventIsMoreRecent && err == nil {
				dataStates.AddCurrentState(entities.AuthModel, data.DataKey(), data)
			} else {
				dataStates.AddHistoricState(entities.AuthModel, data.DataKey(), data.MsgPack())
			}
			go dsquery.UpdateAccountCounter(string(event.Payload.Account))
			// if eventIsMoreRecent && err == nil {
			// 	// update state
			// 	if localState.ID != "" {
			// 		dataStates.AddCurrentState(entities.AuthModel, data.ID, data)

			// 		_, err = dsquery.CreateAuthorizationState(&data, &stateTxn)
			// 		if err != nil {
			// 			// TODO worker that will retry processing unSynced valid events with error
			// 			_, err = saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)
					
			// 		}
			// 	} else {
			// 		_, err = dsquery.CreateAuthorizationState(&data, &stateTxn)
			// 		if err != nil {
			// 			// TODO worker that will retry processing unSynced valid events with error
			// 			_, err = saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)
			// 		}
			// 	}
			// 	if err == nil {
			// 		err = stateTxn.Commit(context.Background())
			// 		if err == nil {
			// 			err = txn.Commit(context.Background())
			// 		} else {
			// 			logger.Infof("Error %v", err)
			// 		}

			// 	}
			// }
			// if err != nil {
			// 	logger.Debug("SaveEroror:::", err)
			// }

			// if err == nil {
			// 	go func() {
			// 		dsquery.IncrementStats(event, nil)
			// 		dsquery.UpdateAccountCounter(utils.IfThenElse(len(data.Account) > 0, data.Account.ToString(), string(data.DeviceKey)))
			// 		event.Subnet = event.Payload.Subnet
			// 		OnFinishProcessingEvent(ctx, event, &models.AuthorizationState{
			// 			Authorization: data,
			// 		})
			// 	}()
			// }

			// if string(event.Validator) != cfg.PublicKeyEDDHex {
			// 	go func() {
			// 		dependent, err := dsquery.GetDependentEvents(event)
			// 		if err != nil {
			// 			logger.Debug("Unable to get dependent events", err)
			// 		}
			// 		for _, dep := range *dependent {
			// 			logger.Debugf("Processing Dependend Event %s", dep.Hash)
			// 			go HandleNewPubSubEvent(dep, ctx)
			// 		}
			// 	}()
			// }

		}
	}
	return nil
}



// {"action":"AuthorizeAgent","network":"84532","identifier":"0xD466f0C2506b69e091b4356cd55b55f6DF00491b","hash":"kXTFkj7NkzQt5VQKvTcxXq6RHd5KvhO7eEDxtcsy+Ec="}
