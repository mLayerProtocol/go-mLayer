package client

import (
	// "errors"
	"context"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
	"gorm.io/gorm"
)



func CreateEvent(payload entities.ClientPayload, ctx *context.Context) (model any, err error) {
	defer utils.TrackExecutionTime(time.Now(), "CreateEvent::")
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	stateDS, _ := (*ctx).Value(constants.ValidStateStore).(*ds.Datastore)
	if !strings.EqualFold(utils.AddressToHex(payload.Validator), utils.AddressToHex(cfg.OwnerAddress.String())) {
		return model, apperror.Forbidden(fmt.Sprintf("Validator (%s) not authorized to procces this request", cfg.OwnerAddress.String()))
	}

	
	var authState *models.AuthorizationState
	var agent *entities.DeviceString
	excludedEvents := []constants.EventType{constants.CreateSubnetEvent, constants.UpdateSubnetEvent, constants.DeleteSubnetEvent, constants.AuthorizationEvent}
	if !slices.Contains(excludedEvents, constants.EventType(payload.EventType)) {
		logger.Infof("ISNOTEXLUCDED: %d",  payload.EventType)
		authState, agent, err = ValidateClientPayload(stateDS, &payload, true, cfg)
		
		// logger.Debugf("New Event for Agent/Device2 %s", (*agent))
		if err != nil && err != gorm.ErrRecordNotFound && !dsquery.IsErrorNotFound(err)  {
			return model, err
		}
		if authState == nil || *authState.Authorization.Priviledge < constants.MemberPriviledge {
			// agent not authorized
			return model, apperror.Unauthorized("agent unauthorized to write in this subnet")
		}

		if *authState.Duration != 0 && uint64(time.Now().UnixMilli()) >
			(uint64(*authState.Timestamp)+uint64(*authState.Duration)) {
			return model, apperror.Unauthorized("Agent authorization expired")
		}
		payload.Agent = *agent
	}

	var assocPrevEvent *entities.EventPath
	var assocAuthEvent *entities.EventPath
	eventPayloadType := entities.GetModelTypeFromEventType(constants.EventType(payload.EventType))
	var subnetState = models.SubnetState{}
	logger.Infof("NewRequest: %v",  payload.EventType)
	if payload.Subnet != "" {
		// query.GetOneState(entities.Subnet{ID: payload.Subnet}, &subnetState)
		snet, _ := dsquery.GetSubnetStateById(payload.Subnet)
		if err != nil {
				logger.Errorf("CreateEvent/GetSubnetError: %v", err)
				return model, err
		}
		subnetState.Subnet = *snet
	}
	
	//Perfom checks base on event types
	logger.Debugf("authState****** 2: %v ", authState)
	
	switch payload.EventType {
	case uint16(constants.AuthorizationEvent):
		// authData := entities.Authorization{}
		// d, _ := json.Marshal(payload.Data)
		// e := json.Unmarshal(d, &authData)
		// if e != nil {
		// 	logger.Errorf("UnmarshalError %v", e)
		// }
		// payload.Data = authData
		// logger.Infof("NewRequest: %v",  "Authorization")
		assocPrevEvent, assocAuthEvent, err = ValidateAuthPayload(cfg, payload)
		logger.Infof("NewRequestProcessed: %v",  "Authorization")
		if err != nil {
			logger.Errorf("AuthDataVerificationError: %v", err)
			return model, err
		}
		
		
	case uint16(constants.CreateTopicEvent), uint16(constants.UpdateNameEvent), uint16(constants.UpdateTopicEvent), uint16(constants.LeaveEvent):
		
		// if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		assocPrevEvent, assocAuthEvent, err = ValidateTopicPayload(payload, authState)
		// return nil, fmt.Errorf("just debugiing")
		if err != nil {
			return model, err
		}
		if *authState.Authorization.Priviledge < constants.MemberPriviledge {
			return model, apperror.Forbidden("Agent not authorized to perform this action")
		}
		if assocPrevEvent == nil {
			assocPrevEvent = &subnetState.Event
		}
		// case uint16(constants.SubscribeTopicEvent):
		// 	if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 		return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// 	}
		// 	eventPayloadType = constants.SubscriptionPayloadType
		// 	assocPrevEvent, assocAuthEvent,  err = ValidateSubscriptionPayload(payload, authState)
		// 	if err != nil {
		// 		return nil, err
		// 	}
	case uint16(constants.CreateSubnetEvent), uint16(constants.UpdateSubnetEvent):
		
		// if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		logger.Infof("ValidatingSubnetPayload: %v", payload)
		assocPrevEvent, assocAuthEvent, err = ValidateSubnetPayload(payload, authState, ctx)
		if err != nil {
			logger.Errorf("InvalidSubnetPayload: %v", err)
			return model, err
		}
		logger.Infof("ValidSubnetPayload: %v", payload)
		
	case uint16(constants.CreateWalletEvent), uint16(constants.UpdateWalletEvent):
	
		// if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		assocPrevEvent, assocAuthEvent, err = ValidateWalletPayload(payload, authState)
		if err != nil {
			return model, err
		}
	case uint16(constants.SubscribeTopicEvent), uint16(constants.ApprovedEvent), uint16(constants.BanMemberEvent), uint16(constants.UnbanMemberEvent):
		if *authState.Authorization.Priviledge < constants.MemberPriviledge {
			return model, apperror.Forbidden("Agent not authorized to perform this action")
		}
		
		logger.Infof("ValidatingTopic...")
		assocPrevEvent, assocAuthEvent, err = ValidateSubscriptionPayload(payload, authState, cfg)
		
		if err != nil {
			logger.Debugf("SubscriptionError: %+v", err)
			return model, err
		}
	case uint16(constants.SendMessageEvent):
		logger.Debugf("authState 2: %d ", *authState.Authorization.Priviledge)
		// 1. Agent message
		// if *authState.Authorization.Priviledge < constants.MemberPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		
		assocPrevEvent, assocAuthEvent, err = ValidateMessagePayload(payload, authState)
		if err != nil {
			logger.Error("ERRRRRRR:::", err)
			return model, err
		}
	default:
	}
	// logger.Debugf("UPDATINGSUBNE1: %v", err)
	payloadHash, err := payload.GetHash()
	
	if err != nil {
		logger.Errorf("UPDATINGSUBNET2 %v", err)
	}
	logger.Debugf("UPDATINGSUBNE_HASH: %v",payloadHash)
	// chainInfo, err := chain.DefaultProvider(cfg).GetChainInfo()
	//chainInfo := chain.NetworkInfo
	// bNum, err := chain.DefaultProvider(cfg).GetCurrentBlockNumber()
	// cycle, err := chain.DefaultProvider(cfg).GetCycle(bNum)
	
	if err != nil {
		return nil, err
	}
	
	subnet := payload.Subnet
	if payload.EventType == uint16(constants.CreateSubnetEvent) {
		subnet, err = entities.GetId(payload.Data.(entities.Subnet), "")
		if err != nil {
			logger.Debugf("Subnet error: %v", err)
			return model, err
		}
		
	}
	if payload.EventType == uint16(constants.UpdateAvatarEvent) {
		subnet = payload.Data.(entities.Subnet).ID
	}
	if payload.EventType == uint16(constants.AuthorizationEvent) {
		subnet = payload.Data.(entities.Authorization).Subnet
	}
	
	event := entities.Event{
		Payload:           payload,
		Timestamp:         uint64(time.Now().UnixMilli()),
		EventType:         uint16(payload.EventType),
		Associations:      []string{},
		PreviousEvent: *utils.IfThenElse(assocPrevEvent == nil, entities.EventPathFromString(""), assocPrevEvent),
		AuthEvent:     *utils.IfThenElse(assocAuthEvent == nil, entities.EventPathFromString(""), assocAuthEvent),
		Synced:            utils.FalsePtr(),
		PayloadHash:       hex.EncodeToString(payloadHash),
		Broadcasted:       false,
		BlockNumber:       chain.NetworkInfo.CurrentBlock.Uint64(),
		Cycle: 				chain.NetworkInfo.CurrentCycle.Uint64(),
		Epoch: 				chain.NetworkInfo.CurrentEpoch.Uint64(),		
		Validator:         entities.PublicKeyString(cfg.PublicKeyEDDHex),
		Subnet: subnet,
	}
	
	 logger.Debugf("NewEvent: %v, %v", event.ID, eventPayloadType)

	b, err := event.EncodeBytes()
	if err != nil {
		return model, apperror.Internal(err.Error())
	}
	// logger.Debugf("Validator 2: %s", event.Validator)
	// logger.Debugf("eventPayloadType 2: %s", eventPayloadType)

	event.Hash = hex.EncodeToString(crypto.Sha256(b))
	_, event.Signature = crypto.SignEDD(b, cfg.PrivateKeyEDD)
	// err = dsquery.CreateEvent(&event, nil)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	logger.Infof("EventCreated...")
	event.ID, err = event.GetId()
	if err != nil {
		return model, err
	}
	service.HandleNewPubSubEvent(event, ctx) 
	event.Broadcasted = true;
	go p2p.PublishEvent(event)
	// dispatch to network

	// if err != nil {
	// 	return nil, err
	// }
	
	return event, nil
}


func GetEventTypeFromModel(eventType entities.EntityModel) constants.EventType {
	// cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// check if client payload is valid
	// if err := payload.Validate(entities.PublicKeyString(cfg.NetworkPublicKey)); err != nil {
	// 	return  err
	// }

	//Perfom checks base on event types
	switch eventType {
	case entities.AuthModel:

		return constants.AuthorizationEvent

	case entities.TopicModel:
		return constants.CreateTopicEvent

	case entities.SubscriptionModel:
		return constants.SubscribeTopicEvent

	case entities.MessageModel:
		return constants.SendMessageEvent

	case entities.SubnetModel:
		return constants.CreateSubnetEvent

	case entities.WalletModel:
		return constants.CreateWalletEvent
	}

	return 0

}



func GetEvent(eventId string, eventType int) (model interface{}, err error) {
	modelType := entities.GetModelTypeFromEventType(constants.EventType(eventType))
		event, err1 := dsquery.GetEventById(eventId, modelType)

		if err1 != nil {
			logger.Error(err)
			return nil, err1
		}
		return event, nil
}

func GetEventByPath(eventHash string, eventType int) (model interface{}, err error) {
	modelType := entities.GetModelTypeFromEventType(constants.EventType(eventType))
		event, err1 := dsquery.GetEventById(eventHash, modelType)

		if err1 != nil {
			logger.Error(err)
			return nil, err1
		}
		return event, nil

	// switch uint16(eventType) {
	// case uint16(constants.CreateTopicEvent), uint16(constants.UpdateNameEvent), uint16(constants.UpdateTopicEvent), uint16(constants.LeaveEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.TopicModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case uint16(constants.CreateSubnetEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.SubnetModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case uint16(constants.SubscribeTopicEvent), uint16(constants.ApprovedEvent), uint16(constants.BanMemberEvent), uint16(constants.UnbanMemberEvent):
	// 	event, err1 :=  dsquery.GetEventById(eventHash, entities.SubscriptionModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil
	// case uint16(constants.SendMessageEvent), uint16(constants.DeleteMessageEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.MessageModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case uint16(constants.AuthorizationEvent), uint16(constants.UnauthorizationEvent):
	// 	event, err1 :=  dsquery.GetEventById(eventHash, entities.AuthModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil
	// default:
	// }

	// return model, nil

}

// func GetTopicEventById(id string) (*models.TopicEvent, error) {
// 	nEvent := models.TopicEvent{}

// 	// err := query.GetOne(&models.TopicEvent{
// 	// 	Event: entities.Event{ID: id},
// 	// }, &nEvent)
// 	err := sql.SqlDb.Where(&models.TopicEvent{
// 			Event: entities.Event{ID: id},
// 		}).Take(&nEvent).Error
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil

// }

// func GetTopicEventByHash(hash string) (*models.TopicEvent, error) {
// 	nEvent := models.TopicEvent{}

// 	event, err := dsquery.GetEventFromPath(&entities.EventPath{EntityPath: entities.EntityPath{Hash: hash}})
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	nEvent.Event = *event
// 	return &nEvent, nil

// }

// func GetSubnetEventById(id string) (*models.SubnetEvent, error) {
// 	nEvent := models.SubnetEvent{}

// 	err := query.GetOne(models.SubnetEvent{
// 		Event: entities.Event{ID: id},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil

// }

// func GetSubnetEventByHash(hash string) (*models.SubnetEvent, error) {
// 	nEvent := models.SubnetEvent{}

// 	err := query.GetOne(models.SubnetEvent{
// 		Event: entities.Event{Hash: hash},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil

// }

// func GetSubEventById(id string) (*models.SubscriptionEvent, error) {
// 	nEvent := models.SubscriptionEvent{}

// 	err := query.GetOne(models.SubscriptionEvent{
// 		Event: entities.Event{ID: id},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil
// }

// func GetSubEventByHash(hash string) (*models.SubscriptionEvent, error) {
// 	nEvent := models.SubscriptionEvent{}

// 	err := query.GetOne(models.SubscriptionEvent{
// 		Event: entities.Event{Hash: hash},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil
// }

// func GetMessageEventById(id string) (*models.MessageEvent, error) {
// 	nEvent := models.MessageEvent{}

// 	event, err := dsquery.GetEventById(id, entities.MessageModel)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	nEvent.Event = *event
// 	return &nEvent, nil
// }

// func GetMessageEventByHash(hash string) (*models.MessageEvent, error) {
// 	event, err := dsquery.GetEventFromPath(&entities.EventPath{EntityPath: entities.EntityPath{Model: entities.MessageModel, Hash: hash}})
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return & models.MessageEvent{Event: *event}, nil
// }

// func GetAuthorizationEventById(id string) (*models.AuthorizationEvent, error) {
// 	nEvent := models.AuthorizationEvent{}

// 	err := query.GetOne(models.AuthorizationEvent{
// 		Event: entities.Event{ID: id},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil
// }

// func GetAuthorizationEventByHash(hash string) (*models.AuthorizationEvent, error) {
// 	nEvent := models.AuthorizationEvent{}

// 	err := query.GetOne(models.AuthorizationEvent{
// 		Event: entities.Event{Hash: hash},
// 	}, &nEvent)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &nEvent, nil
// }



func GetEventModelFromEventType(eventType constants.EventType)  any {
	if eventType < 1000 {
		return &models.AuthorizationEvent{}
	}
	if eventType < 1100 {
		return &models.TopicEvent{}
	}
	if eventType < 1200 {
		return &models.SubscriptionEvent{}
	}
	if eventType < 1300 {
		return &models.MessageEvent{}
	}
	if eventType < 1400 {
		return &models.SubnetEvent{}
	}
	if eventType < 1400 {
		return &models.WalletEvent{}
	}
	return nil
}