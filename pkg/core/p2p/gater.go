package p2p

import (
	"context"
	"io"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
	"github.com/multiformats/go-multiaddr"
)


type NetworkGater struct {
    blockedPeers map[peer.ID]struct{}
    host         host.Host
    config *configs.MainConfiguration
}

// InterceptPeerDial checks if a peer dial should be allowed
func (g *NetworkGater) InterceptPeerDial(p peer.ID) (allow bool) {
    _, blocked := g.blockedPeers[p]
    return !blocked
}

// InterceptAddrDial checks if a dial to a multiaddr should be allowed
func (g *NetworkGater) InterceptAddrDial(peer.ID, multiaddr.Multiaddr) (allow bool) {
    return true
}

// InterceptAccept checks if an incoming connection should be allowed
func (g *NetworkGater) InterceptAccept(network.ConnMultiaddrs) (allow bool) {
    return true
}

// InterceptSecured checks if a secured connection should be allowed
func (g *NetworkGater) InterceptSecured(dir network.Direction, p peer.ID, addrs network.ConnMultiaddrs) (allow bool) {
    if dir == network.DirInbound {
        return g.performHandshake(p)
    }
    return true
}

// InterceptUpgraded checks if an upgraded connection should be allowed
func (g *NetworkGater) InterceptUpgraded(network.Conn) (allow bool, reason string) {
    return true, ""
}

// performHandshake initiates and validates the handshake with the peer
func (g *NetworkGater) performHandshake(p peer.ID) bool {
    ctx := context.Background()
    s, err := g.host.NewStream(ctx, p, protocol.ID(handShakeProtocolId))
    if err != nil {
        log.Printf("Failed to create stream: %v", err)
        return false
    }
    defer s.Close()

    // Send a handshake message
    nodeType := constants.SentryNodeType
		if cfg.Validator {
			nodeType = constants.ValidatorNodeType
		}
        lastSync, _ := ds.GetLastSyncedBlock(g.config.Context)
    message, err := NewNodeHandshake(g.config, handShakeProtocolId, g.config.PrivateKeySECP, g.config.PublicKeyEDD, nodeType, lastSync, utils.RandomAplhaNumString(6))
    if err != nil {
        logger.Error("NewNodeHandshake: ", err)
    }
    msgBytes := message.MsgPack()
    _, err = s.Write(msgBytes)
    if err != nil {
        log.Printf("Failed to write handshake message: %v", err)
        return false
    }

    // Read the response
    buf := make([]byte, 1024)
    _, err = s.Read(buf)
    if err != nil && err != io.EOF {
        log.Printf("Failed to read handshake response: %v", err)
        return false
    }

    handshake, err := UnpackNodeHandshake(buf)
    if err != nil {
        logger.Errorf("Failed to unmarshal handshake response: %v", err)
        return false
    }
    // Validate the handshake response
    if nodeType != constants.ValidatorNodeType {
        return handshake.IsValid(g.config.ChainId)
    }
    // return handshake.IsValid() && handshake.HasValidStake(g.Config)
    return handshake.IsValid(g.config.ChainId) 
}
