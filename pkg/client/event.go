package client

import (
	// "errors"
	"context"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)


type PaddedInt64 struct {
	// once  sync.Once
	value int64
	_     [56]byte // 60 bytes padding + 4 bytes int32 = 64 bytes
}
var messageVectors  sync.Map

func CreateEvent(payload entities.ClientPayload, ctx *context.Context) (model any, err error) {
	defer utils.TrackExecutionTime(time.Now(), "CreateEvent::")
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !strings.EqualFold(utils.AddressToHex(payload.Validator), utils.AddressToHex(cfg.OwnerAddress.String())) {
		return model, apperror.Forbidden(fmt.Sprintf("Validator (%s) not authorized to procces this request", cfg.OwnerAddress.String()))
	}
	stateEvent := []entities.StateEvents{}
	
	var authState *models.AuthorizationState
	var agent *entities.DeviceString
	logger.Infof("SUBSCRIPTION_PAYLOAD2 %+v", payload)
	excludedEvents := []constants.EventType{constants.CreateApplicationEvent, constants.UpdateApplicationEvent, constants.DeleteApplicationEvent, constants.AuthorizationEvent}
	if !slices.Contains(excludedEvents, constants.EventType(payload.EventType)) {
		logger.Infof("ISNOTEXLUCDED: %d",  payload.EventType)
		authState, agent, err = ValidateClientPayload( &payload, true, cfg)
		
		// logger.Debugf("New Event for Agent/Device2 %s", (*agent))
		if err != nil && err != gorm.ErrRecordNotFound && !dsquery.IsErrorNotFound(err)  {
			return model, err
		}
		if authState == nil || *authState.Authorization.Priviledge < constants.MemberPriviledge {
			// agent not authorized
			return model, apperror.Unauthorized("agent unauthorized to write in this app")
		}

		if *authState.Duration != 0 && uint64(time.Now().UnixMilli()) >
			(uint64(*authState.Timestamp)+uint64(*authState.Duration)) {
			return model, apperror.Unauthorized("Agent authorization expired")
		}
		payload.AppKey = *agent
	}

	var assocPrevEvent *entities.EventPath
	var assocAuthEvent *entities.EventPath
	eventPayloadType := entities.GetModelTypeFromEventType(constants.EventType(payload.EventType))
	var appState = models.ApplicationState{}
	logger.Infof("NewRequest: %v",  payload.EventType)
	if payload.Application != "" {
		// query.GetOneState(entities.Application{ID: payload.Application}, &appState)
		app, _ := dsquery.GetApplicationStateById(payload.Application)
		if err != nil {
				logger.Errorf("CreateEvent/GetApplicationError: %v", err)
				return model, err
		}
		appState.Application = *app
		stateEvent = append(stateEvent, entities.StateEvents{ID: app.ID, Event: app.Event})
	}
	
	//Perfom checks base on event types
	logger.Debugf("authState****** 2: %v ", authState)
	
	switch payload.EventType {
	case constants.AuthorizationEvent:
		// authData := entities.Authorization{}
		// d, _ := json.Marshal(payload.Data)
		// e := json.Unmarshal(d, &authData)
		// if e != nil {
		// 	logger.Errorf("UnmarshalError %v", e)
		// }
		// payload.Data = authData
		// logger.Infof("NewRequest: %v",  "Authorization")
		appState := &entities.Application{}
		assocPrevEvent, assocAuthEvent, appState, err = ValidateAuthPayload(cfg, payload)
		if err != nil {
			return nil, err
		}

		stateEvent = append(stateEvent, entities.StateEvents{ID: appState.ID, Event: appState.Event})
		
		if err != nil {
			logger.Errorf("AuthDataVerificationError: %v", err)
			return model, err
		}
		
		
		
	case constants.CreateTopicEvent, constants.UpdateNameEvent, constants.UpdateTopicEvent, constants.LeaveEvent:
		
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
			assocPrevEvent = &appState.Event
		}
		// case constants.SubscribeTopicEvent):
		// 	if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 		return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// 	}
		// 	eventPayloadType = constants.SubscriptionPayloadType
		// 	assocPrevEvent, assocAuthEvent,  err = ValidateSubscriptionPayload(payload, authState)
		// 	if err != nil {
		// 		return nil, err
		// 	}
	case constants.CreateApplicationEvent, constants.UpdateApplicationEvent:
		
		// if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		logger.Infof("ValidatingApplicationPayload: %v", payload)
		assocPrevEvent, assocAuthEvent, err = ValidateApplicationPayload(payload, authState, ctx)
		if err != nil {
			logger.Errorf("InvalidApplicationPayload: %v", err)
			return model, err
		}
		logger.Infof("ValidApplicationPayload: %v", payload)
		
	case constants.CreateWalletEvent,constants.UpdateWalletEvent:
	
		// if authState.Authorization.Priviledge < constants.AdminPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		assocPrevEvent, assocAuthEvent, err = ValidateWalletPayload(payload, authState)
		if err != nil {
			return model, err
		}
	case constants.SubscribeTopicEvent, constants.ApprovedEvent, constants.BanMemberEvent, constants.UnbanMemberEvent:
		if *authState.Authorization.Priviledge < constants.MemberPriviledge {
			return model, apperror.Forbidden("Agent not authorized to perform this action")
		}
		
		logger.Infof("ValidatingTopic...")
		_topic := &entities.Topic{}
		assocPrevEvent, assocAuthEvent, _topic, err = ValidateSubscriptionPayload(payload, authState, cfg)
		
		if err != nil {
			logger.Debugf("SubscriptionError: %+v", err)
			return model, err
		}
		stateEvent = append(stateEvent, entities.StateEvents{ID: _topic.ID, Event: _topic.Event})
	case constants.SendMessageEvent:
		logger.Debugf("authState 2: %d ", *authState.Authorization.Priviledge)
		// 1. Agent message
		// if *authState.Authorization.Priviledge < constants.MemberPriviledge {
		// 	return nil, apperror.Forbidden("Agent not authorized to perform this action")
		// }
		topic := &entities.Topic{}
		_, err := service.SyncTypedStateById(payload.Data.(entities.Message).Topic,  topic, cfg, "")
		if err != nil {
			return model, err
		}
		stateEvent = append(stateEvent, entities.StateEvents{Event: topic.Event, ID: topic.ID})
		subscription := &entities.Subscription{}
		assocPrevEvent, assocAuthEvent, subscription, err = ValidateMessagePayload(payload, authState, topic)
		stateEvent = append(stateEvent, entities.StateEvents{Event: subscription.Event, ID: subscription.ID})
		if err != nil {
			logger.Error("ERRRRRRR:::", err)
			return model, err
		}
	default:
	}
	// logger.Debugf("UPDATINGSUBNE1: %v", err)
	if authState != nil {
		stateEvent = append(stateEvent, entities.StateEvents{Event: authState.Event, ID: authState.ID})
	}
	payloadHash, err := payload.GetHash()
	
	if err != nil {
		logger.Errorf("UPDATINGAPP2 %v", err)
	}
	logger.Debugf("UPDATINGSUBNE_HASH: %v",payloadHash)
	// chainInfo, err := chain.DefaultProvider(cfg).GetChainInfo()
	//chainInfo := chain.NetworkInfo
	// bNum, err := chain.DefaultProvider(cfg).GetCurrentBlockNumber()
	// cycle, err := chain.DefaultProvider(cfg).GetCycle(bNum)
	
	if err != nil {
		return nil, err
	}
	
	app := payload.Application
	if payload.EventType == constants.CreateApplicationEvent {
		app, err = entities.GetId(payload.Data.(entities.Application), "")
		if err != nil {
			logger.Debugf("Application error: %v", err)
			return model, err
		}
		
	}
	if payload.EventType == constants.UpdateAvatarEvent {
		app = payload.Data.(entities.Application).ID
	}
	if payload.EventType == constants.AuthorizationEvent {
		app = payload.Data.(entities.Authorization).Application
	}
	
	event := entities.Event{
		Payload:           payload,
		Timestamp:         uint64(time.Now().UnixMilli()),
		EventType:        payload.EventType,
		// Associations:      []string{},
		PreviousEvent: *utils.IfThenElse(assocPrevEvent == nil, entities.EventPathFromString(""), assocPrevEvent),
		AuthEvent:     *utils.IfThenElse(assocAuthEvent == nil, entities.EventPathFromString(""), assocAuthEvent),
		Synced:            utils.FalsePtr(),
		PayloadHash:       hex.EncodeToString(payloadHash),
		Broadcasted:       false,
		BlockNumber:       chain.NetworkInfo.CurrentBlock.Uint64(),
		Cycle: 				chain.NetworkInfo.CurrentCycle.Uint64(),
		Epoch: 				chain.NetworkInfo.CurrentEpoch.Uint64(),		
		Validator:         entities.PublicKeyString(cfg.PublicKeyEDDHex),
		Application: app,
		StateEvents:  stateEvent,
	}
	
	if constants.SendMessageEvent == event.EventType {
		vecKey := event.VectorKey(event.Payload.Data.(entities.Message).Topic)
		vectorInterface, loaded := messageVectors.LoadOrStore(vecKey, &PaddedInt64{})
		vector := vectorInterface.(*PaddedInt64)
		// var  _err error
		if loaded && vector.value == 0 {
			// 
			time.Sleep(5 * time.Millisecond)
			vectorInterface, loaded = messageVectors.LoadOrStore(vecKey, &PaddedInt64{})
			vector = vectorInterface.(*PaddedInt64)
		}
		if !loaded {
			// load it from your local db
			// vector.once.Do(func() {
				d, err := stores.NetworkStatsStore.Get(context.Background(), datastore.NewKey(vecKey))
				if err != nil && !dsquery.IsErrorNotFound(err) {
					return nil, err
				}
				i, err := strconv.Atoi(string(d))
				if err != nil {
					i = 0
				}
				// messageVectors[vecKey] = &v
				atomic.StoreInt64(&vector.value, int64(i))
			// })
		}
		// if _err != nil {
		// 	return nil, _err
		// }
		logger.Infof("Vector")
		event.Index = atomic.AddInt64(&vector.value, 1)
		
		// err := stores.NetworkStatsStore.Set(context.Background(), datastore.NewKey(vecKey), []byte(fmt.Sprint(vector.value)), true)
		// if err != nil {
		// 	return nil, err
		// }
		// defer func ()  {
		// 	if err == nil {
		// 		stores.EventStore.Get(context.Background(), datastore.NewKey(vecKey))
		// 	}
		// } ()
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

	case entities.ApplicationModel:
		return constants.CreateApplicationEvent

	case entities.WalletModel:
		return constants.CreateWalletEvent
	case entities.SystemModel:
		return constants.SystemMessage
	}

	return 0

}

func PublishInterest(ids []string, _type entities.EntityModel, cfg *configs.MainConfiguration) error {
	logger.Debugf("Pulbishing interest in topics... %+v", ids)
	event := entities.Event{
		Payload:           entities.ClientPayload{
			Data: entities.SystemMessage{
				Type: entities.AnnounceTopicInterest,
				Data: (&entities.NodeInterest{
					Ids: ids,
					Type: _type,
					Expiry: int(time.Now().Add(constants.TOPIC_INTEREST_TTL).UnixMilli()),
				}).MsgPack(),
			},
		},
		Timestamp:         uint64(time.Now().UnixMilli()),
		Validator:         entities.PublicKeyString(cfg.PublicKeyEDDHex),
		EventType: constants.SystemMessage,
	}
	b, err := event.EncodeBytes()
	if err != nil {
		return apperror.Internal(err.Error())
	}
	event.Hash = hex.EncodeToString(crypto.Sha256(b))
	_, event.Signature = crypto.SignEDD(b, cfg.PrivateKeyEDD)
	event.ID, err = event.GetId()
	if err != nil {
		return err
	}
	_, err = service.HandleNewPubSubEvent(event, cfg.Context) 
	return err
}



func GetEvent(eventId string, eventType int) (model interface{}, err error) {
	modelType := entities.GetModelTypeFromEventType(constants.EventType(eventType))
		event, err1 := dsquery.GetEventById(eventId, modelType)

		if err1 != nil {
			logger.Error("GetEvent: ", err1, " ", eventId)
			return nil, err1
		}
		return event, nil
}

func GetEventByPath(eventHash string, eventType int) (model interface{}, err error) {
	modelType := entities.GetModelTypeFromEventType(constants.EventType(eventType))
		event, err1 := dsquery.GetEventById(eventHash, modelType)

		if err1 != nil {
			logger.Error("GetEventByPath: ", err)
			return nil, err1
		}
		return event, nil

	// switch uint16(eventType) {
	// case constants.CreateTopicEvent), constants.UpdateNameEvent), constants.UpdateTopicEvent), constants.LeaveEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.TopicModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case constants.CreateApplicationEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.ApplicationModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case constants.SubscribeTopicEvent), constants.ApprovedEvent), constants.BanMemberEvent), constants.UnbanMemberEvent):
	// 	event, err1 :=  dsquery.GetEventById(eventHash, entities.SubscriptionModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil
	// case constants.SendMessageEvent), constants.DeleteMessageEvent):
	// 	event, err1 := dsquery.GetEventById(eventHash, entities.MessageModel)

	// 	if err1 != nil {
	// 		logger.Error(err)
	// 		return nil, err1
	// 	}
	// 	return event, nil

	// case constants.AuthorizationEvent), constants.UnauthorizationEvent):
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

// func GetApplicationEventById(id string) (*models.ApplicationEvent, error) {
// 	nEvent := models.ApplicationEvent{}

// 	err := query.GetOne(models.ApplicationEvent{
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

// func GetApplicationEventByHash(hash string) (*models.ApplicationEvent, error) {
// 	nEvent := models.ApplicationEvent{}

// 	err := query.GetOne(models.ApplicationEvent{
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
		return &models.ApplicationEvent{}
	}
	if eventType < 1400 {
		return &models.WalletEvent{}
	}
	return nil
}