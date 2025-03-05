package channelpool

import (
	"github.com/mlayerprotocol/go-mlayer/entities"
)
const CHANNEL_SIZE = 50
// channels for moving events received from other nodes through the pubsub channels
var AuthorizationEvent_SubscriptionC = make(chan *entities.Event, CHANNEL_SIZE)
var IncomingTopicEventSubscriptionC = make(chan *entities.Event, CHANNEL_SIZE)

// channels for broadcasting new events to other nodes
var AuthorizationEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var TopicEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var SubnetEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var WalletEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var SubscriptionEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var MessageEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var UnSubscribeEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var ApproveSubscribeEventPublishC = make(chan *entities.Event, CHANNEL_SIZE)
var EventProcessorChannel = make(chan *entities.Event, CHANNEL_SIZE)
var EventCounterChannel = make(chan *entities.Event, CHANNEL_SIZE)
var SystemMessagePublishC = make(chan *entities.Event, CHANNEL_SIZE)
var MempoolC = make(chan *entities.KeyByteValue, CHANNEL_SIZE)
// var RemoteEventProcessorChannel = make(chan *entities.RemoteEvent, CHANNEL_SIZE)

// CLEANUP
// most of these will be deleted
// transmits events received from other nodes in p2p to daemon
// var IncomingMessageEvent_P2P_D_C = make(chan *entities.Event, CHANNEL_SIZE)

// transmits validated events from Daemon to P2P to be broadcasted
// var OutgoingMessageEvents_D_P2P_C = make(chan *entities.Event, CHANNEL_SIZE)

// transmits messages sent through rpc, or other channels to daemon
// var NewPayload_Cli_D_C = make(chan *entities.ClientPayload, CHANNEL_SIZE)

// transmits new subscriptions from RPC to Daemon for processing
var Subscribers_RPC_D_C = make(chan *entities.Subscription, CHANNEL_SIZE)

// transmits valid subscriptions from Daemon to P2P for broadcasting
var Subscription_D_P2P_C = make(chan *entities.Subscription, CHANNEL_SIZE)

// var ClientHandshakeC = make(chan *entities.ClientHandshake, CHANNEL_SIZE)
 var ClientWsSubscriptionChannel = make(chan *entities.ClientWsSubscription, CHANNEL_SIZE)
 var ClientWsSubscriptionChannelV2 = make(chan *entities.ClientWsSubscriptionV2, CHANNEL_SIZE)
// var IncomingDeliveryProofsC = make(chan *entities.DeliveryProof, CHANNEL_SIZE)
// var OutgoingDeliveryProof_BlockC = make(chan *entities.Block, CHANNEL_SIZE)
// var OutgoingDeliveryProofC = make(chan *entities.DeliveryProof, CHANNEL_SIZE)
// var PubSubInputBlockC = make(chan *entities.Block, CHANNEL_SIZE)
// var PubSubInputProofC = make(chan *entities.DeliveryProof, CHANNEL_SIZE)
// var PublishedSubC = make(chan *entities.Subscription, CHANNEL_SIZE)




func init() {

}