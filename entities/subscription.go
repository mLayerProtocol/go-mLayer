package entities

import (
	// "errors"

	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
)

var logger = &log.Logger

type Subscription struct {

	ID         string    `gorm:"primaryKey;type:char(36);not null"  json:"id,omitempty"`
	Topic      string    `json:"top" binding:"required"  gorm:"not null;uniqueIndex:idx_sub_topic;type:char(36);index"`
	Ref        string    `json:"ref,omitempty" gorm:"uniqueIndex:idx_ref_subnet;type:varchar(100);default:null"`
	Meta       string    `json:"meta,omitempty"  gorm:"type:varchar(100);"`
	Subnet     string    `json:"snet"  binding:"required" gorm:"not null;uniqueIndex:idx_ref_subnet;type:varchar(36)"`
	Subscriber DIDString `json:"sub"  gorm:"not null;uniqueIndex:idx_sub_topic;type:varchar(100);index"`
	// Device     DeviceString                  `json:"dev,omitempty" binding:"required"  gorm:"not null;uniqueIndex:idx_acct_dev_topic;type:varchar(100);index"`
	Status *constants.SubscriptionStatus `json:"st"  gorm:"not null;type:smallint;default:2"`
	Role   *constants.SubscriberRole  `json:"rol" gorm:"default:0"`

	//Signature string                         `json:"sig"`
	Timestamp *uint64       `json:"ts,omitempty"`
	Hash      string       `json:"h,omitempty" gorm:"unique" `
	Event     EventPath    `json:"e,omitempty" gorm:"index;char(64);"`
	Agent     DeviceString `json:"agt,omitempty"  gorm:"not null;type:varchar(100);index"`
	BlockNumber uint64          `json:"blk,omitempty"`
	Cycle   	uint64			`json:"cy,omitempty"`
	Epoch		uint64			`json:"ep,omitempty"`
	EventSignature  string    `json:"sig,omitempty"`
}

func (d Subscription) GetSignature() (string) {
	return d.EventSignature
}

func (g *Subscription) GetKeys() (keys []string)  {
	keys = append(keys, fmt.Sprintf("%s/%s", g.SubscriberKey(), utils.IntMilliToTimestampString(int64(utils.SafePointerValue(g.Timestamp, uint64(time.Now().UnixMilli())))),))
	keys = append(keys, fmt.Sprintf("%s/%d/%s", g.SubscriptionStatusKey(), *g.Status, utils.IntMilliToTimestampString(int64(utils.SafePointerValue(g.Timestamp, uint64(time.Now().UnixMilli()))))))
	if g.Status != &constants.UnsubscribedSubscriptionStatus &&  g.Status != &constants.BannedSubscriptionStatus {
		keys = append(keys, g.SubscribedTopicsKey())
	}
	keys = append(keys, fmt.Sprintf("%s/%d/%s", g.SubscriptionStatusKey(), *g.Status, utils.IntMilliToTimestampString(int64(utils.SafePointerValue(g.Timestamp, uint64(time.Now().UnixMilli()))))))
	keys = append(keys, g.Key())	
	keys = append(keys, g.DataKey())
	if len(g.Ref) > 0 {
		keys = append(keys, g.RefKey())
	}
	// keys = append(keys, fmt.Sprintf("%s/%d/%s", AuthModel, g.Cycle, g.ID))
	return keys;
}
func (item *Subscription) DataKey() string {
	return fmt.Sprintf(DataKey, GetModel(item), item.Event.ID )
}
func (item *Subscription) RefKey() string {
	return fmt.Sprintf("%s|ref|%s", SubscriptionModel, item.Ref)
}



func (g *Subscription) SubscriberKey() (string) {
	if (g.Subscriber != "") {
			return fmt.Sprintf("%s/sub/%s/%s", SubscriptionModel, g.Topic, g.Subscriber)
	} else {
		return fmt.Sprintf("%s/sub/%s", SubscriptionModel, g.Topic)
	}
}

func (g *Subscription) SubscribedTopicsKey() (string) {
	if (g.Subscriber != "") {
			return fmt.Sprintf("%s/sub/%s/%s", SubscriptionModel, g.Subscriber, g.Topic,)
	} else {
		return fmt.Sprintf("%s/sub/%s", SubscriptionModel, g.Subscriber)
	}
}
func (g *Subscription) SubscriptionStatusKey() (string) {
	return fmt.Sprintf("%s/st/%s/%s", SubscriptionModel, g.Topic, g.Subscriber)
}
func (sub *Subscription) Key() string {
	key := strings.ReplaceAll(sub.SubscriberKey(), "/", ":")
	return fmt.Sprintf("%s/id/%s", SubscriptionModel, key)
}

func (sub Subscription) ToJSON() []byte {
	m, e := json.Marshal(sub)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}

func (subscription Subscription) ToString() (string, error) {
	values := []string{}
	values = append(values, subscription.Hash)
	values = append(values, subscription.ID)
	// values = append(values, fmt.Sprintf("%d", subscription.Timestamp))
	values = append(values, string(subscription.Subscriber))
	values = append(values, fmt.Sprintf("%d", subscription.Timestamp))
	return strings.Join(values, ","), nil
}

func (sub Subscription) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(sub)
	return b
}

func SubscriptionFromBytes(b []byte) (Subscription, error) {
	var sub Subscription
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &sub)
	return sub, err
}
func UnpackSubscription(b []byte) (Subscription, error) {
	var sub Subscription
	err := encoder.MsgPackUnpackStruct(b, &sub)
	return sub, err
}

func (sub Subscription) GetHash() ([]byte, error) {
	b, err := sub.EncodeBytes()
	if err != nil {
		log.Logger.Errorf("Subscription Hashing error, %v", err)
		return []byte(""), err
	}
	return crypto.Keccak256Hash(b), nil
}

func (sub Subscription) GetEvent() EventPath {
	return sub.Event
}
func (sub Subscription) GetAgent() DeviceString {
	return sub.Agent
}

func (sub Subscription) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(

		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: sub.Meta},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: sub.Ref},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: utils.SafePointerValue(sub.Role, 0)},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: utils.SafePointerValue(sub.Status, 0)},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: sub.Subscriber.ToString()},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(sub.Topic)},
	)
}
