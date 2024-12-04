package api

import (
	"encoding/hex"
	"math/big"

	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
)



const NetworkStartDate = int64(1706468787000)

var API GenericAPI

type GenericAPI struct {
	IChainAPI,
	chainId configs.ChainId
}


// func Init(chainId configs.ChainId, api IChainAPI) {
// 	// API = *NewGenericAPI(chainId)
// 	ChainId configs.ChainId
// }
func NewGenericAPI() *GenericAPI {
	return &GenericAPI{}
}

func (n GenericAPI)  GetStartBlock() (*big.Int, error) {
	return big.NewInt(0), nil
}
func (n GenericAPI)  GetStartTime() (*big.Int, error) {
	return big.NewInt(NetworkStartDate), nil
}
func (n GenericAPI)  GetValidatorLicenseOwnerAddress(publicKey []byte) ([]byte, error) {
	return nil, nil
}

func (n GenericAPI)  GetSentryLicenseOwnerAddress(publicKey []byte) ([]byte, error) {
	return nil, nil
}
func (n GenericAPI) GetChainInfo() (*ChainInfo, error) {
	curBlock, _ := n.GetCurrentBlockNumber()
	startBlock, _ := n.GetStartBlock()
	cycle, _ := n.GetCurrentCycle()
	epoch, _ := n.GetCurrentEpoch()

 return &ChainInfo{ChainId: n.chainId, CurrentBlock: curBlock, StartBlock: startBlock , CurrentCycle: cycle, CurrentEpoch: epoch, StartTime: big.NewInt(NetworkStartDate)}, nil
}
func (n GenericAPI) GetEpoch(blockNumber *big.Int) (*big.Int, error) {
	return new(big.Int).Div(blockNumber, big.NewInt(14400)), nil
}
func (n GenericAPI) GetCycle(blockNumber *big.Int) (*big.Int, error) {
	b, err := n.GetEpoch(blockNumber)
	var cycle *big.Int
	if err == nil {
		cycle = new(big.Int).Add(big.NewInt(1), new(big.Int).Div(b, big.NewInt(1)))
	}
	return cycle, err
}
func (n GenericAPI) GetTotalValueLockedInSubnets() (*big.Int, error) {
	amount, _ := new(big.Int).SetString("2000000000000000000000", 10)
	return amount, nil
}

func (n GenericAPI) GetCurrentBlockNumber() (*big.Int, error) {
	// return big.NewInt(int64(time.Now().UnixMilli() - NetworkStartDate) / 6000), nil
	return big.NewInt(4), nil
}
func (n GenericAPI) GetSentryActiveLicenseCount(cycle *big.Int)  (*big.Int, error)  {
	return big.NewInt(4), nil
}

func (n GenericAPI) GetValidatorActiveLicenseCount(cycle *big.Int) (*big.Int, error) {
	return big.NewInt(2), nil
}
func (n GenericAPI) GetSentryOperatorCycleLicenseCount(operator []byte, cycle  *big.Int) (*big.Int, error) {
	return big.NewInt(100), nil
}
func (n GenericAPI) GetValidatorOperatorCycleLicenseCount(operator []byte, cycle  *big.Int) (*big.Int, error) {
	return big.NewInt(100), nil
}
func (n GenericAPI) GetSentryLicenseOperator(license *big.Int) ([]byte, error) {
	operator1, _ := hex.DecodeString("03d212263468365e70b2d673b06b903216b5e101d8243cdbfac6884369e3c069a0")
	operator2, _ := hex.DecodeString("02733cc67380d6a8f4dad591126e08bd4d8f4471de0cc611f8cbc05461a85fb5c3")
	operators := map[uint64][]byte{1000: operator1, 1001: operator1, 1002: operator2, 1003: operator2}

	return operators[license.Uint64()], nil
}
func (n GenericAPI) GetValidatorLicenseOperator(license *big.Int) ([]byte, error) {
	operator1, _ := hex.DecodeString("03d212263468365e70b2d673b06b903216b5e101d8243cdbfac6884369e3c069a0")
	operator2, _ := hex.DecodeString("02733cc67380d6a8f4dad591126e08bd4d8f4471de0cc611f8cbc05461a85fb5c3")
	operators := map[uint64][]byte{1000: operator1,  1002: operator2}

	return operators[license.Uint64()], nil
}

func (n GenericAPI) GetCurrentEpoch() (*big.Int, error) {
	b, err := n.GetCurrentBlockNumber()
	if err == nil {
		return n.GetEpoch(b)
	}
	return nil, err
}
func (n GenericAPI) IsValidatorNodeOperator(publicKey []byte, cycle *big.Int) (bool, error) {
	return true, nil
}

func (n GenericAPI) IsSentryNodeOperator(publicKey []byte, cycle *big.Int) (bool, error) {
	return true, nil
}

func (n GenericAPI) GetCurrentCycle() (*big.Int, error) {
	b, err := n.GetCurrentBlockNumber()
	if err == nil {
		return n.GetCycle(b)
	}
	return nil, err
}

func (n *GenericAPI) LicenseOperator(license *big.Int) ([]byte, error) {
	
	return hex.DecodeString("02c4435e768b4bae8236eeba29dd113ed607813b4dc5419d33b9294f712ca79ff4")
}




func (n GenericAPI) GetCurrentYear() (*big.Int, error) {
	 r, err := n.GetCurrentCycle()
	return new(big.Int).Div(r, big.NewInt(12)), err
}


func (n GenericAPI) GetStakeBalance(address entities.DIDString) big.Int {
	bal := new(big.Int)
	bal.SetString("100000000000000000000000000", 10)
	return *bal
}

func (n GenericAPI) GetSubnetBalance(id [16]byte) (*big.Int, error) {
	bal := new(big.Int)
	bal.SetString("100000000000000000000000000", 10)
	return bal, nil
}

// func (n GenericAPI) GetMinStakeAmountForValidators() (*big.Int, error) {
// 	bal := new(big.Int)
// 	bal.SetString("1000000000000000", 10)
// 	return bal, nil
// }

func (n GenericAPI) GetMessagePrice(cycle *big.Int) (*big.Int, error) {
	// bal := new(big.Int)
	// bal.SetString("10000000", 10)
	return big.NewInt(10000000), nil
}

func (n GenericAPI) GetCurrentMessagePrice() (*big.Int, error) {
	// bal := new(big.Int)
	// bal.SetString("10000000", 10)
	return big.NewInt(10000000), nil
}



func (n GenericAPI) ClaimReward(claim *entities.ClaimData) (hash []byte, err error) {
	
	return nil, nil
}


func  (n GenericAPI) GetMinStakeAmountForSentry() (*big.Int, error) {
	return new(big.Int), nil
}
func (n GenericAPI) GetSentryLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)  {
	return []*big.Int{big.NewInt(0), big.NewInt(1)}, nil
}
func (n GenericAPI) GetValidatorLicenses(operator []byte, cycle *big.Int)  ([]*big.Int, error)  {
	return []*big.Int{big.NewInt(0), big.NewInt(1)}, nil
}

func (n GenericAPI)  GetValidatorNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error) {
	d1, _:= hex.DecodeString("03d212263468365e70b2d673b06b903216b5e101d8243cdbfac6884369e3c069a0")
	d2, _:=hex.DecodeString("02c4435e768b4bae8236eeba29dd113ed607813b4dc5419d33b9294f712ca79ff4")
	return []OperatorInfo{{PublicKey: d1}, {PublicKey: d2}}, nil
}
func(n GenericAPI)	GetSentryNodeOperators(page *big.Int, perPage *big.Int) ([]OperatorInfo, error) {
	d1, _:= hex.DecodeString("03d212263468365e70b2d673b06b903216b5e101d8243cdbfac6884369e3c069a0")
	d2, _:=hex.DecodeString("02c4435e768b4bae8236eeba29dd113ed607813b4dc5419d33b9294f712ca79ff4")
	return []OperatorInfo{{PublicKey: d1}, {PublicKey: d2}}, nil
}
func (n GenericAPI) Claimed(validator []byte, cycle *big.Int, index *big.Int) (bool, error) {
	return false, nil
}


func (n GenericAPI) IsValidatorLicenseOwner(address string) (bool, error) {
	return true, nil
}
func (n GenericAPI) IsSentryLicenseOwner(address string)  (bool, error) {
	return true, nil
}