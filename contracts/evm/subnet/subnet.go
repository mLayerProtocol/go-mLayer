// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package subnet

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// LibSecp256k1Point is an auto generated low-level Go binding around an user-defined struct.
type LibSecp256k1Point struct {
	X *big.Int
	Y *big.Int
}

// SubnetClaim is an auto generated low-level Go binding around an user-defined struct.
type SubnetClaim struct {
	Validator  []byte
	ClaimData  []SubnetRewardClaimData
	Cycle      *big.Int
	Index      *big.Int
	Signers    []LibSecp256k1Point
	Commitment common.Address
	Signature  []byte
	TotalCost  *big.Int
}

// SubnetLicenseInfo is an auto generated low-level Go binding around an user-defined struct.
type SubnetLicenseInfo struct {
	Id          *big.Int
	Earned      *big.Int
	IsValidator bool
	DelegatedTo []byte
}

// SubnetRewardClaimData is an auto generated low-level Go binding around an user-defined struct.
type SubnetRewardClaimData struct {
	SubnetId [16]byte
	Amount   *big.Int
}

// SubnetStakeStruct is an auto generated low-level Go binding around an user-defined struct.
type SubnetStakeStruct struct {
	Amount    *big.Int
	Timestamp *big.Int
}

// SubnetContractMetaData contains all meta data concerning the SubnetContract contract.
var SubnetContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structSubnet.StakeStruct\",\"name\":\"stake\",\"type\":\"tuple\"}],\"name\":\"StakeEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structSubnet.StakeStruct\",\"name\":\"stake\",\"type\":\"tuple\"}],\"name\":\"UnStakeEvent\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"addressInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"prevCycleReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"allTimeReward\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isValidator\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"delegatedTo\",\"type\":\"bytes\"}],\"internalType\":\"structSubnet.LicenseInfo[]\",\"name\":\"info\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"claimable\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes\",\"name\":\"validator\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.RewardClaimData[]\",\"name\":\"claimData\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"cycle\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structLibSecp256k1.Point[]\",\"name\":\"signers\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"commitment\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"totalCost\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.Claim\",\"name\":\"claim\",\"type\":\"tuple\"}],\"name\":\"claimReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bool\",\"name\":\"_enabled\",\"type\":\"bool\"}],\"name\":\"enableWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes\",\"name\":\"validator\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.RewardClaimData[]\",\"name\":\"claimData\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"cycle\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structLibSecp256k1.Point[]\",\"name\":\"signers\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"commitment\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"totalCost\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.Claim\",\"name\":\"claim\",\"type\":\"tuple\"}],\"name\":\"getClaimHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"cycleNumLicences\",\"type\":\"uint256\"}],\"name\":\"getMinSignerCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"getSubnetAccountBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.RewardClaimData[]\",\"name\":\"claimData\",\"type\":\"tuple[]\"}],\"name\":\"hashRewardData\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_network\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"xTokenAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_sentryContract\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_validatorNodeContract\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"locked\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minStakable\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"network\",\"outputs\":[{\"internalType\":\"contractINetwork\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"processedClaim\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"proofProviderRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sentryBaseReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sentryContract\",\"outputs\":[{\"internalType\":\"contractINodeContract\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"sentryCycleRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"sentryLicenseRevenue\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minStakable\",\"type\":\"uint256\"}],\"name\":\"setMinStakable\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"add\",\"type\":\"address\"}],\"name\":\"setNetworkAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_baseReward\",\"type\":\"uint256\"}],\"name\":\"setSentryBaseReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"add\",\"type\":\"address\"}],\"name\":\"setSentryNodeAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"add\",\"type\":\"address\"}],\"name\":\"setTokenAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_baseReward\",\"type\":\"uint256\"}],\"name\":\"setValidatorBaseReward\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"add\",\"type\":\"address\"}],\"name\":\"setValidatorNodeAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"stakeAddresses\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"}],\"name\":\"subnetBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"\",\"type\":\"bytes16\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"subnetBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"\",\"type\":\"bytes16\"}],\"name\":\"subnetCredit\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"\",\"type\":\"bytes16\"}],\"name\":\"subnetDebt\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"\",\"type\":\"bytes16\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"subnetStakerBalances\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalValueLocked\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"}],\"name\":\"unStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"unstakeOrders\",\"outputs\":[{\"internalType\":\"int32\",\"name\":\"\",\"type\":\"int32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validatorBaseReward\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"validatorCycleRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"validatorNodeContract\",\"outputs\":[{\"internalType\":\"contractINodeContract\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"validatorRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes\",\"name\":\"validator\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"bytes16\",\"name\":\"subnetId\",\"type\":\"bytes16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.RewardClaimData[]\",\"name\":\"claimData\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"cycle\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"x\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"y\",\"type\":\"uint256\"}],\"internalType\":\"structLibSecp256k1.Point[]\",\"name\":\"signers\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"commitment\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"totalCost\",\"type\":\"uint256\"}],\"internalType\":\"structSubnet.Claim\",\"name\":\"claim\",\"type\":\"tuple\"}],\"name\":\"verifyClaim\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawableAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawalEnabled\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// SubnetContractABI is the input ABI used to generate the binding from.
// Deprecated: Use SubnetContractMetaData.ABI instead.
var SubnetContractABI = SubnetContractMetaData.ABI

// SubnetContract is an auto generated Go binding around an Ethereum contract.
type SubnetContract struct {
	SubnetContractCaller     // Read-only binding to the contract
	SubnetContractTransactor // Write-only binding to the contract
	SubnetContractFilterer   // Log filterer for contract events
}

// SubnetContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type SubnetContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SubnetContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SubnetContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SubnetContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SubnetContractSession struct {
	Contract     *SubnetContract   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SubnetContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SubnetContractCallerSession struct {
	Contract *SubnetContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SubnetContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SubnetContractTransactorSession struct {
	Contract     *SubnetContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SubnetContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type SubnetContractRaw struct {
	Contract *SubnetContract // Generic contract binding to access the raw methods on
}

// SubnetContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SubnetContractCallerRaw struct {
	Contract *SubnetContractCaller // Generic read-only contract binding to access the raw methods on
}

// SubnetContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SubnetContractTransactorRaw struct {
	Contract *SubnetContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSubnetContract creates a new instance of SubnetContract, bound to a specific deployed contract.
func NewSubnetContract(address common.Address, backend bind.ContractBackend) (*SubnetContract, error) {
	contract, err := bindSubnetContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SubnetContract{SubnetContractCaller: SubnetContractCaller{contract: contract}, SubnetContractTransactor: SubnetContractTransactor{contract: contract}, SubnetContractFilterer: SubnetContractFilterer{contract: contract}}, nil
}

// NewSubnetContractCaller creates a new read-only instance of SubnetContract, bound to a specific deployed contract.
func NewSubnetContractCaller(address common.Address, caller bind.ContractCaller) (*SubnetContractCaller, error) {
	contract, err := bindSubnetContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetContractCaller{contract: contract}, nil
}

// NewSubnetContractTransactor creates a new write-only instance of SubnetContract, bound to a specific deployed contract.
func NewSubnetContractTransactor(address common.Address, transactor bind.ContractTransactor) (*SubnetContractTransactor, error) {
	contract, err := bindSubnetContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SubnetContractTransactor{contract: contract}, nil
}

// NewSubnetContractFilterer creates a new log filterer instance of SubnetContract, bound to a specific deployed contract.
func NewSubnetContractFilterer(address common.Address, filterer bind.ContractFilterer) (*SubnetContractFilterer, error) {
	contract, err := bindSubnetContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SubnetContractFilterer{contract: contract}, nil
}

// bindSubnetContract binds a generic wrapper to an already deployed contract.
func bindSubnetContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SubnetContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetContract *SubnetContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetContract.Contract.SubnetContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetContract *SubnetContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetContract.Contract.SubnetContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetContract *SubnetContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetContract.Contract.SubnetContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SubnetContract *SubnetContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SubnetContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SubnetContract *SubnetContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SubnetContract *SubnetContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SubnetContract.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SubnetContract *SubnetContractCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SubnetContract *SubnetContractSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _SubnetContract.Contract.DEFAULTADMINROLE(&_SubnetContract.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_SubnetContract *SubnetContractCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _SubnetContract.Contract.DEFAULTADMINROLE(&_SubnetContract.CallOpts)
}

// AddressInfo is a free data retrieval call binding the contract method 0x2126fcb2.
//
// Solidity: function addressInfo(address addr) view returns(uint256 prevCycleReward, uint256 allTimeReward, (uint256,uint256,bool,bytes)[] info, uint256 claimable)
func (_SubnetContract *SubnetContractCaller) AddressInfo(opts *bind.CallOpts, addr common.Address) (struct {
	PrevCycleReward *big.Int
	AllTimeReward   *big.Int
	Info            []SubnetLicenseInfo
	Claimable       *big.Int
}, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "addressInfo", addr)

	outstruct := new(struct {
		PrevCycleReward *big.Int
		AllTimeReward   *big.Int
		Info            []SubnetLicenseInfo
		Claimable       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PrevCycleReward = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.AllTimeReward = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Info = *abi.ConvertType(out[2], new([]SubnetLicenseInfo)).(*[]SubnetLicenseInfo)
	outstruct.Claimable = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// AddressInfo is a free data retrieval call binding the contract method 0x2126fcb2.
//
// Solidity: function addressInfo(address addr) view returns(uint256 prevCycleReward, uint256 allTimeReward, (uint256,uint256,bool,bytes)[] info, uint256 claimable)
func (_SubnetContract *SubnetContractSession) AddressInfo(addr common.Address) (struct {
	PrevCycleReward *big.Int
	AllTimeReward   *big.Int
	Info            []SubnetLicenseInfo
	Claimable       *big.Int
}, error) {
	return _SubnetContract.Contract.AddressInfo(&_SubnetContract.CallOpts, addr)
}

// AddressInfo is a free data retrieval call binding the contract method 0x2126fcb2.
//
// Solidity: function addressInfo(address addr) view returns(uint256 prevCycleReward, uint256 allTimeReward, (uint256,uint256,bool,bytes)[] info, uint256 claimable)
func (_SubnetContract *SubnetContractCallerSession) AddressInfo(addr common.Address) (struct {
	PrevCycleReward *big.Int
	AllTimeReward   *big.Int
	Info            []SubnetLicenseInfo
	Claimable       *big.Int
}, error) {
	return _SubnetContract.Contract.AddressInfo(&_SubnetContract.CallOpts, addr)
}

// GetClaimHash is a free data retrieval call binding the contract method 0xf6b85a9a.
//
// Solidity: function getClaimHash((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bytes32)
func (_SubnetContract *SubnetContractCaller) GetClaimHash(opts *bind.CallOpts, claim SubnetClaim) ([32]byte, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "getClaimHash", claim)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetClaimHash is a free data retrieval call binding the contract method 0xf6b85a9a.
//
// Solidity: function getClaimHash((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bytes32)
func (_SubnetContract *SubnetContractSession) GetClaimHash(claim SubnetClaim) ([32]byte, error) {
	return _SubnetContract.Contract.GetClaimHash(&_SubnetContract.CallOpts, claim)
}

// GetClaimHash is a free data retrieval call binding the contract method 0xf6b85a9a.
//
// Solidity: function getClaimHash((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bytes32)
func (_SubnetContract *SubnetContractCallerSession) GetClaimHash(claim SubnetClaim) ([32]byte, error) {
	return _SubnetContract.Contract.GetClaimHash(&_SubnetContract.CallOpts, claim)
}

// GetMinSignerCount is a free data retrieval call binding the contract method 0x9cee2fe3.
//
// Solidity: function getMinSignerCount(uint256 cycleNumLicences) pure returns(uint256)
func (_SubnetContract *SubnetContractCaller) GetMinSignerCount(opts *bind.CallOpts, cycleNumLicences *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "getMinSignerCount", cycleNumLicences)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinSignerCount is a free data retrieval call binding the contract method 0x9cee2fe3.
//
// Solidity: function getMinSignerCount(uint256 cycleNumLicences) pure returns(uint256)
func (_SubnetContract *SubnetContractSession) GetMinSignerCount(cycleNumLicences *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.GetMinSignerCount(&_SubnetContract.CallOpts, cycleNumLicences)
}

// GetMinSignerCount is a free data retrieval call binding the contract method 0x9cee2fe3.
//
// Solidity: function getMinSignerCount(uint256 cycleNumLicences) pure returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) GetMinSignerCount(cycleNumLicences *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.GetMinSignerCount(&_SubnetContract.CallOpts, cycleNumLicences)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SubnetContract *SubnetContractCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SubnetContract *SubnetContractSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _SubnetContract.Contract.GetRoleAdmin(&_SubnetContract.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_SubnetContract *SubnetContractCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _SubnetContract.Contract.GetRoleAdmin(&_SubnetContract.CallOpts, role)
}

// GetSubnetAccountBalance is a free data retrieval call binding the contract method 0x09774aaf.
//
// Solidity: function getSubnetAccountBalance(bytes16 subnetId, address addr) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) GetSubnetAccountBalance(opts *bind.CallOpts, subnetId [16]byte, addr common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "getSubnetAccountBalance", subnetId, addr)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSubnetAccountBalance is a free data retrieval call binding the contract method 0x09774aaf.
//
// Solidity: function getSubnetAccountBalance(bytes16 subnetId, address addr) view returns(uint256)
func (_SubnetContract *SubnetContractSession) GetSubnetAccountBalance(subnetId [16]byte, addr common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.GetSubnetAccountBalance(&_SubnetContract.CallOpts, subnetId, addr)
}

// GetSubnetAccountBalance is a free data retrieval call binding the contract method 0x09774aaf.
//
// Solidity: function getSubnetAccountBalance(bytes16 subnetId, address addr) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) GetSubnetAccountBalance(subnetId [16]byte, addr common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.GetSubnetAccountBalance(&_SubnetContract.CallOpts, subnetId, addr)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SubnetContract *SubnetContractCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SubnetContract *SubnetContractSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _SubnetContract.Contract.HasRole(&_SubnetContract.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_SubnetContract *SubnetContractCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _SubnetContract.Contract.HasRole(&_SubnetContract.CallOpts, role, account)
}

// HashRewardData is a free data retrieval call binding the contract method 0x5a9a7b87.
//
// Solidity: function hashRewardData((bytes16,uint256)[] claimData) pure returns(bytes32 hash)
func (_SubnetContract *SubnetContractCaller) HashRewardData(opts *bind.CallOpts, claimData []SubnetRewardClaimData) ([32]byte, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "hashRewardData", claimData)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashRewardData is a free data retrieval call binding the contract method 0x5a9a7b87.
//
// Solidity: function hashRewardData((bytes16,uint256)[] claimData) pure returns(bytes32 hash)
func (_SubnetContract *SubnetContractSession) HashRewardData(claimData []SubnetRewardClaimData) ([32]byte, error) {
	return _SubnetContract.Contract.HashRewardData(&_SubnetContract.CallOpts, claimData)
}

// HashRewardData is a free data retrieval call binding the contract method 0x5a9a7b87.
//
// Solidity: function hashRewardData((bytes16,uint256)[] claimData) pure returns(bytes32 hash)
func (_SubnetContract *SubnetContractCallerSession) HashRewardData(claimData []SubnetRewardClaimData) ([32]byte, error) {
	return _SubnetContract.Contract.HashRewardData(&_SubnetContract.CallOpts, claimData)
}

// Locked is a free data retrieval call binding the contract method 0xcf309012.
//
// Solidity: function locked() view returns(bool)
func (_SubnetContract *SubnetContractCaller) Locked(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "locked")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Locked is a free data retrieval call binding the contract method 0xcf309012.
//
// Solidity: function locked() view returns(bool)
func (_SubnetContract *SubnetContractSession) Locked() (bool, error) {
	return _SubnetContract.Contract.Locked(&_SubnetContract.CallOpts)
}

// Locked is a free data retrieval call binding the contract method 0xcf309012.
//
// Solidity: function locked() view returns(bool)
func (_SubnetContract *SubnetContractCallerSession) Locked() (bool, error) {
	return _SubnetContract.Contract.Locked(&_SubnetContract.CallOpts)
}

// MinStakable is a free data retrieval call binding the contract method 0xc0d41476.
//
// Solidity: function minStakable() view returns(uint256)
func (_SubnetContract *SubnetContractCaller) MinStakable(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "minStakable")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStakable is a free data retrieval call binding the contract method 0xc0d41476.
//
// Solidity: function minStakable() view returns(uint256)
func (_SubnetContract *SubnetContractSession) MinStakable() (*big.Int, error) {
	return _SubnetContract.Contract.MinStakable(&_SubnetContract.CallOpts)
}

// MinStakable is a free data retrieval call binding the contract method 0xc0d41476.
//
// Solidity: function minStakable() view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) MinStakable() (*big.Int, error) {
	return _SubnetContract.Contract.MinStakable(&_SubnetContract.CallOpts)
}

// Network is a free data retrieval call binding the contract method 0x6739afca.
//
// Solidity: function network() view returns(address)
func (_SubnetContract *SubnetContractCaller) Network(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "network")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Network is a free data retrieval call binding the contract method 0x6739afca.
//
// Solidity: function network() view returns(address)
func (_SubnetContract *SubnetContractSession) Network() (common.Address, error) {
	return _SubnetContract.Contract.Network(&_SubnetContract.CallOpts)
}

// Network is a free data retrieval call binding the contract method 0x6739afca.
//
// Solidity: function network() view returns(address)
func (_SubnetContract *SubnetContractCallerSession) Network() (common.Address, error) {
	return _SubnetContract.Contract.Network(&_SubnetContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetContract *SubnetContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetContract *SubnetContractSession) Owner() (common.Address, error) {
	return _SubnetContract.Contract.Owner(&_SubnetContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SubnetContract *SubnetContractCallerSession) Owner() (common.Address, error) {
	return _SubnetContract.Contract.Owner(&_SubnetContract.CallOpts)
}

// ProcessedClaim is a free data retrieval call binding the contract method 0x2ad6f7f6.
//
// Solidity: function processedClaim(uint256 , bytes , uint256 ) view returns(bool)
func (_SubnetContract *SubnetContractCaller) ProcessedClaim(opts *bind.CallOpts, arg0 *big.Int, arg1 []byte, arg2 *big.Int) (bool, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "processedClaim", arg0, arg1, arg2)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ProcessedClaim is a free data retrieval call binding the contract method 0x2ad6f7f6.
//
// Solidity: function processedClaim(uint256 , bytes , uint256 ) view returns(bool)
func (_SubnetContract *SubnetContractSession) ProcessedClaim(arg0 *big.Int, arg1 []byte, arg2 *big.Int) (bool, error) {
	return _SubnetContract.Contract.ProcessedClaim(&_SubnetContract.CallOpts, arg0, arg1, arg2)
}

// ProcessedClaim is a free data retrieval call binding the contract method 0x2ad6f7f6.
//
// Solidity: function processedClaim(uint256 , bytes , uint256 ) view returns(bool)
func (_SubnetContract *SubnetContractCallerSession) ProcessedClaim(arg0 *big.Int, arg1 []byte, arg2 *big.Int) (bool, error) {
	return _SubnetContract.Contract.ProcessedClaim(&_SubnetContract.CallOpts, arg0, arg1, arg2)
}

// ProofProviderRewards is a free data retrieval call binding the contract method 0x58f25d13.
//
// Solidity: function proofProviderRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) ProofProviderRewards(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "proofProviderRewards", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProofProviderRewards is a free data retrieval call binding the contract method 0x58f25d13.
//
// Solidity: function proofProviderRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) ProofProviderRewards(arg0 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.ProofProviderRewards(&_SubnetContract.CallOpts, arg0)
}

// ProofProviderRewards is a free data retrieval call binding the contract method 0x58f25d13.
//
// Solidity: function proofProviderRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) ProofProviderRewards(arg0 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.ProofProviderRewards(&_SubnetContract.CallOpts, arg0)
}

// SentryBaseReward is a free data retrieval call binding the contract method 0x25a772a4.
//
// Solidity: function sentryBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SentryBaseReward(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "sentryBaseReward")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SentryBaseReward is a free data retrieval call binding the contract method 0x25a772a4.
//
// Solidity: function sentryBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractSession) SentryBaseReward() (*big.Int, error) {
	return _SubnetContract.Contract.SentryBaseReward(&_SubnetContract.CallOpts)
}

// SentryBaseReward is a free data retrieval call binding the contract method 0x25a772a4.
//
// Solidity: function sentryBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SentryBaseReward() (*big.Int, error) {
	return _SubnetContract.Contract.SentryBaseReward(&_SubnetContract.CallOpts)
}

// SentryContract is a free data retrieval call binding the contract method 0xdbca5c52.
//
// Solidity: function sentryContract() view returns(address)
func (_SubnetContract *SubnetContractCaller) SentryContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "sentryContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SentryContract is a free data retrieval call binding the contract method 0xdbca5c52.
//
// Solidity: function sentryContract() view returns(address)
func (_SubnetContract *SubnetContractSession) SentryContract() (common.Address, error) {
	return _SubnetContract.Contract.SentryContract(&_SubnetContract.CallOpts)
}

// SentryContract is a free data retrieval call binding the contract method 0xdbca5c52.
//
// Solidity: function sentryContract() view returns(address)
func (_SubnetContract *SubnetContractCallerSession) SentryContract() (common.Address, error) {
	return _SubnetContract.Contract.SentryContract(&_SubnetContract.CallOpts)
}

// SentryCycleRewards is a free data retrieval call binding the contract method 0x7773b6e9.
//
// Solidity: function sentryCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SentryCycleRewards(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "sentryCycleRewards", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SentryCycleRewards is a free data retrieval call binding the contract method 0x7773b6e9.
//
// Solidity: function sentryCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SentryCycleRewards(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.SentryCycleRewards(&_SubnetContract.CallOpts, arg0, arg1)
}

// SentryCycleRewards is a free data retrieval call binding the contract method 0x7773b6e9.
//
// Solidity: function sentryCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SentryCycleRewards(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.SentryCycleRewards(&_SubnetContract.CallOpts, arg0, arg1)
}

// SentryLicenseRevenue is a free data retrieval call binding the contract method 0x37b560f3.
//
// Solidity: function sentryLicenseRevenue(uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SentryLicenseRevenue(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "sentryLicenseRevenue", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SentryLicenseRevenue is a free data retrieval call binding the contract method 0x37b560f3.
//
// Solidity: function sentryLicenseRevenue(uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SentryLicenseRevenue(arg0 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.SentryLicenseRevenue(&_SubnetContract.CallOpts, arg0)
}

// SentryLicenseRevenue is a free data retrieval call binding the contract method 0x37b560f3.
//
// Solidity: function sentryLicenseRevenue(uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SentryLicenseRevenue(arg0 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.SentryLicenseRevenue(&_SubnetContract.CallOpts, arg0)
}

// StakeAddresses is a free data retrieval call binding the contract method 0x6a7cce86.
//
// Solidity: function stakeAddresses(address ) view returns(address)
func (_SubnetContract *SubnetContractCaller) StakeAddresses(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "stakeAddresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeAddresses is a free data retrieval call binding the contract method 0x6a7cce86.
//
// Solidity: function stakeAddresses(address ) view returns(address)
func (_SubnetContract *SubnetContractSession) StakeAddresses(arg0 common.Address) (common.Address, error) {
	return _SubnetContract.Contract.StakeAddresses(&_SubnetContract.CallOpts, arg0)
}

// StakeAddresses is a free data retrieval call binding the contract method 0x6a7cce86.
//
// Solidity: function stakeAddresses(address ) view returns(address)
func (_SubnetContract *SubnetContractCallerSession) StakeAddresses(arg0 common.Address) (common.Address, error) {
	return _SubnetContract.Contract.StakeAddresses(&_SubnetContract.CallOpts, arg0)
}

// SubnetBalance is a free data retrieval call binding the contract method 0x4d1e4062.
//
// Solidity: function subnetBalance(bytes16 subnetId) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SubnetBalance(opts *bind.CallOpts, subnetId [16]byte) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "subnetBalance", subnetId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubnetBalance is a free data retrieval call binding the contract method 0x4d1e4062.
//
// Solidity: function subnetBalance(bytes16 subnetId) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SubnetBalance(subnetId [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetBalance(&_SubnetContract.CallOpts, subnetId)
}

// SubnetBalance is a free data retrieval call binding the contract method 0x4d1e4062.
//
// Solidity: function subnetBalance(bytes16 subnetId) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SubnetBalance(subnetId [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetBalance(&_SubnetContract.CallOpts, subnetId)
}

// SubnetBalances is a free data retrieval call binding the contract method 0x5a50b09b.
//
// Solidity: function subnetBalances(bytes16 , address , uint256 ) view returns(uint256 amount, uint256 timestamp)
func (_SubnetContract *SubnetContractCaller) SubnetBalances(opts *bind.CallOpts, arg0 [16]byte, arg1 common.Address, arg2 *big.Int) (struct {
	Amount    *big.Int
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "subnetBalances", arg0, arg1, arg2)

	outstruct := new(struct {
		Amount    *big.Int
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Timestamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// SubnetBalances is a free data retrieval call binding the contract method 0x5a50b09b.
//
// Solidity: function subnetBalances(bytes16 , address , uint256 ) view returns(uint256 amount, uint256 timestamp)
func (_SubnetContract *SubnetContractSession) SubnetBalances(arg0 [16]byte, arg1 common.Address, arg2 *big.Int) (struct {
	Amount    *big.Int
	Timestamp *big.Int
}, error) {
	return _SubnetContract.Contract.SubnetBalances(&_SubnetContract.CallOpts, arg0, arg1, arg2)
}

// SubnetBalances is a free data retrieval call binding the contract method 0x5a50b09b.
//
// Solidity: function subnetBalances(bytes16 , address , uint256 ) view returns(uint256 amount, uint256 timestamp)
func (_SubnetContract *SubnetContractCallerSession) SubnetBalances(arg0 [16]byte, arg1 common.Address, arg2 *big.Int) (struct {
	Amount    *big.Int
	Timestamp *big.Int
}, error) {
	return _SubnetContract.Contract.SubnetBalances(&_SubnetContract.CallOpts, arg0, arg1, arg2)
}

// SubnetCredit is a free data retrieval call binding the contract method 0x8d73c2fa.
//
// Solidity: function subnetCredit(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SubnetCredit(opts *bind.CallOpts, arg0 [16]byte) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "subnetCredit", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubnetCredit is a free data retrieval call binding the contract method 0x8d73c2fa.
//
// Solidity: function subnetCredit(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SubnetCredit(arg0 [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetCredit(&_SubnetContract.CallOpts, arg0)
}

// SubnetCredit is a free data retrieval call binding the contract method 0x8d73c2fa.
//
// Solidity: function subnetCredit(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SubnetCredit(arg0 [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetCredit(&_SubnetContract.CallOpts, arg0)
}

// SubnetDebt is a free data retrieval call binding the contract method 0xe9af524e.
//
// Solidity: function subnetDebt(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SubnetDebt(opts *bind.CallOpts, arg0 [16]byte) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "subnetDebt", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubnetDebt is a free data retrieval call binding the contract method 0xe9af524e.
//
// Solidity: function subnetDebt(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SubnetDebt(arg0 [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetDebt(&_SubnetContract.CallOpts, arg0)
}

// SubnetDebt is a free data retrieval call binding the contract method 0xe9af524e.
//
// Solidity: function subnetDebt(bytes16 ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SubnetDebt(arg0 [16]byte) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetDebt(&_SubnetContract.CallOpts, arg0)
}

// SubnetStakerBalances is a free data retrieval call binding the contract method 0x78d25664.
//
// Solidity: function subnetStakerBalances(bytes16 , address ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) SubnetStakerBalances(opts *bind.CallOpts, arg0 [16]byte, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "subnetStakerBalances", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubnetStakerBalances is a free data retrieval call binding the contract method 0x78d25664.
//
// Solidity: function subnetStakerBalances(bytes16 , address ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) SubnetStakerBalances(arg0 [16]byte, arg1 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetStakerBalances(&_SubnetContract.CallOpts, arg0, arg1)
}

// SubnetStakerBalances is a free data retrieval call binding the contract method 0x78d25664.
//
// Solidity: function subnetStakerBalances(bytes16 , address ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) SubnetStakerBalances(arg0 [16]byte, arg1 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.SubnetStakerBalances(&_SubnetContract.CallOpts, arg0, arg1)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetContract *SubnetContractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetContract *SubnetContractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetContract.Contract.SupportsInterface(&_SubnetContract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_SubnetContract *SubnetContractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _SubnetContract.Contract.SupportsInterface(&_SubnetContract.CallOpts, interfaceId)
}

// TotalValueLocked is a free data retrieval call binding the contract method 0xec18154e.
//
// Solidity: function totalValueLocked() view returns(uint256)
func (_SubnetContract *SubnetContractCaller) TotalValueLocked(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "totalValueLocked")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalValueLocked is a free data retrieval call binding the contract method 0xec18154e.
//
// Solidity: function totalValueLocked() view returns(uint256)
func (_SubnetContract *SubnetContractSession) TotalValueLocked() (*big.Int, error) {
	return _SubnetContract.Contract.TotalValueLocked(&_SubnetContract.CallOpts)
}

// TotalValueLocked is a free data retrieval call binding the contract method 0xec18154e.
//
// Solidity: function totalValueLocked() view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) TotalValueLocked() (*big.Int, error) {
	return _SubnetContract.Contract.TotalValueLocked(&_SubnetContract.CallOpts)
}

// UnstakeOrders is a free data retrieval call binding the contract method 0x9d922aca.
//
// Solidity: function unstakeOrders(address , bytes ) view returns(int32)
func (_SubnetContract *SubnetContractCaller) UnstakeOrders(opts *bind.CallOpts, arg0 common.Address, arg1 []byte) (int32, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "unstakeOrders", arg0, arg1)

	if err != nil {
		return *new(int32), err
	}

	out0 := *abi.ConvertType(out[0], new(int32)).(*int32)

	return out0, err

}

// UnstakeOrders is a free data retrieval call binding the contract method 0x9d922aca.
//
// Solidity: function unstakeOrders(address , bytes ) view returns(int32)
func (_SubnetContract *SubnetContractSession) UnstakeOrders(arg0 common.Address, arg1 []byte) (int32, error) {
	return _SubnetContract.Contract.UnstakeOrders(&_SubnetContract.CallOpts, arg0, arg1)
}

// UnstakeOrders is a free data retrieval call binding the contract method 0x9d922aca.
//
// Solidity: function unstakeOrders(address , bytes ) view returns(int32)
func (_SubnetContract *SubnetContractCallerSession) UnstakeOrders(arg0 common.Address, arg1 []byte) (int32, error) {
	return _SubnetContract.Contract.UnstakeOrders(&_SubnetContract.CallOpts, arg0, arg1)
}

// ValidatorBaseReward is a free data retrieval call binding the contract method 0x0063b8dd.
//
// Solidity: function validatorBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractCaller) ValidatorBaseReward(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "validatorBaseReward")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidatorBaseReward is a free data retrieval call binding the contract method 0x0063b8dd.
//
// Solidity: function validatorBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractSession) ValidatorBaseReward() (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorBaseReward(&_SubnetContract.CallOpts)
}

// ValidatorBaseReward is a free data retrieval call binding the contract method 0x0063b8dd.
//
// Solidity: function validatorBaseReward() view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) ValidatorBaseReward() (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorBaseReward(&_SubnetContract.CallOpts)
}

// ValidatorCycleRewards is a free data retrieval call binding the contract method 0xbf19eb0c.
//
// Solidity: function validatorCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) ValidatorCycleRewards(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "validatorCycleRewards", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidatorCycleRewards is a free data retrieval call binding the contract method 0xbf19eb0c.
//
// Solidity: function validatorCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) ValidatorCycleRewards(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorCycleRewards(&_SubnetContract.CallOpts, arg0, arg1)
}

// ValidatorCycleRewards is a free data retrieval call binding the contract method 0xbf19eb0c.
//
// Solidity: function validatorCycleRewards(address , uint256 ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) ValidatorCycleRewards(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorCycleRewards(&_SubnetContract.CallOpts, arg0, arg1)
}

// ValidatorNodeContract is a free data retrieval call binding the contract method 0x71f2053b.
//
// Solidity: function validatorNodeContract() view returns(address)
func (_SubnetContract *SubnetContractCaller) ValidatorNodeContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "validatorNodeContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ValidatorNodeContract is a free data retrieval call binding the contract method 0x71f2053b.
//
// Solidity: function validatorNodeContract() view returns(address)
func (_SubnetContract *SubnetContractSession) ValidatorNodeContract() (common.Address, error) {
	return _SubnetContract.Contract.ValidatorNodeContract(&_SubnetContract.CallOpts)
}

// ValidatorNodeContract is a free data retrieval call binding the contract method 0x71f2053b.
//
// Solidity: function validatorNodeContract() view returns(address)
func (_SubnetContract *SubnetContractCallerSession) ValidatorNodeContract() (common.Address, error) {
	return _SubnetContract.Contract.ValidatorNodeContract(&_SubnetContract.CallOpts)
}

// ValidatorRewards is a free data retrieval call binding the contract method 0xb1845c56.
//
// Solidity: function validatorRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractCaller) ValidatorRewards(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "validatorRewards", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ValidatorRewards is a free data retrieval call binding the contract method 0xb1845c56.
//
// Solidity: function validatorRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractSession) ValidatorRewards(arg0 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorRewards(&_SubnetContract.CallOpts, arg0)
}

// ValidatorRewards is a free data retrieval call binding the contract method 0xb1845c56.
//
// Solidity: function validatorRewards(address ) view returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) ValidatorRewards(arg0 common.Address) (*big.Int, error) {
	return _SubnetContract.Contract.ValidatorRewards(&_SubnetContract.CallOpts, arg0)
}

// VerifyClaim is a free data retrieval call binding the contract method 0x8b2ed4c8.
//
// Solidity: function verifyClaim((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bool, bytes32)
func (_SubnetContract *SubnetContractCaller) VerifyClaim(opts *bind.CallOpts, claim SubnetClaim) (bool, [32]byte, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "verifyClaim", claim)

	if err != nil {
		return *new(bool), *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)

	return out0, out1, err

}

// VerifyClaim is a free data retrieval call binding the contract method 0x8b2ed4c8.
//
// Solidity: function verifyClaim((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bool, bytes32)
func (_SubnetContract *SubnetContractSession) VerifyClaim(claim SubnetClaim) (bool, [32]byte, error) {
	return _SubnetContract.Contract.VerifyClaim(&_SubnetContract.CallOpts, claim)
}

// VerifyClaim is a free data retrieval call binding the contract method 0x8b2ed4c8.
//
// Solidity: function verifyClaim((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) view returns(bool, bytes32)
func (_SubnetContract *SubnetContractCallerSession) VerifyClaim(claim SubnetClaim) (bool, [32]byte, error) {
	return _SubnetContract.Contract.VerifyClaim(&_SubnetContract.CallOpts, claim)
}

// WithdrawableAmount is a free data retrieval call binding the contract method 0x951303f5.
//
// Solidity: function withdrawableAmount() pure returns(uint256)
func (_SubnetContract *SubnetContractCaller) WithdrawableAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "withdrawableAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WithdrawableAmount is a free data retrieval call binding the contract method 0x951303f5.
//
// Solidity: function withdrawableAmount() pure returns(uint256)
func (_SubnetContract *SubnetContractSession) WithdrawableAmount() (*big.Int, error) {
	return _SubnetContract.Contract.WithdrawableAmount(&_SubnetContract.CallOpts)
}

// WithdrawableAmount is a free data retrieval call binding the contract method 0x951303f5.
//
// Solidity: function withdrawableAmount() pure returns(uint256)
func (_SubnetContract *SubnetContractCallerSession) WithdrawableAmount() (*big.Int, error) {
	return _SubnetContract.Contract.WithdrawableAmount(&_SubnetContract.CallOpts)
}

// WithdrawalEnabled is a free data retrieval call binding the contract method 0xf8ea5daf.
//
// Solidity: function withdrawalEnabled() view returns(bool)
func (_SubnetContract *SubnetContractCaller) WithdrawalEnabled(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _SubnetContract.contract.Call(opts, &out, "withdrawalEnabled")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// WithdrawalEnabled is a free data retrieval call binding the contract method 0xf8ea5daf.
//
// Solidity: function withdrawalEnabled() view returns(bool)
func (_SubnetContract *SubnetContractSession) WithdrawalEnabled() (bool, error) {
	return _SubnetContract.Contract.WithdrawalEnabled(&_SubnetContract.CallOpts)
}

// WithdrawalEnabled is a free data retrieval call binding the contract method 0xf8ea5daf.
//
// Solidity: function withdrawalEnabled() view returns(bool)
func (_SubnetContract *SubnetContractCallerSession) WithdrawalEnabled() (bool, error) {
	return _SubnetContract.Contract.WithdrawalEnabled(&_SubnetContract.CallOpts)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x25f3a9c0.
//
// Solidity: function claimReward((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) returns()
func (_SubnetContract *SubnetContractTransactor) ClaimReward(opts *bind.TransactOpts, claim SubnetClaim) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "claimReward", claim)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x25f3a9c0.
//
// Solidity: function claimReward((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) returns()
func (_SubnetContract *SubnetContractSession) ClaimReward(claim SubnetClaim) (*types.Transaction, error) {
	return _SubnetContract.Contract.ClaimReward(&_SubnetContract.TransactOpts, claim)
}

// ClaimReward is a paid mutator transaction binding the contract method 0x25f3a9c0.
//
// Solidity: function claimReward((bytes,(bytes16,uint256)[],uint256,uint256,(uint256,uint256)[],address,bytes,uint256) claim) returns()
func (_SubnetContract *SubnetContractTransactorSession) ClaimReward(claim SubnetClaim) (*types.Transaction, error) {
	return _SubnetContract.Contract.ClaimReward(&_SubnetContract.TransactOpts, claim)
}

// EnableWithdrawal is a paid mutator transaction binding the contract method 0x5636548f.
//
// Solidity: function enableWithdrawal(bool _enabled) returns()
func (_SubnetContract *SubnetContractTransactor) EnableWithdrawal(opts *bind.TransactOpts, _enabled bool) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "enableWithdrawal", _enabled)
}

// EnableWithdrawal is a paid mutator transaction binding the contract method 0x5636548f.
//
// Solidity: function enableWithdrawal(bool _enabled) returns()
func (_SubnetContract *SubnetContractSession) EnableWithdrawal(_enabled bool) (*types.Transaction, error) {
	return _SubnetContract.Contract.EnableWithdrawal(&_SubnetContract.TransactOpts, _enabled)
}

// EnableWithdrawal is a paid mutator transaction binding the contract method 0x5636548f.
//
// Solidity: function enableWithdrawal(bool _enabled) returns()
func (_SubnetContract *SubnetContractTransactorSession) EnableWithdrawal(_enabled bool) (*types.Transaction, error) {
	return _SubnetContract.Contract.EnableWithdrawal(&_SubnetContract.TransactOpts, _enabled)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.GrantRole(&_SubnetContract.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.GrantRole(&_SubnetContract.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0x1459457a.
//
// Solidity: function initialize(address _network, address tokenAddress, address xTokenAddress, address _sentryContract, address _validatorNodeContract) returns()
func (_SubnetContract *SubnetContractTransactor) Initialize(opts *bind.TransactOpts, _network common.Address, tokenAddress common.Address, xTokenAddress common.Address, _sentryContract common.Address, _validatorNodeContract common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "initialize", _network, tokenAddress, xTokenAddress, _sentryContract, _validatorNodeContract)
}

// Initialize is a paid mutator transaction binding the contract method 0x1459457a.
//
// Solidity: function initialize(address _network, address tokenAddress, address xTokenAddress, address _sentryContract, address _validatorNodeContract) returns()
func (_SubnetContract *SubnetContractSession) Initialize(_network common.Address, tokenAddress common.Address, xTokenAddress common.Address, _sentryContract common.Address, _validatorNodeContract common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.Initialize(&_SubnetContract.TransactOpts, _network, tokenAddress, xTokenAddress, _sentryContract, _validatorNodeContract)
}

// Initialize is a paid mutator transaction binding the contract method 0x1459457a.
//
// Solidity: function initialize(address _network, address tokenAddress, address xTokenAddress, address _sentryContract, address _validatorNodeContract) returns()
func (_SubnetContract *SubnetContractTransactorSession) Initialize(_network common.Address, tokenAddress common.Address, xTokenAddress common.Address, _sentryContract common.Address, _validatorNodeContract common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.Initialize(&_SubnetContract.TransactOpts, _network, tokenAddress, xTokenAddress, _sentryContract, _validatorNodeContract)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetContract *SubnetContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetContract *SubnetContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetContract.Contract.RenounceOwnership(&_SubnetContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SubnetContract *SubnetContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SubnetContract.Contract.RenounceOwnership(&_SubnetContract.TransactOpts)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_SubnetContract *SubnetContractTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_SubnetContract *SubnetContractSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.RenounceRole(&_SubnetContract.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_SubnetContract *SubnetContractTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.RenounceRole(&_SubnetContract.TransactOpts, role, callerConfirmation)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.RevokeRole(&_SubnetContract.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_SubnetContract *SubnetContractTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.RevokeRole(&_SubnetContract.TransactOpts, role, account)
}

// SetMinStakable is a paid mutator transaction binding the contract method 0xf65ff7bd.
//
// Solidity: function setMinStakable(uint256 _minStakable) returns()
func (_SubnetContract *SubnetContractTransactor) SetMinStakable(opts *bind.TransactOpts, _minStakable *big.Int) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setMinStakable", _minStakable)
}

// SetMinStakable is a paid mutator transaction binding the contract method 0xf65ff7bd.
//
// Solidity: function setMinStakable(uint256 _minStakable) returns()
func (_SubnetContract *SubnetContractSession) SetMinStakable(_minStakable *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetMinStakable(&_SubnetContract.TransactOpts, _minStakable)
}

// SetMinStakable is a paid mutator transaction binding the contract method 0xf65ff7bd.
//
// Solidity: function setMinStakable(uint256 _minStakable) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetMinStakable(_minStakable *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetMinStakable(&_SubnetContract.TransactOpts, _minStakable)
}

// SetNetworkAddress is a paid mutator transaction binding the contract method 0x05f5dc95.
//
// Solidity: function setNetworkAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactor) SetNetworkAddress(opts *bind.TransactOpts, add common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setNetworkAddress", add)
}

// SetNetworkAddress is a paid mutator transaction binding the contract method 0x05f5dc95.
//
// Solidity: function setNetworkAddress(address add) returns()
func (_SubnetContract *SubnetContractSession) SetNetworkAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetNetworkAddress(&_SubnetContract.TransactOpts, add)
}

// SetNetworkAddress is a paid mutator transaction binding the contract method 0x05f5dc95.
//
// Solidity: function setNetworkAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetNetworkAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetNetworkAddress(&_SubnetContract.TransactOpts, add)
}

// SetSentryBaseReward is a paid mutator transaction binding the contract method 0x1227d71a.
//
// Solidity: function setSentryBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractTransactor) SetSentryBaseReward(opts *bind.TransactOpts, _baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setSentryBaseReward", _baseReward)
}

// SetSentryBaseReward is a paid mutator transaction binding the contract method 0x1227d71a.
//
// Solidity: function setSentryBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractSession) SetSentryBaseReward(_baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetSentryBaseReward(&_SubnetContract.TransactOpts, _baseReward)
}

// SetSentryBaseReward is a paid mutator transaction binding the contract method 0x1227d71a.
//
// Solidity: function setSentryBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetSentryBaseReward(_baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetSentryBaseReward(&_SubnetContract.TransactOpts, _baseReward)
}

// SetSentryNodeAddress is a paid mutator transaction binding the contract method 0x1e7fc3ea.
//
// Solidity: function setSentryNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactor) SetSentryNodeAddress(opts *bind.TransactOpts, add common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setSentryNodeAddress", add)
}

// SetSentryNodeAddress is a paid mutator transaction binding the contract method 0x1e7fc3ea.
//
// Solidity: function setSentryNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractSession) SetSentryNodeAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetSentryNodeAddress(&_SubnetContract.TransactOpts, add)
}

// SetSentryNodeAddress is a paid mutator transaction binding the contract method 0x1e7fc3ea.
//
// Solidity: function setSentryNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetSentryNodeAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetSentryNodeAddress(&_SubnetContract.TransactOpts, add)
}

// SetTokenAddress is a paid mutator transaction binding the contract method 0x26a4e8d2.
//
// Solidity: function setTokenAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactor) SetTokenAddress(opts *bind.TransactOpts, add common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setTokenAddress", add)
}

// SetTokenAddress is a paid mutator transaction binding the contract method 0x26a4e8d2.
//
// Solidity: function setTokenAddress(address add) returns()
func (_SubnetContract *SubnetContractSession) SetTokenAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetTokenAddress(&_SubnetContract.TransactOpts, add)
}

// SetTokenAddress is a paid mutator transaction binding the contract method 0x26a4e8d2.
//
// Solidity: function setTokenAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetTokenAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetTokenAddress(&_SubnetContract.TransactOpts, add)
}

// SetValidatorBaseReward is a paid mutator transaction binding the contract method 0x8b043d02.
//
// Solidity: function setValidatorBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractTransactor) SetValidatorBaseReward(opts *bind.TransactOpts, _baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setValidatorBaseReward", _baseReward)
}

// SetValidatorBaseReward is a paid mutator transaction binding the contract method 0x8b043d02.
//
// Solidity: function setValidatorBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractSession) SetValidatorBaseReward(_baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetValidatorBaseReward(&_SubnetContract.TransactOpts, _baseReward)
}

// SetValidatorBaseReward is a paid mutator transaction binding the contract method 0x8b043d02.
//
// Solidity: function setValidatorBaseReward(uint256 _baseReward) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetValidatorBaseReward(_baseReward *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetValidatorBaseReward(&_SubnetContract.TransactOpts, _baseReward)
}

// SetValidatorNodeAddress is a paid mutator transaction binding the contract method 0xb791bf39.
//
// Solidity: function setValidatorNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactor) SetValidatorNodeAddress(opts *bind.TransactOpts, add common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "setValidatorNodeAddress", add)
}

// SetValidatorNodeAddress is a paid mutator transaction binding the contract method 0xb791bf39.
//
// Solidity: function setValidatorNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractSession) SetValidatorNodeAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetValidatorNodeAddress(&_SubnetContract.TransactOpts, add)
}

// SetValidatorNodeAddress is a paid mutator transaction binding the contract method 0xb791bf39.
//
// Solidity: function setValidatorNodeAddress(address add) returns()
func (_SubnetContract *SubnetContractTransactorSession) SetValidatorNodeAddress(add common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.SetValidatorNodeAddress(&_SubnetContract.TransactOpts, add)
}

// Stake is a paid mutator transaction binding the contract method 0xf47cfa50.
//
// Solidity: function stake(bytes16 subnetId, uint256 amount) returns()
func (_SubnetContract *SubnetContractTransactor) Stake(opts *bind.TransactOpts, subnetId [16]byte, amount *big.Int) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "stake", subnetId, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xf47cfa50.
//
// Solidity: function stake(bytes16 subnetId, uint256 amount) returns()
func (_SubnetContract *SubnetContractSession) Stake(subnetId [16]byte, amount *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.Stake(&_SubnetContract.TransactOpts, subnetId, amount)
}

// Stake is a paid mutator transaction binding the contract method 0xf47cfa50.
//
// Solidity: function stake(bytes16 subnetId, uint256 amount) returns()
func (_SubnetContract *SubnetContractTransactorSession) Stake(subnetId [16]byte, amount *big.Int) (*types.Transaction, error) {
	return _SubnetContract.Contract.Stake(&_SubnetContract.TransactOpts, subnetId, amount)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetContract *SubnetContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetContract *SubnetContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.TransferOwnership(&_SubnetContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SubnetContract *SubnetContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SubnetContract.Contract.TransferOwnership(&_SubnetContract.TransactOpts, newOwner)
}

// UnStake is a paid mutator transaction binding the contract method 0x7fd72127.
//
// Solidity: function unStake(bytes16 subnetId) returns()
func (_SubnetContract *SubnetContractTransactor) UnStake(opts *bind.TransactOpts, subnetId [16]byte) (*types.Transaction, error) {
	return _SubnetContract.contract.Transact(opts, "unStake", subnetId)
}

// UnStake is a paid mutator transaction binding the contract method 0x7fd72127.
//
// Solidity: function unStake(bytes16 subnetId) returns()
func (_SubnetContract *SubnetContractSession) UnStake(subnetId [16]byte) (*types.Transaction, error) {
	return _SubnetContract.Contract.UnStake(&_SubnetContract.TransactOpts, subnetId)
}

// UnStake is a paid mutator transaction binding the contract method 0x7fd72127.
//
// Solidity: function unStake(bytes16 subnetId) returns()
func (_SubnetContract *SubnetContractTransactorSession) UnStake(subnetId [16]byte) (*types.Transaction, error) {
	return _SubnetContract.Contract.UnStake(&_SubnetContract.TransactOpts, subnetId)
}

// SubnetContractInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SubnetContract contract.
type SubnetContractInitializedIterator struct {
	Event *SubnetContractInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractInitialized represents a Initialized event raised by the SubnetContract contract.
type SubnetContractInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetContract *SubnetContractFilterer) FilterInitialized(opts *bind.FilterOpts) (*SubnetContractInitializedIterator, error) {

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SubnetContractInitializedIterator{contract: _SubnetContract.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetContract *SubnetContractFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SubnetContractInitialized) (event.Subscription, error) {

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractInitialized)
				if err := _SubnetContract.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_SubnetContract *SubnetContractFilterer) ParseInitialized(log types.Log) (*SubnetContractInitialized, error) {
	event := new(SubnetContractInitialized)
	if err := _SubnetContract.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SubnetContract contract.
type SubnetContractOwnershipTransferredIterator struct {
	Event *SubnetContractOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractOwnershipTransferred represents a OwnershipTransferred event raised by the SubnetContract contract.
type SubnetContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetContract *SubnetContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SubnetContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractOwnershipTransferredIterator{contract: _SubnetContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetContract *SubnetContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SubnetContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractOwnershipTransferred)
				if err := _SubnetContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SubnetContract *SubnetContractFilterer) ParseOwnershipTransferred(log types.Log) (*SubnetContractOwnershipTransferred, error) {
	event := new(SubnetContractOwnershipTransferred)
	if err := _SubnetContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the SubnetContract contract.
type SubnetContractRoleAdminChangedIterator struct {
	Event *SubnetContractRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractRoleAdminChanged represents a RoleAdminChanged event raised by the SubnetContract contract.
type SubnetContractRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_SubnetContract *SubnetContractFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*SubnetContractRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractRoleAdminChangedIterator{contract: _SubnetContract.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_SubnetContract *SubnetContractFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *SubnetContractRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractRoleAdminChanged)
				if err := _SubnetContract.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_SubnetContract *SubnetContractFilterer) ParseRoleAdminChanged(log types.Log) (*SubnetContractRoleAdminChanged, error) {
	event := new(SubnetContractRoleAdminChanged)
	if err := _SubnetContract.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the SubnetContract contract.
type SubnetContractRoleGrantedIterator struct {
	Event *SubnetContractRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractRoleGranted represents a RoleGranted event raised by the SubnetContract contract.
type SubnetContractRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*SubnetContractRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractRoleGrantedIterator{contract: _SubnetContract.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *SubnetContractRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractRoleGranted)
				if err := _SubnetContract.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) ParseRoleGranted(log types.Log) (*SubnetContractRoleGranted, error) {
	event := new(SubnetContractRoleGranted)
	if err := _SubnetContract.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the SubnetContract contract.
type SubnetContractRoleRevokedIterator struct {
	Event *SubnetContractRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractRoleRevoked represents a RoleRevoked event raised by the SubnetContract contract.
type SubnetContractRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*SubnetContractRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractRoleRevokedIterator{contract: _SubnetContract.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *SubnetContractRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractRoleRevoked)
				if err := _SubnetContract.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_SubnetContract *SubnetContractFilterer) ParseRoleRevoked(log types.Log) (*SubnetContractRoleRevoked, error) {
	event := new(SubnetContractRoleRevoked)
	if err := _SubnetContract.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractStakeEventIterator is returned from FilterStakeEvent and is used to iterate over the raw logs and unpacked data for StakeEvent events raised by the SubnetContract contract.
type SubnetContractStakeEventIterator struct {
	Event *SubnetContractStakeEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractStakeEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractStakeEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractStakeEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractStakeEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractStakeEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractStakeEvent represents a StakeEvent event raised by the SubnetContract contract.
type SubnetContractStakeEvent struct {
	Account common.Address
	Stake   SubnetStakeStruct
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterStakeEvent is a free log retrieval operation binding the contract event 0x278a86cadccd34d129a359367ad378aae3808cfbe93b8a7f9d4fce6638cd2b87.
//
// Solidity: event StakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) FilterStakeEvent(opts *bind.FilterOpts, account []common.Address) (*SubnetContractStakeEventIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "StakeEvent", accountRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractStakeEventIterator{contract: _SubnetContract.contract, event: "StakeEvent", logs: logs, sub: sub}, nil
}

// WatchStakeEvent is a free log subscription operation binding the contract event 0x278a86cadccd34d129a359367ad378aae3808cfbe93b8a7f9d4fce6638cd2b87.
//
// Solidity: event StakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) WatchStakeEvent(opts *bind.WatchOpts, sink chan<- *SubnetContractStakeEvent, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "StakeEvent", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractStakeEvent)
				if err := _SubnetContract.contract.UnpackLog(event, "StakeEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeEvent is a log parse operation binding the contract event 0x278a86cadccd34d129a359367ad378aae3808cfbe93b8a7f9d4fce6638cd2b87.
//
// Solidity: event StakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) ParseStakeEvent(log types.Log) (*SubnetContractStakeEvent, error) {
	event := new(SubnetContractStakeEvent)
	if err := _SubnetContract.contract.UnpackLog(event, "StakeEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SubnetContractUnStakeEventIterator is returned from FilterUnStakeEvent and is used to iterate over the raw logs and unpacked data for UnStakeEvent events raised by the SubnetContract contract.
type SubnetContractUnStakeEventIterator struct {
	Event *SubnetContractUnStakeEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *SubnetContractUnStakeEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SubnetContractUnStakeEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(SubnetContractUnStakeEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *SubnetContractUnStakeEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SubnetContractUnStakeEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SubnetContractUnStakeEvent represents a UnStakeEvent event raised by the SubnetContract contract.
type SubnetContractUnStakeEvent struct {
	Account common.Address
	Stake   SubnetStakeStruct
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnStakeEvent is a free log retrieval operation binding the contract event 0xb8075e85c8e8c717443eca24e396af055e00019ff8c88d1cf27aeb39cba8353f.
//
// Solidity: event UnStakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) FilterUnStakeEvent(opts *bind.FilterOpts, account []common.Address) (*SubnetContractUnStakeEventIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _SubnetContract.contract.FilterLogs(opts, "UnStakeEvent", accountRule)
	if err != nil {
		return nil, err
	}
	return &SubnetContractUnStakeEventIterator{contract: _SubnetContract.contract, event: "UnStakeEvent", logs: logs, sub: sub}, nil
}

// WatchUnStakeEvent is a free log subscription operation binding the contract event 0xb8075e85c8e8c717443eca24e396af055e00019ff8c88d1cf27aeb39cba8353f.
//
// Solidity: event UnStakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) WatchUnStakeEvent(opts *bind.WatchOpts, sink chan<- *SubnetContractUnStakeEvent, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _SubnetContract.contract.WatchLogs(opts, "UnStakeEvent", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SubnetContractUnStakeEvent)
				if err := _SubnetContract.contract.UnpackLog(event, "UnStakeEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnStakeEvent is a log parse operation binding the contract event 0xb8075e85c8e8c717443eca24e396af055e00019ff8c88d1cf27aeb39cba8353f.
//
// Solidity: event UnStakeEvent(address indexed account, (uint256,uint256) stake)
func (_SubnetContract *SubnetContractFilterer) ParseUnStakeEvent(log types.Log) (*SubnetContractUnStakeEvent, error) {
	event := new(SubnetContractUnStakeEvent)
	if err := _SubnetContract.contract.UnpackLog(event, "UnStakeEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
