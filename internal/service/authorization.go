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

func ValidateAuthPayloadData(clientPayload *entities.ClientPayload, chainId configs.ChainId) (prevAuthState *models.AuthorizationState, grantorAuthState *models.AuthorizationState, subnet *models.SubnetState, err error) {
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
	_subnet, err := dsquery.GetSubnetStateById(auth.Subnet)
	if err != nil {
		if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
			return nil, nil, nil, apperror.NotFound("subnet not found")
		}
		return nil, nil, nil, apperror.Internal(err.Error())
	}
	subnet = &models.SubnetState{Subnet: *_subnet}

	if auth.Account != subnet.Account && *auth.Priviledge > *subnet.DefaultAuthPrivilege {
		return nil, nil, subnet, apperror.Internal("invalid auth priviledge. Cannot be higher than subnets default")
	}
	account := entities.AddressFromString(string(auth.Account))
	grantor := entities.AddressFromString(string(auth.Grantor))
	agent := entities.AddressFromString(string(auth.Agent))
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

	if err = VerifyAuthDataSignature(auth, msg, chainId); err != nil {
		return nil, nil, subnet, apperror.Unauthorized("Invalid authorization data signature")
	}
	if auth.Grantor != auth.Account {
		// grantorAuthState, err = query.GetOneAuthorizationState(entities.Authorization{Account: entities.DIDString(string(account.ToDeviceString())), Subnet: auth.Subnet, Agent: grantor.ToDeviceString()})
		_grantorAuthState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Account: entities.DIDString(string(account.ToDeviceString())), Subnet: auth.Subnet, Agent: grantor.ToDeviceString()}, dsquery.DefaultQueryLimit, nil)

		if err == gorm.ErrRecordNotFound || dsquery.IsErrorNotFound(err) {
			return nil, nil, subnet, apperror.Unauthorized("Grantor not authorized agent")
		}
		grantorAuthState = &models.AuthorizationState{Authorization: *_grantorAuthState[0]}
		if *grantorAuthState.Authorization.Priviledge != constants.AdminPriviledge {
			return nil, grantorAuthState, subnet, apperror.Forbidden(" Grantor does not have enough permission")
		}
	}
	// prevAuthState, err = query.GetOneAuthorizationState(entities.Authorization{Agent:  agent.ToDeviceString(), Subnet: auth.Subnet})
	_prevAuthState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Account: entities.DIDString(string(account.ToDeviceString())), Subnet: auth.Subnet, Agent: agent.ToDeviceString()}, dsquery.DefaultQueryLimit, nil)
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

	account := entities.AddressFromString(string(auth.Account))
	grantor := entities.AddressFromString(string(auth.Grantor))
	agent := entities.AddressFromString(string(auth.Agent))

	valid := false

	action := "write_authorization"
	switch auth.SignatureData.Type {
	case entities.EthereumPubKey:
		authMsg := fmt.Sprintf(constants.SignatureMessageString, action, chainId, agent.Addr, encoder.ToBase64Padded(msg))
		logger.Debug("MSG:: ", authMsg)

		msgByte := crypto.EthMessage([]byte(authMsg))
		signer := utils.IfThenElse(len(string(auth.Grantor)) == 0, account.Addr, grantor.Addr)
		// if len(agent.Addr) > 0 {
		// 	signer = agent.Addr
		// }
		// logger.Debug("Signer:: ", signer)
		valid = crypto.VerifySignatureECC(signer, &msgByte, auth.SignatureData.Signature)

		if !valid {
			// check if the grantor is authorized
			return apperror.Unauthorized("invalid auth signature")
			// check if agent is authorized by grantor
			// if agent.Addr != "" {
			// 	agentAuthState, err := query.GetOneAuthorizationState(entities.Authorization{Account: entities.DIDString(grantor.ToDeviceString()), Subnet: auth.Subnet, Agent: agent.ToDeviceString()})
			// 	if err == gorm.ErrRecordNotFound {
			// 		return nil, nil, subnet, apperror.Unauthorized("Agent not authorized to act on behalf of grantor")
			// 	}
			// 	if *agentAuthState.Authorization.Priviledge != constants.AdminPriviledge {
			// 		return nil, grantorAuthState,  subnet, apperror.Forbidden("Agent does not have enough permission")
			// 	}
			// }
		}

	case entities.TendermintsSecp256k1PubKey:

		decodedSig, err := base64.StdEncoding.DecodeString(auth.SignatureData.Signature)
		if err != nil {
			return err
		}
		publicKeyBytes, err := base64.RawStdEncoding.DecodeString(auth.SignatureData.PublicKey)

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
	data := event.Payload.Data.(entities.Authorization)
	data.BlockNumber = event.BlockNumber
	data.Cycle = event.Cycle
	data.Epoch = event.Epoch
	data.Event = *event.GetPath()
	hash, err := data.GetHash()
	data.EventSignature = event.Signature
	data.Agent = entities.AddressFromString(string(data.Agent)).ToDeviceString()
	if err != nil {
		return err
	}
	data.Hash = hex.EncodeToString(hash)
	var subnet = data.Subnet

	var localState *models.AuthorizationState
	// err := query.GetOne(&models.TopicState{Topic: entities.Topic{ID: id}}, &localTopicState)
	// err = sql.SqlDb.Where(&models.AuthorizationState{Authorization: entities.Authorization{Subnet: subnet, Agent: entities.AddressFromString(string(data.Agent)).ToDeviceString()}}).Take(&localState).Error
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
	_localState, err := dsquery.GetAccountAuthorizations(entities.Authorization{Subnet: subnet, Account: entities.AddressFromString(string(data.Account)).ToString(), Agent: entities.AddressFromString(string(data.Agent)).ToDeviceString()}, dsquery.DefaultQueryLimit, &stateTxn)
	if err != nil {
		logger.Error(err)
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
			Hash:      localState.Hash,
			Event:     &localState.Event,
			Timestamp: *localState.Timestamp,
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

	previousEventUptoDate, authEventUpToDate, _, eventIsMoreRecent, err := ProcessEvent(event, eventData, false, saveAuthorizationEvent, &txn, tx, ctx)
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
		_, _, _, err = ValidateAuthPayloadData(&event.Payload, cfg.ChainId)
		if err != nil {
			logger.Infof("ErrorValidatingAuth %v", err)
			saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.FalsePtr(), Synced: utils.TruePtr()}, &txn, nil)
		} else {
			// TODO if event is older than our state, just save it and mark it as synced

			savedEvent, err := saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)

			if eventIsMoreRecent && err == nil {
				// update state
				if localState.ID != "" {
					_, err = dsquery.CreateAuthorizationState(&data, &stateTxn)
					if err != nil {
						// TODO worker that will retry processing unSynced valid events with error
						_, err = saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)

					}
				} else {
					_, err = dsquery.CreateAuthorizationState(&data, &stateTxn)
					if err != nil {
						// TODO worker that will retry processing unSynced valid events with error
						_, err = saveAuthorizationEvent(entities.Event{ID: event.ID}, nil, &entities.Event{Error: err.Error(), IsValid: utils.TruePtr(), Synced: utils.TruePtr()}, &txn, nil)
					}
				}
				if err == nil {
					err = stateTxn.Commit(context.Background())
					if err == nil {
						err = txn.Commit(context.Background())
					} else {
						logger.Infof("Error %v", err)
					}

				}
			}
			if err != nil {
				logger.Debug("SaveEroror:::", err)
			}

			if err == nil {
				go func() {
					dsquery.IncrementStats(event, nil)
					dsquery.UpdateAccountCounter(utils.IfThenElse(len(data.Account) > 0, data.Account.ToString(), string(data.Agent)))
					event.Subnet = event.Payload.Subnet
					OnFinishProcessingEvent(ctx, event, &models.AuthorizationState{
						Authorization: data,
					}, &savedEvent.ID)
				}()
			}

			if string(event.Validator) != cfg.PublicKeyEDDHex {
				go func() {
					dependent, err := dsquery.GetDependentEvents(event)
					if err != nil {
						logger.Debug("Unable to get dependent events", err)
					}
					for _, dep := range *dependent {
						logger.Debugf("Processing Dependend Event %s", dep.Hash)
						go HandleNewPubSubEvent(dep, ctx)
					}
				}()
			}

		}
	}
	return nil
}

// {"action":"AuthorizeAgent","network":"84532","identifier":"0xD466f0C2506b69e091b4356cd55b55f6DF00491b","hash":"kXTFkj7NkzQt5VQKvTcxXq6RHd5KvhO7eEDxtcsy+Ec="}
