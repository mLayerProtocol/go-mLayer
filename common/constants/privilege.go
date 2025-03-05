package constants

type AuthorizationPrivilege uint8

const (
	UnauthorizedPriviledge  AuthorizationPrivilege = 0
	GuestPriviledge  AuthorizationPrivilege = 10
	MemberPriviledge AuthorizationPrivilege = 20
	ManagerPriviledge AuthorizationPrivilege = 30
	AdminPriviledge AuthorizationPrivilege = 40
)
type SubscriberRole uint8

var (
	TopicUnsubscribed SubscriberRole = 0
	TopicWriterRole  SubscriberRole = 10
	TopicManagerRole SubscriberRole = 20
	TopicAdminRole  SubscriberRole = 30
)

type SubscriptionStatus int16

var (
	UnsubscribedSubscriptionStatus SubscriptionStatus = 0
	InvitedSubscriptionStatus      SubscriptionStatus = 10
	PendingSubscriptionStatus      SubscriptionStatus = 20
	SubscribedSubscriptionStatus   SubscriptionStatus = 30
	// ApprovedSubscriptionStatus      SubscriptionStatus = "approved"
	BannedSubscriptionStatus SubscriptionStatus = 40
	// UNBANNED     SubscriptionStatus = "unbanned"
)