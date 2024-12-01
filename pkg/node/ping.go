package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
)

type Message struct {
	Timestamp int64  `json:"timestamp"`
	Data      string `json:"data"`
}

type Payload struct {
	Signer    string  `json:"signer"`
	Message   Message `json:"message"`
	Signature string  `json:"signature"`
}

func (msg *Message) EncodeBytes() []byte {

	// strData := "";
	timestamp := big.NewInt(int64(msg.Timestamp)).Bytes()
	data := []byte(msg.Data)

	return append(timestamp, data...)
}

func LivelinessPing(ctx *context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop() // Stop the ticker when we're done

	// Infinite loop to keep printing messages
	for {
		select {
		case <-ticker.C:
			fmt.Println("This message Before every 30 seconds! ")
			body, err := makeHTTPRequest(ctx)
			fmt.Printf("This message prints every 30 seconds! %v %v\n", body, err)
		}
	}
}

func makeHTTPRequest(ctx *context.Context) (string, error) {
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	fmt.Printf("makeHTTPRequest every 30 seconds! cfg.PingUrl: %s  cfg.PublicKeySECPHex: %s \n", cfg.PingUrl, cfg.PublicKeySECPHex)
	message := Message{
		Timestamp: time.Now().Unix(),
		Data:      "Hello word",
	}

	privateKey, _ := btcec.PrivKeyFromBytes(cfg.PrivateKeySECP)
	hash := sha256.Sum256(message.EncodeBytes())

	// Sign the hashed message using the Schnorr signature scheme
	signature, err := schnorr.Sign(privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message with Schnorr: %v", err)
	}

	// Serialize the signature to hex
	signatureHex := hex.EncodeToString(signature.Serialize())
	payload := Payload{
		Signer:    cfg.PublicKeySECPHex,
		Signature: signatureHex,
		Message:   message,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	resp, err := http.Post(cfg.PingUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to make GET request: %v", err)
	}
	defer resp.Body.Close() // Close the response body when the function exits

	// Check if the status code is not 200 (OK)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Return the body as a string
	return string(body), nil
}
