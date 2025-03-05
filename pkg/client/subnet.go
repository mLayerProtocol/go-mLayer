package client

import (
	// "errors"

	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
)

/// When a client connects, they are interested in 2 things
// 1. To create messages
// 2. To subscribe
// Therefore
// In case of one
// - check to see if node has interacted with subnet
// - if not, sync from bootstrap
// - if client sends a message to a topic not found locally, check bootstrap for topic
// - load all nodes interested in the topic from bootstrap
// - after processing message locally, broadcast to all nodes interested

// In case of 2
// publish your interest in the topic to all locally connected nodes. They will rebroadcast
// to their own locally connected nodes
// that node will receive any message sent to the topic


func GetSubscribedSubnets(item models.SubnetState) (state *[]models.SubnetState, err error) {

	// _, cacheError := dsquery.GetCacheKey(dsquery.AccountConnectedKey, string(item.Account))
	// if dsquery.IsErrorNotFound(cacheError) {
	// 	// never interacted with
	// 	defer func () {
	// 		if err == nil {
	// 			dsquery.SetCacheKey(dsquery.AccountConnectedKey,string(item.Account), []byte{} )
	// 		}
	// 	}()

	// 	// go the p2p route
	// 	payload := p2p.NewP2pPayload(config, p2p.P2pActionGetAccountSubscriptions, []byte(item.Account))
		
	// 	return
	// }

	 SubnetStates := []models.SubnetState{}
	_subnets, err := dsquery.GetAccountSubnets(item.Account, *dsquery.DefaultQueryLimit)
	if err != nil {
		return &SubnetStates, err
	}
	logger.Infof("AccountSubnets: %v", _subnets)
	for _, _sub := range _subnets {
		SubnetStates = append(SubnetStates, models.SubnetState{Subnet: *_sub})
	}
	subscriptionStates, err := dsquery.GetSubscriptions(entities.Subscription{Subscriber: entities.AddressString(item.Account)}, nil, nil)

	if err != nil {

		return &SubnetStates, err
	}
	var subnetIds = []string{}
	


	for _, sub := range subscriptionStates {
		subnetIds = append(subnetIds, sub.Subnet)
	}
	var subSubnetStates []models.SubnetState
	if len(subnetIds) > 0 {
		subSubnetErr := query.GetWithIN(models.SubnetState{}, &subSubnetStates, subnetIds)
		if subSubnetErr != nil {
			return &SubnetStates, err
		}
	}

	SubnetStates = append(SubnetStates, subSubnetStates...)

	return &SubnetStates, nil
}

// func GetSubnetEvents() (*[]models.SubnetEvent, error) {
// 	var SubnetEvents []models.SubnetEvent

// 	err := query.GetMany(models.SubnetEvent{
// 		Event: entities.Event{
// 			BlockNumber: 1,
// 		},
// 	}, &SubnetEvents, nil)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &SubnetEvents, nil
// }

// func ListenForNewSubnetEventFromPubSub (mainCtx *context.Context) {
// 	ctx, cancel := context.WithCancel(*mainCtx)
// 	defer cancel()

//		incomingSubnetC, ok := (*mainCtx).Value(constants.IncomingSubnetEventChId).(*chan *entities.Event)
//		if !ok {
//			logger.Errorf("incomingSubnetC closed")
//			return
//		}
//		for {
//			event, ok :=  <-*incomingSubnetC
//			if !ok {
//				logger.Fatal("incomingSubnetC closed for read")
//				return
//			}
//			go service.HandleNewPubSubSubnetEvent(event, ctx)
//		}
//	}
func ValidateSubnetPayload(payload entities.ClientPayload, authState *models.AuthorizationState, ctx *context.Context) (assocPrevEvent *entities.EventPath, assocAuthEvent *entities.EventPath, err error) {

	payloadData := entities.Subnet{}
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &payloadData)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
		return nil, nil, apperror.BadRequest(e.Error())
	}

	payload.Data = payloadData


	if uint64(payloadData.Timestamp) == 0 || uint64(payloadData.Timestamp) > uint64(time.Now().UnixMilli())+15000 || uint64(payloadData.Timestamp) < uint64(time.Now().UnixMilli())-15000 {
		return nil, nil, apperror.BadRequest("Invalid event timestamp")
	}
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	currentState, err2 := service.ValidateSubnetData(&payload, cfg.ChainId)
	logger.Infof("IVLAIDERR %v, %v", err2, currentState)
	if err2 != nil {
		return nil, nil, err2
	}
	if payload.EventType == constants.CreateSubnetEvent {
		// dont worry validating the AuthHash for Authorization requests
		// if entities.AddressFromString(payloadData.Owner.ToString()).Addr == "" {
		// 	return nil, nil, apperror.BadRequest("You must specify the owner of the subnet")
		// }
		if payloadData.ID != "" {
			return nil, nil, apperror.BadRequest("You cannot set an id when creating a subnet")
		}
		// var found []models.SubnetState
		// query.GetMany(&models.SubnetState{Subnet: entities.Subnet{Ref: payloadData.Ref}}, &found, nil)
		
		refExists, err := dsquery.RefExists(entities.SubnetModel, payloadData.Ref, "")
		
		if err != nil {
			return nil, nil, err
		}
		// if len(found) > 0 {
		// 	return nil, nil, apperror.BadRequest(fmt.Sprintf("Subnet with reference %s already exists", payloadData.Ref))
		// }
		if refExists {
			return nil, nil, apperror.BadRequest(fmt.Sprintf("Subnet with reference %s already exists", payloadData.Ref))
		}
		keySecP := "/ml/snetref/" + hex.EncodeToString(crypto.Keccak256Hash([]byte(payloadData.Ref)))
		v, err := p2p.GetDhtValue(keySecP)
		if err != nil {
			logger.Debugf("DHTERROR: %v", err)
			// return nil, nil, err
		}
		if  len(v) > 0 {
			return nil, nil, apperror.BadRequest(fmt.Sprintf("Subnet with reference %s already exists", payloadData.Ref))
		}
		// logger.Debug("FOUNDDDDD", found, payloadData.Ref)

	}
	if payload.EventType == constants.UpdateSubnetEvent {
		if payloadData.ID == "" {
			
			return nil, nil, apperror.BadRequest("Subnet ID must be provided")
		}
	}
	
	
	// generate associations
	if currentState != nil {
		//logger.Debugf("SUBNETINFO %v, %s, %s", strings.EqualFold(currentState.Account.ToString(), payloadData.Account.ToString()), currentState.Account.ToString(), payloadData.Account.ToString())
		if !strings.EqualFold(string(currentState.Account), string(payloadData.Account)) {
			return nil, nil, apperror.BadRequest("subnet account do not match")
		}
		assocPrevEvent = &currentState.Event
	}
	if authState != nil {
		assocAuthEvent = &authState.Event
	}
	
	return assocPrevEvent, assocAuthEvent, nil
}
