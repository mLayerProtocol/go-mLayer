package entities

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	// "math"
	"strings"

	cryptoEth "github.com/ethereum/go-ethereum/crypto"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	PlatformEthereum string = "ethereum"
	PlatformBitcoin string = "bitcoin"
	PlatformSolana          = "solana"
	PlatformCosmos          = "cosmos"
)


var buf bytes.Buffer
var msgPackEncoder = msgpack.NewEncoder(&buf)

func init() {
	msgPackEncoder.SetCustomStructTag("json")
}

/*
*
CHAT MESSAGE
*
*/
// type MessageHeader struct {
// 	Length   int    `json:"l"`
// 	Sender   DIDString `json:"s"`
// 	Receiver string `json:"r"`
// 	// ChainId configs.ChainId      string `json:"cId"`
// 	// Platform      string `json:"p"`
// 	Timestamp      uint64 `json:"ts"`
// 	// ApprovalExpiry uint64 `json:"apExp"`
// 	// Wildcard      bool   `json:"wildcard"`
// 	// Channels      []string `json:"chs"`
// 	// SenderAddress string   `json:"sA"`
// 	// OwnerAddress  string `json:"oA"`
// }

// func (h MessageHeader) EncodeBytes() []byte {
// 	b, _ := encoder.EncodeBytes(
// 		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: h.Length},
// 		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: h.Sender},
// 		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: h.Receiver},
// 		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: h.Timestamp},
// 	)
// 	return b
// }

// TODO! platform enum channel
// ! receiver field is name of channel u are sending to
// ! look for all subscribers to the channel
// ! channel subscribers store
// type MessageBody struct {
// 	DataHash string `json:"mH"`
// 	Url         string `json:"url"`
// 	Data 		json.RawMessage `json:"d,omitempty"`
// }
// func (b MessageBody) EncodeBytes() []byte {
// 	e, _ := encoder.EncodeBytes(
// 		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: b.DataHash},
// 		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: b.Url},
// 		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: b.Data},
// 	)
// 	return e
// }

type MessageAttachment struct {
	CID  string `json:"cid"`
	Hash string `json:"h"`
}

func (b MessageAttachment) EncodeBytes() []byte {
	e, _ := encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: b.CID},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: b.Hash},
	)
	return e
}

type MessageAction struct {
	Contract   string   `json:"c"`
	Abi        string   `json:"abi"`
	Action     string   `json:"a"`
	Parameters []string `json:"pa"`
}

func (a MessageAction) EncodeBytes() []byte {
	var b []byte
	for _, d := range a.Parameters {
		data, _ := encoder.EncodeBytes(encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: d})
		b = append(b, data...)
	}
	encoded, _ := encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: a.Contract},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: a.Abi},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: a.Action},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: b},
	)
	return encoded
}

type Message struct {
	ID string `json:"id" gorm:"type:uuid;primaryKey;not null"`
	// Timestamp      uint64   `json:"ts"`
	Topic string        `json:"top,omitempty"`
	
	
	// OwnerAddress  string              `json:"oA"`
	Receiver DIDString   `json:"r,omitempty"`
	Data     string          `json:"d"`
	DataType     constants.DataType      `json:"dTy"`
	Actions  []MessageAction `json:"a,omitempty" gorm:"json;"`
	Sender  DIDString `json:"s,omitempty"`
	// Length int `json:"len"`
	
	Nonce uint64 `json:"nonce,omitempty" binding:"required"`

	/// DERIVED
	
	Agent DeviceString `json:"agt,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`
	Event       EventPath           `json:"e,omitempty" gorm:"index;char(64);"`
	Hash        string              `json:"h"`
	// Attachments []MessageAttachment `json:"atts" gorm:"json;"`
	// Subject     string              `json:"s"`
	Signature string `json:"sig,omitempty"`
	// Origin      string              `json:"o"`
	DataHash string `json:"dH,omitempty"`
	Url      string `json:"url,omitempty"`
	BlockNumber uint64          `json:"blk,omitempty"`
	Cycle   	uint64			`json:"cy,omitempty"`
	Epoch		uint64			`json:"ep,omitempty"`
	Subnet		string			`json:"snet,omitempty" gorm:"-"`
	EventSignature  string    `json:"csig,omitempty"`
	EventTimestamp uint64 		`json:"ets,omitempty"`
	// DEPRECATED COLUMNS
	// TopicId string        `json:"-" gorm:"-" msgpack:"-"`
	// Attachments  string `json:"-" gorm:"-" msgpack:"-"`
}

func (d Message) GetSignature() (string) {
	return d.EventSignature
}
func (chatMessage Message) ToString() (string, error) {
	values := []string{}

	values = append(values, string(chatMessage.Receiver))
	// values = append(values, fmt.Sprintf("%s", chatMessage.Approval))
	// values = append(values, fmt.Sprintf("%d", chatMessage.ApprovalExpiry))
	// values = append(values, fmt.Sprintf("%s", chatMessage.ChainId))
	// values = append(values, fmt.Sprintf("%s", chatMessage.Platform))
	// values = append(values, fmt.Sprintf("%d", chatMessage.Timestamp))

	// values = append(values, fmt.Sprintf("%s", chatMessage.SubjectHash))
	values = append(values, chatMessage.DataHash)
	values = append(values, chatMessage.Url)
	_action := []string{}
	for i := 0; i < len(chatMessage.Actions); i++ {
		_action = append(_action, fmt.Sprintf("[%d]:%s", i, chatMessage.Actions[i].Contract))
		_action = append(_action, fmt.Sprintf("[%d]:%s", i, chatMessage.Actions[i].Abi))
		_action = append(_action, fmt.Sprintf("[%d]:%s", i, chatMessage.Actions[i].Action))

		_parameter := []string{}
		for j := 0; j < len(chatMessage.Actions[i].Parameters); j++ {
			_parameter = append(_parameter, fmt.Sprintf("[%d][%d]:%s", i, j, chatMessage.Actions[i].Parameters[j]))
		}

		_action = append(_action, fmt.Sprintf("[%d]:%s", i, _parameter))
	}

	values = append(values, fmt.Sprintf("%s", _action))

	return strings.Join(values, ""), nil
}

func (msg Message) GetHash() ([]byte, error) {
	if msg.Hash != "" {
		return hex.DecodeString(msg.Hash)
	}
	b, err := msg.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	return cryptoEth.Keccak256Hash(b).Bytes(), nil
}

func (item *Message) DataKey() string {
	return fmt.Sprintf(DataKey, GetModel(item), item.Event.Hash )
}


func (g *Message) GetKeys() (keys []string)  {
	
	if g.EventTimestamp == 0 {
		g.EventTimestamp = uint64(time.Now().UnixMilli())
	}
	keys = append(keys, g.Key())
	keys = append(keys, g.DataKey())
	keys = append(keys,fmt.Sprintf("%s/%s/%s", g.MessageReceiverKey(), utils.IntMilliToTimestampString(int64(g.EventTimestamp)), g.Event.Hash ))
	keys = append(keys,fmt.Sprintf("%s/%s/%s", g.MessageSenderKey(), utils.IntMilliToTimestampString(int64(g.EventTimestamp)),  g.Event.Hash))
	keys = append(keys,fmt.Sprintf("%s/%s/%s", g.MessageSenderReceiverKey(), utils.IntMilliToTimestampString(int64(g.EventTimestamp)),  g.Event.Hash))
	keys = append(keys,fmt.Sprintf("%s/%s/%s", g.TopicMessageKey(), utils.IntMilliToTimestampString(int64(g.EventTimestamp)),  g.Event.Hash))
	keys = append(keys, g.UniqueId())
	
	// keys = append(keys, g.GetEventStateKey())
	// keys = append(keys, fmt.Sprintf("%s/%d/%s", AuthModel, g.Cycle, g.ID))
	return keys;
}
func (g Message) TopicMessageKey() (string) {
   return fmt.Sprintf("top/%s", g.Topic )
}

func (msg *Message) Key() string {
	if msg.ID == "" {
		msg.ID, _ = GetId(msg)
	}
	return fmt.Sprintf("id/%s",  msg.ID)
}

func (msg *Message) UniqueId() string {
	return msg.Hash[0:32]
}


func (g *Message) MessageSenderKey() (string) {
	if (g.Topic != "") {
		return fmt.Sprintf("s/%s/%s", g.Sender, g.Topic)
	} else {
		return fmt.Sprintf("s/%s", g.Sender)
	}
}

func (g *Message) MessageReceiverKey() (string) {
	if (g.Topic != "") {
		return fmt.Sprintf("%s/s/%s/%s", MessageModel, g.Receiver, g.Topic)
	} else {
		return fmt.Sprintf("%s/s/%s", MessageModel, g.Receiver)
	}
}

func (g *Message) MessageSenderReceiverKey() (string) {
	if (g.Topic != "") {
		return fmt.Sprintf("%s/s/%s/%s/%s", MessageModel, g.Sender, g.Receiver, g.Topic)
	} else {
		return fmt.Sprintf("%s/s/%s/%s", MessageModel, g.Receiver, g.Receiver)
	}
}


func (msg Message) EncodeBytes() ([]byte, error) {
	// var attachments []byte
	var actions []byte

	// for _, at := range msg.Actions {
	// 	attachments = append(actions, at.EncodeBytes()...)
	// }
	for _, ac := range msg.Actions {
		actions = append(actions, ac.EncodeBytes()...)
	}

	dataByte, _ := hex.DecodeString(msg.Data)
	// logger.Debugf("DataBytes: %s %s %s %s", dataByte, msg.DataType, msg.Receiver, hex.EncodeToString(utils.UuidToBytes(msg.Topic)))
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: actions},
		// encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: attachments},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: dataByte},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: msg.DataType},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Nonce},
		encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: msg.Receiver},
		encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: msg.Sender},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(msg.Topic)},
	)
}

// func (channel MessageHeader) ToApprovalBytes() ([]byte, error) {
// 	values := []string{}
// 	values = append(values, fmt.Sprintf("%d", channel.ApprovalExpiry))
// 	// values = append(values, fmt.Sprintf("%s", channel.Wildcard))
// 	values = append(values, fmt.Sprintf("%s", channel.Channels))
// 	values = append(values, fmt.Sprintf("%s", channel.SenderAddress))
// 	// values = append(values, fmt.Sprintf("%s", channel.OwnerAddress))
// 	b, err := encoder.EncodeBytes(
// 		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: channel.ApprovalExpiry},
// 		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: channel.Channels},
// 		encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: channel.SenderAddress},
// 	)
// 	if err != nil {
// 		return []byte(""), err
// 	}
// 	return b, nil
// }

type SuccessResponse struct {
	Body ClientPayload
	Meta Meta
}

type ErrorResponse struct {
	statusCode int
	meta       Meta
}

type Meta struct {
	statusCode int
	success    bool
}

func ReturnError(msg string, code int) *ErrorResponse {
	meta := Meta{statusCode: code}
	meta.success = false
	e := ErrorResponse{statusCode: code}
	e.meta = meta
	return &e
}

func UnpackMessage(b []byte) (Message, error) {
	var item Message
	err := encoder.MsgPackUnpackStruct(b, &item)
	return item, err
}

// func JsonMessageFromBytes(b []byte) (MessageJsonInput, error) {
// 	var message MessageJsonInput
// 	// if err := json.Unmarshal(b, &message); err != nil {
// 	// 	panic(err)
// 	// }
// 	err := json.Unmarshal(b, &message)
// 	return message, err
// }

// func UnpackJsonMessage(b []byte) (MessageJsonInput, error) {
// 	var message MessageJsonInput
// 	err := encoder.MsgPackUnpackStruct(b, message)
// 	return message, err
// }

func (msg *Message) ToJSON() string {
	m, _ := json.Marshal(msg)
	return string(m)
}

func (msg *Message) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(msg)
	return b
}

func (entity Message) GetEvent() EventPath {
	return entity.Event
}
func (entity Message) GetAgent() DeviceString {
	return entity.Agent
}

func MessageFromBytes(b []byte) *Message {
	var message Message
	if err := json.Unmarshal(b, &message); err != nil {
		logger.Error(err)
	}
	return &message
}

func MessageFromString(msg string) *Message {
	return MessageFromBytes([]byte(msg))
}
