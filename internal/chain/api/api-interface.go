package api

import (
	"math/big"

	"github.com/mlayerprotocol/go-mlayer/pkg/log"

	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
)

var logger = &log.Logger

type ChainInfo struct {
	StartTime *big.Int
	StartBlock *big.Int
	CurrentBlock *big.Int
	CurrentEpoch *big.Int
	CurrentCycle *big.Int
	ChainId configs.ChainId
	ValidatorLicenseCount *big.Int
	ValidatorActiveLicenseCount *big.Int
	SentryLicenseCount *big.Int
	SentryActiveLicenseCount *big.Int
	ValidatorOperatorCount *big.Int
	SentryOperatorCount *big.Int
}
type OperatorInfo struct {
	PublicKey  []byte
	LicenseOwner string
	EddKey [32]byte
}
type IChainAPI interface {
	// general
	GetChainInfo() (*ChainInfo, error)
	GetStartTime() (*big.Int, error) 
	GetStartBlock() (*big.Int, error) 
	GetEpoch(blockNumber *big.Int) (*big.Int, error) 
	GetCycle(blockNumber *big.Int) (*big.Int, error) 
	GetCurrentCycle() (*big.Int, error) 
	GetCurrentBlockNumber() (*big.Int, error) 
	GetCurrentEpoch() (*big.Int, error) 
	GetCurrentYear() (*big.Int, error) 

	// license
	GetSentryActiveLicenseCount(cycle *big.Int)  (*big.Int, error)
	GetValidatorActiveLicenseCount(cycle *big.Int) (*big.Int, error)
	GetSentryOperatorCycleLicenseCount(operator []byte, cycle *big.Int)  (*big.Int, error)
	GetValidatorOperatorCycleLicenseCount(operator []byte, cycle *big.Int) (*big.Int, error)
	GetSentryLicenseOperator(license *big.Int) ([]byte, error)
	GetValidatorLicenseOperator(license *big.Int) ([]byte, error)


	GetValidatorNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error)
	GetSentryNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error)

	IsValidatorLicenseOwner(address string) (bool, error)
	IsSentryLicenseOwner(address string) (bool, error)

	IsValidatorNodeOperator(publicKey []byte, cycle *big.Int) (bool, error)
	IsSentryNodeOperator(publicKey []byte, cycle *big.Int) (bool, error)

	GetValidatorLicenseOwnerAddress(publicKey []byte) ([]byte, error)
	GetSentryLicenseOwnerAddress(publicKey []byte) ([]byte, error)
	
	
	// GetStakeBalance(address entities.AccountString) big.Int

	// subnet
	GetSubnetBalance(id [16]byte) (*big.Int, error)

	GetTotalValueLockedInSubnets() (*big.Int, error)

	// sentry
	// GetMinStakeAmountForValidators() (*big.Int, error)
	GetMinStakeAmountForSentry() (*big.Int, error)
	GetCurrentMessagePrice() (*big.Int, error)
	GetMessagePrice(cycle *big.Int) (*big.Int, error)
	// GetChannelBalance(address entities.AccountString) *big.Int
	ClaimReward(claim *entities.ClaimData) ([]byte, error) 
	Claimed(validator []byte, cycle *big.Int, index *big.Int) (bool, error) 
	GetSentryLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)
	GetValidatorLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)
	GetValidatorOperatorCount(cycle *big.Int) (*big.Int, error)
	GetSentryOperatorCount(cycle *big.Int) (*big.Int, error)
}


