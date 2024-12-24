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

type MessagePrice struct {
	Cycle         json.RawMessage        `json:"cy"`
	ChainId configs.ChainId `json:"pre"`
	Price             json.RawMessage        `json:"pr"`
	Signature            json.RawMessage        `json:"sig"`
	Signer json.RawMessage        `json:"sign"`
	Timestamp uint64 `json:"ts"`
	config *configs.MainConfiguration `json:"-"`
}


func (mp *MessagePrice) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(mp)
	return b
}


func UnpackMessagePrice(b []byte) (MessagePrice, error) {
	var mp MessagePrice
	err := encoder.MsgPackUnpackStruct(b,  &mp)
	return mp, err
}

func (mp MessagePrice) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.Cycle},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.ChainId.Bytes()},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: mp.Price},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: mp.Timestamp},
	)
}

func (mp *MessagePrice) IsValid(prefix configs.ChainId) bool {
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


func NewMessagePrice(config *configs.MainConfiguration, privateKey []byte, price []byte, cycle []byte) (*MessagePrice, error) {
	_, pubKey := crypto.GetPublicKeySECP(privateKey)
	mp := MessagePrice{config: config, Cycle: cycle,  ChainId: config.ChainId, Price: price, Timestamp: uint64(time.Now().UnixMilli())}
	b, err := mp.EncodeBytes();
	if(err != nil) {
		return nil, err
	}
	_, signature := crypto.SignSECP(b, privateKey)
    mp.Signature, err = hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}
    mp.Signer = pubKey
	return &mp, nil
}