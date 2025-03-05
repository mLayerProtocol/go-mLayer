package entities

import (
	// "errors"

	"encoding/hex"
	"encoding/json"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto/schnorr"
)


type RegisterationData struct {
	ChainId configs.ChainId  `json:"cId"`
	Timestamp uint64 `json:"ts"`
	PubKeyEDD json.RawMessage `json:"pubEDD"`
}

func (regData RegisterationData) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: regData.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.ToUint256(new(big.Int).SetUint64(regData.Timestamp))},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: regData.PubKeyEDD},
	)
}

func (regData *RegisterationData) Sign(privkBytes []byte) ([]byte, schnorr.EthAddress, error) {
	if regData.Timestamp == 0 {
		regData.Timestamp = uint64(time.Now().UnixMilli())
	}
	_, p := btcec.PrivKeyFromBytes(privkBytes)

	logger.Debugf("PUBKEY_X %d | %d", p.X(), p.Y())
	logger.Debugf("REGDATAHASH: %s", hex.EncodeToString(regData.GetHash()))
	signature, commitment, _, _ := schnorr.SignSingle(privkBytes, [32]byte(regData.GetHash()))
	return signature, commitment, nil
}

func (regData *RegisterationData) GetHash() []byte {
	d, _ := regData.EncodeBytes()
	return crypto.Keccak256Hash(d)
}
