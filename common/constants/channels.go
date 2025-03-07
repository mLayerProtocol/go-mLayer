package constants

// Channel Ids within main context
type ChannelsId string
const (
	// BroadcastAuthorizationEventChId 		ChannelsId = "BroadcastAuthorizationEventChannel"
	// BroadcastTopicEventChId             	ChannelsId  = "BroadcastTopicEventChannel"
	// BroadcastApplicationEventChId               			= "BroadcastApplicationEventChannel"
	// IncomingAuthorizationEventChId         = "IncomingAuthorizationEventChannel"
	// IncomingTopicEventChId 							= "IncomingTopicEventChannel"

	// OutgoingMessageChId     						= "OutgoingMessageChannel"
	// OutgoingMessageDP2PChId 						= "OutgoingMessageDP2PChannel"
	// IncomingMessageChId     						= "IncomingMessageChannel"

	// PublishMessageChId              				= "PublishMessageChannel"
	// SubscribeChId                   				= "SubscribeChannel"
	// SubscriptionDP2PChId            				= "SubscriptionDP2PChannel"
	ClientHandShackChId             			ChannelsId	= "ClientHandshakeChannel"
	// OutgoingDeliveryProof_BlockChId 				= "OutgoingDeliveryProofBlockChannel"
	// OutgoingDeliveryProofChId       				= "OutgoinDeliveryProofChannel"
	// PubSubBlockChId                 				= "PubSubBlockChannel"
	// PubsubDeliverProofChId          				= "PubsubProofChannel"
	// PublishedSubChId                				= "PublishedSubChannel"
	EventCountChId                					= "EventCountChannel"
	WSClientLogId					ChannelsId		= "WSClientLogChannel"
)
