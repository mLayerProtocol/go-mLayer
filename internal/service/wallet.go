package service

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	query "github.com/mlayerprotocol/go-mlayer/internal/sql/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"github.com/sirupsen/logrus"
)

/*
Validate an agent authorization
*/
func ValidateWalletData(Wallet *entities.Wallet) (currentWalletState *models.WalletState, err error) {
	// check fields of Wallet
	// logger.Debug("Walletcc", Wallet.Ref)
	// if len(Wallet.Ref) > 60 {
	// 	return nil, apperror.BadRequest("Wallet ref cannont be more than 40 characters")
	// }
	// if len(Wallet.Ref) > 0 && !utils.IsAlphaNumericDot(Wallet.Ref) {
	// 	return nil, apperror.BadRequest("Ref must be alphanumeric, _ and . but cannot start with a number")
	// }
	return nil, nil
}

func HandleNewPubSubWalletEvent(event *entities.Event, ctx *context.Context) error {
	logger.WithFields(logrus.Fields{"event": event}).Debug("New Wallet event from pubsub channel")
	markAsSynced := false
	updateState := false
	var eventError string
	// hash, _ := event.GetHash()
	err := ValidateEvent(*event)

	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debugf("Event is a valid event %s", event.PayloadHash)
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)

	// Extract and validate the Data of the paylaod which is an Events Payload Data,
	data := event.Payload.Data.(*entities.Wallet)
	hash, _ := data.GetHash()
	data.Hash = hex.EncodeToString(hash)
	authEventHash := event.AuthEvent
	authState, authError := query.GetOneAuthorizationState(entities.Authorization{Event: authEventHash})

	currentState, err := ValidateWalletData(data)
	if err != nil {
		// penalize node for broadcasting invalid data
		logger.Debugf("Invalid Wallet data %v. Node should be penalized", err)
		return err
	}

	// check if we are upto date on this event
	prevEventUpToDate := query.EventExist(&event.PreviousEvent) || (currentState == nil && event.PreviousEvent.Hash == "") || (currentState != nil && currentState.Event.Hash == event.PreviousEvent.Hash)
	authEventUpToDate := query.EventExist(&event.AuthEvent) || (authState == nil && event.AuthEvent.Hash == "") || (authState != nil && authState.Event == authEventHash)

	// Confirm if this is an older event coming after a newer event.
	// If it is, then we only have to update our event history, else we need to also update our current state
	isMoreRecent := false
	if currentState != nil && currentState.Hash != data.Hash {
		var currentStateEvent = &models.WalletEvent{}
		query.GetOne(entities.Event{Hash: currentState.Event.Hash}, currentStateEvent)
		isMoreRecent, markAsSynced = IsMoreRecent(
			currentStateEvent.ID,
			currentState.Event.Hash,
			currentStateEvent.Payload.Timestamp,
			event.Hash,
			event.Payload.Timestamp,
			markAsSynced,
		 )
	}

	if authError != nil {
		// check if we are upto date. If we are, then the error is an actual one
		// the error should be attached when saving the event
		// But if we are not upto date, then we might need to wait for more info from the network

		if prevEventUpToDate && authEventUpToDate {
			// we are upto date. This is an actual error. No need to expect an update from the network
			eventError = authError.Error()
			markAsSynced = true
		} else {
			if currentState == nil || (currentState != nil && isMoreRecent) { // it is a morer ecent event
				if strings.HasPrefix(authError.Error(), constants.ErrorForbidden) || strings.HasPrefix(authError.Error(), constants.ErrorUnauthorized) {
					markAsSynced = false
				} else {
					// entire event can be considered bad since the payload data is bad
					// this should have been sorted out before broadcasting to the network
					// TODO penalize the node that broadcasted this
					eventError = authError.Error()
					markAsSynced = true
				}

			} else {
				// we are upto date. We just need to store this event as well.
				// No need to update state
				markAsSynced = true
				eventError = authError.Error()
			}
		}

	}

	// If no error, then we should act accordingly as well
	// If are upto date, then we should update the state based on if its a recent or old event
	if len(eventError) == 0 {
		if prevEventUpToDate && authEventUpToDate { // we are upto date
			if currentState == nil || isMoreRecent {
				updateState = true
				markAsSynced = true
			} else {
				// Its an old event
				markAsSynced = true
				updateState = false
			}
		} else {
			updateState = false
			markAsSynced = false
		}

	}

	// Save stuff permanently
	tx := sql.SqlDb.Begin()
	logger.Debug(":::::updateState: Db Error", updateState, currentState == nil)

	// If the event was not signed by your node
	if string(event.Validator) != (*cfg).PublicKeyEDDHex  {
		// save the event
		event.Error = eventError
		event.IsValid = utils.BoolPtr(markAsSynced && len(eventError) == 0)
		event.Synced = &markAsSynced
		event.Broadcasted = true
		_, _, err := query.SaveRecord(models.WalletEvent{
			Event: entities.Event{
				PayloadHash: event.PayloadHash,
			},
		}, &models.WalletEvent{
			Event: *event,
		}, nil, tx)
		if err != nil {
			tx.Rollback()
			logger.Error("1000: Db Error", err)
			return err
		}
	} else {
		if markAsSynced {
			_, _, err := query.SaveRecord(models.WalletEvent{
				Event: entities.Event{PayloadHash: event.PayloadHash},
			}, 
			&models.WalletEvent{
				Event: *event,
			},
			&models.WalletEvent{
				Event: entities.Event{Synced: utils.TruePtr(), Broadcasted: true, Error: eventError, IsValid: utils.BoolPtr(len(eventError) == 0)},
			}, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		} else {
			// mark as broadcasted
			_, _, err := query.SaveRecord(models.WalletEvent{
				Event: entities.Event{PayloadHash: event.PayloadHash, Broadcasted: false},
			},
			&models.WalletEvent{
				Event: *event,
			}, 
				&models.WalletEvent{
					Event: entities.Event{Broadcasted: true},
				}, tx)
			if err != nil {
				logger.Error("DB error", err)
			}
		}
	}

	// d, err := event.Payload.EncodeBytes()
	// if err != nil {
	// 	logger.Errorf("Invalid event payload")
	// }
	// agent, err := crypto.GetSignerECC(&d, &event.Payload.Signature)
	// if err != nil {
	// 	logger.Errorf("Invalid event payload")
	// }
	data.Event = *entities.NewEventPath(event.Validator, entities.WalletModel, event.ID)
	// data.Agent = entities.DIDString(agent)
	data.Account = event.Payload.Account
	// logger.Error("data.Public ", data.Public)

	if updateState {
		_, _, err := query.SaveRecord(models.WalletState{
			Wallet: entities.Wallet{ID: data.ID},
		}, &models.WalletState{
			Wallet: *data,
		}, utils.IfThenElse(event.EventType == uint16(constants.UpdateWalletEvent), &models.WalletState{
			Wallet: *data,
		}, nil ), tx)
		if err != nil {
			tx.Rollback()
			logger.Error("7000: Db Error", err)
			return err
		}
	}
	tx.Commit()

	if string(event.Validator) != (*cfg).PublicKeyEDDHex  {
		dependent, err := query.GetDependentEvents(event)
		if err != nil {
			logger.Debug("Unable to get dependent events", err)
		}
		for _, dep := range *dependent {
			go HandleNewPubSubWalletEvent(&dep, ctx)
		}
	}
return nil
	// TODO Broadcast the updated state
}
