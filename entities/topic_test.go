package entities

import (
	"encoding/hex"
	"testing"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestEncodeTopicBytes(t *testing.T) {
	agent := "0xe652d28F89A28adb89e674a6b51852D0C341Ebe9"
	topicIds := "*"
	ts := uint64(1705392177894)
	privilege := constants.AdminPriviledge
    auth := Authorization{Agent: StringToDeviceString(agent), TopicIds: topicIds, Timestamp: ts,  Priviledge: privilege, }
	hash := hex.EncodeToString(auth.GetHash())
	expected := "7da1e026b7a8dc545e2f84cca2eea963807e701aa03ad4b64b1ab447304460e9"
    if hash != expected {
        t.Fatalf("Hash should match \n%s \n%s", hash, expected )
    }
}

func TestCreateTopic(t *testing.T) {
	agent := "0xe652d28F89A28adb89e674a6b51852D0C341Ebe9"
	topicIds := "*"
	ts := uint64(1705392177894)
	privilege := constants.AdminPriviledge
    auth := Authorization{Agent: agent, TopicIds: topicIds, Timestamp: ts,  Priviledge: privilege, }
	// hash := hex.EncodeToString(auth.GetHash())
	encoded, err := auth.EncodeBytes();
	sign := "3044022073761e96c0d29d109ba2290fea1821bf093155d7549956a8144d04f9209f6c1b022048d4635f7e2f3e2864b2c1fe695badad548483ce1988f8bf0c04d5caf4131829"
	valid, err := crypto.VerifySignatureEDD("02ebec9d95769bb3d71712f0bf1e7e88b199fc945f67f908bbab81e9b7cb1092d8", encoded, sign)
    if !valid {
        t.Fatalf("Invalid signature signer: %v", err )
    }
}