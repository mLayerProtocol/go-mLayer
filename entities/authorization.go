package entities

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
)

type PubKeyType string

const (
	// EDD PubKeyType = "Edd"
	// GenericSecP PubKeyType = "PubKeySecp256k1"
	TendermintsSecp256k1PubKey PubKeyType = "tendermint/PubKeySecp256k1"
	EthereumPubKey             PubKeyType = "eth"
)

type Authorization struct {
	Version float32 `json:"_v"`
	ID            string                           	`json:"id" gorm:"type:uuid;not null;primaryKey"`
	Authorized    AddressString                 	`json:"auth" gorm:"uniqueIndex:idx_agent_account_subnet;index:idx_authorization_states_agent"`
	Meta          string                           	`json:"meta,omitempty"`
	Account       AccountString                     `json:"acct" gorm:"varchar(40);"`
	Grantor       AccountString                       	`json:"gr" gorm:"index"`
	Priviledge    *constants.AuthorizationPrivilege	`json:"privi"  gorm:""`
	TopicIds      string                           	`json:"topIds"`
	Timestamp     *uint64                           `json:"ts"`
	Duration      *uint64                           `json:"du"`
	SignatureData SignatureData                    	`json:"sigD" gorm:"json;"`
	Hash          string                           	`json:"h" gorm:"unique" `
	Event         EventPath                        	`json:"e,omitempty" gorm:"index;varchar;"`
	Subnet        string                           	`json:"snet" gorm:"uniqueIndex:idx_agent_account_subnet;char(36)"`
	BlockNumber uint64          `json:"blk"`
	Cycle   	uint64			`json:"cy"`
	Epoch		uint64			`json:"ep"`
	// AuthorizationEventID string                           `json:"authEventId,omitempty"`
	EventSignature  string    `json:"sig,omitempty"`
}

func (d Authorization) GetSignature() (string) {
	return d.EventSignature
}  
func (g Authorization) GetHash() ([]byte, error) {
	if g.Hash != "" {
		return hex.DecodeString(g.Hash)
	}
	b, err := (g.EncodeBytes())
	logger.Debug("EncodeBytes:: ", b)
	if err != nil {
		logger.Errorf("Error endoding Authorization: %v", err)
		return []byte(""), err
	}
	bs := crypto.Sha256(b)
	return bs, nil
}

func (entity Authorization) GetEvent() EventPath {
	return entity.Event
}
// func (entity Authorization) GetAgent() DeviceString {
// 	return entity.Agent
// }
func (g Authorization) ToJSON() []byte {
	b, _ := json.Marshal(g)
	return b
}

func (g Authorization) ToString() (string, error) {
	return fmt.Sprintf("TopicIds:%s, Priviledge: %d, Grantor: %s, Timestamp: %d", g.TopicIds, g.Priviledge, g.Grantor, g.Timestamp), nil
}



func (g *Authorization) GetKeys() (keys []string)  {
	if g.ID == "" {
		g.ID, _ = GetId(g, "")
	}
	 // keys = append(keys, fmt.Sprintf("%s/acct/%s/%s/%s/%s", AuthModel, g.Account, g.Subnet, g.Agent, g.ID))
	 keys = append(keys, fmt.Sprintf("%s/%s",  g.AuthorizedAgentStateKey(), utils.IntMilliToTimestampString(int64(*g.Timestamp))))
	 keys = append(keys, fmt.Sprintf("%s/%s", g.AccountAuthorizationsKey(), utils.IntMilliToTimestampString(int64(*g.Timestamp))))
	 keys = append(keys, g.Key())
	 keys = append(keys, g.DataKey())
	 if (g.Account != g.Grantor) {
		keys = append(keys, fmt.Sprintf("%s/%s/%s/%s", AuthModel, g.Grantor, g.Subnet, g.ID))
	 }
	 return keys;
}
// func (g *Authorization) GetEventStateKey() (string) {
// 	return fmt.Sprintf("ev/%s", g.Event.ToString() )
// }


func (g *Authorization) AuthorizedAgentStateKey() (string) {
	if (g.Authorized.IsDevice()) {
		return fmt.Sprintf("%s/agt/%s/%s", AuthModel, g.Authorized, g.Subnet)
	}
	return fmt.Sprintf("%s/agt/%s/%s/%s", AuthModel, g.Authorized, g.Subnet, g.Account)
}

func (g *Authorization) AccountAuthorizationsKey() (string) {
	if (g.TopicIds != "" && g.TopicIds != "*") {
			return fmt.Sprintf("%s/agt/%s/%s/%s/%s", AuthModel, g.Account, g.Subnet, g.Authorized, g.TopicIds)
	} 

	if (g.Subnet != "") {
		if g.Authorized != ""  {
			return fmt.Sprintf("%s/agt/%s/%s/%s", AuthModel, g.Account, g.Subnet, g.Authorized)
		}
		return fmt.Sprintf("%s/agt/%s/%s", AuthModel, g.Account, g.Subnet)
	} else {
		return fmt.Sprintf("%s/agt/%s", AuthModel, g.Account)
	}
}

func AccountAuthorizationsKeyToAuthorization(key string) (*Authorization, error) {
	parts := strings.Split(key, "/")
	if len(parts) > 3 {
		return nil, fmt.Errorf("auth key too long")
	}
	account, err := AccountFromString(parts[0])
	if err != nil {
		return nil, err
	}
	if !account.IsValid() {
		return nil, fmt.Errorf("invalid account")
	}
	auth := &Authorization{Account: account.ToString(), Subnet: parts[1]}
	if len(parts) > 2 {
		addr, _ := AddressFromString(parts[2])
		auth.Authorized = addr.ToString()
	}
	return auth, nil
}
func (item *Authorization) ToAccountAuthKey() string {
	return fmt.Sprintf("%s/%s/%s", item.Subnet, item)
}

func (item *Authorization) Key() string {
	// if item.ID == "" {
	// 	item.ID, _ = GetId(item)
	// }
	key := strings.ReplaceAll(item.AccountAuthorizationsKey(), "/", ":")
	return fmt.Sprintf("%s/id/%s", GetModel(item), key)
}

func (item *Authorization) DataKey() string {
	return fmt.Sprintf(DataKey, GetModel(item), item.Event.ID )
}

func (item *Authorization) MsgPack() []byte {
	b, _ := encoder.MsgPackStruct(item)
	return b
}


func UnpackAuthorization(b []byte) (Authorization, error) {
	var auth Authorization
	err := encoder.MsgPackUnpackStruct(b, &auth)
	return auth, err
}

func AgentCountKey() string {
	return fmt.Sprintf("%s/agents", SubscriptionModel)
}
func (g Authorization) EncodeBytes() ([]byte, error) {

	b, e := encoder.EncodeBytes(
		encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: g.Account},
		encoder.EncoderParam{Type: encoder.AddressEncoderDataType, Value: g.Authorized},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: *g.Duration},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: g.Meta},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: *g.Priviledge},
		encoder.EncoderParam{Type: encoder.ByteEncoderDataType, Value: utils.UuidToBytes(g.Subnet)},
		encoder.EncoderParam{Type: encoder.IntEncoderDataType, Value: *g.Timestamp},
		encoder.EncoderParam{Type: encoder.StringEncoderDataType, Value: g.TopicIds},
	)

	return b, e
}
