package client

import (
	// "errors"

	"encoding/json"
	"fmt"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
)

func GetSubscriptions(payload entities.Subscription) (*[]models.SubscriptionState, error) {
	subscriptionStates  := []models.SubscriptionState{}
	subs, err := dsquery.GetSubscriptions(payload, dsquery.DefaultQueryLimit, nil)
	if err !=nil && !dsquery.IsErrorNotFound(err) {
			
		return nil, err
	}
	for _, sub := range subs {
		subscriptionStates = append(subscriptionStates, models.SubscriptionState{Subscription: *sub})
	}
	
	return &subscriptionStates, nil
}


func GetAccountSubscriptionsV2(payload entities.Subscription) (*[]models.SubscriptionState, error) {
	var states []models.SubscriptionState

	_states, err := dsquery.GetSubscriptions(payload, dsquery.DefaultQueryLimit, nil)
	if err != nil {
		if dsquery.IsErrorNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, _state := range _states {
		states = append(states, models.SubscriptionState{Subscription: *_state})
	}
	return &states, nil
	
	// var subscriptionStates []models.SubscriptionState
	// var subTopicStates []models.TopicState
	// var topicStates []models.TopicState
	// order := map[string]query.Order{"timestamp": query.OrderDec, "created_at": query.OrderDec}
	// err := query.GetMany(models.SubscriptionState{Subscription: payload},
	// 	&subscriptionStates, &order)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }

	// var topicIds []string

	// for _, sub := range subscriptionStates {
	// 	if sub.Topic == "" {
	// 		continue
	// 	}
	// 	topicIds = append(topicIds, sub.Topic)
	// }

	// if len(topicIds) > 0 {
	// 	subTopErr := query.GetWithIN(models.TopicState{}, &subTopicStates, topicIds)
	// 	if subTopErr != nil {
	// 		if subTopErr == gorm.ErrRecordNotFound {
	// 			return nil, nil
	// 		}
	// 		return nil, err
	// 	}
	// }

	// // topErr := query.GetMany(models.TopicState{Topic: entities.Topic{Account: payload.Subscriber}}, &topicStates, nil)
	// // if topErr != nil {
	// // 	if err == gorm.ErrRecordNotFound {
	// // 		return nil, nil
	// // 	}
	// // 	return nil, err
	// // }

	// topicStates = append(topicStates, subTopicStates...)

	// return &topicStates, nil
}

// func GetSubscription(id string) (*models.SubscriptionState, error) {
// 	subscriptionState := models.SubscriptionState{}

// 	err := query.GetOne(models.SubscriptionState{
// 		Subscription: entities.Subscription{ID: id},
// 	}, &subscriptionState)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &subscriptionState, nil

// }

func ValidateSubscriptionPayload(payload entities.ClientPayload, authState *models.AuthorizationState, cfg *configs.MainConfiguration) (
	assocPrevEvent *entities.EventPath,
	assocAuthEvent *entities.EventPath,
	err error,
) {

	payloadData := entities.Subscription{}
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &payloadData)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
	}
	// if authState == nil {
	// 	return nil, nil,  apperror.Unauthorized("Agent not authorized")
	// }
	// if authState.Priviledge != constants.AdminPriviledge {
	// 	return nil, nil,  apperror.Unauthorized("Agent does not have sufficient privileges to perform this action")
	// }
	payload.Data = payloadData

	// var topicData *models.TopicState
	// query.GetOne(models.TopicState{
	// 	Topic: entities.Topic{ID: payloadData.Topic, Subnet: payload.Subnet},
	// }, &topicData)
	_topic, err := dsquery.GetTopicById(payloadData.Topic)
	
	if dsquery.IsErrorNotFound(err) || _topic == nil {
		logger.Infof("ValidatingTopic 2... %v", err)
		state, _, err := service.SyncStateFromPeer(payloadData.Topic, entities.TopicModel, cfg, "")
		// logger.Infof("ValidatingTopic 3... %v, %v", err, state)
		if err != nil || state == nil {
			return nil, nil, apperror.BadRequest(fmt.Sprintf("Topic %s does not exist in subnet %s", payloadData.Topic, payload.Subnet))
		}
		_topic = state.(*entities.Topic)
	}

	currentState, err := service.ValidateSubscriptionData(&payload, _topic)
	logger.Infof("SubscriptionError: %+v", err)
	if err != nil && (!dsquery.IsErrorNotFound(err) && payload.EventType == uint16(constants.SubscribeTopicEvent)) {
		return nil, nil, err
	}

	if currentState == nil && payload.EventType != uint16(constants.SubscribeTopicEvent) {
		return nil, nil, apperror.BadRequest("Account not subscribed")
	}

	// if payload.EventType == uint16(constants.SubscribeTopicEvent) && payload.Account == topicData.Account && !slices.Contains([]constants.SubscriptionStatus{1, 3}, *payloadData.Status) {
	// 	return nil, nil, apperror.BadRequest("Invalid Subscription Status")
	// }

	// if payload.EventType == uint16(constants.SubscribeTopicEvent) && payload.Account != topicData.Account && slices.Contains([]constants.SubscriptionStatus{3}, *payloadData.Status) {
	// 	return nil, nil, apperror.BadRequest("Invalid Subscription Status 01")
	// }

	if currentState != nil && payloadData.Subscriber == _topic.Account && payload.EventType == uint16(constants.SubscribeTopicEvent) {
		return nil, nil, apperror.BadRequest("Topic already owned by account")
	}

	// if currentState != nil && currentState.Status != &constants.UnsubscribedSubscriptionStatus && payload.EventType == uint16(constants.SubscribeTopicEvent) {
	// 	return nil, nil, apperror.BadRequest("Account already subscribed")
	// }
	logger.Infof("SubscriptionError2: %+v", err)
	// generate associations
	if currentState != nil {
		assocPrevEvent = &currentState.Event
	} else {
		assocPrevEvent = &_topic.Event
	}

	if authState != nil {
		assocAuthEvent = &authState.Event
	}
	logger.Infof("SubscriptionError3: %+v", err)
	return assocPrevEvent, assocAuthEvent, nil
}
