package p2p

import (
	// "errors"

	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
)

type ApplicationValidator struct {
	ID         json.RawMessage        `json:"id"`
	Cycle         uint64        `json:"cy"`
	Validators         json.RawMessage        `json:"v"`
	ChainId configs.ChainId `json:"pre"`
	Signature            json.RawMessage        `json:"sig"`
	Signer json.RawMessage        `json:"sign"`
	Timestamp uint64 `json:"ts"`
	config *configs.MainConfiguration `json:"-"`
}


func (mp *ApplicationValidator) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(mp)
	return b
}


func UnpackApplicationValidator(b []byte) (ApplicationValidator, error) {
	var mp ApplicationValidator
	err := encoder.MsgPackUnpackStruct(b,  &mp)
	return mp, err
}

func (mp ApplicationValidator) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: mp.Cycle},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.ID},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.Validators},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: mp.Timestamp},
	)
}

func (mp *ApplicationValidator) IsValid(prefix configs.ChainId) bool {
	// Important security update. Do not remove. 
	// Prevents cross chain replay attack
	mp.ChainId = prefix // Important security update. Do not remove

	signer, err := hex.DecodeString(string(mp.Signer));
	if err != nil {
		logger.Error("Unable to decode signer")
		return false
	}
	data, err := mp.EncodeBytes()
	if err != nil {
		logger.Error("Unable to decode signer")
		return false
	}
	// signature, err := hex.DecodeString(string(mp.Signature));
	// if err != nil {
	// 	logger.Error(err)
	// 	return false
	// }
	isValid, err := crypto.VerifySignatureSECP(signer, data, mp.Signature)
	if err != nil {
		logger.Error("VerifySignatureSECP: ", err)
		return false
	}
	if !isValid {
	//	logger.WithFields(logrus.Fields{"message": mp.Protocol, "signature": mp.Signature}).Warnf("Invalid signer %s", mp.Signer)
		return false
	}



	return true
}


func NewApplicationValidator(config *configs.MainConfiguration, id []byte, validators []byte, cycle uint64) (*ApplicationValidator, error) {

	mp := ApplicationValidator{config: config, Cycle: cycle,  ID: id, ChainId: config.ChainId, Validators: validators, Timestamp: uint64(time.Now().UnixMilli())}
	b, err := mp.EncodeBytes();
	if(err != nil) {
		return nil, err
	}
	_, signature := crypto.SignSECP(b, cfg.PrivateKeySECP)
    mp.Signature, err = hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}
    mp.Signer = cfg.PublicKeySECP
	return &mp, nil
}