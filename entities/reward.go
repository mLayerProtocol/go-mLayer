package entities

import (
	"encoding/json"
	"math/big"
	"time"

	// "math"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
)

type ApplicationCount struct {
	Application     string `json:"sNet"`
	EventCount uint64 `json:"eC"`
	Cost json.RawMessage `json:"cost"`
}
const MaxBatchSize = 100

func (msg *ApplicationCount) EncodeBytes() ([]byte, error) {
	d := utils.UuidToBytes(msg.Application)
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: d[:6]},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.Cost},
	)
}


// RewardBatch
type RewardBatch struct {
	Id           string          `json:"id"`
	Index        int           `json:"idx"`
	Data         []ApplicationCount   `json:"d"`
	DataBoundary [2]ApplicationCount  `json:"dBound"`
	DataHash     json.RawMessage `json:"dH"`
	ChainId      configs.ChainId `json:"pre"`
	Closed       bool            `json:"cl"`
	Cycle        uint64          `json:"nh"`
	MessageCost  json.RawMessage `json:"cost"`
	TotalValue   json.RawMessage `json:"tv"`
	// Proofs       Proof           `json:"proof"`
	Timestamp    uint64          `json:"ts"`
	Validator  json.RawMessage          `json:"val"`
	// Signature    json.RawMessage `json:"sig"`
	// Signer       string          `json:"sign"`
	config       *configs.MainConfiguration  `msgpack:"_"`
	cycleSize	int `msgpack:"_"`
	// sync.Mutex
}

// func (msg *Block) ToJSON() []byte {
// 	var buf bytes.Buffer
// 	enc := msgpack.NewEncoder(&buf)
// 	enc.SetCustomStructTag("json")
// 	enc.Encode(msg)
// 	return buf.Bytes()
// }

func (msg *RewardBatch) Append(appCount ApplicationCount) {
	if len(msg.Data) == 0 {
		msg.DataHash = nil
	}
	appTotal := big.NewInt(0).SetBytes(msg.MessageCost).Mul(big.NewInt(int64(appCount.EventCount)), new(big.Int).SetBytes(msg.MessageCost))
	appCount.Cost = utils.ToUint256(appTotal)
	msg.Data = append(msg.Data, appCount)
	encoded, err := appCount.EncodeBytes()
	if err != nil {
		panic("Unable to encode ApplicationCount data")
	}
	msg.DataHash = crypto.Keccak256Hash(append(msg.DataHash, encoded...))
	logger.Debugf("TOTALVLAUE:::%s", big.NewInt(0).Add(new(big.Int).SetBytes(msg.TotalValue), appTotal))
	msg.TotalValue = utils.ToUint256(big.NewInt(0).Add(new(big.Int).SetBytes(msg.TotalValue), appTotal))
	length := len(msg.Data)
	if length == 1 {
		msg.DataBoundary[0] = msg.Data[0]
	}
	if len(msg.Data) >= MaxBatchSize || len(msg.Data) == int(msg.cycleSize) {
		msg.Closed = true
		msg.DataBoundary[1] = msg.Data[length-1]
	}
}

func (msg *RewardBatch) Clear() {
	msg.Data =[]ApplicationCount{}
	msg.TotalValue = nil
	msg.DataHash = []byte{}
}

func (msg *RewardBatch) GetProofData(chainId configs.ChainId) *ProofData {
	return &ProofData{
		DataHash: msg.DataHash,
		ChainId: chainId,
		Cycle: msg.Cycle,
		Index: msg.Index,
		Validator: msg.Validator,
		TotalCost: msg.TotalValue,
	}
}
func (msg *RewardBatch) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Cycle},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.DataHash},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: msg.Id},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Index},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Timestamp},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.TotalValue},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.Validator},
	)
}
func (rb *RewardBatch) GetHash(chainId configs.ChainId) ([32]byte, error) {
	rb.ChainId = chainId
	b, err := rb.EncodeBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return [32]byte(crypto.Keccak256Hash(b)), nil
}
// func (rb *RewardBatch) Sign(privateKey string) (error) {
// 	pubKey := crypto.GetPublicKeySECP(privateKey)
// 	rb.Timestamp = uint64(time.Now().UnixMilli())
// 	b, err := rb.EncodeBytes();
// 	if(err != nil) {
// 		return err
// 	}
// 	rb.Signature, _ = crypto.SignSECP(b, privateKey)
//     rb.Signer = pubKey
// 	return nil
// }

// func (rb *RewardBatch) IsValid() (bool) {
// 	// Important security update. Do not remove.
// 	// Prevents cross chain replay attack
// 	rb.Prefix = rb.config.AddressPrefix  // Important security update. Do not remove
// 	//
// 	// if math.Abs(float64(uint64(time.Now().UnixMilli()) - mp.Timestamp)) > constants.VALID_HANDSHAKE_SECONDS {
// 	// 	logger.WithFields(logrus.Fields{"data": mp}).Warnf("Batch Expired: %d", uint64(time.Now().UnixMilli()) - mp.Timestamp)
// 	// 	return false
// 	// }
// 	signer, err := hex.DecodeString(string(rb.Signer));
// 	if err != nil {
// 		logger.Error("Unable to decode signer")
// 		return false
// 	}
// 	data, err := rb.EncodeBytes()
// 	if err != nil {
// 		logger.Error("Unable to decode signer")
// 		return false
// 	}
// 	// signature, err := hex.DecodeString(string(mp.Signature));
// 	// if err != nil {
// 	// 	logger.Error(err)
// 	// 	return false
// 	// }
// 	isValid, err := crypto.VerifySignatureSECP(signer, data, rb.Signature)
// 	if err != nil {
// 		logger.Error(err)
// 		return false
// 	}
// 	if !isValid {
// 	//	logger.WithFields(logrus.Fields{"message": mp.Protocol, "signature": mp.Signature}).Warnf("Invalid signer %s", mp.Signer)
// 		return false
// 	}

// 	return true
// }

// func (msg *Block) ToString() string {
// 	values := []string{}
// 	values = append(values, fmt.Sprintf("%s", string(msg.BlockId)))
// 	values = append(values, fmt.Sprintf("%d", msg.Size))
// 	values = append(values, fmt.Sprintf("%d", msg.NodeHeight))
// 	values = append(values, fmt.Sprintf("%s", strconv.Itoa(msg.Timestamp)))
// 	values = append(values, fmt.Sprintf("%s", msg.Hash))
// 	return strings.Join(values, "")
// }

// func (msg *Block) Sign(privateKey string) Block {

// 	msg.Timestamp = int(time.Now().Unix())
// 	_, sig := Sign(msg.ToString(), privateKey)
// 	msg.Signature = sig
// 	return *msg
// }

func NewRewardBatch(config *configs.MainConfiguration, cycle uint64, index int, cycleMessageCost *big.Int, cycleSize int, validator []byte) *RewardBatch {
	id := utils.RandomHexString(32)
	return &RewardBatch{Id: id,
		config:      config,
		Timestamp:   uint64(time.Now().UnixMilli()),
		ChainId: config.ChainId,
		Cycle:       uint64(cycle),
		MessageCost: cycleMessageCost.Bytes(),
		Index: index,
		cycleSize: cycleSize,
		Validator: validator,
	}
}

func RewardBatchFromBytes(b []byte) (*RewardBatch, error) {
	var message RewardBatch
	err := json.Unmarshal(b, &message)
	return &message, err
}

func UnpackRewardBatch(b []byte) (*RewardBatch, error) {
	var message RewardBatch
	err := encoder.MsgPackUnpackStruct(b, &message)
	return &message, err
}

func (msg *RewardBatch) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(msg)
	return b
}


type SignatureRequestData struct {
	ProofHash json.RawMessage `json:"pH"`
	Challenge json.RawMessage `json:"chal"`
	// AggPubKey json.RawMessage `json:"aggPub"`
	// Message [32]byte `json:"msg"`
}
func (sr *SignatureRequestData) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(sr)
	return b
}
func UnpackSignatureRequestData(b []byte) (*SignatureRequestData, error) {
	var message SignatureRequestData
	err := encoder.MsgPackUnpackStruct(b, &message)
	return &message, err
}

type ProofData struct {
	DataHash json.RawMessage `json:"bH"`
	ChainId configs.ChainId `json:"chId"`
	Cycle uint64 `json:"cyc"`
	Index int `json:"idx"`
	Validator json.RawMessage `json:"val"`
	Signature  json.RawMessage `json:"sig"`
	Signers [][]byte `json:"signers"`
	Commitment json.RawMessage `json:"com"`
	TotalCost json.RawMessage `json:"mCo"`
}
func (sr *ProofData) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(sr)
	return b
}
func UnpackProofData(b []byte) (*ProofData, error) {
	var message ProofData
	err := encoder.MsgPackUnpackStruct(b, &message)
	return &message, err
}


func (msg *ProofData) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.DataHash},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.ToUint256(new(big.Int).SetUint64(msg.Cycle))},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.ToUint256(big.NewInt(int64(msg.Index)))},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.TotalCost},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.Validator},
	)
}
func (rb *ProofData) GetHash() ([32]byte, error) {
	
	b, err := rb.EncodeBytes()
	logger.Debugf("ENCODERD %v", b)
	if err != nil {
		return [32]byte{}, err
	}
	return [32]byte(crypto.Keccak256Hash(b)), nil
}