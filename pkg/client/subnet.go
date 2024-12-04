package client

import (
	// "errors"

	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
)

// type SubnetService struct {
// 	Ctx context.Context
// 	Cfg configs.MainConfiguration
// }

// func NewSubnetService(mainCtx *context.Context) *SubnetService {
// 	ctx := *mainCtx
// 	cfg, _ := ctx.Value(constants.ConfigKey).(*configs.MainConfiguration)
// 	return &SubnetService{
// 		Ctx: ctx,
// 		Cfg: *cfg,
// 	}
// }

// func (p *SubnetService) NewSubnetSubscription(sub *entities.Subscription) error {
// 	// subscribersc, ok := p.Ctx.Value(utils.SubscribeChId).(*chan *entities.Subscription)

// 	// validate before storing
// 	if entities.IsValidSubscription(*sub, true) {
// 		SubnetSubscriberStore, ok := p.Ctx.Value(constants.NewSubnetSubscriptionStore).(*ds.Datastore)
// 		if !ok {
// 			return errors.New("Could not connect to subscription datastore")
// 		}
// 		error := SubnetSubscriberStore.Set(p.Ctx, db.Key(sub.Key()), sub.MsgPack(), false)
// 		if error != nil {
// 			return error
// 		}
// 	}
// 	return nil
// }

/*
Validate and Process the Subnet request
*/

// func GetSubnetById(id string) (*models.SubnetState, error) {
// 	SubnetState := models.SubnetState{}

// 	err := query.GetOne(models.SubnetState{
// 		Subnet: entities.Subnet{ID: id},
// 	}, &SubnetState)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &SubnetState, nil

// }
// func GetSubnetByHash(hash string) (*models.SubnetState, error) {
// 	SubnetState := models.SubnetState{}

// 	err := query.GetOne(models.SubnetState{
// 		Subnet: entities.Subnet{Hash: hash},
// 	}, &SubnetState)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &SubnetState, nil

// }

// func GetSubnets(item models.SubnetState) (*[]models.SubnetState, error) {
// 	var SubnetStates []models.SubnetState

// 	err := query.GetMany(item, &SubnetStates, nil)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return &SubnetStates, nil
// }

func GetSubscribedSubnets(item models.SubnetState) (*[]models.SubnetState, error) {

	 SubnetStates := []models.SubnetState{}

	// err := query.GetMany(item, &SubnetStates, nil)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return nil, nil
	// 	}
	// 	return nil, err
	// }

	_subnets, err := dsquery.GetAccountSubnets(item.Account, *dsquery.DefaultQueryLimit)
	if err != nil {
		return nil, err
	}
	logger.Infof("AccountSubnets: %v", _subnets)
	for _, _sub := range _subnets {
		SubnetStates = append(SubnetStates, models.SubnetState{Subnet: *_sub})
	}


	//var subscriptionStates []models.SubscriptionState
	// err = query.GetMany(models.SubscriptionState{Subscription: entities.Subscription{Subscriber: item.Account}},
	// 	&subscriptionStates, nil)
	subscriptionStates, err := dsquery.GetSubscriptions(entities.Subscription{Subscriber: item.Account}, nil, nil)

	if err != nil {

		return nil, err
	}
	var subnetIds = []string{}

	for _, sub := range subscriptionStates {
		subnetIds = append(subnetIds, sub.Subnet)
	}
	var subSubnetStates []models.SubnetState
	if len(subnetIds) > 0 {
		subSubnetErr := query.GetWithIN(models.SubnetState{}, &subSubnetStates, subnetIds)
		if subSubnetErr != nil {
			return nil, err
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
	logger.Infof("BeforeValidate 1: %s", payload.Signature)
	currentState, err2 := service.ValidateSubnetData(&payload, cfg.ChainId)
	if err2 != nil {
		return nil, nil, err2
	}
	if payload.EventType == uint16(constants.CreateSubnetEvent) {
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
		// logger.Debug("FOUNDDDDD", found, payloadData.Ref)

	}
	if payload.EventType == uint16(constants.UpdateSubnetEvent) {
		if payloadData.ID == "" {
			return nil, nil, apperror.BadRequest("Subnet ID must be provided")
		}
	}
	
	
	// generate associations
	if currentState != nil {
		//logger.Debugf("SUBNETINFO %v, %s, %s", strings.EqualFold(currentState.Account.ToString(), payloadData.Account.ToString()), currentState.Account.ToString(), payloadData.Account.ToString())
		if !strings.EqualFold(currentState.Account.ToString(), payloadData.Account.ToString()) {
			return nil, nil, apperror.BadRequest("subnet account do not match")
		}
		assocPrevEvent = &currentState.Event
	}
	if authState != nil {
		assocAuthEvent = &authState.Event
	}
	
	return assocPrevEvent, assocAuthEvent, nil
}
