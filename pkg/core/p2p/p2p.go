package p2p

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/channelpool"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/db"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/p2p/notifee"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"

	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	noise "github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/sirupsen/logrus"
	// rest "messagingprotocol/pkg/core/rest"
	// dhtConfig "github.com/libp2p/go-libp2p-kad-dht/internal/config"
)

var logger = &log.Logger
var Delimiter = []byte{'0'}
var Host host.Host

// var config configs.MainConfiguration
type P2pChannelFlow int8

const (
	P2pChannelOut P2pChannelFlow = 1
	P2pChannelIn  P2pChannelFlow = 2
)


// var privKey crypto.PrivKey
var config *configs.MainConfiguration
var handShakeProtocolId = "mlayer/handshake/1.0.0"
var p2pProtocolId string
var syncProtocolId = "mlayer/sync/1.0.0"
// var P2pComChannels = make(map[string]map[P2pChannelFlow]chan P2pPayload)


const (
	AuthorizationChannel string = "ml-authorization-channel"
	TopicChannel         string = "ml-topic-channel"
	SubnetChannel        string = "ml-sub-network-channel"
	WalletChannel        string = "ml-wallet-channel"
	MessageChannel       string = "ml-message-channel"
	SubscriptionChannel         = "ml-subscription-channel"
	// UnSubscribeChannel                = "ml-unsubscribe-channel"
	// ApproveSubscriptionChannel        = "ml-approve-subscription-channel"
	BatchChannel         = "ml-batch-channel"
	DeliveryProofChannel = "ml-delivery-proof"
)

// var PeerStreams = make(map[string]peer.ID)
var PeerPubKeys = make(map[peer.ID][]byte)
var DisconnectFromPeer = make(map[peer.ID]bool)
var MainContext *context.Context

// var node *host.Host
var idht *dht.IpfsDHT

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	// TODO load name from flag/config
	return fmt.Sprintf("%s-%s", os.Getenv("USER"), shortID(p))
}

// shortID returns the last 8 chars of a base58-encoded peer id.
func shortID(p peer.ID) string {
	pretty := p.String()
	return pretty[len(pretty)-12:]
}

func discover(ctx context.Context, h host.Host, kdht *dht.IpfsDHT, rendezvous string) {
	// kdht.PutValue(ctx, "user/name", []byte("femi"))
	// v, err := kdht.GetValue(ctx, "user/name")
	// if err != nil {
	// 	logger.Error("KDHTERROR", err)
	// }
	// logger.Debugf("VALUEEEEE %s", string(v))
	routingDiscovery := drouting.NewRoutingDiscovery(kdht)
	dutil.Advertise(ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := routingDiscovery.FindPeers(ctx, rendezvous)
			if err != nil {
				logger.Error(err)
				continue
			}

			for p := range peers {

				if p.ID == h.ID() {
					continue
				}

				if h.Network().Connectedness(p.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, p.ID)
					if err != nil {
						logger.Debugf("Failed to connect to peer: %s \n%s", p.ID.String(), err.Error())
						h.Peerstore().RemovePeer(p.ID)
						kdht.ForceRefresh()
						continue
					}
					if len(p.ID) == 0 {
						continue
					}
					logger.Debugf("Connected to discovered peer: %s at %s \n", p.ID.String(), p.Addrs)
					handleConnect(&h, &p)
				}
			}
		}
	}
}

func Run(mainCtx *context.Context) {
	// fmt.Printf("publicKey %s", privateKey)
	// The context governs the lifetime of the libp2p node.
	// Cancelling it will stop the the host.
	logger.Debugf("LoadedChainInfo: %v", chain.NetworkInfo)
	ctx, cancel := context.WithCancel(*mainCtx)
	MainContext = &ctx
	defer cancel()

	cfg, ok := ctx.Value(constants.ConfigKey).(*configs.MainConfiguration)
	config = cfg
	if !ok {
		panic("Unable to load config from context")
	}

	p2pDataStore := db.New(&ctx, string(constants.P2PDataStore))
	defer p2pDataStore.Close()

	if !ok {
		panic("Unable to load data store from context")
	}

	p2pProtocolId = config.ProtocolVersion

	// incomingAuthorizationC, ok := ctx.Value(constants.IncomingAuthorizationEventChId).(*chan *entities.Event)
	// if !ok {
	// 	panic(apperror.Internal("incomingAuthorizationC channel closed"))
	// }

	// incomingTopicEventC, ok := ctx.Value(constants.IncomingTopicEventChId).(*chan *entities.Event)
	// if !ok {
	// 	panic(apperror.Internal("incomingTopicEventC channel closed"))
	// }

	// incomingMessagesC, ok := ctx.Value(constants.IncomingMessageChId).(*chan *entities.ClientPayload)
	// if !ok {

	// }
	// outgoinMessageC, ok := ctx.Value(utils.OutgoingMessageDP2PChId).(*chan *entities.ClientPayload)
	// if !ok {

	// }

	// subscriptionC, ok := ctx.Value(constants.SubscriptionDP2PChId).(*chan *entities.Subscription)
	// if !ok {

	// }

	// outgoingDPBlockCh, ok := ctx.Value(constants.OutgoingDeliveryProof_BlockChId).(*chan *entities.Block)
	// outgoingProofCh, ok := ctx.Value(utils.OutgoingDeliveryProofCh).(*chan *utils.DeliveryProof)
	// publishedSubscriptionC, ok := ctx.Value(constants.SubscribeChId).(*chan *entities.Subscription)
	// if !ok {

	// }

	privKey, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		2048,           // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

	// if len(config.NodePrivateKey) == 0 {
	// 	priv, _, err := crypto.GenerateKeyPair(
	// 		crypto.Ed25519, // Select your key type. Ed25519 are nice short
	// 		-1,             // Select key length when possible (i.e. RSA).
	// 	)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	privKey = priv
	// } else {
	// 	priv, err := crypto.UnmarshalECDSAPrivateKey(hexutil.MustDecode(config.NodePrivateKey))
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	privKey = priv
	// }

	// conMgr := connmgr.NewConnManager(
	// 	100,         // Lowwater
	// 	400,         // HighWater,
	// 	time.Minute, // GracePeriod
	// )

	Host, err = libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(privKey),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(config.ListenerAdresses...),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// libp2p.Transport(ws.New),
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.

		// libp2p.ConnectionManager(connmgr.NewConnManager(
		// 	100,         // Lowwater
		// 	400,         // HighWater,
		// 	time.Minute, // GracePeriod
		// )),

		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts

		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {

			var bootstrapPeers []peer.AddrInfo

			for _, addr := range config.BootstrapPeers {
				addr, _ := multiaddr.NewMultiaddr(addr)
				pi, _ := peer.AddrInfoFromP2pAddr(addr)
				bootstrapPeers = append(bootstrapPeers, *pi)
			}
			var dhtOptions []dht.Option
			dhtOptions = append(dhtOptions,
				dht.BootstrapPeers(bootstrapPeers...),
				dht.ProtocolPrefix(protocol.ID(p2pProtocolId)),
				dht.ProtocolPrefix(protocol.ID(handShakeProtocolId)),
				dht.ProtocolPrefix(protocol.ID(syncProtocolId)),
				dht.Datastore(p2pDataStore),
				dht.NamespacedValidator("pk", record.PublicKeyValidator{}),
				dht.NamespacedValidator("ipns", record.PublicKeyValidator{}),
				dht.NamespacedValidator("ml", &DhtValidator{config: cfg}),
			)
			if !config.BootstrapNode {
				dhtOptions = append(dhtOptions, dht.Mode(dht.ModeServer))
			}
			// dhtOptions = append(dhtOptions,  dht.Datastore(syncDatastore))

			kdht, err := dht.New(ctx, h,
				dhtOptions...)
			if err != nil {
				panic(err)
			}

			// validator = {
			// 	// Validate validates the given record, returning an error if it's
			// 	// invalid (e.g., expired, signed by the wrong key, etc.).
			// 	Validate(key string, value []byte) error

			// 	// Select selects the best record from the set of records (e.g., the
			// 	// newest).
			// 	//
			// 	// Decisions made by select should be stable.
			// 	Select(key string, values [][]byte) (int, error)
			// }
			// dhtOptions = append(dhtOptions, dht.NamespacedValidator("subsc", customValidator))

			//if cfg.BootstrapNode {
			if err = kdht.Bootstrap(ctx); err != nil {
				logger.Fatalf("Error starting bootstrap node %o", err)
				return nil, err
			}
			// }

			idht = kdht

			// for _, addr := range config.BootstrapPeers {
			// 	addr, _ := multiaddr.NewMultiaddr(addr)
			// 	pi, err := peer.AddrInfoFromP2pAddr(addr)
			// 	if err != nil {
			// 		logger.Warnf("Invalid boostrap peer address (%s): %s \n", addr, err)
			// 	} else {
			// 		error := h.Connect(ctx, *pi)
			// 		if error != nil {
			// 			logger.Debugf("Unable connect to boostrap peer (%s): %s \n", addr, err)
			// 			continue
			// 		}
			// 		logger.Debugf("Connected to boostrap peer (%s)", addr)
			// 		h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
			// 		handleConnect(&h, pi)
			// 	}
			// }

			// routingOptions := routing.Options{
			// 	Expired: true,
			// 	Offline: true,
			// }
			// var	routingOptionsSlice []routing.Option;
			// routingOptionsSlice = append(routingOptionsSlice, routingOptions.ToOption())
			// key := "/$name/$first"
			// putErr := kdht.PutValue(ctx, key, []byte("femi"), routingOptions.ToOption())

			// if putErr != nil {
			// 	logger.Debugf("Put the error %o", putErr)
			// }
			return idht, err
		}),
		// libp2p.Relay(options...),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		// libp2p.DefaultEnableRelay(),
		//libp2p.EnableAutoRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
	)

	// gater := NetworkGater{host: h, config: config, blockPeers: make(map[peer.ID]struct{})}

	go discover(ctx, Host, idht, fmt.Sprintf("%s-%s", constants.NETWORK_NAME, config.ChainId))
	if err != nil {
		panic(err)
	}
	Host.Network().Notify(&notifee.ConnectionNotifee{Dht: idht})

	Host.SetStreamHandler(protocol.ID(handShakeProtocolId), handleHandshake)
	Host.SetStreamHandler(protocol.ID(p2pProtocolId), handlePayload)
	Host.SetStreamHandler(protocol.ID(syncProtocolId), handleSync)
	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, Host)
	if err != nil {
		panic(err)
	}
	// setup local mDNS discovery
	err = setupDiscovery(Host, fmt.Sprintf("%s-%s", constants.NETWORK_NAME, config.ChainId))
	if err != nil {
		panic(err)
	}

	// node = &h

	// The last step to get fully up and running would be to connect to
	// bootstrap peers (or any other peers). We leave this commented as
	// this is an example and the peer will die as soon as it finishes, so
	// it is unnecessary to put strain on the network.
	fmt.Println("------------------------------- MLAYER -----------------------------------")
	fmt.Println("- Licence Operator Public Key (SECP): ", hex.EncodeToString(cfg.PublicKeySECP))
	fmt.Println("- Network Public Key (EDD): ", cfg.PublicKey)
	fmt.Println("- Host started with ID: ", Host.ID())
	fmt.Println("- Host Network: ", p2pProtocolId)
	fmt.Println("- Host Listening on: ", Host.Addrs())

	// Subscrbers
	authPubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), AuthorizationChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.AuthorizationPubSub = *authPubSub

	topicPubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), TopicChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.TopicPubSub = *topicPubSub

	subnetPubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), SubnetChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.SubnetPubSub = *subnetPubSub

	walletPubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), WalletChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.WalletPubSub = *walletPubSub

	subscriptionPubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), SubscriptionChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.SubscriptionPubSub = *subscriptionPubSub

	messagePubSub, err := entities.JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), MessageChannel, config.ChannelMessageBufferSize)
	if err != nil {
		panic(err)
	}
	entities.MessagePubSub = *messagePubSub

	// unsubscribePubSub, err := JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), UnSubscribeChannel, config.ChannelMessageBufferSize)
	// if err != nil {
	// 	panic(err)
	// }

	// approveSubscriptionPubSub, err := JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), ApproveSubscriptionChannel, config.ChannelMessageBufferSize)
	// if err != nil {
	// 	panic(err)
	// }

	// Publishers
	go publishChannelEventToNetwork(channelpool.AuthorizationEventPublishC, &entities.AuthorizationPubSub, mainCtx)
	go publishChannelEventToNetwork(channelpool.TopicEventPublishC, &entities.TopicPubSub, mainCtx)
	go publishChannelEventToNetwork(channelpool.SubnetEventPublishC, &entities.SubnetPubSub, mainCtx)
	go publishChannelEventToNetwork(channelpool.WalletEventPublishC, &entities.WalletPubSub, mainCtx)
	go publishChannelEventToNetwork(channelpool.SubscriptionEventPublishC, &entities.SubscriptionPubSub, mainCtx)
	go publishChannelEventToNetwork(channelpool.MessageEventPublishC, &entities.MessagePubSub, mainCtx)
	// go PublishChannelEventToNetwork(channelpool.UnSubscribeEventPublishC, unsubscribePubSub, mainCtx)
	// go PublishChannelEventToNetwork(channelpool.ApproveSubscribeEventPublishC, approveSubscriptionPubSub, mainCtx)

	// Subscribers

	
	// go ProcessEventsReceivedFromOtherNodes(&entities.Subscription{}, unsubscribePubSub, mainCtx, service.HandleNewPubSubUnSubscribeEvent)
	// go ProcessEventsReceivedFromOtherNodes(&entities.Subscription{}, approveSubscriptionPubSub, mainCtx, service.HandleNewPubSubApproveSubscriptionEvent)

	// messagePubSub, err := JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), MessageChannel, config.ChannelMessageBufferSize)
	// if err != nil {
	// 	panic(err)
	// }

	// batchPubSub, err := JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), BatchChannel, config.ChannelMessageBufferSize)
	// if err != nil {
	// 	panic(err)
	//}
	// delieveryProofPubSub, err := JoinChannel(ctx, ps, Host.ID(), defaultNick(Host.ID()), DeliveryProofChannel, config.ChannelMessageBufferSize)
	// if err != nil {
	// 	panic(err)
	// }

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	for {
	// 		select {

	// 		case authEvent, ok := <-authorizationPubSub.Messages:
	// 			if !ok {
	// 				cancel()
	// 				logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 				return
	// 			}
	// 			// !validating message
	// 			// !if not a valid message continue
	// 			// _, err := inMessage.MsgPack()
	// 			// if err != nil {
	// 			// 	continue
	// 			// }
	// 			//TODO:
	// 			// if not a valid message, continue

	// 			logger.Debugf("Received new message %s\n", authEvent.ToString())
	// 			cm := models.AuthorizationEvent{}
	// 			err = encoder.MsgPackUnpackStruct(authEvent.Data, cm)
	// 			if err != nil {

	// 			}
	// 			*incomingAuthorizationC <- &cm
	// 		case inMessage, ok := <-batchPubSub.Messages:
	// 			if !ok {
	// 				cancel()
	// 				logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 				return
	// 			}
	// 			// !validating message
	// 			// !if not a valid message continue
	// 			// _, err := inMessage.MsgPack()
	// 			// if err != nil {
	// 			// 	continue
	// 			// }
	// 			//TODO:
	// 			// if not a valid message, continue

	// 			logger.Debugf("Received new message %s\n", inMessage.ToString())
	// 			cm, err := entities.MsgUnpackClientPayload(inMessage.Data)
	// 			if err != nil {

	// 			}
	// 			*incomingMessagesC <- &cm
	// 		case sub, ok := <-subscriptionPubSub.Messages:
	// 			if !ok {
	// 				cancel()
	// 				logger.Fatalf("Primary Message channel closed. Please restart server to try or adjust buffer size in config")
	// 				return
	// 			}
	// 			// logger.Debug("Received new message %s\n", inMessage.Message.Body.Message)
	// 			cm, err := entities.UnpackSubscription(sub.Data)
	// 			if err != nil {

	// 			}
	// 			logger.Debug("New subscription updates:::", string(cm.ToJSON()))
	// 			// *incomingMessagesC <- &cm
	// 			cm.Broadcasted = false
	// 			*publishedSubscriptionC <- &cm
	// 		}
	// 	}
	// }()
	// if config.Validator {
		
		storeAddress(&ctx, &Host)
	// }
	defer forever()

}

func forever() {
	for {
		time.Sleep(time.Hour)
	}
}

// func handleHandshake(stream network.Stream) {

// 	// // if len(PeerStreams[stream.ID()]) == 0 {
// 	// // 	logger.Debugf("No peer for stream %s Peer %s", stream.ID(), PeerStreams[stream.ID()])
// 	// // 	return
// 	// // }
// 	// logger.Debugf("Got a new stream 2! %s Peer %s", stream.ID(), PeerStreams[stream.ID()])
// 	// // stream.SetReadDeadline()
// 	// // Create a buffer stream for non blocking read and write.
// 	// peer := stPeerStreams[stream.ID()]
// 	// rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
// 	// logger.Debugf("Got a new stream 3! %s Peer %s", stream.ID(), PeerStreams[stream.ID()])
// 	host := idht.Host()
// 	verifyHandshake(&host, &stream)
// 	// go sendData(rw)

// }

func handleHandshake(stream network.Stream) {
	// ctx, _ := context.WithCancel(context.Background())
	// config, _ := ctx.Value(constants.ConfigKey).(*configs.MainConfiguration)
	// defer delete(DisconnectFromPeer, p )
	peerId := (stream).Conn().RemotePeer()

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	for {
		hsData, err := rw.ReadBytes('\n')
		if err != nil {
			logger.Errorf("Error reading from buffer %o", err)
			return
		}
		if hsData == nil {
			//break
			return
		}

		handshake, err := UnpackNodeHandshake(hsData)

		if err != nil {
			logger.WithFields(logrus.Fields{"data": handshake}).Warnf("Failed to parse handshake: %o", err)
			return
			// break
		}
		validHandshake := handshake.IsValid(config.ChainId)

		logger.Debugf("Validating peer %s", (stream).Conn().RemotePeer())
		if !validHandshake {
			disconnect((stream).Conn().RemotePeer())
			return
		}
		if handshake.NodeType == constants.ValidatorNodeType {
			validHandshake, err = chain.NetworkInfo.IsValidator(handshake.Signer)
			if err != nil || !validHandshake {
				disconnect((stream).Conn().RemotePeer())
				return
			}
			// Validate stake as well

			// validStake := isValidStake(handshake, p, config)
			// if !validStake {
			// 	// disconnect(*node, p)
			// 	logger.WithFields(logrus.Fields{"address": handshake.Signer, "data": hsData}).Infof("Disconnecting from peer (%s) with inadequate stake in network", p)
			// 	return
			// }
		} else {
			validHandshake, err = chain.NetworkInfo.IsSentry(handshake.Signer, nil)
			if err != nil || !validHandshake {
				disconnect((stream).Conn().RemotePeer())
				return
			}
		}
		if !chain.NetworkInfo.Synced {
			go chain.NetworkInfo.Sync(MainContext, func () bool {
				return true
			})
		}
		// b, _ := hexutil.Decode(handshake.Signer)
		// PeerPubKeys[p] = b
		// break
		logger.WithFields(logrus.Fields{"peer": peerId, "pubKey": handshake.Signer}).Info("Successfully connected peer with valid handshake")
		delete(DisconnectFromPeer, peerId)
	}
}

func sendHandshake(stream network.Stream, data []byte) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	_, err := rw.WriteString(fmt.Sprintf("%s\n", string(data)))
	if err != nil {
		logger.Warn("Error writing to to stream")
		return
	}

	err = rw.Flush()
	logger.Debugf("Flushed data to stream %s", stream.ID())
	if err != nil {
		// fmt.Println("Error flushing buffer")
		// panic(err)
		logger.Error("Error flushing to handshake stream")
		return
	}
}

func storeAddress(ctx *context.Context, h *host.Host) {
	for {
		if (*h).Peerstore().PeersWithAddrs().Len() == 1 {
			time.Sleep(1 * time.Second)
			continue
		}
		// logger.Debug("Iamavalidator")
		
		mad, err := NewNodeMultiAddressData(config, config.PrivateKeyEDD, GetMultiAddresses(*h), config.PublicKeyEDD)
		if err != nil {
			logger.Error(err)
		}
		key := "/ml/val/" + config.PublicKey
		keySecP := "/ml/val/" + hex.EncodeToString(config.PublicKeySECP)
		// v, err := idht.GetValue(*ctx, key)
		// 	if err != nil {
		// 		logger.Error("KDHT_GET_ERROR: ", err)
		// 	} else {
		// 		logger.Debugf("VALURRRR %s", string(v))
		// 	}
		
		packed :=  mad.MsgPack()
		err = idht.PutValue(*ctx, key, packed)
		
		if err != nil {
			logger.Error("KDHT_PUT_ERROR", err)
		} else {
			logger.Debugf("Successfully stored key: %s", key)
		}

		logger.Debugf("ADDING_SECP_ADDRESS: %d", config.PublicKeySECP)
		err = idht.PutValue(*ctx, keySecP, packed)
		if err != nil {
			logger.Error("KDHT_PUT_ERROR", err)
		} else {
			logger.Debugf("Successfully stored key: %s", keySecP)
		}
		time.Sleep(1 * time.Hour)
		// else {
		// 	time.Sleep(2 * time.Second)
		// 	v, err := idht.GetValue(ctx, key)
		// 	if err != nil {
		// 		logger.Error("KDHT_GET_ERROR", err)
		// 	} else {
		// 		logger.Debugf("VALURRRR %s", string(v))
		// 	}
		// }
	}
}

func GetNodeMultiAddressData(ctx *context.Context, key string) (*NodeMultiAddressData, error) {
	
		key = "/ml/val/" + key
		data, err := idht.GetValue(*ctx, key, nil)
		if err != nil {
			return  nil, err
		}
		mad, err := UnpackNodeMultiAddressData(data)
		if err != nil {
			return nil, err
		}
		cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
		if !mad.IsValid(cfg.ChainId) {
			return nil, fmt.Errorf("invalid multiaddress data")
		}
		return &mad, err
}

// called when a peer connects
func handleConnect(h *host.Host, pairAddr *peer.AddrInfo) {
	// pi := *pa
	logger.Debugf("My multiaddress: %s", GetMultiAddresses(*h))
	if pairAddr == nil {
		return
	}
	DisconnectFromPeer[pairAddr.ID] = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handshakeStream, err := (*h).NewStream(ctx, pairAddr.ID, protocol.ID(handShakeProtocolId))

	if err != nil {
		logger.Warningf("Unable to establish stream with peer: %s %o", pairAddr.ID, err)
	} else {
		nodeType := constants.SentryNodeType
		if config.Validator {
			nodeType = constants.ValidatorNodeType
		}
		hs, _ := NewNodeHandshake(config, handShakeProtocolId, config.PrivateKeyEDD, nodeType)
		// b, _ := hs.EncodeBytes()
		logger.Debugf("Created handshake with salt %s", hs.Salt)
		logger.Debugf("Created new stream %s with peer %s", handshakeStream.ID(), pairAddr.ID)
		defer func(pairID peer.ID) {
			time.Sleep(3 * time.Second)
			if DisconnectFromPeer[pairAddr.ID] {
				handshakeStream.Close()
				disconnect(pairAddr.ID)
			}
		}(pairAddr.ID)
		go sendHandshake(handshakeStream, (*hs).MsgPack())
		go handleHandshake(handshakeStream)

		host := idht.Host()
		networkStream, err := host.NewStream(idht.Context(), pairAddr.ID, protocol.ID(p2pProtocolId))
		if err != nil {
			(networkStream).Reset()
			return
		}
		go handlePayload(networkStream)

		syncStream, err := host.NewStream(idht.Context(), pairAddr.ID, protocol.ID(syncProtocolId))
		if err != nil {
			(syncStream).Reset()
			return
		}
		go handleSync(syncStream)

		// _, pub, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
		//time.Sleep(5 * time.Second)
		// peerID, _ := peer.IDFromPublicKey(pub)
		// logger.Debugf("Waiting to send data to peer")
		// time.Sleep(5 * time.Second)
		// logger.Debugf("Starting send process...")
		// eventBytes := (&entities.EventPath{entities.EntityPath{Model: entities.TopicModel,Hash:"23032", Validator: "02c4435e768b4bae8236eeba29dd113ed607813b4dc5419d33b9294f712ca79ff4"}}).MsgPack()
		// payload := NewP2pPayload(config, P2pActionGetEvent, eventBytes)
		// if err != nil {
		// 	logger.Errorf("ERrror: %s", err)
		// 	return
		// }
		// err = payload.Sign(config.PrivateKeyEDD)
		// if err != nil {
		// 	logger.Debugf("Error SIgning: %v", err)
		// 	return
		// }
		// logger.Debugf("Payload data signed: %s", payload.Id)
		// resp, err := payload.SendRequestToAddress(config.PrivateKeyEDD, multiaddr.StringCast("/ip4/127.0.0.1/tcp/6000/ws/p2p/12D3KooWH7Ch4EETUDfCZAG1aBDUD2WmXukXuDVfpJqxxbVx7jBm"))
		// if err != nil {
		// 	logger.Errorf("Error :%v", err)
		// }
		// logger.Debugf("Resopnse :%v", resp)

	}
}

func disconnect(id peer.ID) {
	idht.Host().Network().ClosePeer(id)
}

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host, serviceName string) error {
	logger.Debugf("Setting up Discovery on %s ....", serviceName)
	n := notifee.DiscoveryNotifee{Host: h, HandleConnect: handleConnect, Dht: idht}

	disc := mdns.NewMdnsService(h, serviceName, &n)
	return disc.Start()
}


func connectToNode(targetAddr multiaddr.Multiaddr, ctx context.Context) (pid *peer.AddrInfo, p2pStream *network.Stream, syncStream *network.Stream,  err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	
	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		logger.Errorf("Failed to get peer info: %v", err)
		return targetInfo, nil, nil, err
	}
	// logger.Debugf("P2PCHANNELIDS %s", connectionId)
	// if P2pComChannels[connectionId] == nil {
		
	// }
	// logger.Debugf("P2PCHANNELIDS %v", P2pComChannels[targetInfo.ID.String()][P2pChannelOut] == nil)
	h := idht.Host()
	if  h.Network().Connectedness(targetInfo.ID) != network.Connected {
		err = h.Connect(ctx, *targetInfo)
	}
	if err != nil {
		logger.Errorf("ErrorConnectingToPeer %v", err)
		h.Peerstore().RemovePeer(targetInfo.ID)
		//delete(P2pComChannels, connectionId)
		return nil, nil, nil, err
	}
	h.Peerstore().AddAddrs(targetInfo.ID, targetInfo.Addrs, peerstore.PermanentAddrTTL)
	// Add the target peer to the host's peerstore
	// logger.Debug("Connectedness: %s", h.Network().Connectedness(targetInfo.ID).String())
	if h.Network().Connectedness(targetInfo.ID) == network.Connected {
		streamz, err := h.NewStream(ctx, targetInfo.ID, protocol.ID(p2pProtocolId))
		if err != nil {
			logger.Errorf("ConnectionError: %v", err)
		}
		syncStreamz, err := h.NewStream(ctx, targetInfo.ID, protocol.ID(syncProtocolId))
		if err != nil {
			logger.Errorf("ConnectionError: %v", err)
		}
		p2pStream = &streamz
		syncStream = &syncStreamz
		logger.Debugf("CreateNewPeerStream")
		
	} 
	

	// Connect to the target node
	return targetInfo, p2pStream, syncStream, nil
}

func GetMultiAddresses(h host.Host) []string {
	m := []string{}
	addrs := h.Addrs()

	for _, addr := range addrs {
		m = append(m, fmt.Sprintf("%s/p2p/%s", addr, h.ID().String()))
	}
	logger.Debugf("MULTI %v", m)
	return m
}

func handlePayload(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	peerId := stream.Conn().RemotePeer()
	// go writePayload(rw, peerId, stream)
	go readPayload(rw, peerId, stream)
}

func handleSync(stream network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	peerId := stream.Conn().RemotePeer()
	// go writePayload(rw, peerId, stream)
	go readPayload(rw, peerId, stream)
}


func readPayload(rw *bufio.ReadWriter, peerId peer.ID, stream network.Stream) {
	
	// defer stream.Close()
	for {
		var payloadBuf bytes.Buffer
		bufferLen := 1024
		buf := make([]byte, bufferLen)
		for {	
			n, err := rw.Read(buf)
			if n > 0 {
				payloadBuf.Write(buf[:n])
			}
			// logger.Debugf("BYTESREAD: %d %d",  n, payloadBuf.Len())
			if err != nil {
				if err == io.EOF {
					break
				}
				break
			}
			if n < bufferLen {
				break;
			}
		}
		pData := payloadBuf.Bytes()
		if len(pData) == 0 {
			return
		}
		payload, err := UnpackP2pPayload(pData[:len(pData)-1])
		if err != nil {
			logger.WithFields(logrus.Fields{"data": len(pData)}).Warnf("Failed to parse payload: %o", err)
			return
			// break
		}
		validPayload := payload.IsValid(config.ChainId)
		logger.Debugf("Received Data from remote peer: %v", validPayload)
		if !validPayload {
			logger.Debugf("Invalid payload received from peer %s", peerId)
			// delete(P2pComChannels, payload.Id)
			(stream).Reset()
			return
		}
		response, err := processP2pPayload(config, payload)
		if err != nil {
			logger.Debugf("readPayload: %v", err)
		}
		delimeter := []byte{'\n'}
		b := response.MsgPack()
		resp, _ := UnpackP2pPayload(b)
		logger.Debugf("BYTESSSSS: %s, %v", resp.Id, err)
		_, err = stream.Write(append(b, delimeter...))
		if err != nil {
			logger.Errorf("readPayload: %v", err)
		}
	
		err = rw.Flush()
		logger.Debugf("Flushed response data to payload stream %s", stream.ID())
		if err != nil {
			// fmt.Println("Error flushing buffer")
			// panic(err)
			logger.Error("Error flushing response to stream")
			return
		}
	}
}
func GetDhtValue(key string)  ([]byte, error) {
	return idht.GetValue(*MainContext, key)
}

// func GetOperatorMultiAddress(pubKey string, chainId configs.ChainId ) (multiaddr.Multiaddr, error) {
// 	key := "/ml/val/" + pubKey
// 	d, err := GetDhtValue(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	md, err := UnpackNodeMultiAddressData(d)
// 	if err != nil {
// 		return nil, err
// 	}
	
// 	if md.IsValid(chainId) {
// 		return multiaddr.NewMultiaddr(md.Addresses[0])
// 	}
// 	return nil, fmt.Errorf("invalid multiaddress ")
// }
// check the dht before going onchain
func GetCycleMessageCost(ctx context.Context, cycle uint64) (*big.Int, error) {
	_, ok := (ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	if !ok {
		return nil, fmt.Errorf("failed to load config")
	}
	claimedRewardStore, ok := (ctx).Value(constants.ClaimedRewardStore).(*db.Datastore)
	if !ok {
		return nil, fmt.Errorf("failed to load store")
	}
	priceKey := datastore.NewKey(fmt.Sprintf("/ml/cost/%d", cycle))
	priceByte, err := claimedRewardStore.Get(ctx, priceKey)
	//
	if err != nil && err != datastore.ErrNotFound {
		return nil, err
	}
	if len(priceByte) > 0 && err == nil {
		// priceData, err := UnpackMessagePrice(priceByte)
		// if err != nil {
		// 	return GetAndSaveMessageCostFromChain(ctx, cycle)
		// }
		return big.NewInt(0).SetBytes(priceByte), nil
	} else {
		return GetAndSaveMessageCostFromChain(&ctx, cycle)
	}
}

func GetAndSaveMessageCostFromChain(ctx *context.Context, cycle uint64) (*big.Int, error) {
	
	cfg, _ := (*ctx).Value(constants.ConfigKey).(*configs.MainConfiguration)
	claimedRewardStore, ok := (*ctx).Value(constants.ClaimedRewardStore).(*db.Datastore)
	if !ok {
		return nil, fmt.Errorf("GetAndSaveMessageCostFromChain: failed to load store")
	}
	price, err := chain.DefaultProvider(cfg).GetMessagePrice(big.NewInt(int64(cycle)))
	if err != nil {
		return nil, err
	}
	logger.Debugf("ITEMPRICE: %s", price)
	priceKey := datastore.NewKey(fmt.Sprintf("/ml/cost/%d", cycle))
	
	err = claimedRewardStore.Put(*ctx, priceKey, utils.ToUint256(price))
	return price, err
}

func GetNodeAddress(ctx *context.Context, pubKey string) (multiaddr.Multiaddr, error) {
	mad, err := GetNodeMultiAddressData(ctx, pubKey)
	if err != nil {
		logger.Error("KDHT_GET_ERROR: ", err)
		return nil, err
	}

	return multiaddr.StringCast(mad.Addresses[0]), nil
}
