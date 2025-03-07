package client

import (
	// "errors"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"

	// query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
)

type RequestType string



const (
	GetNodeInfoRequest   RequestType = "READ:info"
	FindApplicationsRequest   RequestType = "READ:apps"
	GetApplicationByIdRequest             = "READ:apps/:id"
	WriteApplicationRequest               = "WRITE:apps"
	FindAuthorizationsRequest RequestType = "READ:authorizations"
	WriteAuthorizationRequest RequestType = "WRITE:authorizations"
	FindTopicsRequest          = "READ:topics"
	WriteTopicRequest          = "WRITE:topics"
	GetTopicSubscribersRequest = "READ:topics/subscribers"
	GetTopicByIdRequest        = "READ:/topics"
	WriteSubscriptionRequest   = "WRITE:subscriptions"
	GetSubscriptionByIdRequest = "READ:subscription/:id"
	GetAccountSubscriptionsRequest = "READ:accounts/:acct/subscriptions"
	WriteMessageRequest     = "WRITE:messages"
	GetTopicMessagesRequest = "READ:topics/:id/messages"
	SyncClientRequest          = "READ:sync"
	BlockStatsRequest          = "READ:block-stats"
	GetEventByTypeAndIdRequest = "READ:event/:type/:id"
	GetMainStatsRequest        = "READ:main-stats"
)

var requestPatterns = []RequestType{
	GetNodeInfoRequest,
	FindApplicationsRequest,
	GetApplicationByIdRequest,
	WriteApplicationRequest,

	FindAuthorizationsRequest,
	WriteAuthorizationRequest,

	FindTopicsRequest,
	WriteTopicRequest,
	GetTopicSubscribersRequest,
	GetTopicByIdRequest,

	WriteSubscriptionRequest,
	GetSubscriptionByIdRequest,

	GetAccountSubscriptionsRequest,

	WriteMessageRequest,
	GetTopicMessagesRequest,

	SyncClientRequest,
	BlockStatsRequest,
	GetEventByTypeAndIdRequest,
	GetMainStatsRequest,
}

type ClientRequestHandler struct {
	Ctx *context.Context
	Cfg *configs.MainConfiguration
}

var (
	ErrorInvalidRequest error = fmt.Errorf("invalid request type")
)

func NewClientRequestHandler(mainCtx *context.Context) *ClientRequestHandler {
	cfg, _ := (*mainCtx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	return &ClientRequestHandler{
		Ctx: mainCtx,
		Cfg: cfg,
	}
}

func getNodeInfo(cfg *configs.MainConfiguration) (interface{}, error) {
	info, err := Info(cfg)
	if err != nil {
		return nil, err
	}
	return info, nil

}




func parseEntity[M any](_type M, payload *entities.ClientPayload) {
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &_type)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
	}
	payload.Data = _type
}

func parseClientPayload(payload *entities.ClientPayload, requestType RequestType) {
	switch requestType {
	case GetNodeInfoRequest:
		parseEntity(entities.Application{}, payload)
	case WriteAuthorizationRequest:
		parseEntity(entities.Authorization{}, payload)
		// data := entities.Authorization{}
		// d, _ := json.Marshal(payload.Data)
		// e := json.Unmarshal(d, &data)
		// if e != nil {
		// 	logger.Errorf("UnmarshalError %v", e)
		// }
		// payload.Data = data
	case WriteTopicRequest:
		parseEntity(entities.Topic{}, payload)
		//data := entities.Topic{}
		// d, _ := json.Marshal(payload.Data)
		// e := json.Unmarshal(d, &data)
		// if e != nil {
		// 	logger.Errorf("UnmarshalError %v", e)
		// }
		// payload.Data = data
	case WriteSubscriptionRequest:
		parseEntity(entities.Subscription{}, payload)
	case WriteMessageRequest:
		parseEntity(entities.Message{}, payload)

	}
}

func (p *ClientRequestHandler) Process(requestPath RequestType, params map[string]interface{}, payload interface{}) (interface{}, error) {
	var request RequestType
	queryLimit := dsquery.DefaultQueryLimit
	for _, pattern := range requestPatterns {
		match, par := utils.MatchUrlPath(string(pattern), string(requestPath))
		if match {
			request = pattern;
			for k, v := range par {
				params[k] = v
			}
			break;
		}
	}
	page := 1;
	perPage := queryLimit.Limit
	if params["page"] != nil {
		d, err := strconv.Atoi(fmt.Sprint(params["page"]))
		if err != nil {
			page = 1
		} else {
			page = d
		}
	}
	if params["perPage"] != nil {
		if page == 0 {
			page = 1
		}
		queryLimit.Offset = page - 1 * perPage
	}
	
	switch request {
	case "READ:ping":
		return entities.ClientResponse{}, nil
	case GetNodeInfoRequest:
		return getNodeInfo(p.Cfg)
	case FindAuthorizationsRequest:
		return GetAuthorizations(payload.(*entities.Authorization))
		// b, parseError := utils.ParseQueryString(c)
		// if parseError != nil {
		// 	logger.Error(parseError)
		// 	return nil, parseError
		// }
		// var authEntity entities.Authorization
		// json.Unmarshal(*payload, &authEntity)
		// return getAuthorizations(&authEntity)
	case FindTopicsRequest:
		return dsquery.GetAccountTopics(payload.(entities.Topic), dsquery.DefaultQueryLimit, nil)
	case WriteApplicationRequest: // "WRITE:topics/subscribers/approve", "PATCH:topics/unsubscribe", "PATCH:topics/ban":
		cpl := payload.(entities.ClientPayload)
		data := entities.Application{}
		d, _ := json.Marshal(cpl.Data)
		e := json.Unmarshal(d, &data)
		if e != nil {
			logger.Errorf("UnmarshalError %v", e)
		}
		cpl.Data = data
		return CreateEvent(cpl, p.Ctx)
	case WriteAuthorizationRequest:
		cpl := payload.(entities.ClientPayload)
		parseClientPayload(&cpl, request)
		data := entities.Authorization{}
		d, _ := json.Marshal(cpl.Data)
		e := json.Unmarshal(d, &data)
		if e != nil {
			logger.Errorf("UnmarshalError %v", e)
		}
		cpl.Data = data
		return CreateEvent(cpl, p.Ctx)
	case WriteTopicRequest:
		cpl := payload.(entities.ClientPayload)
		parseClientPayload(&cpl, request)
		data := entities.Topic{}
		d, _ := json.Marshal(cpl.Data)
		e := json.Unmarshal(d, &data)
		if e != nil {
			logger.Errorf("UnmarshalError %v", e)
		}
		cpl.Data = data
		return CreateEvent(cpl, p.Ctx)
	case WriteSubscriptionRequest:
		cpl := payload.(entities.ClientPayload)
		parseClientPayload(&cpl, request)
		
		data := entities.Subscription{}
		d, _ := json.Marshal(cpl.Data)
		e := json.Unmarshal(d, &data)
		if e != nil {
			logger.Errorf("UnmarshalError %v", e)
		}
		cpl.Data = data
		
		return CreateEvent(cpl, p.Ctx)
	case WriteMessageRequest:
		cpl := payload.(entities.ClientPayload)
		parseClientPayload(&cpl, request)
		data := entities.Message{}
		d, _ := json.Marshal(cpl.Data)
		e := json.Unmarshal(d, &data)
		if e != nil {
			logger.Errorf("UnmarshalError %v", e)
		}
		cpl.Data = data
		return CreateEvent(cpl, p.Ctx)
	// case GetSubscriptionByIdRequest:
	// 	return dsquery.GetSubsc(params["id"].(string))
	case GetTopicSubscribersRequest:
		subPayload := payload.(entities.Subscription)
		subPayload.Topic = params["topic"].(string)
		return GetSubscriptions(subPayload)
	case GetTopicMessagesRequest:
		return GetMessages(params["id"].(string))
	case GetTopicByIdRequest:
		return dsquery.GetTopicById(params["id"].(string))
	case GetAccountSubscriptionsRequest:

		status := params["status"]
		if status != "" {
			iStatus, parseError := strconv.Atoi(status.(string))
			if parseError != nil {
				return nil, parseError
			}
			statusConst := constants.SubscriptionStatus(iStatus)
			params["status"] = statusConst
		}
		var subs entities.Subscription
		pB, err := json.Marshal(params)
		if err != nil {

		}
		json.Unmarshal(pB, &subs)
		return GetAccountSubscriptionsV2(p.Cfg, subs)
	// case SyncClientRequest:
	// 	//var authEntity entities.Authorization
	// 	// var payload entities.ClientPayload
	// 	// json.Unmarshal(*b, &authEntity)

	// 	syncResponse := entities.SyncResponse{}
	// 	SyncAgent(&entities.SyncRequest{}, &entities.ClientPayload{})

	// 	return syncResponse, nil
	case BlockStatsRequest:
		block := params["block"].(string)
	
		blockNumber := uint64(0)
		if len(block) > 0 {
			n, err := strconv.Atoi(block)
			if err != nil {
				return nil, err
			}
			blockNumber = uint64(n)
		}
		blockStats, err := GetBlockStats(blockNumber,  queryLimit)
		if err != nil {

			return nil, err
		}
		return blockStats, nil
	case GetMainStatsRequest:
		mainStats, err := GetMainStats(p.Cfg)

		if err != nil {
			return nil, err
		}
		return mainStats, nil
	case "GET:event-path/:hash/:type/:id":
		id := params["id"].(string)
		typeParam := params["type"].(string)
		// typeParamInt := GetEventTypeFromModel(entities.EntityModel(typeParam))

		topic, err := dsquery.GetStateFromEventPath(&entities.EventPath{EntityPath: entities.EntityPath{ID: id, Model: entities.EntityModel(typeParam)}})

		if err != nil {
			return nil, err
		}
		return topic, nil

	case GetEventByTypeAndIdRequest:
		id := params["id"].(string)
		typeParam := params["type"].(string)
		typeParamInt, err := strconv.Atoi(typeParam)
		if err != nil {
			return nil, err
		}
		event, err := GetEvent(id, typeParamInt)
		if err != nil {
			return nil, err
		}
		return event, nil
	case FindApplicationsRequest:

		b, parseError := json.Marshal(params)
		if parseError != nil {
			return nil, parseError
		}

		var appState models.ApplicationState

		json.Unmarshal(b, &appState)

		return GetSubscribedApplications(appState)
	case GetApplicationByIdRequest:
		app, err := dsquery.GetApplicationStateById(params["id"].(string))
		if err != nil {
			return nil, err
		}
		return models.ApplicationState{Application: *app}, nil
	default:
		return nil, ErrorInvalidRequest
	}

	// get info about the node
	// return nil, ErrorInvalidRequest
}
