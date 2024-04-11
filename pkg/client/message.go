package client

import (
	// "errors"
	"context"
	"encoding/json"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
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
	var messageStates []models.MessageState

	err := query.GetMany(models.MessageState{
		Message: entities.Message{TopicId: topicId},
	}, &messageStates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
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
// 	// if strings.ToLower(chatMsg.Validator) != strings.ToLower(crypto.GetPublicKeyEDD(p.Cfg.NetworkPrivateKey)) {
// 	// 	return nil, errors.New("Invalid Origin node address: " + chatMsg.Validator + " is not")
// 	// }
// 	if service.IsValidMessage(chatMsg, senderSignature) {

// 		//if utils.Contains(chatMsg.Header.Channels, "*") || utils.Contains(chatMsg.Header.Channels, strings.ToLower(channel[0])) {

// 			privateKey := p.Cfg.NetworkPrivateKey

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

	topicData, err := GetTopicById(payloadData.TopicId)
	if err != nil {
		return nil, nil, err
	}

	if topicData == nil {
		return nil, nil, apperror.BadRequest("Invalid topic id")
	}

	// pool = channelpool.SubscriptionEventPublishC

	var subscription models.SubscriptionState
	err = query.GetOne(models.SubscriptionState{
		Subscription: entities.Subscription{Account: payload.Account, Topic: topicData.ID},
	}, &subscription)
	logger.Info()
	if err != nil && payload.Account != topicData.Account {
		if err != gorm.ErrRecordNotFound {
			if payload.Account != topicData.Account {
				return nil, &currentAuthState.Event, apperror.Forbidden("Now subscribed to topic")
			}
		}
		return nil, &currentAuthState.Event, apperror.Internal(err.Error())
	}
	if *topicData.ReadOnly && payload.Account != topicData.Account && subscription.Role != constants.AdminSubPriviledge {
		return nil, nil, apperror.Unauthorized("Not allowed to post to this topic")
	}

	_, err = service.ValidateMessageData(&payloadData, &payload)
	if err != nil {
		return nil, nil, err
	}

	// dont worry validating the AuthHash for Authorization requests
	// if uint64(payloadData.Timestamp) > uint64(time.Now().UnixMilli())+15000 {
	// 	return  errors.New("Authorization timestamp exceeded")
	// }

	// generate associations
	assocPrevEvent = &subscription.Event

	if currentAuthState != nil {
		assocAuthEvent = &currentAuthState.Event
	}
	return assocPrevEvent, assocAuthEvent, nil
}
