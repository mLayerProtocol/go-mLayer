package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/looplab/fsm"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"gorm.io/gorm"
)



const (
	PreEvUptoDateMetaData string = "PreEvUptoDate"
	AuthEvUptoDateMetaData string = "AuthEvUptoDate"
	AuthStateMetaData   string = "AuthState"
	IsRecentMetaData string = "IsRecent"
)
type EventStateMachine struct {
	*fsm.FSM
	event        *entities.Event
	data         PayloadData
	txn          *datastore.Txn
	tx           *gorm.DB
	cfg 	*configs.MainConfiguration
	dataDataStates    *dsquery.DataStates
	validAgentRequired bool
}

type ValidationState struct {
	PreEvUptoDate bool
	AuthEvUptoDate bool
	AuthSt *models.AuthorizationState
	IsRecent bool
	// Err error
}

func NewEventFSM(event *entities.Event, validAgentRequired bool, data PayloadData, txn *datastore.Txn, tx *gorm.DB, cfg *configs.MainConfiguration, dataDataStates *dsquery.DataStates) *EventStateMachine {
	return &EventStateMachine{
		FSM: fsm.NewFSM(
			"init",
			fsm.Events{
				{Name: "validate", Src: []string{"init"}, Dst: "validated"},
				{Name: "sync_app", Src: []string{"validated"}, Dst: "app_synced"},
				{Name: "sync_authorization", Src: []string{"app_synced"}, Dst: "authorization_synced"},
				{Name: "validate_agent", Src: []string{"authorization_synced"}, Dst: "agent_validated"},
				{Name: "process_previous", Src: []string{"agent_validated"}, Dst: "previous_processed"},
				{Name: "process_auth", Src: []string{"previous_processed"}, Dst: "auth_processed"},
				// {Name: "check_local", Src: []string{"auth_processed"}, Dst: "local_checked"},
				{Name: "save_event", Src: []string{"local_checked"}, Dst: "event_saved"},
				{Name: "update_state", Src: []string{"event_saved"}, Dst: "state_updated"},
				{Name: "done", Src: []string{"event_saved"}, Dst: "done"},
				{Name: "error", Src: []string{"init", "validated", "previous_processed", "auth_processed", "local_checked", "event_saved", "state_updated"}, Dst: "error"},
			},
			fsm.Callbacks{
				"before_validate": func(_ context.Context, e *fsm.Event) {
					e.Args[0].(*EventStateMachine).BeforeValidateEvent()
					fmt.Printf("Entering state %s\n", e.Dst)
				},
				"enter_state": func(_ context.Context, e *fsm.Event) {
					fmt.Printf("Entering state %s\n", e.Dst)
				},
			},
		),
		event:        event,
		data:         data,
		txn:          txn,
		tx:           tx,
		cfg:          cfg,
		dataDataStates:    dataDataStates,
		validAgentRequired: validAgentRequired,
	}
}

func (esm *EventStateMachine) setMetaData(preEvUptoDate bool, authEvUptoDate bool, authSt *models.AuthorizationState, isRecent bool, err error) error {
	esm.FSM.SetMetadata(string(PreEvUptoDateMetaData), preEvUptoDate)
	esm.FSM.SetMetadata(string(AuthEvUptoDateMetaData), authEvUptoDate)
	esm.FSM.SetMetadata(string(AuthStateMetaData), authSt)
	esm.FSM.SetMetadata(string(IsRecentMetaData), isRecent)
	return err
}

func (esm *EventStateMachine) BeforeValidateEvent()  error  {
	if esm.event.IsLocal(esm.cfg) {
		if len(esm.event.AuthEvent.ID) > 0 {
			_, stateByte, _, err := GetEventByPath(&esm.event.AuthEvent, esm.cfg, string(esm.event.Validator))
			if err != nil {
				return esm.setMetaData(true, false, nil, true, err)
			}
			state, err := entities.UnpackAuthorization(stateByte)
			if err != nil {
				return esm.setMetaData(true, true, nil, true,  nil)
			}
			return esm.setMetaData( true, true, &models.AuthorizationState{Authorization: state}, true, nil)
		}
	
		esm.FSM.SetState("done")
		return esm.setMetaData( true, true, nil, true, nil)
	}
	return nil
	
}

func (esm *EventStateMachine) ValidateEvent()  error  {
	if esm.validAgentRequired && esm.event.Payload.Timestamp > 0 && 
	(uint64(esm.event.Payload.Timestamp) > uint64(esm.event.Timestamp)+15000 || uint64(esm.event.Payload.Timestamp) < uint64(esm.event.Timestamp)-15000) {
		return esm.setMetaData(false, false, nil, false, errors.New("event timestamp exceeds payload timestamp"))
	}

	err := ValidateEvent(*esm.event)

	if err != nil {
		logger.Errorf("ValidateEventError: %v", err)
		return esm.setMetaData(false, false, nil, false, err)
	}
	d, err := esm.event.Payload.EncodeBytes()
	if err != nil || len(d) == 0 {
		logger.Debug("Invalid event payload")
		return esm.setMetaData(false, false, nil, false, fmt.Errorf("invalid event payload"))
	}

	return nil
	
}
