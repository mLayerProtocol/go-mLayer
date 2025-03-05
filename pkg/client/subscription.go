package client

import (
	// "errors"

	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
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


func GetAccountSubscriptionsV2(cfg *configs.MainConfiguration, payload entities.Subscription) (*[]models.SubscriptionState, error) {
	var states []models.SubscriptionState = []models.SubscriptionState{}
	_states, err := dsquery.GetSubscriptions(payload, dsquery.DefaultQueryLimit, nil)
	if len(payload.Subscriber) > 0 && len(_states) == 0  && !cfg.BootstrapNode  {
		cacheKey := stores.AccountConnectedKey.NewKey(string(payload.Subscriber))
		_, seen := stores.SystemCache.Get(cacheKey)
		if !seen {
			// TODO sync with a bootstrap node
			payload := p2p.NewP2pPayload(cfg, p2p.P2pActionGetAccountSubscriptions, []byte(payload.Subscriber))
			
			data, err := p2p.SendSecureQuicRequest(cfg, p2p.GetRandomBootstrapPeer(cfg.BootstrapPeers), "", payload.MsgPack())
			if err != nil {
				return &states, err
			}
			response, err := p2p.UnpackP2pPayload(data)
			if err != nil ||  response == nil || len(response.Error) > 0 {
				return nil, err
			}
			if len(response.Error) > 0 {
				return nil, fmt.Errorf("GetAccountSubscriptionsV2: %v", response.Error)
			}
			dataBytes := bytes.Split(response.Data, p2p.Delimiter)
			txn, _ := dsquery.InitTx(stores.StateStore, nil)
			defer txn.Discard(context.Background())
			topics := []string{}
			for _, subsBytes:= range(dataBytes) {
				s, err := entities.UnpackSubscription(subsBytes)
				if err != nil {
					return &states, err
				}
				dsquery.CreateSubscriptionState(&s, &txn)
				states = append(states, models.SubscriptionState{Subscription: s})

				// check if in own interest list, if there ignore. Use this to exclude repeated topic publishing - ie. multiple account subscribed to a topic
				if b, err := dsquery.IsInterestedIn(s.Topic, entities.TopicModel); err == nil && !b {
					topics = append(topics, s.Topic)
				}
			
			}
			if err == nil {
				err = txn.Commit(context.Background())
			}
			// Register interest in topics
			errInterest := dsquery.SetOwnInterests(topics, entities.TopicModel)
			if errInterest != nil && errInterest != dsquery.ErrorKeyExist {
				panic(errInterest)
			}
			if errInterest == nil {
				// broadcast your interest
				go PublishInterest(topics, entities.TopicModel, cfg)
			}
			
			
			

			stores.SystemCache.Set(cacheKey, []byte{})
			return &states, err
		} 

	}
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
	topic *entities.Topic,
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
			return nil, nil, nil, apperror.BadRequest(fmt.Sprintf("Topic %s does not exist in subnet %s", payloadData.Topic, payload.Subnet))
		}
		_topic = state.(*entities.Topic)
	}

	currentState, err := service.ValidateSubscriptionData(&payload, _topic)
	logger.Infof("SubscriptionError: %+v", err)
	if err != nil && (!dsquery.IsErrorNotFound(err) && payload.EventType == constants.SubscribeTopicEvent) {
		return nil, nil, _topic, err
	}

	if currentState == nil && payload.EventType != constants.SubscribeTopicEvent {
		return nil, nil, _topic, apperror.BadRequest("Account not subscribed")
	}

	// if payload.EventType == constants.SubscribeTopicEvent && payload.Account == topicData.Account && !slices.Contains([]constants.SubscriptionStatus{1, 3}, *payloadData.Status) {
	// 	return nil, nil, apperror.BadRequest("Invalid Subscription Status")
	// }

	// if payload.EventType == constants.SubscribeTopicEvent) && payload.Account != topicData.Account && slices.Contains([]constants.SubscriptionStatus{3}, *payloadData.Status) {
	// 	return nil, nil, apperror.BadRequest("Invalid Subscription Status 01")
	// }

	if currentState != nil && payloadData.Subscriber == entities.AddressString(_topic.Account) && payload.EventType == constants.SubscribeTopicEvent {
		return nil, nil, _topic, apperror.BadRequest("Topic already owned by account")
	}

	// if currentState != nil && currentState.Status != &constants.UnsubscribedSubscriptionStatus && payload.EventType == constants.SubscribeTopicEvent) {
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
	logger.Infof("SubscriptionError3: %+v", _topic,  err)
	return assocPrevEvent, assocAuthEvent, _topic, nil
}
