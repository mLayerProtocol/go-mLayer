package client

import (
	// "errors"
	"context"
	"encoding/json"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"gorm.io/gorm"
)

var logger = &log.Logger

type Flag string

// !sign web3 m
// type msgError struct {
// 	code int
// 	message string
// }

type MessageService struct {
	Ctx context.Context
	Cfg configs.MainConfiguration
}

// type Subscribe struct {
// 	channel   string
// 	timestamp string
// }

func GetMessages(topicId string) (*[]models.MessageState, error) {
	messageStates := []models.MessageState{}
	// order := map[string]query.Order{"created_at": query.OrderDec}
	// err := query.GetMany(models.MessageState{
	// 	Message: entities.Message{Topic: topicId},
	// }, &messageStates, &order)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }
	messages, err := dsquery.GetMessages(entities.Message{Topic: topicId}, dsquery.DefaultQueryLimit, nil)
	if err != nil && !dsquery.IsErrorNotFound(err) {
		return &messageStates, err
	}
	for _, msg :=  range messages {
		messageStates = append(messageStates, models.MessageState{Message: *msg})
	}
	return &messageStates, nil
}

func NewMessageService(mainCtx *context.Context) *MessageService {
	ctx := *mainCtx
	cfg, _ := ctx.Value(constants.ConfigKey).(*configs.MainConfiguration)
	return &MessageService{
		Ctx: ctx,
		Cfg: *cfg,
	}
}

// func (p *MessageService) Send(chatMsg entities.Message, senderSignature string) (*entities.Event, error) {
// 	// if strings.ToLower(chatMsg.Validator) != strings.ToLower(crypto.GetPublicKeyEDD(p.Cfg.PrivateKey)) {
// 	// 	return nil, errors.New("Invalid Origin node address: " + chatMsg.Validator + " is not")
// 	// }
// 	if service.IsValidMessage(chatMsg, senderSignature) {

// 		//if utils.Contains(chatMsg.Header.Channels, "*") || utils.Contains(chatMsg.Header.Channels, strings.ToLower(channel[0])) {

// 			privateKey := p.Cfg.PrivateKey

// 			// TODO:
// 			// if its an array check the channels .. if its * allow
// 			// message server::: store messages, require receiver to request message through an endpoint
// 			hash, _ := chatMsg.GetHash()
// 			signature, _ := crypto.SignECC(hash, privateKey)
// 			message := entities.Event{}
// 			message.Payload.Data = &chatMsg
// 			message.Signature = hexutil.Encode(signature)
// 			outgoingMessageC, ok := p.Ctx.Value(constants.OutgoingMessageChId).(*chan *entities.Event)
// 			if !ok {
// 				logger.Error("Could not connect to outgoing channel")
// 				panic("outgoing channel fail")
// 			}
// 			*outgoingMessageC <- &message
// 			fmt.Printf("Testing my function%s, %s", chatMsg.ToString(), string(chatMsg.Data))
// 			return &message, nil
// 		//}
// 	}
// 	return nil, errors.New("INVALID MESSAGE SIGNER")
// }

func ValidateMessagePayload(payload entities.ClientPayload, currentAuthState *models.AuthorizationState) (assocPrevEvent *entities.EventPath, assocAuthEvent *entities.EventPath, err error) {

	payloadData := entities.Message{}
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &payloadData)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
	}
	payload.Data = payloadData

	topicData, err := dsquery.GetTopicById(payloadData.Topic)
	if err != nil {
		return nil, nil, err
	}

	if topicData == nil {
		return nil, nil, apperror.BadRequest("Invalid topic id")
	}

	
	// pool = channelpool.SubscriptionEventPublishC

	
	


	subscription, err := service.ValidateMessageData(&payload, topicData)
	
	
	
	if err != nil {
		if  payload.Account != topicData.Account {
			if err == gorm.ErrRecordNotFound {
				if payload.Account != topicData.Account {
					return nil, &currentAuthState.Event, apperror.Forbidden(err.Error())
				}
			}
			return nil, &currentAuthState.Event, err
		}
		return nil, nil, err
	}

	// generate associations
		if subscription != nil  {
			assocPrevEvent = &subscription.Event
		} else {
			if payload.Account == topicData.Account {
				assocPrevEvent = &topicData.Event
			} else {
				subState, err := dsquery.GetSubnetStateById(topicData.Subnet)
				if err != nil {
					if err == gorm.ErrRecordNotFound  || dsquery.IsErrorNotFound(err){
						return nil, nil, apperror.Forbidden("Invalid subnet id")
					}
					return nil, nil, apperror.Internal(err.Error())
				}
				assocPrevEvent = &subState.Event
			
			}

	}

	if currentAuthState != nil {
		assocAuthEvent = &currentAuthState.Event
	}
	return assocPrevEvent, assocAuthEvent, nil
}
