package chain

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain/api"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/exp/rand"
)

type NetworkParams struct {
	StartTime *big.Int 	`json:"start_time"`
	StartBlock *big.Int `json:"start_block"`
	CurrentCycle *big.Int `json:"current_cycle"`
	CurrentBlock *big.Int `json:"current_block"`
	CurrentEpoch *big.Int `json:"current_epoch"`
	ActiveValidatorLicenseCount uint64 `json:"active_validator_license_count"`
	ActiveSentryLicenseCount uint64 `json:"active_sentry_license_count"`
	Validators map[string]string `json:"-"`
	SyncedValidators map[string]multiaddr.Multiaddr `json:"-"`
	Sentries map[string]uint64 `json:"-"`
	Config *configs.MainConfiguration `json:"-"`
	Synced bool `json:"synced"`
	ValidatorOperatorCount *big.Int `json:"validator_count"`
	SentryOperatorCount *big.Int `json:"sentry_count"`
	syncMu sync.Mutex
}

func (n *NetworkParams) GetValidatorKeys(key entities.PublicKeyString) (edd entities.PublicKeyString, secp entities.PublicKeyString) {
	secp = ""
	if len(NetworkInfo.Validators[fmt.Sprintf("secp/%s/edd", key)]) > 0 {
		edd = entities.PublicKeyString(NetworkInfo.Validators[fmt.Sprintf("secp/%s/edd", key)])
		secp = key
	} else {
		if len(NetworkInfo.Validators[fmt.Sprintf("edd/%s/secp", key)]) > 0 {
			secp = entities.PublicKeyString(NetworkInfo.Validators[fmt.Sprintf("edd/%s/secp", key)])
			edd = key
		} 
	}
	return edd, secp
}

func (n *NetworkParams) GetRandomSyncedNode() string {
		rand.Seed(rand.Uint64())
		keys := make([]string, 0, len(n.SyncedValidators))
		for key := range n.SyncedValidators {
			keys = append(keys, key)
		}
		if len(keys) == 0 {
			return ""
		}
		randomKey := keys[rand.Intn(len(keys))]
		return randomKey
}

func (n *NetworkParams) IsValidator(key string) (bool, error) {
	if len(key) == 0 {
		return false, fmt.Errorf("IsValidator: invalid key")
	}
	
	
	if len(n.Validators[key]) > 0 || len(n.Validators[fmt.Sprintf("edd/%s/addr",key)]) > 0 || len(n.Validators[fmt.Sprintf("secp/%s/addr",key)]) > 0 {
		return true, nil
	} 
		// check chain
		// p, err := hex.DecodeString(key)
		// if err != nil {
		// 	return false, err
		// }
		return false, fmt.Errorf("IsValidator: not a validator: %s", key)
		// return DefaultProvider(n.Config).IsValidatorNodeOperator(p, n.CurrentCycle)	
}
func (n *NetworkParams) IsValidatorOwner(address string) (bool, error) {
	if len(n.Validators[address]) > 0 {
		return true, nil
	} 
		// check chain
	
		return DefaultProvider(n.Config).IsValidatorLicenseOwner(address)	
}



func (n *NetworkParams) IsSentry(pubKey string, cycle *big.Int) (bool, error)  {
	if n.Sentries[pubKey] > 0 &&  uint64(time.Now().UnixMicro()) - n.Sentries[pubKey] < uint64(10 * time.Minute.Microseconds()) {
		return true, nil
	}
	b, err := hex.DecodeString(pubKey)
	if err != nil {
		return false, err
	}
	if cycle == nil {
		cycle = n.CurrentCycle
	}
	count, err := DefaultProvider(n.Config).GetSentryOperatorCycleLicenseCount(b, cycle)
	if err != nil {
		return false, err
	}
	hasLicense := count.Cmp(big.NewInt(0)) != 0
	if hasLicense {
		n.Sentries[pubKey] = uint64(time.Now().UnixMicro())
	}
	return count.Cmp(big.NewInt(0)) != 0, nil
}

func (n *NetworkParams) Sync(ctx *context.Context, syncFunc func () bool) {
		n.syncMu.Lock()
		defer n.syncMu.Unlock()
		if n.Synced {
			return 
		}
		n.Synced = syncFunc()
}
var networkInfo NetworkParams 
var NetworkInfo = &networkInfo

var APIs map[configs.ChainId]*api.IChainAPI = map[configs.ChainId]*api.IChainAPI{}

func DefaultProvider(cfg *configs.MainConfiguration) api.IChainAPI {
	return *APIs[cfg.ChainId]
}
// func (n MLChainAPI)  RegisterNetwork(chainId configs.ChainId, api api.IChainAPI) {
// 	n.apis[chainId] = &api
// }

func RegisterProvider (chainId configs.ChainId, api api.IChainAPI) {
	APIs[chainId] = &api
}
func Provider (chainId configs.ChainId) api.IChainAPI {
	return *APIs[chainId]
}


// func (n MLChainAPI) GetEpoch(blockNumber uint64) uint64 {
// 	return blockNumber / 14400
// }
// func GetCycle(blockNumber uint64) uint64 {
// 	cycle := 1 + (GetEpoch(blockNumber) / 30)
// 	return cycle
// }

// func (n *MLChainAPI) GetCurrentBlockNumber() uint64 {
// 	return (uint64(time.Now().UnixMilli()) - NetworkStartDate) / 6000
// }

// func (n *MLChainAPI) GetLicenseCount(cycle uint64) *big.Int {
// 	return big.NewInt(100)
// }

// func (n *MLChainAPI) GetCurrentEpoch() uint64 {
// 	return GetEpoch(n.GetCurrentBlockNumber())
// }


// func (n *MLChainAPI) GetCurrentCycle() uint64 {
// 	return GetCycle(n.GetCurrentBlockNumber())
// }

// func (n *MLChainAPI) LicenceOperator(license *big.Int) ([]byte, error) {
	
// 	return hex.DecodeString("02c4435e768b4bae8236eeba29dd113ed607813b4dc5419d33b9294f712ca79ff4")
// }




// func (n *MLChainAPI) GetCurrentYear() uint64 {
	
// 	return n.GetCurrentCycle() / 12
// }

// func (n *MLChainAPI) Get() big.Int {
// 	bal := new(big.Int)
// 	bal.SetString("1000000000000000", 10)
// 	return *bal
// }

// func (n *MLChainAPI) GetStakeBalance(address entities.AccountString) big.Int {
// 	bal := new(big.Int)
// 	bal.SetString("100000000000000000000000000", 10)
// 	return *bal
// }

// func (n *MLChainAPI) GetApplicationBalance(hashOrId string) big.Int {
// 	bal := new(big.Int)
// 	bal.SetString("100000000000000000000000000", 10)
// 	return *bal
// }

// func (n *MLChainAPI) GetMinStakeAmountForValidators() big.Int {
// 	bal := new(big.Int)
// 	bal.SetString("1000000000000000", 10)
// 	return *bal
// }

// func (n *MLChainAPI) GetCurrentMessageCost() (*big.Int, error) {
// 	bal := new(big.Int)
// 	bal.SetString("10000000", 10)
// 	return bal, nil
// }

// func (n *MLChainAPI) GetMessageCost(cycle uint64) (*big.Int, error) {
// 	bal := new(big.Int)
// 	bal.SetString("10000000", 10)
// 	return bal, nil
// }

// func (n *MLChainAPI) GetChannelBalance(address entities.AccountString) *big.Int {
// 	bal := new(big.Int)
// 	bal.SetString("100000000000000000000000000", 10)
// 	return bal
// }

// func (n *MLChainAPI) ClaimReward(claim entities.ClaimData) (bool, error) {
	
// 	return true, nil
// }

