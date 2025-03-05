package global

import (
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/entities"
)
var subnetMemberPrivi = constants.MemberPriviledge
var subnetStatus = uint8(1)

const TIMESTAMP = 1737460800000 // first crypto friendly government sworn in
const (
  SUBNET_UUID_PRIFIX string = "534e4554"
)


var GlobalSubnets = []entities.Subnet {
  {
    ID: SUBNET_UUID_PRIFIX+"-0000-0000-0000-000000000000",
    Meta: "{\"name\":\"global subnet\"}",
    Ref: "global.x1",
    Status: &subnetStatus,
    Balance: 0,
    Account: "mid:0x0000000000000000000000000000000000000000",
    DefaultAuthPrivilege: &subnetMemberPrivi,
    BlockNumber: 0,
    Cycle  : 0,
    Epoch:		0,
    Agent:  "did:0x0000000000000000000000000000000000000000",
    Event: entities.EventPath{EntityPath: entities.EntityPath{ Model: "snet", ID: "00000000-0000-0000-0000-000000000000", Validator: ""}},
    Timestamp: TIMESTAMP,
  },
}

var GlobalEvent = []entities.Event {
  {
    ID: "00000000-0000-0000-0000-000000000000",
    Payload: entities.ClientPayload{
      Data: GlobalSubnets[0],
    },
    BlockNumber: 0,
    Cycle  : 0,
    Epoch:		0,
    Synced: utils.TruePtr(),
    Broadcasted: true,
    IsValid: utils.TruePtr(),
    Timestamp: TIMESTAMP,
    EventType: constants.CreateSubnetEvent,
    Signature: "00000000000000000000000000000000",
    Hash: "00000000000000000000000000000000",
  },
  {
    ID: "00000000-0000-0000-0000-000000000001",
    Payload: entities.ClientPayload{
      Data: GlobalTopics[0],
    },
    BlockNumber: 0,
    Cycle  : 0,
    Epoch:		0,
    Synced: utils.TruePtr(),
    Broadcasted: true,
    IsValid: utils.TruePtr(),
    Timestamp: TIMESTAMP,
    EventType: constants.CreateTopicEvent,
    Signature: "00000000000000000000000000000001",
    Hash: "00000000000000000000000000000001",
    PreviousEvent: entities.EventPath{EntityPath: entities.EntityPath{ Model: entities.SubnetModel, ID: "00000000-0000-0000-0000-000000000000", Validator: ""}},
  },
}

const (
  TOPIC_UUID_PRIFIX string = "746f7069"
  REGISTERY_TOPIC_REF string  = "global.x1.registry"
  HANDSHAKE_TOPIC_REF string  = "global.x1.hanshake"
)
var RegistryTopicRef = []byte("global.x1.registry")

var GlobalTopics = []entities.Topic {
  {
    ID: TOPIC_UUID_PRIFIX+"-0000-0000-0000-000000000000",
    Subnet: SUBNET_UUID_PRIFIX+"-0000-0000-0000-000000000000",
    Meta: "{\"name\":\"global topic registry\"}",
    Ref: REGISTERY_TOPIC_REF,
    DefaultSubscriberRole: &constants.TopicWriterRole,
    Timestamp: TIMESTAMP, 
    Account: "mid:0x0000000000000000000000000000000000000000",
    Public: utils.BoolPtr(true),
    Handler: RegistryTopicRef,
    BlockNumber: 0,
    Cycle  : 0,
    Epoch:		0,
    
    Agent:  "did:0x0000000000000000000000000000000000000000",
    Event: entities.EventPath{EntityPath: entities.EntityPath{ Model: "top", ID: "00000000-0000-0000-0000-000000000001", Validator: ""}},
  },
  {
    ID: TOPIC_UUID_PRIFIX+"-0000-0000-0000-000000000001",
    Subnet: SUBNET_UUID_PRIFIX+"-0000-0000-0000-000000000000",
    Meta: "{\"name\":\"global handshake registry\"}",
    Ref: HANDSHAKE_TOPIC_REF,
    DefaultSubscriberRole: &constants.TopicWriterRole,
    Timestamp: TIMESTAMP, 
    Account: "mid:0x0000000000000000000000000000000000000000",
    Public: utils.BoolPtr(true),
    Handler: RegistryTopicRef,
    BlockNumber: 0,
    Cycle  : 0,
    Epoch:		0,
    
    Agent:  "did:0x0000000000000000000000000000000000000000",
    Event: entities.EventPath{EntityPath: entities.EntityPath{ Model: "top", ID: "00000000-0000-0000-0000-000000000001", Validator: ""}},
  },
}