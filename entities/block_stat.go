package entities

import (
	"fmt"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
)

// "math"

type BlockStats struct {
	ID                 string `json:"id" gorm:"type:uuid;primaryKey;not null"`
	BlockNumber        uint64 `json:"blk"`
	EventCount         uint64 `json:"eC"`
	MessageCount       uint64 `json:"mC"`
	TopicCount         uint64 `json:"tC"`
	AuthorizationCount uint64 `json:"authC"`
	SubnetCount uint64 `json:"snetC"`
	SubscriptionCount uint64 `json:"subC"`
	Cycle uint64 `json:"cy"`
	Epoch uint64 `json:"ep"`
	MessageCost string `json:"mCo"`
	Event string `json:"e"`
	// Count              uint64 `json:"c" gorm:"default:1"`
}


func GetBlockStatsKeys(event *Event) []string {
	keys := []string{}
	keys = append(keys, "e")
	// logger.Infof("EventType %s, %d",  GetModelTypeFromEventType(constants.EventType(event.EventType)), event.EventType)
	keys = append(keys, fmt.Sprintf("e/%s", GetModelTypeFromEventType(constants.EventType(event.EventType))))
	keys = append(keys, fmt.Sprintf("b/%015d/%s", event.BlockNumber, GetModelTypeFromEventType(constants.EventType(event.EventType))))
	keys = append(keys, fmt.Sprintf("b/%015d", event.BlockNumber))
	keys = append(keys, fmt.Sprintf("e/%015d", event.Epoch))
	keys = append(keys, fmt.Sprintf("c/%015d", event.Cycle))
	keys = append(keys, RecentEventKey(event.Cycle))
	return keys
}

func RecentEventKey(cycle uint64) string {
	return fmt.Sprintf("c/%015d/e", cycle)
}
// func (chatStats Stats) ToString() string {
// 	values := []string{}

// 	values = append(values, string(chatStats.BlockNumber))
// 	// values = append(values, fmt.Sprintf("%d", chatStats.EventType))
// 	values = append(values, fmt.Sprintf("%d", chatStats.EventCount))
// 	values = append(values, fmt.Sprintf("%d", chatStats.MessageCount))
// 	values = append(values, fmt.Sprintf("%d", chatStats.TopicCount))
// 	values = append(values, fmt.Sprintf("%d", chatStats.AuthorizationCount))
// 	// values = append(values, fmt.Sprintf("%d", chatStats.ApprovalExpiry))
// 	// values = append(values, fmt.Sprintf("%s", chatStats.ChainId))
// 	// values = append(values, fmt.Sprintf("%s", chatStats.Platform))
// 	// values = append(values, fmt.Sprintf("%d", chatStats.Timestamp))

// 	// values = append(values, fmt.Sprintf("%s", chatStats.SubjectHash))

// 	return strings.Join(values, "")
// }

// func (msg Stats) GetHash() ([]byte, error) {
// 	b, err := msg.EncodeBytes()
// 	if err != nil {
// 		return []byte(""), err
// 	}
// 	return cryptoEth.Keccak256Hash(b).Bytes(), nil
// }

// func (msg Stats) EncodeBytes() ([]byte, error) {
// 	var attachments []byte
// 	var actions []byte

// 	// for _, at := range msg.Actions {
// 	// 	attachments = append(actions, at.EncodeBytes()...)
// 	// }
// 	// for _, ac := range msg.Actions {
// 	// 	actions = append(actions, ac.EncodeBytes()...)
// 	// }

// 	// logger.Debug("Mesage....", string(msg.Data))
// 	// dataByte, _ := hex.DecodeString(msg.Data)
// 	return encoder.EncodeBytes(
// 		// encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: msg.TopicId},
// 		// encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: msg.Sender},
// 		// encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: msg.Receiver},
// 		// encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: crypto.Keccak256Hash(dataByte)},
// 		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: attachments},
// 		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: actions},
// 	)
// }
