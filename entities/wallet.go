package entities

import (
	// "errors"

	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"gorm.io/gorm"
)

type Wallet struct {
	// Primary
	ID        string        `gorm:"primaryKey;type:uuid;not null" json:"id,omitempty"`
	Account   DIDString `json:"acct"`
	Subnet    string        `json:"snet" gorm:"type:varchar(32);index;not null" msgpack:",noinline"`
	Name      string        `json:"n" gorm:"type:varchar(12);not null"`
	Symbol      string        `json:"sym" gorm:"type:varchar(8);not null"`
	Timestamp uint64        `json:"ts"`
	Agent DeviceString `json:"agt,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`

	// Derived
	Event EventPath `json:"e,omitempty" gorm:"index;varchar;"`
	Hash  string    `json:"h,omitempty" gorm:"type:char(64)"`
	BlockNumber uint64          `json:"blk"`
	Cycle   	uint64			`json:"cy"`
	Epoch		uint64			`json:"ep"`
	Signature  string    `json:"sig,omitempty" gorm:"type:char(64)"`
}

func (d Wallet) GetSignature() (string) {
	return d.Signature
}
func (d *Wallet) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		uuid, err := GetId(*d, "")
		if err != nil {
			logger.Error(err)
			panic(err)
		}

		d.ID = uuid
	}

	return nil
}

// func (e *Wallet) Key() string {
// 	return fmt.Sprintf("/%s", e.Hash)
// }

func (e *Wallet) ToJSON() []byte {
	m, err := json.Marshal(e)
	if err != nil {
		logger.Errorf("Unable to parse event to []byte")
	}
	return m
}

func (e *Wallet) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(e)
	return b
}

func WalletFromJSON(b []byte) (Event, error) {
	var e Event
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &e)
	return e, err
}

func (e Wallet) GetHash() ([]byte, error) {
	if e.Hash != "" {
		return hex.DecodeString(e.Hash)
	}
	b, err := e.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return crypto.Keccak256Hash(b), nil
}

func (entity Wallet) GetEvent() (EventPath) {
	return entity.Event
}
func (entity Wallet) GetAgent() (DeviceString) {
	return entity.Agent
}

func (e Wallet) ToString() (string, error) {
	values := []string{}
	values = append(values, e.Hash)
	values = append(values, e.Name)
	values = append(values, e.Subnet)
	values = append(values, e.Account.ToString())

	return strings.Join(values, ""), nil
}

func (e Wallet) EncodeBytes() ([]byte, error) {

	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: e.Name},
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: e.Subnet},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: e.Account.ToString()},
	)
}
