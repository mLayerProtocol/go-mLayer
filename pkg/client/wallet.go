package client

import (
	// "errors"

	"encoding/json"
	"errors"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"gorm.io/gorm"
)

// type WalletService struct {
// 	Ctx context.Context
// 	Cfg configs.MainConfiguration
// }

// func NewWalletService(mainCtx *context.Context) *WalletService {
// 	ctx := *mainCtx
// 	cfg, _ := ctx.Value(constants.ConfigKey).(*configs.MainConfiguration)
// 	return &WalletService{
// 		Ctx: ctx,
// 		Cfg: *cfg,
// 	}
// }

// func (p *WalletService) NewWalletSubscription(sub *entities.Subscription) error {
// 	// subscribersc, ok := p.Ctx.Value(utils.SubscribeChId).(*chan *entities.Subscription)

// 	// validate before storing
// 	if entities.IsValidSubscription(*sub, true) {
// 		WalletSubscriberStore, ok := p.Ctx.Value(constants.NewWalletSubscriptionStore).(*ds.Datastore)
// 		if !ok {
// 			return errors.New("Could not connect to subscription datastore")
// 		}
// 		error := WalletSubscriberStore.Set(p.Ctx, db.Key(sub.Key()), sub.MsgPack(), false)
// 		if error != nil {
// 			return error
// 		}
// 	}
// 	return nil
// }

/*
Validate and Process the Wallet request
*/

func GetWalletById(id string) (*models.WalletState, error) {
	WalletState := models.WalletState{}

	err := query.GetOne(models.WalletState{
		Wallet: entities.Wallet{ID: id},
	}, &WalletState)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &WalletState, nil

}
func GetWalletByHash(hash string) (*models.WalletState, error) {
	WalletState := models.WalletState{}

	err := query.GetOne(models.WalletState{
		Wallet: entities.Wallet{Hash: hash},
	}, &WalletState)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &WalletState, nil

}

func GetWallets() (*[]models.WalletState, error) {
	var WalletStates []models.WalletState

	err := query.GetMany(models.WalletState{}, &WalletStates, nil)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &WalletStates, nil
}

func GetWalletEventById(id string) (*models.WalletEvent, error) {
	var WalletEvent models.WalletEvent

	err := query.GetOne(models.WalletEvent{
		Event: entities.Event{
			ID: id,
		},
	}, &WalletEvent)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &WalletEvent, nil
}

func GetWalletEvents() (*[]models.WalletEvent, error) {
	var WalletEvents []models.WalletEvent

	err := query.GetMany(models.WalletEvent{
		Event: entities.Event{
			BlockNumber: 1,
		},
	}, &WalletEvents, nil)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &WalletEvents, nil
}

// func ListenForNewWalletEventFromPubSub (mainCtx *context.Context) {
// 	ctx, cancel := context.WithCancel(*mainCtx)
// 	defer cancel()

//		incomingWalletC, ok := (*mainCtx).Value(constants.IncomingWalletEventChId).(*chan *entities.Event)
//		if !ok {
//			logger.Errorf("incomingWalletC closed")
//			return
//		}
//		for {
//			event, ok :=  <-*incomingWalletC
//			if !ok {
//				logger.Fatal("incomingWalletC closed for read")
//				return
//			}
//			go service.HandleNewPubSubWalletEvent(event, ctx)
//		}
//	}
func ValidateWalletPayload(payload entities.ClientPayload, authState *models.AuthorizationState) (assocPrevEvent *entities.EventPath, assocAuthEvent *entities.EventPath, err error) {

	payloadData := entities.Wallet{}
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &payloadData)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
	}

	payload.Data = payloadData
	if payload.EventType == constants.CreateWalletEvent {
		// dont worry validating the AuthHash for Authorization requests
		if uint64(payloadData.Timestamp) > uint64(time.Now().UnixMilli())+15000 {
			return nil, nil, errors.New("Authorization timestamp exceeded")
		}

	}

	currentState, err := service.ValidateWalletData(&payloadData)
	if err != nil {
		return nil, nil, err
	}

	// generate associations
	if currentState != nil {
		assocPrevEvent = &currentState.Event

	}
	if authState != nil {
		assocAuthEvent = &authState.Event
	}
	return assocPrevEvent, assocAuthEvent, nil
}
