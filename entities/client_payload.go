package entities

import (
	"context"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	// "math"

	"github.com/google/uuid"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Payload interface {
	GetHash() ([]byte, error)
	ToString() (string, error)
	EncodeBytes() ([]byte, error)
	// GetEvent() EventPath
	GetSignature() string
}
func (g ClientPayload) GetSignature()  string {
	return g.Signature
}

func (g ClientPayload) GetId()  (string, error) {
	return GetId(g, "")
}

func GetId(d Payload, id string) (string, error) {
	if len(id) > 0 {
		return id, nil
	}
	sig := d.GetSignature()
	if len(sig) == 0 {
		return "", fmt.Errorf("payload has no signature")
	}
	b, err := hex.DecodeString(strings.ReplaceAll(sig, "0x","")[:32])
	if err != nil {
		return "", err
	}
	u, err := uuid.FromBytes(b)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

type ClientPayload struct {
	// Primary
	Id string `json:"id"`
	Data      interface{}   `json:"d"`
	Timestamp uint64           `json:"ts"`
	EventType uint16        `json:"ty"`
	Nonce     uint64        `json:"nonce"`
	Account   DIDString `json:"acct,omitempty"` // optional public key of sender
	ChainId   configs.ChainId `json:"chId"` // optional public key of sender

	Validator string `json:"val,omitempty"`
	// Secondary																								 	AA	`							qaZAA	`q1aZaswq21``		`	`
	Signature string       `json:"sig"`
	Hash      string       `json:"h,omitempty"`
	Agent     DeviceString `gorm:"-" json:"agt"`
	Subnet    string       `json:"snet" gorm:"index;"`
	Page      uint16       `json:"page,omitempty" gorm:"_"`
	PerPage   uint16       `json:"perPage,omitempty" gorm:"_"`
}

func (msg ClientPayload) ToJSON() []byte {
	m, _ := json.Marshal(&msg)
	return m
}

// func (msg ClientPayload) EventNonce() string {
// 	return fmt.Sprintf("%s:%s", string(msg.Account), msg.Nonce)
// 	// d, err := msg.EncodeBytes()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// agent, _ := crypto.GetSignerECC(&d, &msg.Signature)
// 	// return hex.EncodeToString(crypto.Keccak256Hash([]byte(fmt.Sprintf("%s:%d",  agent, msg.Nonce))))
// 	return
// }
// func (s *ClientPayload) Encode() []byte {
// 	b, _ := s.Data.ToString()
// 	return b
// }

func (s *ClientPayload) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(s)
	return b
}

func MsgUnpackClientPayload(b []byte) (ClientPayload, error) {
	var p ClientPayload
	err := encoder.MsgPackUnpackStruct(b, &p)
	return p, err
}

func (msg ClientPayload) ToString() (string, error) {
	// return fmt.Sprintf("Data: %s, EventType: %d, Authority: %s", (msg.Data).(Payload).ToString(), msg.EventType, msg.Account)
	b, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (msg ClientPayload) GetHash() ([]byte, error) {
	if msg.Hash != "" {
		return hex.DecodeString(msg.Hash)
	}
	b, err := msg.EncodeBytes()
	if err != nil {
		logger.Debugf("ENCODBYTEERROR: %v",err)
		return nil, err
	}
	
	bs := crypto.Keccak256Hash(b)
	
	return bs, nil
}

func (msg *ClientPayload) GetSigner() (DeviceString, error) {

		b, err := msg.EncodeBytes()
		if err != nil {
			return "", err
		}
		
		agent, _ := crypto.GetSignerECC(&b,  &msg.Signature)
		msg.Agent = AddressFromString(agent).ToDeviceString()
		return msg.Agent, nil
}


func (msg ClientPayload) EncodeBytes() ([]byte, error) {
	hashed := []byte("")
	if msg.Data != nil {
		d, _ :=  json.Marshal(msg.Data.(Payload))
		logger.Debugf(string(d))
		b, err := msg.Data.(Payload).EncodeBytes()
		if err != nil {
			
			return nil, err
		}
		hashed = crypto.Keccak256Hash(b)
	}
	var params []encoder.EncoderParam
	params = append(params, encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: msg.ChainId.Bytes()})
	params = append(params, encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: hashed})
	params = append(params, encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.EventType})
	
	params = append(params, encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: msg.Id})
	
	if msg.Subnet != "" {
		params = append(params, encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(msg.Subnet)})
	}
	if msg.Account != "" {
		params = append(params, encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: msg.Account})
	}
	
	params = append(params, encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: msg.Validator})
	params = append(params, encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Nonce})
	params = append(params, encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Timestamp})
	
	return encoder.EncodeBytes(
		params...,
	)
}

func ClientPayloadFromBytes(b []byte) (ClientPayload, error) {
	var message ClientPayload
	err := json.Unmarshal(b, &message)
	return message, err
}

/** SYNC REQUEST PAYLOAD **/
type SyncRequest struct {
	Interval ResponseInterval `json:"inter"`
	TopicIds string           `json:"topIds"`
}


func (eP ClientPayload) GormDataType() string {
	return "varchar"
}
func (eP ClientPayload) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {

	// asString := encoder.ToBase64Padded(eP.MsgPack())
	asString := string(eP.ToJSON())
	return clause.Expr{
		SQL:                "?",
		Vars:               []interface{}{asString},
		WithoutParentheses: false,
	}
}

func (sD *ClientPayload) Scan(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Value not instance of string:", value))
	}
	//if strings.HasPrefix(data, "{") {
		d, err := ClientPayloadFromBytes([]byte(data))
		if err != nil {
			return err
		}
		*sD = d
	// } else {
	// 	d, err := base64.RawStdEncoding.DecodeString(data)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	b, err := MsgUnpackClientPayload(d)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	*sD = b
	// }
	
	return nil
}

func (sD *ClientPayload) Value() (driver.Value, error) {
	// return encoder.ToBase64Padded(sD.MsgPack()), nil
	return string(sD.ToJSON()), nil
}


