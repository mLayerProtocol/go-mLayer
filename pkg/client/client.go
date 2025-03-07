package client

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/apperror"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	dsquery "github.com/mlayerprotocol/go-mlayer/internal/ds/query"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p"
)

type NodeInfo struct {
	Account string `json:"account"` 
	NodeType constants.NodeType `json:"node_type"` 
	NodePublicKey string `json:"node_pubkey"` 
	// ChainPublicKey string `json:"chain_pubkey"` 
	ChainId string `json:"chain_id"`
	CurrentCycle uint64 `json:"current_cycle"`
	CurrentBlock uint64 `json:"current_block"`
	CurrentEpoch uint64 `json:"current_epoch"`
	Listeners []string `json:"listeners"`
	Client string `json:"client"`
	ClientVersion string `json:"client_version"`
	ClientReleaseDate string `json:"client_release_date"`
}

func Info(cfg *configs.MainConfiguration) (*NodeInfo, error) {
	provider := chain.Provider(cfg.ChainId)
	info, err := provider.GetChainInfo()
	if err != nil  {
		return  nil, err
	}
	var owner []byte
	if cfg.Validator {
		owner, err = provider.GetValidatorLicenseOwnerAddress(cfg.PublicKeySECP)
	} else {
		owner, err = provider.GetSentryLicenseOwnerAddress(cfg.PublicKeySECP)
	}
	if err != nil {
		return nil, err
	}
	nodeType := constants.ValidatorNodeType
	if !cfg.Validator {
		nodeType = constants.SentryNodeType
	}
	return &NodeInfo{
		Account: hex.EncodeToString(owner),
		NodeType: nodeType,
		NodePublicKey: hex.EncodeToString(cfg.PublicKeySECP),
		//ChainPublicKey: hex.EncodeToString(cfg.PublicKeyEDD),
		ChainId: string(cfg.ChainId),
		Listeners: p2p.GetMultiAddresses(p2p.Host),
		CurrentCycle: info.CurrentCycle.Uint64(),
		CurrentEpoch: info.CurrentEpoch.Uint64(),
		CurrentBlock: info.CurrentBlock.Uint64(),
		Client: "goml",
		ClientVersion: os.Getenv("CLIENT_VERSION"),
		ClientReleaseDate: os.Getenv("RELEASE_DATE"),
	}, nil
	

}
func ValidateClientPayload(
	payload *entities.ClientPayload,
	strictAuth bool,
	cfg *configs.MainConfiguration,
) (*models.AuthorizationState, *entities.DeviceString, error) {
	
	// _, err := payload.EncodeBytes()
	// if err != nil {
	// 	logger.Error(err)
	// 	return nil, apperror.Internal(err.Error())
	// }
	// logger.Debug("ENCODEDBYTESSS"," ", hex.EncodeToString(d), " ", hex.EncodeToString(crypto.Keccak256Hash(d)))

	if payload.Application == ""  {
		return nil, nil, apperror.Forbidden("Application Id is required")
	}
	if string(payload.ChainId) != string(cfg.ChainId) {
		return nil, nil, apperror.Forbidden("Invalid chain Id")
	}
	// payload.ChainId = chainId
	agent, err := payload.GetSigner()
	
	if err != nil {
		return nil, nil, err
	}
	
	// if agent != payload.AppKey {
	// 	return nil, nil, apperror.BadRequest("Agent is required")
	// }
	// logger.Debugf("AGENTTTT %s", agent)
	// app := models.ApplicationState{}
	// err = query.GetOne(models.ApplicationState{Application: entities.Application{ID: payload.Application}}, &app)
	// app , err := dsquery.GetApplicationStateById(payload.Application)
	 app := entities.Application{}
	_, err = service.SyncTypedStateById(payload.Application, &app, cfg, "")
	if err != nil {
		return nil, nil, err
	}
	// if err != nil {
	// 	// if err == gorm.ErrRecordNotFound {
	// 	if dsquery.IsErrorNotFound(err){
	// 		logger.Infof("GettingApplicationFrom %s", payload.Application )
	// 		app, err = service.UpdateApplicationFromPeer(payload.Application, cfg, "" )
	// 		if err != nil || app == nil || app.ID == "" {
	// 			return nil,  nil, apperror.Forbidden("Invalid app id")
	// 		}
	// 	} else {
	// 		return nil, nil, apperror.Internal(err.Error())
	// 	}
	// }
	//app := sub.(entities.Application)
	
	if app.ID == "" {
		return nil, nil, apperror.Forbidden("Invalid app")
	}
	if *app.Status ==  0 {
		return nil, nil, apperror.Forbidden("Application is disabled")
	}

	

	// check if device is authorized
	if agent != ""  {
		
		agent = entities.DeviceString(agent)
		
		if strictAuth || string(payload.Account) != ""  {
			if !agent.IsValid() {
				return nil, nil, apperror.BadRequest("agent address is invalid")
			}
			
			var authData models.AuthorizationState
			// err := query.GetOne(models.AuthorizationState{
			// 	Authorization: entities.Authorization{Account: payload.Account,
			// 		Application: payload.Application,
			// 		Agent:  agent},
			// }, &authData)
			filter := entities.Authorization{
				Account: payload.Account,
				Application: payload.Application,
				Authorized: entities.AddressString(agent),
			}
			
			authDatas, err := dsquery.GetAccountAuthorizations(filter, dsquery.DefaultQueryLimit, nil)
			if err != nil {
				// if err == gorm.ErrRecordNotFound {
				// 	return nil, nil
				// }
				logger.Errorf("GetAuthError: %v", err)
				return nil, &agent, err
			} else {
				if len(authDatas) == 0 {
					return nil, &agent, apperror.Unauthorized("agent not authorized")
				}
				for _, a := range authDatas {
					logger.Debugf("AuthStatesss::: %s", a.Event.ID)
				}
				
				authData = models.AuthorizationState{Authorization: *authDatas[0]};
				logger.Debugf("New Event for Agent/Device %+v, %d, %d", authData, *authData.Duration , *authData.Timestamp)
				if *authData.Duration != 0 && (*authData.Duration + *authData.Timestamp) < uint64(time.Now().UnixMilli()) {
					return nil, &agent,  fmt.Errorf("invalid authdata")
				}
				return &authData, &agent, err
			}
		} else {
			return nil, &agent,  nil
		}
	}
	return nil, nil, apperror.BadRequest("Unable to resolve agent")
}

func SyncRequest(payload *entities.ClientPayload) entities.SyncResponse {
	var response = entities.SyncResponse{}
	return response
}


func validateAgent(payload *entities.ClientPayload) error {
	if len(payload.Account) > 0 &&  len(payload.AppKey) > 0  {
		// check if its an admin
		// auth := models.AuthorizationState{}
		// err = query.GetOneState(entities.Authorization{Agent: payload.AppKey, Account: payload.Account}, &auth)
		auth, err := dsquery.GetAccountAuthorizations(entities.Authorization{
			Authorized: entities.AddressString(payload.AppKey), Account: payload.Account, Application: payload.Application,
		}, nil, nil)
		if err != nil {
			return  apperror.Unauthorized("Invalid app")
		}
		
		if len(auth) == 0 || *auth[0].Priviledge  < constants.MemberPriviledge {
			return  apperror.Unauthorized("agent not authorized")
		}
	}
		return nil
}