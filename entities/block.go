package entities

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	// "math"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
)

// Block
type Block struct {
	Version float32 `json:"_v"`
	BlockId    string `json:"blId"`
	Size       int    `json:"s"`
	Closed     bool   `json:"c"`
	NodeHeight uint64    `json:"nh"`
	Hash       string `json:"h"`
	Timestamp  int    `json:"ts"`
	sync.Mutex
}

// func (msg *Block) ToJSON() []byte {
// 	var buf bytes.Buffer
// 	enc := msgpack.NewEncoder(&buf)
// 	enc.SetCustomStructTag("json")
// 	enc.Encode(msg)
// 	return buf.Bytes()
// }

func (msg *Block) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(msg)
	return b
}

func (msg *Block) ToString() string {
	values := []string{}
	values = append(values, fmt.Sprintf("%s", string(msg.BlockId)))
	values = append(values, fmt.Sprintf("%d", msg.Size))
	values = append(values, fmt.Sprintf("%d", msg.NodeHeight))
	values = append(values, fmt.Sprintf("%s", strconv.Itoa(msg.Timestamp)))
	values = append(values, fmt.Sprintf("%s", msg.Hash))
	return strings.Join(values, "")
}

func (msg *Block) Key() string {
	return fmt.Sprintf("/%s", msg.BlockId)
}

// func (msg *Block) Sign(privateKey string) Block {

// 	msg.Timestamp = int(time.Now().Unix())
// 	_, sig := Sign(msg.ToString(), privateKey)
// 	msg.Signature = sig
// 	return *msg
// }

func NewBlock() *Block {
	id, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_", 32)
	return &Block{BlockId: id,
		Size:   0,
		Closed: false}
}

func BlockFromBytes(b []byte) (*Block, error) {
	var message Block
	err := json.Unmarshal(b, &message)
	return &message, err
}

func UnpackBlock(b []byte) (*Block, error) {
	var message Block
	err := encoder.MsgPackUnpackStruct(b, &message)
	return &message, err
}

