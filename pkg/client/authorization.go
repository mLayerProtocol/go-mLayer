package client

import (
	"encoding/json"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
)



func ValidateAuthPayload(cfg *configs.MainConfiguration, payload entities.ClientPayload) (assocPrevEvent *entities.EventPath, assocAuthEvent *entities.EventPath, appState *entities.Application, err error) {
	authData := entities.Authorization{}
	
	d, _ := json.Marshal(payload.Data)
	e := json.Unmarshal(d, &authData)
	if e != nil {
		logger.Errorf("UnmarshalError %v", e)
	}
	
	payload.Data = authData
	if uint64(*authData.Timestamp) == 0 || uint64(*authData.Timestamp) > uint64(time.Now().UnixMilli())+15000 || uint64(*authData.Timestamp) < uint64(time.Now().UnixMilli())-15000 {
		return nil, nil, nil, apperror.BadRequest("Invalid event timestamp")
	}
	logger.Debugf("CurrentStateDD: %+v", payload.Data)
	if *authData.Duration != 0 && uint64(time.Now().UnixMilli()) >
		(uint64(*authData.Timestamp)+uint64(*authData.Duration)) {
		return nil, nil, nil, apperror.BadRequest("Authorization duration exceeded")
	}
	logger.Debugf("CurrentStateEE")
	// dataStates := dsquery.NewDataStates(cfg)
	currentState, grantorAuthState, _, err := service.ValidateAuthPayloadData(&payload, cfg, "")
	// if !dataStates.Empty() {
	// 	dataStates.Commit(nil, nil, nil)
	// }
	
	// TODO If error is because the app was not found, check the dht for the app
	if err != nil {
		logger.Error("ValidateuthPayload: ", err)
		return nil, nil, nil, err
	}
	appState, _ = dsquery.GetApplicationStateById(authData.Application)
	
	// generate associations
	if currentState != nil {
		
		assocPrevEvent = &currentState.Event
		// assocPrevEvent = entities.EventPath{
		// 	Relationship: entities.PreviousEventAssoc,
		// 	Hash: currentState.Event,
		// 	Model: entities.AuthorizationEventModel,
		// }.ToString()
	} else {
		// Get the apps state event
		// appState := &models.ApplicationState{}
		//err = query.GetOne(&models.ApplicationState{Application: entities.Application{ID: authData.Application }}, appState)
		
		if appState == nil {
			// find ways to get the app
		} else {
			assocPrevEvent = &appState.Event
		}


	}
	if grantorAuthState != nil {
		assocAuthEvent = &grantorAuthState.Event
		// assocAuthEvent =  entities.EventPath{
		// 	Relationship: entities.AuthorizationEventAssoc,
		// 	Hash: grantorAuthState.Event,
		// 	Model: entities.AuthorizationEventModel,
		// }
	}
	return assocPrevEvent, assocAuthEvent, appState, nil
}

func GetAuthorizations(auth *entities.Authorization) (*[]models.AuthorizationState, error) {
	authState := []models.AuthorizationState{}
	auths, err := dsquery.GetAccountAuthorizations(*auth, dsquery.DefaultQueryLimit, nil)
	if err != nil {
		if dsquery.IsErrorNotFound(err) {
			return &authState, nil
		}
		return &authState, err
	}
	for _, auth := range auths {
		authState = append(authState, models.AuthorizationState{Authorization: *auth})
	}
	return &authState, nil
}
