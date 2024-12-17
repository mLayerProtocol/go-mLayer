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
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
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
	ds *ds.Datastore,
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

	if payload.Subnet == ""  {
		return nil, nil, apperror.Forbidden("Subnet Id is required")
	}
	if string(payload.ChainId) != string(cfg.ChainId) {
		return nil, nil, apperror.Forbidden("Invalid chain Id")
	}
	// payload.ChainId = chainId
	agent, err := payload.GetSigner()
	
	if err != nil {
		return nil, nil, err
	}
	
	// if agent != payload.Agent {
	// 	return nil, nil, apperror.BadRequest("Agent is required")
	// }
	// logger.Debugf("AGENTTTT %s", agent)
	// subnet := models.SubnetState{}
	// err = query.GetOne(models.SubnetState{Subnet: entities.Subnet{ID: payload.Subnet}}, &subnet)
	// subnet , err := dsquery.GetSubnetStateById(payload.Subnet)
	subnet := &entities.Subnet{}
	_, err = service.SyncTypedStateById(payload.Subnet, subnet, cfg, "")
	// if err != nil {
	// 	// if err == gorm.ErrRecordNotFound {
	// 	if dsquery.IsErrorNotFound(err){
	// 		logger.Infof("GettingSubnetFrom %s", payload.Subnet )
	// 		subnet, err = service.UpdateSubnetFromPeer(payload.Subnet, cfg, "" )
	// 		if err != nil || subnet == nil || subnet.ID == "" {
	// 			return nil,  nil, apperror.Forbidden("Invalid subnet id")
	// 		}
	// 	} else {
	// 		return nil, nil, apperror.Internal(err.Error())
	// 	}
	// }
	if err != nil {
		return nil, nil, err
	}
	if subnet == nil {
		return nil, nil, apperror.Forbidden("Invalid subnet")
	}
	if *subnet.Status ==  0 {
		return nil, nil, apperror.Forbidden("Subnet is disabled")
	}


	// check if device is authorized
	if agent != "" {
		
		agent = entities.DeviceString(agent)
		
		if strictAuth || string(payload.Account) != ""  {
			
			var authData models.AuthorizationState
			// err := query.GetOne(models.AuthorizationState{
			// 	Authorization: entities.Authorization{Account: payload.Account,
			// 		Subnet: payload.Subnet,
			// 		Agent:  agent},
			// }, &authData)
			filter := entities.Authorization{
				Account: payload.Account,
				Subnet: payload.Subnet,
				Agent: agent,
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
					logger.Debugf("AuthStatesss::: %s", a.Event.Hash)
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
	if len(payload.Account) > 0 &&  len(payload.Agent) > 0  {
		// check if its an admin
		// auth := models.AuthorizationState{}
		// err = query.GetOneState(entities.Authorization{Agent: payload.Agent, Account: payload.Account}, &auth)
		auth, err := dsquery.GetAccountAuthorizations(entities.Authorization{
			Agent: payload.Agent, Account: payload.Account, Subnet: payload.Subnet,
		}, nil, nil)
		if err != nil {
			return  apperror.Unauthorized("Invalid subnet")
		}
		
		if len(auth) == 0 || *auth[0].Priviledge  < constants.MemberPriviledge {
			return  apperror.Unauthorized("agent not authorized")
		}
	}
		return nil
}