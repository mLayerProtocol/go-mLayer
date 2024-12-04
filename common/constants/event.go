package constants

//* Matrix Network Events *//
// m.room.message: The most common event for sending messages in a conversation. It can include text messages, images, videos, etc.
// m.room.encrypted: Used for sending encrypted messages.
// m.room.redaction: Used for redacting (or "deleting") previously sent entities.
// m.call.invite, m.call.candidates, m.call.answer, m.call.hangup: Events used for VoIP calls.
// m.reaction: Used for adding reactions to messages.
// m.room.member:

// m.room.create: The first event in a room, which creates the room.
// m.room.name: Sets the room's name.
// m.room.topic: Sets the room's topic or description.
// m.room.avatar: Sets the room's avatar (image).
// m.room.power_levels: Defines the power levels (permissions) for users in the room.
// m.room.history_visibility: Controls the visibility of the room's history.
// m.room.guest_access: Controls whether guests can join the room.
// m.room.join_rules: Defines how users can join the room (e.g., public, invite-only).
// m.room.third_party_invite: Represents an invitation to the room via a third-party service.
// m.room.pinned_events: Indicates messages that are pinned in the room.

// type EventPayloadType string

// const (
// 	AuthorizationPayloadType EventPayloadType = "auth"
// 	TopicPayloadType         EventPayloadType = "topic"
// 	SubscriptionPayloadType  EventPayloadType = "subscription"
// 	MessagePayloadType       EventPayloadType = "message"
// 	SubnetPayloadType        EventPayloadType = "sub_network"
// 	WalletPayloadType        EventPayloadType = "wallet"
// )



type EventType uint16

// Authrization


// Administrative Subnet Actions
const (
	DeleteSubnetEvent EventType = 500
	CreateSubnetEvent EventType = 501 // m.room.create
	// PrivacySetEvent        EventType = 1002
	// BanMemberEvent         EventType = 1003
	// UnbanMemberEvent       EventType = 1004
	// ContractSetEvent       EventType = 1005
	// UpdateNameEvent        EventType = 1006 //  m.room.name
	// UpdateDescriptionEvent EventType = 1007 //  m.room.topic
	// UpdateAvatarEvent      EventType = 1008 //  m.room.avatar
	// PinMessageEvent        EventType = 1008 //  m.room.avatar
	UpdateSubnetEvent EventType = 509
	// UpgradeSubscriberEvent EventType = 1010
)
const (
	AuthorizationEvent   EventType = 600
	UnauthorizationEvent EventType = 601
)

// Administrative Topic Actions
const (
	DeleteTopicEvent       EventType = 1000
	CreateTopicEvent       EventType = 1001 // m.room.create
	PrivacySetEvent        EventType = 1002
	BanMemberEvent         EventType = 1003
	UnbanMemberEvent       EventType = 1004
	ContractSetEvent       EventType = 1005
	UpdateNameEvent        EventType = 1006 //  m.room.name
	UpdateDescriptionEvent EventType = 1007 //  m.room.topic
	UpdateAvatarEvent      EventType = 1008 //  m.room.avatar
	PinMessageEvent        EventType = 1008 //  m.room.avatar
	UpdateTopicEvent       EventType = 1009
	UpgradeSubscriberEvent EventType = 1010
)

// Subscription Actions
const (
	LeaveEvent          EventType = 1100
	SubscribeTopicEvent EventType = 1101
	RequestedEvent      EventType = 1102
	ApprovedEvent       EventType = 1103
	InvitedEvent        EventType = 1104
)

// Message Actions
const (
	DeleteMessageEvent EventType = 1200 //m.room.encrypted
	SendMessageEvent   EventType = 1201 // m.room.message
	// CreateReactionEvent EventType = 1202 // m.reaction
	// IsTypingEvent       EventType = 1203
)



// Administrative Wallet Actions
const (
	DeleteWalletEvent EventType = 1400
	CreateWalletEvent EventType = 1401 // m.room.create

	UpdateWalletEvent EventType = 1409
)
