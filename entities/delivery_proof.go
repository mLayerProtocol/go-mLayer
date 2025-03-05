package entities

import (
	"encoding/json"
	"fmt"
	"strconv"

	// "math"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
)

// DeliveryProof
type DeliveryProof struct {
	Version float32 `json:"_v"`
	MessageHash   string `json:"mH"`
	MessageSender string `json:"mS"`
	OperatorAddress   string `json:"nA"`
	Timestamp     int    `json:"ts"`
	Signature     string `json:"sig"`
	Block         string `json:"bl"`
	Index         int    `json:"i"`
}

func (msg *DeliveryProof) ToJSON() []byte {
	m, _ := json.Marshal(msg)
	return m
}
func (msg *DeliveryProof) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(msg)
	return b
}

func (msg *DeliveryProof) Key() string {
	return fmt.Sprintf("/%s/%s", msg.MessageHash, msg.MessageSender)
}
func (msg *DeliveryProof) BlockKey() string {
	return fmt.Sprintf("/%s", msg.Block)
}

func (msg DeliveryProof) ToString() string {
	values := []string{}
	values = append(values, fmt.Sprintf("%s", string(msg.MessageHash)))
	values = append(values, fmt.Sprintf("%s", msg.OperatorAddress))
	values = append(values, fmt.Sprintf("%s", strconv.Itoa(msg.Timestamp)))
	return strings.Join(values, "")
}

func (msg DeliveryProof) EncodeBytes() ([]byte, error) {
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: msg.MessageHash},
		encoder.EncoderParam{Type: encoder.HexEncoderDataType, Value: msg.OperatorAddress},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: msg.Timestamp},
	)
}

// func NewSignedDeliveryProof(data []byte, privateKey string) DeliveryProof {
// 	message, _ := DeliveryProofFromBytes(data)
// 	_, sig := Sign(message.ToString(), privateKey)
// 	message.Signature = sig
// 	return message
// }

func DeliveryProofFromBytes(b []byte) (DeliveryProof, error) {
	var message DeliveryProof
	err := json.Unmarshal(b, &message)
	return message, err
}

// DeliveryClaim
type DeliveryClaim struct {
	NodeHeight int      `json:"nh"`
	Signature  string   `json:"sig"`
	Amount     string   `json:"a"`
	Proofs     []string `json:"prs"`
}

func (msg *DeliveryClaim) ToJSON() []byte {
	m, _ := json.Marshal(msg)
	return m
}

func DeliveryClaimFromBytes(b []byte) (DeliveryClaim, error) {
	var message DeliveryClaim
	err := json.Unmarshal(b, &message)
	return message, err
}

func UnpackDelvieryClaim(b []byte) (DeliveryClaim, error) {
	var message DeliveryClaim
	err := encoder.MsgPackUnpackStruct(b, &message)
	return message, err
}
