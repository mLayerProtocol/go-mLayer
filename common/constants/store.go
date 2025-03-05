package constants

type DataStore string

const (
	NewClientPayloadStore DataStore = "new-client-payload-store"
	EventStore DataStore = "event-store"
	ValidStateStore DataStore = "valid-state-store"
	MessageStateStore DataStore = "message-state-store"
	RefDataStore DataStore = "entity-refs"
	// ValidMessageStore             DataStore = "valid-messages"
	// UnsentMessageStore                      = "unsent-messages"
	// SentMessageStore                        = "sent-messages"
	NewTopicSubscriptionStore          DataStore     = "new-topic-subscription"
	TopicSubscriptionStore              DataStore    = "top-subscriptions"
	TopicSubscriptionCountStore             = "topic-subscription-count"
	DeliveryProofStore                      = "delivery-proof-store"
	UnconfirmedDeliveryProofStore           = "unconfirmed-delivery-proof-store"
	DeliveryProofBlockStateStore            = "delivery-proof-block-state-store"
	SubscriptionBlockStateStore             = "subscription-block-state-store"
	DeliveryProofBlockStore                 = "dp-block-store"
	SubscriptionBlockStore                  = "sub-block-store"
	
	EventCountStore                 DataStore		= "event-count-store"
	ClaimedRewardStore               DataStore  		= "claimed-reward-store"
	P2PDhtStore                 DataStore		= "p2p-data-store"
	SystemStore                 DataStore		= "system-store"
	NetworkStatsStore                 DataStore		= "network-stats-store"
	NodeTopicsStore                 DataStore		= "node-topics-store"
	GlobalHandlerStore	DataStore		= "global-handler-store"
	MempoolStore	DataStore		= "mempool-store"
)
