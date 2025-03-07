package api

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/contracts/evm/app"
	"github.com/mlayerprotocol/go-mlayer/contracts/evm/network"
	"github.com/mlayerprotocol/go-mlayer/contracts/evm/sentry"
	"github.com/mlayerprotocol/go-mlayer/contracts/evm/validator"
	"github.com/mlayerprotocol/go-mlayer/entities"
)





type Protocol string
const (
	Ws Protocol = "ws"
	Https Protocol = "https"
)


type EthereumAPI struct {
	// Protocol Protocol
	EthConfig configs.EthConfig
	sentryContract *sentry.SentryContract
	validatorContract *validator.ValidatorContract
	chainInfoContract *network.NetworkContract
	appContract *app.ApplicationContract
	// chainId *configs.ChainId
	client *ethclient.Client
	signer *[]byte
}



func NewEthAPI(chainId configs.ChainId, ethConfig configs.EthConfig, privateKey *[]byte) (*EthereumAPI, error) {
	if len(ethConfig.Http) == 0 {
		logger.Fatal("Invalid chain config for chainId ", chainId)
	}
	api := EthereumAPI{EthConfig: ethConfig}
	client, err := ethclient.Dial(ethConfig.Http)
    if err != nil {
        return nil, err
    }
	api.client = client
	api.signer = privateKey
	api.chainInfoContract, err = network.NewNetworkContract(common.HexToAddress(ethConfig.ChainInfoContract), api.client)
	if err != nil {
		return nil, err
	}
	
	
	api.sentryContract, err = sentry.NewSentryContract(common.HexToAddress(ethConfig.SentryNodeContract), api.client)
	if err != nil {
		return nil, err
	}

	api.validatorContract, err = validator.NewValidatorContract(common.HexToAddress(ethConfig.ValidatorNodeContract), api.client)
	if err != nil {
		return nil, err
	}
	api.appContract, err = app.NewApplicationContract(common.HexToAddress(ethConfig.ApplicationContract), api.client)
	if err != nil {
		return nil, err
	}
	chainInfo, err := api.GetChainInfo()
	if err != nil {
		return nil, err
	}
	
	if !chainInfo.ChainId.Equals(chainId)  {
		return nil, fmt.Errorf("invalid chain ids")
	}
	return &api, nil
}

func (n EthereumAPI) GetEpoch(blockNumber *big.Int) (*big.Int, error) {
	return n.chainInfoContract.GetEpoch( nil,blockNumber)
	// return new(big.Int).Div(blockNumber , big.NewInt(14400)), nil
}
func (n EthereumAPI) GetCycle(blockNumber *big.Int) (*big.Int, error) {
	return n.chainInfoContract.GetCycle( nil,blockNumber)
}

func (n EthereumAPI) GetCurrentBlockNumber() (*big.Int, error) {
	return n.chainInfoContract.GetCurrentBlockNumber( nil)
}
func (n EthereumAPI) GetSentryActiveLicenseCount(cycle *big.Int)  (*big.Int, error)  {
	return n.sentryContract.GetActiveLicenseCount( nil)
}
func (n EthereumAPI) GetSentryLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)  {
	return n.sentryContract.GetOperatorLicenses( nil, operator)
}
func (n EthereumAPI) GetValidatorLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)  {
	return n.validatorContract.GetOperatorLicenses( nil, operator)
}
func (n EthereumAPI) GetValidatorActiveLicenseCount(cycle *big.Int) (*big.Int, error) {
	return n.validatorContract.GetActiveLicenseCount( nil)
}
func (n EthereumAPI) GetSentryOperatorCycleLicenseCount(operator []byte, cycle  *big.Int) (*big.Int, error) {
	return n.sentryContract.GetOperatorCycleLicenseCount( nil, operator, cycle)
}
func (n EthereumAPI) GetValidatorOperatorCycleLicenseCount(operator []byte, cycle  *big.Int) (*big.Int, error) {
	return n.validatorContract.GetOperatorCycleLicenseCount( nil, operator, cycle)
}
func (n EthereumAPI) GetSentryLicenseOperator(license *big.Int) ([]byte, error) {
	return n.sentryContract.LicenseOperator(nil, license)
}
func (n EthereumAPI) GetValidatorLicenseOperator(license *big.Int) ([]byte, error) {
	return n.validatorContract.LicenseOperator(nil, license)
}

func (n EthereumAPI) IsValidatorNodeOperator(publicKey []byte,  cycle *big.Int) (bool, error) {
	licenseCount, err := n.validatorContract.GetOperatorCycleLicenseCount(nil, publicKey, cycle)
	if err != nil {
		return false, err
	}
	if licenseCount.Cmp(big.NewInt(0)) == 1 {
		return false, nil
	}
	return true, nil
}

func (n EthereumAPI) IsSentryNodeOperator(publicKey []byte, cycle *big.Int) (bool, error) {
	licenseCount, err := n.sentryContract.GetOperatorCycleLicenseCount(nil, publicKey, cycle)
	if err != nil {
		return false, err
	}
	if licenseCount.Cmp(big.NewInt(0)) == 1 {
		return false, nil
	}
	return true, nil
}

func (n EthereumAPI) GetCurrentEpoch() (*big.Int, error) {
	return n.chainInfoContract.GetCurrentEpoch(nil)
}


func (n EthereumAPI) GetCurrentCycle() (*big.Int, error) {
	return n.chainInfoContract.GetCurrentCycle(nil)
}

func (n *EthereumAPI) LicenseOperator(license *big.Int) ([]byte, error) {
	return n.sentryContract.LicenseOperator(nil, license)
}


func (n EthereumAPI) GetCurrentYear() (*big.Int, error) {
	 r, err := n.chainInfoContract.GetStartTime(nil)
	return new(big.Int).Div(r, big.NewInt(int64((((time.Hour)*24*365)+(time.Hour*6))/time.Millisecond))), err
}


func (n EthereumAPI) GetApplicationBalance(id [16]byte) (*big.Int, error) {
	// bal := new(big.Int)
	// bal.SetString("100000000000000000000000000", 10)
	// return bal, nil
	return n.appContract.ApplicationBalance(nil, id)
}

func (n EthereumAPI)  GetValidatorLicenseOwnerAddress(publicKey []byte) ([]byte, error) {
	addr, err := n.validatorContract.OperatorsOwner(nil, publicKey)
	if err != nil {
		return nil, err
	}
	return addr.Bytes(), nil
}

func (n EthereumAPI)  GetSentryLicenseOwnerAddress(publicKey []byte) ([]byte, error) {
	addr, err := n.sentryContract.OperatorsOwner(nil, publicKey)
	if err != nil {
		return nil, err
	}
	return addr.Bytes(), nil
}

func (n EthereumAPI) GetCurrentMessagePrice() (*big.Int, error) {
	return n.chainInfoContract.GetCurrentMessagePrice(nil)
}

func (n EthereumAPI) GetMessagePrice(cycle *big.Int) (*big.Int, error) {
	return n.chainInfoContract.GetMessagePrice(nil, cycle)
}

func (n EthereumAPI) GetValidatorOperatorCount(cycle *big.Int) (*big.Int, error) {
	return big.NewInt(10), nil
}

func (n EthereumAPI) GetSentryOperatorCount(cycle *big.Int) (*big.Int, error) {
	return big.NewInt(10), nil
}

func (n EthereumAPI) GetTotalValueLockedInApplications() (*big.Int, error) {
	return n.appContract.TotalValueLocked(nil)
}

func (n EthereumAPI) GetChainInfo() (info *ChainInfo, err error) {
	chain, err := n.chainInfoContract.GetChainInfo(nil)
	if err != nil {
		return nil, err
	}
	validatorLicenseCount, err := n.validatorContract.GetCycleLicenseCount(nil, chain.CurrentCycle);
	if err != nil {
		return nil, err
	}
	validatorOperatorCount := big.NewInt(10)
	// validatorOperatorCount, err := n.validatorContract.CycleOperatorCount(nil, chain.CurrentCycle);
	// if err != nil {
	// 	return nil, err
	// }
	validatorActiveLicenseCount, err := n.validatorContract.GetCycleActiveLicenseCount(nil, chain.CurrentCycle);
	if err != nil {
		return nil, err
	}
	sentryLicenseCount, err := n.sentryContract.GetCycleLicenseCount(nil, chain.CurrentCycle);
	if err != nil {
		return nil, err
	}
	sentryActiveLicenseCount, err := n.sentryContract.GetCycleActiveLicenseCount(nil, chain.CurrentCycle);
	if err != nil {
		return nil, err
	}
	sentryOperatorCount := big.NewInt(10)
	// sentryOperatorCount, err := n.sentryContract.CycleOperatorCount(nil, chain.CurrentCycle);
	// if err != nil {
	// 	return nil, err
	// }
 return &ChainInfo{
	StartTime: chain.StartTime,
	CurrentBlock: chain.CurrentBlock,
	StartBlock: chain.StartBlock ,
	ChainId: configs.ChainId(fmt.Sprint(chain.ChainId)),
	CurrentCycle: chain.CurrentCycle,
	CurrentEpoch: chain.CurrentEpoch,
	ValidatorLicenseCount: validatorLicenseCount,
	ValidatorActiveLicenseCount: validatorActiveLicenseCount,
	ValidatorOperatorCount: validatorOperatorCount,
	SentryLicenseCount: sentryLicenseCount,
	SentryActiveLicenseCount: sentryActiveLicenseCount,
	SentryOperatorCount: sentryOperatorCount,
	}, err
}
func (n EthereumAPI)  GetStartTime() (*big.Int, error) {
	time, err := n.chainInfoContract.GetStartTime(nil)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Mul(time, big.NewInt(1000)), nil
}
func (n EthereumAPI)  GetStartBlock() (*big.Int, error) {
	return n.chainInfoContract.GetStartBlock(nil)
}

func (n EthereumAPI) ClaimReward(claim *entities.ClaimData) (hash []byte, err error) {
		// Create an authenticated transactor
		privateKey, err := crypto.ToECDSA(*n.signer)
		if err != nil {
			return nil, err
		}
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1)) // Mainnet chain ID
		if err != nil {
			return nil, err
		}
		auth.GasLimit = uint64(300000) // Adjust gas limit if needed
	
		
	signers := []app.LibSecp256k1Point{}
	for _, d := range claim.Signers {
		signers = append(signers, app.LibSecp256k1Point{X: d.X, Y: d.Y})
	}

	counts := []app.ApplicationRewardClaimData{}
	for _, d := range claim.ClaimData {
		counts = append(counts, app.ApplicationRewardClaimData{ApplicationId: [16]byte(utils.UuidToBytes(d.Application)), Amount: new(big.Int).SetBytes(d.Cost)})
	}
	claimData := app.ApplicationClaim{
		Validator: claim.Validator,
		ClaimData: counts,
		Index: claim.Index,
		Cycle:  claim.Cycle,
		Signers: signers,
		Commitment: common.BytesToAddress(claim.Commitment),
		Signature: claim.Signature,
		TotalCost: claim.TotalCost,
	}
	
	abi, err := app.ApplicationContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	data, err := abi.Pack("rewardValidator", claimData) // Function name and parameters
	if err != nil {
		log.Fatalf("Failed to encode method call: %v", err)
	}
	ctx := context.Background()
	address := common.HexToAddress(n.EthConfig.ApplicationContract)
	gasLimit, err := n.client.EstimateGas(ctx, ethereum.CallMsg{
		From: auth.From,      // Sender address
		To:   &address,  // The contract's address
		Gas:  0,       
		Data: data, 
	})
	if err != nil {
		return nil, err
	}
	auth.GasLimit = gasLimit
	gasPrice, err := n.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	auth.GasPrice = gasPrice
	// Suggest the base fee per gas (from the network) and the tip (priority fee)
	gasTipCap, err := n.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	auth.GasTipCap = gasTipCap
	// Suggest the gas fee cap (maxFeePerGas)
	gasFeeCap, err := n.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	auth.GasFeeCap = gasFeeCap
	
	result, err := n.appContract.ClaimReward(auth, claimData)
	
	if err != nil {
		return nil, err
	}
	return result.Hash().Bytes(), nil
}


func  (n EthereumAPI) GetMinStakeAmountForSentry() (*big.Int, error) {
	return new(big.Int), nil
}

func (n EthereumAPI)  GetValidatorNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error) {
	operators, err := n.validatorContract.GetOperators(nil, page, perPage)
	if err != nil {
		return nil, err
	}
	infos := []OperatorInfo{}
	for _, info := range operators {
		infos = append(infos, OperatorInfo{PublicKey: info.PubKey, LicenseOwner: info.Owner.String(), EdaKey: info.EdaKey})
	}
	return infos, nil
}
func(n EthereumAPI)	GetSentryNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error) {
	operators, err := n.sentryContract.GetOperators(nil, page, perPage)
	if err != nil {
		return nil, err
	}
	infos := []OperatorInfo{}
	for _, info := range operators {
		infos = append(infos, OperatorInfo{PublicKey: info.PubKey, LicenseOwner: info.Owner.String()})
	}
	return infos, nil
}

func (n EthereumAPI) Claimed(validator []byte, cycle *big.Int, index *big.Int) (bool, error) {
	return n.appContract.ProcessedClaim(nil, cycle, validator, index)
}

func (n EthereumAPI) IsValidatorLicenseOwner(address string) (bool, error) {
	info, err := n.validatorContract.AccountInfo(nil, common.HexToAddress(address))
	return len(info.Licenses) > 0, err
}
func (n EthereumAPI) IsSentryLicenseOwner(address string)  (bool, error) {
	info, err := n.sentryContract.AccountInfo(nil, common.HexToAddress(address))
	return len(info.Licenses) > 0, err
}