package entities

import (
	// "errors"

	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"gorm.io/gorm"
)



type WalletBalance struct {
	// Primary
	ID string `gorm:"primaryKey;type:uuid;not null" json:"id,omitempty"`
	Account           DIDString `json:"acct"  gorm:"type:varchar(64);uniqueIndex:idx_wallet_acct;not null"`
	Wallet string `json:"wal"  gorm:"type:char(36);uniqueIndex:idx_wallet_acct;not null"`
	Balance             big.Int        `json:"bal" gorm:"index;not null;default:0"`
	Agent DeviceString `json:"agt,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`

	// Derived
	Event EventPath `json:"e,omitempty" gorm:"index;not null;varchar;"`
	Timestamp            uint64                           `json:"ts"`
	BlockNumber uint64          `json:"blk"`
	Cycle   	uint64			`json:"cy"`
	Epoch		uint64			`json:"ep"`
	Signature		string			`json:"sig"`
}

func (d WalletBalance) GetSignature() (string) {
	return d.Signature
}
func (d *WalletBalance) BeforeCreate(tx *gorm.DB) (err error) {
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

func (entity WalletBalance) GetEvent() (EventPath) {
	return entity.Event
}
func (entity WalletBalance) GetAgent() (DeviceString) {
	return entity.Agent
}

// func (e *WalletBalance) Key() string {
// 	return fmt.Sprintf("/%s", e.ID)
// }

func (e *WalletBalance) ToJSON() []byte {
	m, err := json.Marshal(e)
	if err != nil {
		logger.Errorf("Unable to parse event to []byte")
	}
	return m
}

func (e *WalletBalance) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(e)
	return b
}


func WalletBalanceFromJSON(b []byte) (Event, error) {
	var e Event
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &e)
	return e, err
}

func (e WalletBalance) GetHash() ([]byte, error) {
	b, err := e.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return crypto.Sha256(b), nil
}

func (e WalletBalance) ToString() (string, error) {
	values := []string{}
	values = append(values, fmt.Sprintf("%s", e.Account))
	values = append(values, fmt.Sprintf("%s", e.Wallet))
	
	return strings.Join(values, ""), nil
}

func (e WalletBalance) EncodeBytes() ([]byte, error) {

	
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: e.Account},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: e.Wallet},
	)
}

