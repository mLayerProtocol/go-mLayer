package entities

import (
	// "errors"

	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/internal/crypto"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
)

type Topic struct {
	Version float32 `json:"_v"`
	ID string `json:"id" gorm:"type:uuid;primaryKey;not null"`
	// Name            string        `json:"n,omitempty" binding:"required"`
	Ref             string        `json:"ref,omitempty" binding:"required" gorm:"uniqueIndex:idx_unique_subnet_ref;type:varchar(64);default:null"`
	Meta            string        `json:"meta,omitempty"`
	ParentTopic string        `json:"pT,omitempty" gorm:"type:char(64)"`
	SubscriberCount uint64        `json:"sC,omitempty"`
	Account         AccountString `json:"acct,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`

	DeviceKey DeviceString `json:"dKey,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`
	//
	Public   *bool `json:"pub,omitempty" gorm:"default:false"`

	DefaultSubscriberRole   *constants.SubscriberRole `json:"dSubRol,omitempty"`

	Invite []Subscription `json:"inv,omitempty"`

	Handler json.RawMessage `json:"hdlr,omitempty"`

	ReadOnly *bool `json:"rO,omitempty" gorm:"default:false"`
	// InviteOnly bool `json:"invO" gorm:"default:false"`

	// Derived
	Event   EventPath `json:"e,omitempty" gorm:"index;varchar;"`
	Hash    string    `json:"h,omitempty" gorm:"type:char(64)"`
	// Signature   string    `json:"sig,omitempty" binding:"required"  gorm:"non null;"`
	// Broadcasted   bool      `json:"br,omitempty"  gorm:"default:false;"`
	Timestamp uint64 `json:"ts,omitempty" binding:"required"`
	Subnet    string `json:"snet" gorm:"uniqueIndex:idx_unique_subnet_ref;type:char(36);"`
	// Subnet string `json:"snet" gorm:"index;varchar(36)"`
	BlockNumber uint64          `json:"blk"`
	Cycle   	uint64			`json:"cy"`
	Epoch		uint64			`json:"ep"`
	EventSignature  string    `json:"csig,omitempty"`
}

func (d Topic) GetSignature() (string) {
	return d.EventSignature
}

func (item *Topic) DataKey() string {
	return fmt.Sprintf(DataKey, GetModel(item), item.Event.ID )
}

func (item *Topic) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(item)
	return b
}

func (item *Topic) Key() string {
	if item.ID == "" {
		item.ID, _ = GetId(item, "")
	}
	return fmt.Sprintf("%s/id/%s", GetModel(item), item.ID)
}


func (g *Topic) GetKeys() (keys []string)  {
	keys = append(keys, fmt.Sprintf("%s/%s/%s",  g.GetAccountTopicsKey(), utils.IntMilliToTimestampString(int64(g.Timestamp)), g.ID))
	// keys = append(keys, fmt.Sprintf("%s/acct/%s/%s/%s", TopicModel, g.Account, g.Subnet, g.ID))
	keys = append(keys, g.Key())
	keys = append(keys, g.DataKey())
	keys = append(keys, g.RefKey())
	return keys;
}


func (item *Topic) RefKey() string {
	if item.Ref == "" {
		return ""
	}
	return fmt.Sprintf("%s|ref|%s|%s", TopicModel, item.Subnet, item.Ref)
}

func (g *Topic) GetAccountTopicsKey() (string) {
	if (g.Subnet != "") {
		if g.DeviceKey != ""  {
			return fmt.Sprintf("%s/sub/%s/%s/%s", TopicModel, g.Subnet, g.Account, g.DeviceKey)
		}
		return fmt.Sprintf("%s/sub/%s/%s", TopicModel, g.Subnet, g.Account)
	} else {
		return fmt.Sprintf("%s/sub/%s", TopicModel, g.Subnet)
	}
}



func (topic *Topic) ToJSON() []byte {
	m, e := json.Marshal(topic)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}


func TopicToByte(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)

	return b
}

func TopicFromBytes(b []byte) (Topic, error) {
	var topic Topic
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &topic)
	return topic, err
}
func UnpackTopic(b []byte) (Topic, error) {
	var topic Topic
	err := encoder.MsgPackUnpackStruct(b, &topic)
	return topic, err
}

func (p *Topic) CanSend(channel string, sender AccountString) bool {
	// check if user can send
	return true
}

func (p *Topic) IsMember(channel string, sender AccountString) bool {
	// check if user can send
	return true
}

func (topic Topic) GetHash() ([]byte, error) {
	if topic.Hash != "" {
		return hex.DecodeString(topic.Hash)
	}
	b, err := topic.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return crypto.Sha256(b), nil
}

func (topic Topic) ToString() (string, error) {
	values := []string{}
	values = append(values, topic.Hash)
	values = append(values, topic.Meta)
	// values = append(values, fmt.Sprintf("%d", topic.Timestamp))
	values = append(values, fmt.Sprintf("%d", topic.SubscriberCount))
	values = append(values, string(topic.Account))
	values = append(values, fmt.Sprintf("%t", topic.Public))
	// values = append(values, fmt.Sprintf("%s", topic.Signature))
	return strings.Join(values, ","), nil
}

func (topic Topic) GetEvent() EventPath {
	return topic.Event
}
func (topic Topic) GetAgent() DeviceString {
	return topic.DeviceKey
}

func (topic Topic) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: utils.SafePointerValue(topic.DefaultSubscriberRole, 0)},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(topic.ID)},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: topic.Handler},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: topic.Meta},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(topic.ParentTopic)},
		encoder.EncoderParam{Type: encoder.BoolEncoderDataType, Value: utils.SafePointerValue(topic.Public, false)},
		encoder.EncoderParam{Type: encoder.BoolEncoderDataType, Value:  utils.SafePointerValue(topic.ReadOnly, false)},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: topic.Ref},
		// encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: *topic.DefaultSubscriptionStatus},
		 encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(topic.Subnet)},
	)
}


type NodeInterest struct {
	Ids []string `json:"ids"`
	Type EntityModel`json:"t"`
	Hash string  `json:"h,omitempty"`
	EventSignature  string    `json:"sig,omitempty"`
	Expiry int `json:"exp"`
	//PublicKey string `json:"pub"`
	// Signature string `json:"sig"`
	// Timestamp int64 `json:"ts"`
}

func (intr NodeInterest) EncodeBytes() ([]byte, error) {
	topicBytes := []byte{}
	for _, topic := range intr.Ids {
		topicBytes = append(topicBytes, utils.UuidToBytes(topic)... )
	}
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: topicBytes},
		// encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: intr.PublicKey},
		// encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: intr.Timestamp},
	)
}


func (intr NodeInterest) GetHash() ([]byte, error) {
	if len(intr.Hash) > 0 {
		return hex.DecodeString(intr.Hash)
	}
	b, err := intr.EncodeBytes()
	if err != nil {
		return nil, err
	}
	return crypto.Sha256(b), nil
}

func (intr NodeInterest) ToString() (string, error) {
	return fmt.Sprintf("%s",intr.Ids), nil
}

func (intr NodeInterest) GetSignature() (string) {
	return intr.EventSignature
}

func (item *NodeInterest) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(item)
	return b
}

func UnpackNodeInterest(b []byte) (NodeInterest, error) {
	intr := NodeInterest{}
	err := encoder.MsgPackUnpackStruct(b, &intr)
	return intr, err
}