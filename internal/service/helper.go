package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/chain/ring"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/internal/system"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"github.com/mlayerprotocol/go-mlayer/smartlet"
)

var logger = &log.Logger
var eventCounterStore *ds.Datastore

func ConnectClient(cfg *configs.MainConfiguration, handshake *entities.ClientHandshake, protocol constants.Protocol) (*entities.ClientHandshake, error) {
	if utils.Abs(uint64(time.Now().UnixMilli()), uint64(handshake.Timestamp)) > uint64(15*time.Millisecond) {
		return nil, fmt.Errorf("handshake expired")
	}

	if !handshake.IsValid(cfg.ChainId) {
		return nil, fmt.Errorf("invalid handshake data")
	}

	handshake.Protocol = protocol
	// logger.Debug("VerifiedRequest.Message: ", verifiedRequest.Message)
	vByte, err := handshake.EncodeBytes()
	if err != nil {
		return nil, apperror.Internal("Invalid client handshake")
	}
	if crypto.VerifySignatureECC(string(handshake.Signer), &vByte, handshake.Signature) {
		// verifiedConn = append(verifiedConn, c)
		logger.Debug("Verification was successful: ", handshake)
		return handshake, nil
	}
	return nil, apperror.Forbidden("Invaliad handshake")

}

func GetEventByPath(path *entities.EventPath, cfg *configs.MainConfiguration, validator string) (event *entities.Event, stateResp []byte, local bool, err error) {
	event, err = dsquery.GetEventById(path.ID, path.Model)
	logger.Debugf("GetEventByPathError1 %s, %v, %v %v", path.ID, path.Model, path, err)
	if err == nil {
		state, err := dsquery.GetStateFromEventPath(path)
		if err == nil {
			data, _ := encoder.MsgPackStruct(state)
			return event, data, true, err
		} else {
			logger.Debugf("GetEventByPathError2 %v", err)
		}

	}

	if validator == "" {
		validator = string(path.Validator)
	}
	if err != nil {
		if !dsquery.IsErrorNotFound(err) {
			return nil, nil, false, err
		}
		event, resp, err := p2p.GetEvent(cfg, *path, (*entities.PublicKeyString)(&validator))
		if err != nil {
			return nil, nil, false, err
		}

		return event, resp.States[0], false, err

	}

	return event, nil, false, err
}

// func SyncTypedStateById(did string, modelType entities.EntityModel, cfg *configs.MainConfiguration, validator string) (event *entities.Event, state any, err error) {
// 	var stateData []byte
// 	stateData, err = dsquery.GetStateById(did, modelType)
// 	if err != nil {
// 		if !dsquery.IsErrorNotFound(err) {
// 			return nil, state, err
// 		}
// 		d, event, err := SyncStateFromPeer(did, modelType,  cfg,  validator)
// 		if err != nil {
// 			return nil, state, err
// 		}

// 		return event, d, err

// 	}

// 	err = encoder.MsgPackUnpackStruct(stateData, &state)
// 	return event, state, err
// }

func SyncTypedStateById[M any](did string, state M, cfg *configs.MainConfiguration, validator string) (event *entities.Event, err error) {
	var stateData []byte
	modelType := entities.GetModel(state)

	stateData, err = dsquery.GetStateById(did, modelType)

	if err != nil {
		if !dsquery.IsErrorNotFound(err) {
			return nil, err
		}
		d, event, err := SyncStateFromPeer(did, modelType, cfg, validator)
		if err != nil {
			return nil, err
		}
		temp, err := encoder.MsgPackStruct(d)
		encoder.MsgPackUnpackStruct(temp, state)
		return event, err

	}

	err = encoder.MsgPackUnpackStruct(stateData, state)
	return event, err
}

func SyncStateFromPeer(id string, modelType entities.EntityModel, cfg *configs.MainConfiguration, validator string) (any, *entities.Event, error) {
	state := entities.GetStateModelFromEntityType(modelType)

	if validator == "" {
		validator = chain.NetworkInfo.GetRandomSyncedNode()
	}
	if len(validator) == 0 {
		return nil, nil, apperror.NotFound(string(modelType) + " state not found")
	}
	logger.Infof("GettingTopic 1:::")
	subPath := entities.NewEntityPath(entities.PublicKeyString(validator), modelType, id)
	var pp *p2p.P2pEventResponse
	var err error
	switch modelType {
	case entities.ApplicationModel:
		newState := state.(entities.Application)
		pp, err = p2p.GetState(cfg, *subPath, nil, &newState)
		state = newState
	case entities.AuthModel:
		newState := state.(entities.Authorization)
		pp, err = p2p.GetState(cfg, *subPath, nil, &newState)
		state = newState
	case entities.TopicModel:
		logger.Infof("GettingTopic:::")
		newState := state.(entities.Topic)
		pp, err = p2p.GetState(cfg, *subPath, nil, &newState)
		state = newState
	case entities.SubscriptionModel:
		newState := state.(entities.Subscription)
		pp, err = p2p.GetState(cfg, *subPath, nil, &newState)
		state = newState
	case entities.MessageModel:
		newState := state.(entities.Message)
		pp, err = p2p.GetState(cfg, *subPath, nil, &newState)
		state = newState
	default:

	}
	logger.Infof("NEWSTATE %v", state)
	if err != nil {
		return nil, nil, err
	}
	if len(pp.Event) < 2 {
		return nil, nil, apperror.NotFound(string(modelType) + " state not found")
	}
	event, err := entities.UnpackEvent(pp.Event, modelType)
	if err != nil {
		logger.Errorf("UnpackError: %v", err)
		return nil, nil, err
	}
	err = dsquery.UpdateEvent(event, nil, nil, true)
	if err != nil {
		return nil, nil, err
	}
	switch modelType {
	case entities.ApplicationModel:
		newState := state.(entities.Application)
		newState.ID = id
		_, err = dsquery.CreateApplicationState(&newState, nil)
	case entities.AuthModel:
		newState := state.(entities.Authorization)
		newState.ID = id
		_, err = dsquery.CreateAuthorizationState(&newState, nil)
	case entities.TopicModel:
		newState := state.(entities.Topic)
		newState.ID = id
		logger.Infof("TopicState: %v", newState)
		_, err = dsquery.CreateTopicState(&newState, nil)
	case entities.SubscriptionModel:
		newState := state.(entities.Subscription)
		newState.ID = id
		_, err = dsquery.CreateSubscriptionState(&newState, nil)
	case entities.MessageModel:
		newState := state.(entities.Message)
		newState.ID = id
		_, err = dsquery.CreateMessageState(&newState, nil)
	default:

	}
	if err != nil {
		return nil, nil, err
	}
	// for _, data := range pp.States {
	// 	state, err := entities.UnpackApplication(snetData)
	// 	logger.Infof("FoundApplication %v", _app)
	// 	if err != nil {
	// 		return  nil, apperror.NotFound("unable to retrieve app")
	// 	}
	// 		s, err := dsquery.CreateApplicationState(&_app, nil)
	// 		logger.Infof("FoundApplication 2 %v", _app)
	// 		if err != nil {
	// 			return  nil, apperror.NotFound("app not saved")
	// 		}
	// 		_app = *s;

	// }
	return &state, event, nil

}

func ValidateEvent(event interface{}) error {
	e := event.(entities.Event)
	b, err := e.EncodeBytes()
	if err != nil {
		logger.Errorf("Invalid Encoding %v", err)
		return err
	}
	logger.Debugf("Payload Validator: %s; Event Signer: %s; Validatos: %v", e.Payload.Validator, e.GetValidator(), chain.NetworkInfo.Validators[fmt.Sprintf("edd/%s/addr", string(e.GetValidator()))])
	if !strings.EqualFold(utils.AddressToHex(chain.NetworkInfo.Validators[fmt.Sprintf("edd/%s/addr", string(e.GetValidator()))]), utils.AddressToHex(e.Payload.Validator)) {
		return apperror.Forbidden("payload validator does not match event validator")
	}

	sign, _ := hex.DecodeString(e.GetSignature())

	valid, err := crypto.VerifySignatureEDD(e.GetValidator().Bytes(), &b, sign)
	if err != nil {
		logger.Error("ValidateEvent: ", err)
		return err
	}
	if !valid {
		return apperror.Forbidden("Invalid node signature")
	}
	// TODO check to ensure that signer is an active validator, if not drop the event
	return nil
}

func ValidateMessageClient(
	ctx *context.Context,
	clientHandshake *entities.ClientHandshake,
) error {
	connectedSubscribers, ok := (*ctx).Value(constants.ConnectedSubscribersMap).(*map[string]map[string][]interface{})
	if !ok {
		return errors.New("could not connect to subscription datastore")
	}

	var subscriptionStates []models.SubscriptionState
	query.GetMany(models.SubscriptionState{Subscription: entities.Subscription{
		Subscriber: entities.AddressString(clientHandshake.Signer),
	}}, &subscriptionStates, nil)

	// VALIDATE AND DISTRIBUTE
	// logger.Debugf("Signer:  %s\n", clientHandshake.Signer)
	// results, err := channelSubscriberStore.Query(ctx, dsQuery.Query{
	// 	Prefix: "/" + clientHandshake.Signer,
	// })
	// if err != nil {
	// 	logger.Errorf("Channel Subscriber Store Query Error %o", err)
	// 	return
	// }
	// entries, _err := results.Rest()
	for i := 0; i < len(subscriptionStates); i++ {
		_sub := subscriptionStates[i]
		_topic := _sub.Subscription.Topic
		_subscriber := string(_sub.Subscriber)
		if (*connectedSubscribers)[_topic] == nil {
			(*connectedSubscribers)[_topic] = make(map[string][]interface{})
		}
		(*connectedSubscribers)[_topic][_subscriber] = append((*connectedSubscribers)[_topic][_subscriber], clientHandshake.ClientSocket)
	}
	logger.Debugf("results:  %v  \n", subscriptionStates[0])
	return nil
}

func HandleNewPubSubEvent(event entities.Event, ctx *context.Context) (*entities.EventProcessorResponse, error) {
	go func() {
		channelpool.EventCounterChannel <- &event
	}()
	switch event.Payload.Data.(type) {
	case entities.Application,entities.Authorization,entities.Topic, entities.Subscription,  entities.Message  :
		return processEvent(&event, ctx)
		//resp, err := HandleNewPubSubApplicationEvent(&event, ctx)
		// return nil, nil
		//	return resp, broadcastEvent(&event, ctx, err)
		// case entities.Authorization:
		// 	return nil, broadcastEvent(&event, ctx, HandleNewPubSubAuthEvent(&event, ctx))
		// case entities.Topic:
		// 	return nil, broadcastEvent(&event, ctx, HandleNewPubSubTopicEvent(&event, ctx))
		// case entities.Subscription:
		// 	return nil, broadcastEvent(&event, ctx, HandleNewPubSubSubscriptionEvent(&event, ctx))
		// case entities.Message:
		// 	return nil, broadcastEvent(&event, ctx, HandleNewPubSubMessageEvent(&event, ctx))
		case entities.SystemMessage:

			return nil, HandleNewNodeSystemMessageEvent(&event, ctx)
	}

	return nil, nil
}
func processEvent(event *entities.Event, ctx *context.Context) (localResponse *entities.EventProcessorResponse, err error) {
	var wg sync.WaitGroup
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	var minValidators = uint8(1)
	var same bool
	wg.Add(1)
	//var localResponse *entities.EventProcessorResponse
	remoteResponse := []*entities.EventDelta{}
	go func() {
		defer wg.Done()
		l, errh := HandleNewPubSubApplicationEvent(event, ctx)
		if errh == nil {
			localResponse = l
			logger.Debugf("LOCALHOSREPOS: %v", l)
		} else {
			logger.Errorf("NEWOSODIOSD: %v", err)
			err = errh
		}
	}()
	if event.Validator != entities.PublicKeyString(cfg.PublicKeyEDDHex) {
		wg.Wait()
		return
	}
	isTopicMessage := event.GetDataModelType() == entities.MessageModel 
	if isTopicMessage {
		// get the topic
		topic := entities.Topic{}
		_, err := SyncTypedStateById(event.Payload.Data.(entities.Message).Topic, &topic, cfg, "" )
		if err != nil {
			wg.Wait()
			return localResponse, err
		}
		minValidators = topic.MinimumValidators
		
	}
	if minValidators > 0 {
		wg.Add(1)
	}
		go func() {
			if minValidators > 0 {
				defer wg.Done()
			}
			responses, err := broadcastEvent(event, ctx, nil)
			// logger.Infof("REMOTERESPONSE %+v", responses)
			if err == nil {
				for _, resp := range responses {
					delta := entities.EventDelta{}
					err := encoder.MsgPackUnpackStruct(resp.Data, &delta)
					if err != nil {
						logger.Errorf("REMOTERESPONSEERRROR %d, %v", len(resp.Data), err)
						continue
					}
					remoteResponse = append(remoteResponse, &delta)
				}
			}
		}()
	
	
	wg.Wait()
	
	sameMap := map[string]bool{}
	if minValidators > 0 {
		for _, deltaResp := range remoteResponse {
			delta := map[string]interface{}{}
			err = encoder.MsgPackUnpackStruct(deltaResp.Delta, &delta)
			if err == nil && delta["h"] == localResponse.Hash {
				sameMap[string(deltaResp.Signatures[0].PublicKey)] = true
				same = true
			}
		}
	}
	// logger.Infof("REMOTERESPONSE %+v, %+v", remoteResponse, sameMap)
	if minValidators > 0  && len(sameMap) < int(minValidators) {
		return nil, fmt.Errorf("minimum commitment count not reached")
	}
	if minValidators == 0 || ((len(sameMap) > 0 && len(sameMap) == len(remoteResponse)))  {
		go func() {
			v, err := system.Mempool.GetData(event.ID)
			if err != nil {
				logger.Errorf("ErrorGettingDataStateFromDB: %v", same)
				return
			}
			dstate := dsquery.DataStates{}
			err = encoder.MsgPackUnpackStruct(*v, &dstate)
			if err != nil {
				logger.Errorf("ErrorUnpackingDataState: %v", err)
				return
			}
			err = dstate.Commit(nil, nil, nil, event.ID, err)
			if err != nil {

				logger.Errorf("ErrorSavingCommitment: %v", same)
			}
		}()

		// notify all validators that it is valid
		go func() {
			for k := range sameMap {
				logger.Infof("NOTIFYNODE %v", k)
				validPaylaod := p2p.NewP2pPayload(cfg, p2p.P2pActionNotifyValidEvent, event.GetPath().MsgPack())
				go p2p.SendSecureQuicRequestToValidator(cfg, string(k), validPaylaod)
			}
		}()
		// ALL SAME (sync and notify validators)
		
		go func() {
			sigData := remoteResponse[0]
			for i, r := range remoteResponse {
				if i == 0 {
					continue
				}
				sigData.Signatures = append(sigData.Signatures, r.Signatures...)
			}

			// FIND NODES INTERESTED IN THE APP AND SEND THEM THE DELTA
			app := event.Application

			dsquery.GetInterestedNodes(app, func(publicKey string) {
				p2p.SendSecureQuicRequestToValidator(cfg, publicKey, p2p.NewP2pPayload(cfg, p2p.P2pActionSyncState, sigData.MsgPack()))
			})
		}()
		// encoded, err := encoder.MsgPackStruct(sigData)
		// go p2p.PublishEvent(entities.Event{
		// 	Payload: entities.ClientPayload{
		// 		Data:  entities.SystemMessage{Data: encoded, Type: entities.SystemMessageType()}
		// 	},
		// 	EventType: constants.SystemMessage,
		// })
		return nil, err
	} else {
		// find who is not same and notify
		logger.Debugf("NotAllResponseIsSame") // TODO verify why not same and notify who needs to be
	}

	logger.Infof("SAMMMMMME: %v", same)
	return localResponse, err
	// broadcastEvent(&event, ctx, err)
}
func broadcastEvent(event *entities.Event, ctx *context.Context, err error) (responses []p2p.P2pPayload, er error) {
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !event.Broadcasted && event.Validator == entities.PublicKeyString(cfg.PublicKeyEDDHex) {
		event.Broadcasted = true
		// go p2p.PublishEvent(*event)
		node := ring.GlobalHashRing.GetNode(event.ID)
		logger.Infof("GettingNodes %v", node)
		eventPayload := entities.EventPayload{
			EventType: event.EventType,
			Event:     event.MsgPack(),
		}
		packed, err := encoder.MsgPackStruct(eventPayload)
		if err != nil {
			return nil, err
		}
		payload := p2p.NewP2pPayload(cfg, p2p.P2pActionPostEvent, packed)
		// payload.Sign(cfg.PrivateKeyEDD)
		responses = []p2p.P2pPayload{}
		var wg sync.WaitGroup
		var mu sync.Mutex
		 
		
		for _, vNode := range node.VirtualNodes {
			if vNode.Node.PubKey == cfg.PublicKeySECPHex {
				continue
			}
			logger.Infof("VIRTUALNODESSS %v", vNode.Node.PubKey)
			// d, err := p2p.GetNodeMultiAddressData(ctx, vNode.Node.PubKey)
			// if err != nil {
			// 	logger.Errorf("UnableToGetNodeAddress %v", err)
			// } else {
			//logger.Info(d.Hostname, d.IP, d.QuicPort)
			// 	// conn, err := p2p.NodeQuicPool.GetConnection(*ctx, address)
			//if uint8(i) < minCommitments {
				wg.Add(1)
			// }
			
			go func(vNode ring.VirtualNode) error {
				defer wg.Done()

				respData, err := p2p.SendSecureQuicRequestToValidator(cfg, vNode.Node.PubKey, payload)
				if err != nil {
					logger.Errorf("P2pQuicReqeuestError %v", err)
					return err
				}
				response := p2p.P2pPayload{}
				err = encoder.MsgPackUnpackStruct(respData, &response)
				if err != nil {
					logger.Errorf("P2pReqeuestResponseUnpackError %v", err)
					return err
				}
				if len(response.Error) != 0 {
					return fmt.Errorf("PostEventP2p2ResonseError: %s", response.Error)
				}
				mu.Lock()
				defer mu.Unlock()
				responses = append(responses, response)

				return nil
			}(*vNode)
			//}
			
		}
		wg.Wait()
			
		// go ring.EventDistribtor.HandleEvent(event, func(primaryNode *ring.Node, backupVNodes []*ring.Node) error {
		// 	// Get quic connection to node
		// 	// address , err := p2p.GetNodeQuicAddress(ctx, primaryNode.ID)
		// 	// if err != nil {
		// 	// 	// TODO store for future process
		// 	// 	return err
		// 	// }
		// 	// conn, err := p2p.NodeQuicPool.GetConnection(*ctx, address)
		// 	// if err != nil {
		// 	// 	// TODO access for redelaging primary node
		// 	// 	return err
		// 	// }
		// 	payload := p2p.NewP2pPayload(cfg, p2p.P2pActionPostEvent, event.MsgPack())
		// 	_,  err := p2p.SendSecureQuicRequestToValidator(cfg, primaryNode.ID, payload )
		// 	if err != nil {
		// 		logger.Errorf("P2pReqeuestError %v", err)
		// 		return err
		// 	}

		// 	// for _, n := range backupVNodes {

		// 	// }
		// 	return nil
		// })
	}
	return responses, err
}

func OnFinishProcessingEvent(cfg *configs.MainConfiguration, event *entities.Event, state interface{}, smartletResult *smartlet.Result) {

	go func(event *entities.Event) {
		if event.EventType == constants.SendMessageEvent {
			// distribute to interested validators validators
			post := true
			attempts := 5
			for post && attempts > 0 {
				attempts--
				message := state.(*entities.Message)
				payload := p2p.NewP2pPayload(cfg, p2p.P2pActionPostEvent, event.MsgPack())
				dsquery.GetInterestedNodes(message.Topic, func(publickey string) {
					resp, err := payload.SendDataRequest(publickey)
					if err != nil {
						return
					}
					if len(resp.Error) > 0 {
						return
					}
					post = false
				})

			}

		}
	}(event)
	go func(event *entities.Event) {
		wsClientList, ok := (*cfg.Context).Value(constants.WSClientLogId).(*entities.WsClientLog)
		if !ok {
			panic("Unable to connect to counter wsClients list")
		}
		config, ok := (*cfg.Context).Value(constants.ConfigKey).(*configs.MainConfiguration)
		if !ok {
			panic("Unable to retrieve config")
		}
		eventModelType := event.GetDataModelType()
		payload := entities.SocketSubscriptoinResponseData{
			Event: map[string]interface{}{
				"id":        event.ID,
				"app":      event.Application,
				"blk":       event.BlockNumber,
				"cy":        event.Cycle,
				"ep":        event.Epoch,
				"h":         event.Hash,
				"preE":      event.PreviousEvent,
				"authE":     event.AuthEvent,
				"modelType": eventModelType,
				"t":         event.EventType,
				"pld":       event.Payload,
			},
			Result: smartletResult,
		}
		// logger.Infof("REUSLTTTT: %v", payload.Result)
		if eventModelType == entities.MessageModel {
			message := state.(*entities.Message)
			payload.Event["topic"] = message.Topic
			for _, subs := range wsClientList.GetClients(event.Application, message.Topic) {
				if subs != nil {
					payload.SubscriptionId = subs.Id
					subs.Conn.WriteJSON(payload)
				}
			}
		}
		for _, subs := range wsClientList.GetClients(event.Application, string(eventModelType)) {
			if subs != nil {
				payload.SubscriptionId = subs.Id
				subs.Conn.WriteJSON(payload)
			}
		}
		for _, subs := range wsClientList.GetClientsV2(string(eventModelType), event.Payload.Data) {
			if subs != nil {
				payload.SubscriptionId = subs.Id
				subs.Conn.WriteJSON(payload)
			}
		}
		if string(event.Validator) != config.PublicKeyEDDHex {
			go func() {
				dependent, err := dsquery.GetDependentEvents(event)
				if err != nil {
					logger.Debug("Unable to get dependent events", err)
				}
				for _, dep := range *dependent {
					HandleNewPubSubEvent(dep, cfg.Context)
				}
			}()
		}
	}(event)
	// event, err := query.GetEventFromPath(&eventPath)
	// eventCounterStore, ok := (*ctx).Value(constants.EventCountStore).(*ds.Datastore)

	// if !ok {
	// 	panic("Unable to connect to counter data store")
	// }
	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	// if err == nil || event != nil {

	// 	// increment count
	// 	currentApplicationCount := uint64(0);
	// 	currentCycleCount := uint64(0);

	// 	batch, err :=	eventCounterStore.Batch(*ctx)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	// TODO consider storing cycle with event so we dont need another network call here
	// 	cycle, err :=  chain.DefaultProvider(cfg).GetCycle(big.NewInt(int64(event.BlockNumber)))
	// 	if err != nil {
	// 		// TODO
	// 		panic(err)
	// 	}
	// 	appKey :=  datastore.NewKey(fmt.Sprintf("%s/%d/%s", event.Payload.Validator, cycle, utils.IfThenElse(event.Payload.Application == "", *appId, event.Payload.Application)))
	// 	cycleKey :=  datastore.NewKey(fmt.Sprintf("%s/%d", event.Payload.Validator, cycle))
	// 	val, err := eventCounterStore.Get(*ctx, appKey)

	// 	if err != nil  {
	// 		if err != datastore.ErrNotFound {
	// 			logger.Error(err)
	// 			return;
	// 		}
	// 	} else {
	// 		currentApplicationCount = encoder.NumberFromByte(val)
	// 	}

	// 	cycleCount, err := eventCounterStore.Get(*ctx, cycleKey)

	// 	if err != nil  {
	// 		if err != datastore.ErrNotFound {
	// 			logger.Error(err)
	// 			return;
	// 		}
	// 	} else {
	// 		currentCycleCount = encoder.NumberFromByte(cycleCount)
	// 	}
	// 	logger.Debugf("CURRENTCYCLE %d, %d", cycleCount, currentCycleCount)
	// 	// if event.Payload.Validator == entities.PublicKeyString(cfg.NetworkPublicKey) {
	// 	// 	appCycleClaimed := uint64(0);
	// 	// 	appClaimStatusKey :=  datastore.NewKey(fmt.Sprintf("%s/%d/%s/claimed", event.Payload.Validator, chain.GetCycle(event.BlockNumber), utils.IfThenElse(event.Payload.Application == "", *stateId, event.Payload.Application)))
	// 	// 	claimStatus, err := eventCounterStore.Get(*ctx, appClaimStatusKey)
	// 	// 	logger.Debugf("CURRENTCYCLECLAIM %d", claimStatus)
	// 	// 	if err != nil  {
	// 	// 		if err != badger.ErrKeyNotFound {
	// 	// 			logger.Error(err)
	// 	// 			return;
	// 	// 		} else {
	// 	// 			err = batch.Put(*ctx, appClaimStatusKey, encoder.NumberToByte(0))
	// 	// 			if err != nil {
	// 	// 				panic(err)
	// 	// 			}
	// 	// 		}
	// 	// 	} else {
	// 	// 		appCycleClaimed = encoder.NumberFromByte(claimStatus)
	// 	// 	}
	// 	// 	if appCycleClaimed == 0 {
	// 	// 		err = batch.Put(*ctx, appClaimStatusKey, encoder.NumberToByte(1))
	// 	// 		if err != nil {
	// 	// 			panic(err)
	// 	// 		}
	// 	// 	}
	// 	// }

	// 	err = batch.Put(*ctx, appKey, encoder.NumberToByte(1+currentApplicationCount))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	err = batch.Put(*ctx, cycleKey, encoder.NumberToByte(1+currentCycleCount))
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// err = eventCounterStore.Set(*ctx, appKey, encoder.NumberToByte(1+currentApplicationCount), true)
	// 	err = batch.Commit(*ctx)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// } else {
	// 	logger.Error(err)
	// }

}

func IsMoreRecentEvent(
	oldEventHash string,
	oldEventTimestamp int,
	recentEventHash string,
	recentEventTimestamp int,
) bool {

	if oldEventTimestamp < recentEventTimestamp {
		return true
	}
	if oldEventTimestamp > recentEventTimestamp {
		return false
	}
	// // if the authorization was created at exactly the same time but their hash is different
	// // use the last 4 digits of their event hash
	// csN := new(big.Int)
	// csN.SetString(oldEventHash[50:], 16)
	// nsN := new(big.Int)
	// nsN.SetString(recentEventHash[50:], 16)
	if oldEventHash < recentEventHash {
		return true
	}
	if oldEventHash > recentEventHash {
		return false
	}
	return false
}

func IsMoreRecent(
	currenStatetEventId string,
	currenStatetHash string,
	currentStateEventTimestamp uint64,
	eventHash string,
	eventPayloadTimestamp uint64,
	markedAsSynced bool,
) (isMoreRecent bool, markAsSynced bool) {
	isMoreRecent = false
	markAsSynced = markedAsSynced
	if currentStateEventTimestamp < eventPayloadTimestamp {
		isMoreRecent = true
	}
	if currentStateEventTimestamp > eventPayloadTimestamp {
		isMoreRecent = false
	}
	// if the authorization was created at exactly the same time but their hash is different
	// use the last 4 digits of their event hash
	if currentStateEventTimestamp == eventPayloadTimestamp {
		// get the event payload of the current state

		// if err != nil && err != gorm.ErrRecordNotFound {
		// 	logger.Error("DB error", err)
		// }
		if currenStatetEventId == "" {
			markAsSynced = false
		} else {
			// if currentStateEventTimestamp < event.Payload.Timestamp {
			// 	isMoreRecent = true
			// }
			// if currentStateEvent.Payload.Timestamp == event.Payload.Timestamp {
			// logger.Debugf("Current state %v", currentStateEvent.Payload)
			// csN := new(big.Int)
			// csN.SetString(currenStatetHash[56:], 16)
			// nsN := new(big.Int)
			// nsN.SetString(eventHash[56:], 16)

			// if csN.Cmp(nsN) < 1 {
			// 	isMoreRecent = true
			// }
			//}
			if currenStatetHash < eventHash {
				isMoreRecent = true
			}
			if currenStatetHash > eventHash {
				isMoreRecent = false
			}
		}
	}
	return isMoreRecent, markAsSynced
}

// func ValidateAndAddToDeliveryProofToBlock(ctx context.Context,
// 	proof *entities.DeliveryProof,
// 	deliveryProofStore *ds.Datastore,
// 	channelSubscriberStore *ds.Datastore,
// 	stateStore *ds.Datastore,
// 	localBlockStore *ds.Datastore,
// 	MaxBlockSize int,
// 	mutex *sync.RWMutex,
// ) {

// 	err := deliveryProofStore.Set(ctx, db.Key(proof.Key()), proof.MsgPack(), true)
// 	if err == nil {
// 		// msg, err := validMessagesStore.Get(ctx, db.Key(fmt.Sprintf("/%s/%s", proof.MessageSender, proof.MessageHash)))
// 		// if err != nil {
// 		// 	// invalid proof or proof has been tampered with
// 		// 	return
// 		// }
// 		// get signer of proof
// 		b, err := proof.EncodeBytes()
// 		if err != nil {
// 			return
// 		}
// 		susbscriber, err := crypto.GetSignerECC(&b, &proof.Signature)
// 		if err != nil {
// 			// invalid proof or proof has been tampered with
// 			return
// 		}
// 		// check if the signer of the proof is a member of the channel
// 		isSubscriber, err := channelSubscriberStore.Has(ctx, db.Key("/"+susbscriber+"/"+proof.MessageHash))
// 		if isSubscriber {
// 			// proof is valid, so we should add to a new or existing batch
// 			var block *entities.Block
// 			var err error
// 			txn, err := stateStore.NewTransaction(ctx, false)
// 			if err != nil {
// 				logger.Errorf("State query errror %o", err)
// 				// invalid proof or proof has been tampered with
// 				return
// 			}
// 			blockData, err := txn.Get(ctx, db.Key(constants.CurrentDeliveryProofBlockStateKey))
// 			if err != nil {
// 				logger.Errorf("State query errror %o", err)
// 				// invalid proof or proof has been tampered with
// 				txn.Discard(ctx)
// 				return
// 			}
// 			if len(blockData) > 0 && block.Size < MaxBlockSize {
// 				block, err = entities.UnpackBlock(blockData)
// 				if err != nil {
// 					logger.Errorf("Invalid batch %o", err)
// 					// invalid proof or proof has been tampered with
// 					txn.Discard(ctx)
// 					return
// 				}
// 			} else {
// 				// generate a new batch
// 				block = entities.NewBlock()

// 			}
// 			block.Size += 1
// 			if block.Size >= MaxBlockSize {
// 				block.Closed = true
// 				block.NodeHeight = chain.API.GetCurrentBlockNumber()
// 			}
// 			// save the proof and the batch
// 			block.Hash = hexutil.Encode(crypto.Keccak256Hash([]byte(proof.Signature + block.Hash)))
// 			err = txn.Put(ctx, db.Key(constants.CurrentDeliveryProofBlockStateKey), block.MsgPack())
// 			if err != nil {
// 				logger.Errorf("Unable to update State store error %o", err)
// 				txn.Discard(ctx)
// 				return
// 			}
// 			proof.Block = block.BlockId
// 			proof.Index = block.Size
// 			err = deliveryProofStore.Put(ctx, db.Key(proof.Key()), proof.MsgPack())
// 			if err != nil {
// 				txn.Discard(ctx)
// 				logger.Errorf("Unable to save proof to store error %o", err)
// 				return
// 			}
// 			err = localBlockStore.Put(ctx, db.Key(constants.CurrentDeliveryProofBlockStateKey), block.MsgPack())
// 			if err != nil {
// 				logger.Errorf("Unable to save batch error %o", err)
// 				txn.Discard(ctx)
// 				return
// 			}
// 			err = txn.Commit(ctx)
// 			if err != nil {
// 				logger.Errorf("Unable to commit state update transaction errror %o", err)
// 				txn.Discard(ctx)
// 				return
// 			}
// 			// dispatch the proof and the batch
// 			if block.Closed {
// 				channelpool.OutgoingDeliveryProof_BlockC <- block
// 			}
// 			channelpool.OutgoingDeliveryProofC <- proof

// 		}

// 	}

// }

/*
type Model struct {
	Event entities.Event
}
func FinalizeEvent [ T entities.Payload, State any] (
	payloadType constants.EventPayloadType,
	event entities.Event,
	currentStateHash string,
	currentStateEventHash string,
	dataHash string,
	currentStateEvent *entities.Event,
	emptyState State,
	currentState  *State, finalState map[string]interface{},
) {
	markAsSynced := false
	updateState := false
	var eventError string
	// Confirm if this is an older event coming after a newer event.
	// If it is, then we only have to update our event history, else we need to also update our current state

	prevEventUpToDate := query.EventExist(&event.PreviousEvent) || (currentState == nil && event.PreviousEvent.ID == "") || (currentState != nil && currentStateEventHash == event.PreviousEvent.ID)
	// authEventUpToDate := query.EventExist(&event.AuthEvent) || (authState == nil && event.AuthEvent.ID == "") || (authState != nil && authState.Event == authEventHash)
	isMoreRecent := false
	if currentState != nil && currentStateHash != dataHash {
		err := query.GetOne(entities.Event{Hash: currentStateEventHash}, currentStateEvent)
		if uint64(currentStateEvent.Payload.Timestamp) < uint64(event.Payload.Timestamp) {
			isMoreRecent = true
		}
		if uint64(currentStateEvent.Payload.Timestamp) > uint64(event.Payload.Timestamp) {
			isMoreRecent = false
		}
		// if the authorization was created at exactly the same time but their hash is different
		// use the last 4 digits of their event hash
		if uint64(currentStateEvent.Payload.Timestamp) == uint64(event.Payload.Timestamp) {
			// get the event payload of the current state

			if err != nil && err != gorm.ErrRecordNotFound {
				logger.Error("DB error", err)
			}
			if currentStateEvent.ID == "" {
				markAsSynced = false
			} else {
				if currentStateEvent.Payload.Timestamp < event.Payload.Timestamp {
					isMoreRecent = true
				}
				if currentStateEvent.Payload.Timestamp == event.Payload.Timestamp {
					// logger.Debugf("Current state %v", currentStateEvent.Payload)
					csN := new(big.Int)
					csN.SetString(currentStateEventHash[56:], 16)
					nsN := new(big.Int)
					nsN.SetString(event.Hash[56:], 16)

					if csN.Cmp(nsN) < 1 {
						isMoreRecent = true
					}
				}
			}
		}
	}


	// If no error, then we should act accordingly as well
	// If are upto date, then we should update the state based on if its a recent or old event
	if len(eventError) == 0 {
		if prevEventUpToDate { // we are upto date
			if currentState == nil || isMoreRecent {
				updateState = true
				markAsSynced = true
			} else {
				// Its an old event
				markAsSynced = true
				updateState = false
			}
		} else {
			updateState = false
			markAsSynced = false
		}

	}

	// Save stuff permanently
	tx := sql.Db.Begin()
	logger.Debug(":::::updateState: Db Error", updateState, currentState == nil)

	// If the event was not signed by your node
	if string(event.Validator) != (*cfg).PublicKey  {
		// save the event
		event.Error = eventError
		event.IsValid = markAsSynced && len(eventError) == 0.
		event.Synced = markAsSynced
		event.Broadcasted = true
		// _, _, err := query.SaveRecord(models.ApplicationEvent{
		// 	Event: entities.Event{
		// 		PayloadHash: event.PayloadHash,
		// 	},
		// }, models.ApplicationEvent{
		// 	Event: event,
		// }, false, tx)
		_, _, err := saveEvent(payloadType, entities.Event{
					PayloadHash: event.PayloadHash,
				}, &event, false, tx)
		if err != nil {
			tx.Rollback()
			logger.Error("1000: Db Error", err)
			return
		}
	} else {
		if markAsSynced {
			// _, _, err := query.SaveRecord(Model{
			// 	Event: entities.Event{PayloadHash: event.PayloadHash},
			// }.(), Model{
			// 	Event: entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0},
			// }.(), true, tx)
			_, _, err := saveEvent(payloadType, entities.Event{
				PayloadHash: event.PayloadHash,
			},  &entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0}, false, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		} else {
			// mark as broadcasted
			// _, _, err := query.SaveRecord(models.ApplicationEvent{
			// 	Event: entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},
			// },
			// 	models.ApplicationEvent{
			// 		Event: entities.Event{Broadcasted: true},
			// 	}, true, tx)
				_, _, err := saveEvent(payloadType, entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},  &entities.Event{Broadcasted: true}, false, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		}
	}

	// d, err := event.Payload.EncodeBytes()
	// if err != nil {
	// 	logger.Errorf("Invalid event payload")
	// }
	// agent, err := crypto.GetSignerECC(&d, &event.Payload.Signature)
	// if err != nil {
	// 	logger.Errorf("Invalid event payload")
	// }
	//data.Event = *entities.NewEventPath(event.Validator, entities.ApplicationModel, event.Hash)
	//state["event"] = *entities.NewEventPath(event.Validator, entities.ApplicationModel, event.Hash)
	//data.Account = event.Payload.Account
	//state["account"] = *entities.NewEventPath(event.Validator, entities.ApplicationModel, event.Hash)
	// logger.Error("data.Public ", data.Public)

	if updateState {
		// _, _, err := query.SaveRecord(models.ApplicationState{
		// 	Application: entities.Application{ID: data.ID},
		// }, models.ApplicationState{
		// 	Application: *data,
		// }, event.EventType == uint16(constants.UpdateApplicationEvent), tx)
		// if err != nil {
		// 	tx.Rollback()
		// 	logger.Error("7000: Db Error", err)
		// 	return
		// }
		_, err := query.SaveRecordWithMap()
	}
	tx.Commit()

	if string(event.Validator) != (*cfg).PublicKey  {
		dependent, err := query.GetDependentEvents(*event)
		if err != nil {
			logger.Debug("Unable to get dependent events", err)
		}
		for _, dep := range *dependent {
			go HandleNewPubSubApplicationEvent(&dep, ctx)
		}
	}
}



func saveEvent(payloadType constants.EventPayloadType, where entities.Event, data *entities.Event,  update bool, tx *gorm.DB) (interface{}, bool, error) {
	switch (payloadType) {
	case constants.AuthorizationPayloadType:
		return query.SaveRecord(models.AuthorizationEvent{
			Event: where,
		}, models.AuthorizationEvent{
			Event: *data,
		}, update, tx)


	case constants.ApplicationPayloadType:
		return query.SaveRecord(models.ApplicationEvent{
			Event: where,
		}, models.ApplicationEvent{
			Event: *data,
		}, update, tx)
	}


}
*/
/*
func HandleNewPubSubEvent(event *entities.Event, ctx *context.Context, validator func(p entities.Payload)(*entities.Payload, error)) {
	logger.WithFields(logrus.Fields{"event": event}).Debug("New topic event from pubsub channel")
	markAsSynced := false
	updateState := false
	var eventError string
	// hash, _ := event.GetHash()


	logger.Debugf("Event is a valid event %s", event.PayloadHash)
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// Extract and validate the Data of the paylaod which is an Events Payload Data,
	data := event.Payload.Data.(entities.Payload)
	stateMap := map[string]interface{}{}
	logger.Debugf("NEWTOPICEVENT: %s", event.Hash)

	err := ValidateEvent(*event)

	if err != nil {
		logger.Error(err)
		return
	}
	dataHash, _ := data.GetHash()
	stateMap["hash"] = hex.EncodeToString(dataHash)
	authEventHash := event.AuthEvent
	authState, authError := query.GetOneAuthorizationState(entities.Authorization{Event: authEventHash})

	currentState, err := validator(data)
	if err != nil {
		// penalize node for broadcasting invalid data
		logger.Debugf("Invalid topic data %v. Node should be penalized", err)
		return
	}

	// check if we are upto date on this event

	prevEventUpToDate := query.EventExist(&event.PreviousEvent) || (currentState == nil && event.PreviousEvent.ID == "") || (currentState != nil && (*currentState).GetEvent().Hash == event.PreviousEvent.ID)
	authEventUpToDate := true
	if event.AuthEvent.ID != "" {
		authEventUpToDate = query.EventExist(&event.AuthEvent) || (authState == nil && event.AuthEvent.ID == "") || (authState != nil && authState.Event == authEventHash)
	}

	// Confirm if this is an older event coming after a newer event.
	// If it is, then we only have to update our event history, else we need to also update our current state
	isMoreRecent := false
	currentStateHash, _ := (*currentState).GetHash()
	if currentState != nil && hex.EncodeToString(currentStateHash) != stateMap["hash"] {
		var currentStateEvent = &models.Event{}
		err := query.GetOne(entities.Event{Hash: (*currentState).GetEvent().Hash}, currentStateEvent)
		if uint64(currentStateEvent.Payload.Timestamp) < uint64(event.Payload.Timestamp) {
			isMoreRecent = true
		}
		if uint64(currentStateEvent.Payload.Timestamp) > uint64(event.Payload.Timestamp) {
			isMoreRecent = false
		}
		// if the authorization was created at exactly the same time but their hash is different
		// use the last 4 digits of their event hash
		if uint64(currentStateEvent.Payload.Timestamp) == uint64(event.Payload.Timestamp) {
			// get the event payload of the current state

			if err != nil && err != gorm.ErrRecordNotFound {
				logger.Error("DB error", err)
			}
			if currentStateEvent.ID == "" {
				markAsSynced = false
			} else {
				if currentStateEvent.Payload.Timestamp < event.Payload.Timestamp {
					isMoreRecent = true
				}
				if currentStateEvent.Payload.Timestamp == event.Payload.Timestamp {
					// logger.Debugf("Current state %v", currentStateEvent.Payload)
					csN := new(big.Int)
					csN.SetString(currentState.Event.ID[56:], 16)
					nsN := new(big.Int)
					nsN.SetString(event.Hash[56:], 16)

					if csN.Cmp(nsN) < 1 {
						isMoreRecent = true
					}
				}
			}
		}
	}

	if authError != nil {
		// check if we are upto date. If we are, then the error is an actual one
		// the error should be attached when saving the event
		// But if we are not upto date, then we might need to wait for more info from the network

		if prevEventUpToDate && authEventUpToDate {
			// we are upto date. This is an actual error. No need to expect an update from the network
			eventError = authError.Error()
			markAsSynced = true
		} else {
			if currentState == nil || (currentState != nil && isMoreRecent) { // it is a morer ecent event
				if strings.HasPrefix(authError.Error(), constants.ErrorForbidden) || strings.HasPrefix(authError.Error(), constants.ErrorUnauthorized) {
					markAsSynced = false
				} else {
					// entire event can be considered bad since the payload data is bad
					// this should have been sorted out before broadcasting to the network
					// TODO penalize the node that broadcasted this
					eventError = authError.Error()
					markAsSynced = true
				}

			} else {
				// we are upto date. We just need to store this event as well.
				// No need to update state
				markAsSynced = true
				eventError = authError.Error()
			}
		}

	}

	// If no error, then we should act accordingly as well
	// If are upto date, then we should update the state based on if its a recent or old event
	if len(eventError) == 0 {
		if prevEventUpToDate && authEventUpToDate { // we are upto date
			if currentState == nil || isMoreRecent {
				updateState = true
				markAsSynced = true
			} else {
				// Its an old event
				markAsSynced = true
				updateState = false
			}
		} else {
			updateState = false
			markAsSynced = false
		}

	}

	// Save stuff permanently
	tx := sql.Db.Begin()
	logger.Debug(":::::updateState: Db Error", updateState, currentState == nil)

	// If the event was not signed by your node
	if string(event.Validator) != (*cfg).PublicKey  {
		// save the event
		event.Error = eventError
		event.IsValid = markAsSynced && len(eventError) == 0.
		event.Synced = markAsSynced
		event.Broadcasted = true
		_, _, err := query.SaveRecord(models.TopicEvent{
			Event: entities.Event{
				PayloadHash: event.PayloadHash,
			},
		}, models.TopicEvent{
			Event: *event,
		}, false, tx)
		if err != nil {
			tx.Rollback()
			logger.Error("1000: Db Error", err)
			return
		}
	} else {
		if markAsSynced {
			_, _, err := query.SaveRecord(models.TopicEvent{
				Event: entities.Event{PayloadHash: event.PayloadHash},
			}, models.TopicEvent{
				Event: entities.Event{Synced: true, Broadcasted: true, Error: eventError, IsValid: len(eventError) == 0},
			}, true, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		} else {
			// mark as broadcasted
			_, _, err := query.SaveRecord(models.TopicEvent{
				Event: entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},
			},
				models.TopicEvent{
					Event: entities.Event{Broadcasted: true},
				}, true, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		}
	}

	d, err := event.Payload.EncodeBytes()
	if err != nil {
		logger.Errorf("Invalid event payload")
	}
	agent, err := crypto.GetSignerECC(&d, &event.Payload.Signature)
	if err != nil {
		logger.Errorf("Invalid event payload")
	}
	data.Event = *entities.NewEventPath(event.Validator, entities.TopicModel, event.Hash)
	data.AppKey = entities.AccountString(agent)
	data.Account = event.Payload.Account
	// logger.Error("data.Public ", data.Public)

	if updateState {
		_, _, err := query.SaveRecord(models.TopicState{
			Topic: entities.Topic{ID: data.ID},
		}, models.TopicState{
			Topic: *data,
		}, event.EventType == uint16(constants.UpdateTopicEvent), tx)
		if err != nil {
			tx.Rollback()
			logger.Error("7000: Db Error", err)
			return
		}
	}
	tx.Commit()

	if string(event.Validator) != (*cfg).PublicKey  {
		dependent, err := query.GetDependentEvents(*event)
		if err != nil {
			logger.Debug("Unable to get dependent events", err)
		}
		for _, dep := range *dependent {
			go HandleNewPubSubTopicEvent(&dep, ctx)
		}
	}

	// TODO Broadcast the updated state
}
*/
