package entities

import (
	// "errors"

	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
)



const DataKey = "data/%s/%s"
type Subnet struct {
	ID            string        `json:"id" gorm:"type:uuid;primaryKey;not null"`
	Meta          string        `json:"meta,omitempty"`
	Ref           string        `json:"ref,omitempty" binding:"required"  gorm:"unique;type:varchar(64);default:null"`
	Categories    pq.Int32Array `gorm:"type:integer[]"`
	
	SignatureData SignatureData `json:"sigD,omitempty" gorm:"json;"`
	Status        *uint8        `json:"st" gorm:"boolean;default:0"`
	Timestamp     uint64        `json:"ts,omitempty" binding:"required"`
	Balance       uint64        `json:"bal" gorm:"default:0"`
	// Readonly
	Account DIDString    `json:"acct,omitempty" binding:"required"  gorm:"not null;type:varchar(100)"`
	

	// CreateTopicPrivilege   *constants.AuthorizationPrivilege `json:"cTopPriv"` //
	DefaultAuthPrivilege *constants.AuthorizationPrivilege `json:"dAuthPriv"` // privilege for external users who joins the subnet. 0 indicates people cant join

	// Derived
	Event EventPath `json:"e,omitempty" gorm:"index;varchar;"`
	Hash  string    `json:"h,omitempty" gorm:"type:char(64)"`
	BlockNumber uint64          `json:"blk"`
	Cycle   	uint64			`json:"cy"`
	Epoch		uint64			`json:"ep"`
	Agent   	DeviceString `json:"agent"  gorm:"agent" msgpack:"agent"`

	//Deprecated
	Owner         string     `json:"-" gorm:"-" msgpack:"-"`
	EventSignature  string    `json:"csig,omitempty"`
}

func (d Subnet) GetSignature() (string) {
	// return string(d.SignatureData.Type)
	// if d.Hash != "" {
	// 	return d.Hash
	// }
	// hash, _ := d.GetHash()
	// return hex.EncodeToString(hash)
	if d.SignatureData.Type == TendermintsSecp256k1PubKey {
		val, _ := base64.StdEncoding.DecodeString(d.SignatureData.Signature)
		 return hex.EncodeToString(val)
	}
	if d.SignatureData.Type == EthereumPubKey {
		return strings.ReplaceAll(d.SignatureData.Signature, "0x", "")
	}
	return ""
}  
// func (g Subnet) GetId() (string) {
// 	// return g.Event.Hash[:32]
// 	return g.ID
// }


func (g *Subnet) GetKeys() (keys []string)  {
	if g.ID == "" {
		g.ID, _ = GetId(g)
	}
	keys = append(keys, fmt.Sprintf("%s/%s/%s", g.AccountSubnetsKey(), utils.IntMilliToTimestampString(int64(g.Timestamp)), g.ID))
	keys = append(keys, g.Key())
	keys = append(keys, g.RefKey())
	// keys = append(keys, fmt.Sprintf("%s/%d/%s", SubnetModel, g.Cycle, g.ID))
	// keys = append(keys,fmt.Sprintf("%s/%s/%s", g.Event.Hash, SubnetModel, g.Hash ))
	keys = append(keys, g.DataKey())
	keys = append(keys, g.ArchiveKey())
	// keys = append(keys, g.GetEventStateKey())
	// keys = append(keys, fmt.Sprintf("%s/%d/%s", AuthModel, g.Cycle, g.ID))
	return keys;
}
// func (g Subnet) GetEventStateKey() (string) {
//    return fmt.Sprintf("ev/%s", g.Event.ToString() )
// }
func (item *Subnet) DataKey() string {
	return fmt.Sprintf(DataKey, SubnetModel, item.Event.Hash )
}

func (item *Subnet) ArchiveKey() string {
	return fmt.Sprintf("arc/%010d/%s", item.Cycle, item.Hash )
}


func (g *Subnet) Key() string {
	if g.ID == "" {
		g.ID, _ = GetId(g)
	}
	return fmt.Sprintf("%s/id/%s", GetModel(g), g.ID)
}

func (item *Subnet) RefKey() string {
	return fmt.Sprintf("%s|ref|%s", SubnetModel, item.Ref)
}




func (g *Subnet) AccountSubnetsKey() string {
	return fmt.Sprintf("%s/acct/%s", SubnetModel, g.Account)
}



func (item *Subnet) ToJSON() []byte {
	m, e := json.Marshal(item)
	if e != nil {
		logger.Errorf("Unable to parse subscription to []byte")
	}
	return m
}

func (item *Subnet) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(item)
	return b
}




func SubnetToByte(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func SubnetFromBytes(b []byte) (Subnet, error) {
	var item Subnet
	// if err := json.Unmarshal(b, &message); err != nil {
	// 	panic(err)
	// }
	err := json.Unmarshal(b, &item)
	return item, err
}
func UnpackSubnet(b []byte) (Subnet, error) {
	var item Subnet
	err := encoder.MsgPackUnpackStruct(b, &item)
	return item, err
}

func (p *Subnet) CanSend(channel string, sender DIDString) bool {
	// check if user can send
	return true
}

func (p *Subnet) IsMember(channel string, sender DIDString) bool {
	// check if user can send
	return true
}

func (item Subnet) GetHash() ([]byte, error) {
	if item.Hash != "" {
		return hex.DecodeString(item.Hash)
	}
	b, err := item.EncodeBytes()
	if err != nil {
		return []byte(""), err
	}
	logger.Debugf("GetHash crypto.Sha256(b) : %v", crypto.Sha256(b))
	return crypto.Sha256(b), nil
}

func (item Subnet) ToString() (string, error) {
	values := []string{}
	values = append(values, item.Hash)
	values = append(values, item.Meta)
	// values = append(values, fmt.Sprintf("%d", item.Timestamp))
	// values = append(values, fmt.Sprintf("%d", item.SubscriberCount))
	values = append(values, string(item.Account))
	// values = append(values, fmt.Sprintf("%s", item.Signature))
	return strings.Join(values, ","), nil
}

func (entity Subnet) GetEvent() EventPath {
	return entity.Event
}
func (entity Subnet) GetAgent() DeviceString {
	return entity.Agent
}

func (item Subnet) EncodeBytes() ([]byte, error) {
	cats := []byte{}
	for _, d := range item.Categories {
		b, err := encoder.EncodeBytes(encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: d})
		if err != nil {
			return cats, err
		}
		cats = append(cats, b...)
	}
	return encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: item.Account},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: utils.SafePointerValue(item.DefaultAuthPrivilege, 0)},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: item.Meta},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: item.Ref},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: utils.SafePointerValue(item.Status, 0)},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: item.Timestamp},
	)
}
