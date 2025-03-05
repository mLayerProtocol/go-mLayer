package entities

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/sirupsen/logrus"
)

type ClientHandshake struct {
	Version float32 `json:"_v"`
	Signature    string             `json:"sig"`
	Signer       DeviceString       `json:"sigr"`
	Account      AccountString      `json:"acct"`
	Validator    PublicKeyString    `json:"val"`
	Protocol     constants.Protocol `json:"proto"`
	ClientSocket interface{}        `json:"ws"`
	ChainId      configs.ChainId    `json:"chId"`
	Timestamp    uint64             `json:"ts"`
}

type ServerIdentity struct {
	Signature string `json:"sig"`
	Signer    string `json:"sigr"`
	Message   string `json:"m"`
	Address   string `json:"addr"`
}

func (cs ClientHandshake) ToJSON() []byte {
	m, e := json.Marshal(cs)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}

func (cs ClientHandshake) FromJSON() []byte {
	m, e := json.Marshal(cs)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}

func (hs *ClientHandshake) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(hs)
	return b
}

func ClientHandshakeFromJson(b []byte) (*ClientHandshake, error) {
	var verMsg ClientHandshake
	err := json.Unmarshal(b, &verMsg)
	return &verMsg, err
}

func (nma *ClientHandshake) IsValid(chainId configs.ChainId) bool {
	// Important security update. Do not remove.
	// Prevents cross chain replay attack
	if nma == nil {
		return false
	}
	nma.ChainId = chainId // Important security update. Do not remove
	// now := uint64(time.Now().UnixMilli())

	// if utils.Abs(now, nma.Timestamp) > uint64(15 * time.Second.Milliseconds()) {
	// 	logger.WithFields(logrus.Fields{"data": *nma, "d": 15 * time.Second.Milliseconds(), "t": utils.Abs(now, nma.Timestamp)}).Debugf("ClientHandshake: Expired -> %d", uint64(time.Now().UnixMilli()) - nma.Timestamp)
	// 	return false
	// }
	// signer, err := hex.DecodeString(string(nma.Signer));
	// if err != nil {
	// 	logger.Error("Unable to decode signer")
	// 	return false
	// }

	data, err := nma.EncodeBytes()
	if err != nil {
		logger.Error("Unable to encode message", err)
		return false
	}
	add, err := AddressFromString(string(nma.Signer))
	if err != nil {
		return false
	}

	isValid := crypto.VerifySignatureECC(add.Addr, &data, nma.Signature)
	if err != nil {
		logger.Error("VerifySignatureECC:", err)
		return false
	}

	if !isValid {
		logger.WithFields(logrus.Fields{"address": nma.Signer, "signature": nma.Signature}).Warnf("Invalid signer %s", nma.Signer)
		return false
	}
	return true
}

func (cs ClientHandshake) EncodeBytes() ([]byte, error) {
	
	return encoder.EncodeBytes(
		 encoder.EncoderParam{
			Type:  encoder.ByteEncoderDataType,
			Value: cs.ChainId.Bytes(),
		}, 
		encoder.EncoderParam{
			Type:  encoder.AddressEncoderDataType,
			Value: cs.Signer,
		},
		encoder.EncoderParam{
			Type:  encoder.ByteEncoderDataType,
			Value: cs.Account,
		}, 
		encoder.EncoderParam{
			Type:  encoder.ByteEncoderDataType,
			Value: cs.Validator.Bytes(),
		}, 
		encoder.EncoderParam{
			Type:  encoder.IntEncoderDataType,
			Value: cs.Timestamp,
		})
}

func UnpackClientHandshake(b []byte) (ClientHandshake, error) {
	var message ClientHandshake
	err := encoder.MsgPackUnpackStruct(b, &message)
	return message, err
}

func (hs *ServerIdentity) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(hs)
	return b
}

func UnpackServerIdentity(b []byte) (ServerIdentity, error) {
	var id ServerIdentity
	err := encoder.MsgPackUnpackStruct(b, &id)
	return id, err
}


type ClientWsSubscription struct {
	Conn   *websocket.Conn
	Filter map[string][]string
	Account string
	Id string
}

type ClientWsSubscriptionV2 struct {
	Conn   *websocket.Conn
	Filter map[string]map[string]string
	Account string
	Id string
}

// func SignIdentity(message []byte) bool {
// 	verifiedRequest, _ := UnpackClientHandshake(message)
// 	verifiedRequest.ClientSocket = &client
// 	verifiedRequest.Protocol = protocol;
// 	log.Println("verifiedRequest.Message: ", verifiedRequest.Message)
// 	hasVerified := false
// 	if VerifySignature(verifiedRequest.Signer, verifiedRequest.Message, verifiedRequest.Signature) {
// 		// verifiedConn = append(verifiedConn, c)
// 		hasVerified = true
// 		log.Println("Verification was successful: ", verifiedRequest)
// 		*ch <- &verifiedRequest
// 	}

// 	return hasVerified
// }
