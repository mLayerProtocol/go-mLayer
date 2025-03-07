package entities

import (
	// "errors"

	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/internal/crypto"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
)


type SystemMessageType uint8

const (
	AnnounceSelf SystemMessageType = 1
	AnnounceApplicationInterest SystemMessageType = 2
	AnnounceTopicInterest SystemMessageType = 3
	// NewStateDelta SystemMessageType = 4
)

type SystemMessage struct {
	Version float32 `json:"_v"`
	Data          json.RawMessage        `json:"d"`
	SignatureData SignatureData `json:"sigD,omitempty" gorm:"json;"`
	Type SystemMessageType `json:"t,omitempty"`
	Timestamp     uint64        `json:"ts,omitempty" binding:"required"`
	Hash  string    `json:"h,omitempty" gorm:"type:char(64)"`
	BlockNumber uint64          `json:"blk,omitempty"`
	Cycle   	uint64			`json:"cy,omitempty"`
	Epoch		uint64			`json:"ep,omitempty"`
}


func (d SystemMessage) GetSignature() (string) {
	if d.SignatureData.Type == TendermintsSecp256k1PubKey {
		val, _ := base64.StdEncoding.DecodeString(string(d.SignatureData.Signature))
		 return hex.EncodeToString(val)
	}
	if d.SignatureData.Type == EthereumPubKey {
		return strings.ReplaceAll(string(d.SignatureData.Signature), "0x", "")
	}
	return ""
}  


func (item SystemMessage) ToString() (string, error) {
	json.Marshal(item)
	b, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func (item *SystemMessage) ToJSON() []byte {
	m, e := json.Marshal(item)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}

func (item *SystemMessage) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(item)
	return b
}

func UnpackSystemMessage(b []byte) (SystemMessage, error) {
	var item SystemMessage
	err := encoder.MsgPackUnpackStruct(b, &item)
	return item, err
}

func (item SystemMessage) GetHash() ([]byte, error) {
	if item.Hash != "" {
		return hex.DecodeString(item.Hash)
	}
	b, err := item.EncodeBytes()
	
	if err != nil {
		return []byte(""), err
	}
	logger.Debugf("GetHash crypto.Sha256(b) : %v", crypto.Sha256(b))
	return crypto.Sha256(b), nil
}

func (item SystemMessage) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: item.Data},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: item.Timestamp},
	)
}
