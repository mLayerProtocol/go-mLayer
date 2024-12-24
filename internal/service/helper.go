package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
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


func SyncEventByPath(path *entities.EventPath, cfg *configs.MainConfiguration, validator string) (event *entities.Event, stateResp []byte, err error) {
	event, err = dsquery.GetEventById(path.ID, path.Model)
	if err != nil {
		if !dsquery.IsErrorNotFound(err) {
			return nil, nil, err
		}
		 event, resp, err := p2p.GetEvent(cfg, *path, (*entities.PublicKeyString)(&validator)) 
		if err != nil {
			return nil, nil, err
		}
		
		return event, resp.States[0], err
		
	}
	state, err := dsquery.GetStateFromEventPath(path)
	data, err := encoder.MsgPackStruct(state)
	return event, data, err
}


func SyncTypedStateById[M any](did string, model *M, cfg *configs.MainConfiguration, validator string) (event *entities.Event, err error) {
	var stateData []byte
	modelType :=  entities.GetModel(model)
	stateData, err = dsquery.GetStateById(did, modelType)
	if err != nil {
		if !dsquery.IsErrorNotFound(err) {
			return nil, err
		}
		d, event, err := SyncStateFromPeer(did, modelType,  cfg,  validator) 
		if err != nil {
			return nil, err
		}
		model = d.(*M)
		return event, err
		
	}

	err = encoder.MsgPackUnpackStruct(stateData, model)
	return event, err
}

func SyncStateFromPeer(id string , modelType  entities.EntityModel,  cfg *configs.MainConfiguration,  validator string ) (any, *entities.Event, error) {
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
	switch(modelType) {
	case entities.SubnetModel:
		newState := state.(entities.Authorization)
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
	logger.Infof("NEWSTATE %v", state, )
	if err != nil {
		return nil, nil, err
	}
	if len(pp.Event) < 2 {
		return nil, nil, apperror.NotFound(string(modelType) + " state not found")
	}
	event, err := entities.UnpackEvent(pp.Event, modelType)
	if err != nil {
		logger.Errorf("UnpackError: %v", err)
		return  nil,nil, err
	}
	err = dsquery.UpdateEvent(event, nil, true)
	if err != nil {
		return nil,nil, err
	}
	switch(modelType) {
	case entities.SubnetModel:
		newState := state.(entities.Subnet)
		newState.ID = id
		_, err = dsquery.CreateSubnetState(&newState, nil);
	case entities.AuthModel:
		newState := state.(entities.Authorization)
		newState.ID = id
		_, err = dsquery.CreateAuthorizationState(&newState, nil);
	case entities.TopicModel:
		newState := state.(entities.Topic)
		newState.ID = id
		logger.Infof("TopicState: %v", newState)
		_, err = dsquery.CreateTopicState(&newState, nil);
	case entities.SubscriptionModel:
		newState := state.(entities.Subscription)
		newState.ID = id
		_, err = dsquery.CreateSubscriptionState(&newState, nil);
	case entities.MessageModel:
		newState := state.(entities.Message)
		newState.ID = id
		_, err = dsquery.CreateMessageState(&newState, nil);
	default:
		



	}
	if err != nil {
		return nil, nil, err
	}
	// for _, data := range pp.States {
	// 	state, err := entities.UnpackSubnet(snetData)
	// 	logger.Infof("FoundSubnet %v", _subnet)
	// 	if err != nil {
	// 		return  nil, apperror.NotFound("unable to retrieve subnet")
	// 	}
	// 		s, err := dsquery.CreateSubnetState(&_subnet, nil)
	// 		logger.Infof("FoundSubnet 2 %v", _subnet)
	// 		if err != nil {
	// 			return  nil, apperror.NotFound("subnet not saved")
	// 		}
	// 		_subnet = *s;
		
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
	// logger.Debugf("Payload Validator: %s; Event Signer: %s; Validatos: %v", e.Payload.Validator, e.GetValidator(), chain.NetworkInfo.Validators)
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
		Subscriber: entities.DIDString(string(clientHandshake.Signer)),
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

func HandleNewPubSubEvent(event entities.Event, ctx *context.Context) error {
	go func () {
		channelpool.EventCounterChannel <- &event
	}()
	
	switch  event.Payload.Data.(type) {
	case entities.Subnet:
		return broadcastEvent(&event, ctx, HandleNewPubSubSubnetEvent(&event, ctx))
	case entities.Authorization:
		return broadcastEvent(&event, ctx, HandleNewPubSubAuthEvent(&event, ctx))
	case entities.Topic:
		return broadcastEvent(&event, ctx, HandleNewPubSubTopicEvent(&event, ctx))
	case entities.Subscription:
		return broadcastEvent(&event, ctx, HandleNewPubSubSubscriptionEvent(&event, ctx))
	case entities.Message:
		return broadcastEvent(&event, ctx, HandleNewPubSubMessageEvent(&event, ctx)) 
	}
	return nil
}

func broadcastEvent(event *entities.Event, ctx *context.Context, err error)  error {
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if err == nil && !event.Broadcasted && event.Validator == entities.PublicKeyString(cfg.PublicKeyEDDHex) {
		event.Broadcasted = true;
		go p2p.PublishEvent(*event)
	}
	return err
}

func OnFinishProcessingEvent(ctx *context.Context, event  *entities.Event, state  interface{}) {
	

	wsClientList, ok := (*ctx).Value(constants.WSClientLogId).(*entities.WsClientLog)
	if !ok {
		panic("Unable to connect to counter wsClients list")
	}
	config, ok := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		panic("Unable to retrieve config")
	}
	eventModelType := event.GetDataModelType()
	payload := entities.SocketSubscriptoinResponseData{
		Event: map[string]interface{}{
			"id": event.ID,
			"snet": event.Subnet,
			 "blk": event.BlockNumber,
			 "cy": event.Cycle,
			 "ep": event.Epoch,
			 "h": event.Hash,
			 "preE": event.PreviousEvent,
			 "authE": event.AuthEvent,
			 "modelType": eventModelType,
			 "t": event.EventType,
			 "pld": event.Payload,
		},
	}
	if eventModelType == entities.MessageModel {
		message := state.(*entities.Message)
		payload.Event["topic"] = message.Topic
		for _, subs := range wsClientList.GetClients(event.Subnet, message.Topic) {
			if subs != nil {
				payload.SubscriptionId = subs.Id
				subs.Conn.WriteJSON(payload)
			}
		}
	}
	for _, subs := range wsClientList.GetClients(event.Subnet, string(eventModelType)) {
		if subs != nil {
			payload.SubscriptionId = subs.Id
			subs.Conn.WriteJSON(payload)
		}
	}
	if string(event.Validator) != config.PublicKeyEDDHex {
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
	// event, err := query.GetEventFromPath(&eventPath)
	// eventCounterStore, ok := (*ctx).Value(constants.EventCountStore).(*ds.Datastore)

	// if !ok {
	// 	panic("Unable to connect to counter data store")
	// }
	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	// if err == nil || event != nil {

	// 	// increment count
	// 	currentSubnetCount := uint64(0);
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
	// 	subnetKey :=  datastore.NewKey(fmt.Sprintf("%s/%d/%s", event.Payload.Validator, cycle, utils.IfThenElse(event.Payload.Subnet == "", *subnetId, event.Payload.Subnet)))
	// 	cycleKey :=  datastore.NewKey(fmt.Sprintf("%s/%d", event.Payload.Validator, cycle))
	// 	val, err := eventCounterStore.Get(*ctx, subnetKey)

	// 	if err != nil  {
	// 		if err != datastore.ErrNotFound {
	// 			logger.Error(err)
	// 			return;
	// 		}
	// 	} else {
	// 		currentSubnetCount = encoder.NumberFromByte(val)
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
	// 	// 	subnetCycleClaimed := uint64(0);
	// 	// 	subnetClaimStatusKey :=  datastore.NewKey(fmt.Sprintf("%s/%d/%s/claimed", event.Payload.Validator, chain.GetCycle(event.BlockNumber), utils.IfThenElse(event.Payload.Subnet == "", *stateId, event.Payload.Subnet)))
	// 	// 	claimStatus, err := eventCounterStore.Get(*ctx, subnetClaimStatusKey)
	// 	// 	logger.Debugf("CURRENTCYCLECLAIM %d", claimStatus)
	// 	// 	if err != nil  {
	// 	// 		if err != badger.ErrKeyNotFound {
	// 	// 			logger.Error(err)
	// 	// 			return;
	// 	// 		} else {
	// 	// 			err = batch.Put(*ctx, subnetClaimStatusKey, encoder.NumberToByte(0))
	// 	// 			if err != nil {
	// 	// 				panic(err)
	// 	// 			}
	// 	// 		}
	// 	// 	} else {
	// 	// 		subnetCycleClaimed = encoder.NumberFromByte(claimStatus)
	// 	// 	}
	// 	// 	if subnetCycleClaimed == 0 {
	// 	// 		err = batch.Put(*ctx, subnetClaimStatusKey, encoder.NumberToByte(1))
	// 	// 		if err != nil {
	// 	// 			panic(err)
	// 	// 		}
	// 	// 	}
	// 	// }

	// 	err = batch.Put(*ctx, subnetKey, encoder.NumberToByte(1+currentSubnetCount))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	err = batch.Put(*ctx, cycleKey, encoder.NumberToByte(1+currentCycleCount))
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// err = eventCounterStore.Set(*ctx, subnetKey, encoder.NumberToByte(1+currentSubnetCount), true)
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
		// _, _, err := query.SaveRecord(models.SubnetEvent{
		// 	Event: entities.Event{
		// 		PayloadHash: event.PayloadHash,
		// 	},
		// }, models.SubnetEvent{
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
			// _, _, err := query.SaveRecord(models.SubnetEvent{
			// 	Event: entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},
			// },
			// 	models.SubnetEvent{
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
	//data.Event = *entities.NewEventPath(event.Validator, entities.SubnetModel, event.Hash)
	//state["event"] = *entities.NewEventPath(event.Validator, entities.SubnetModel, event.Hash)
	//data.Account = event.Payload.Account
	//state["account"] = *entities.NewEventPath(event.Validator, entities.SubnetModel, event.Hash)
	// logger.Error("data.Public ", data.Public)

	if updateState {
		// _, _, err := query.SaveRecord(models.SubnetState{
		// 	Subnet: entities.Subnet{ID: data.ID},
		// }, models.SubnetState{
		// 	Subnet: *data,
		// }, event.EventType == uint16(constants.UpdateSubnetEvent), tx)
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
			go HandleNewPubSubSubnetEvent(&dep, ctx)
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


	case constants.SubnetPayloadType:
		return query.SaveRecord(models.SubnetEvent{
			Event: where,
		}, models.SubnetEvent{
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
	data.Agent = entities.DIDString(agent)
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
