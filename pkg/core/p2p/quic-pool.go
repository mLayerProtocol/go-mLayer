package p2p

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/mlayerprotocol/go-mlayer/internal/crypto"
	"github.com/quic-go/quic-go"
)

// ConnectionState tracks the state of a connection
type ConnectionState int

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
)

// ConnectionInfo tracks metadata about a connection
type ConnectionInfo struct {
    Conn      quic.Connection
    State     ConnectionState
    LastUsed  time.Time
    LastPing  time.Time
    Attempts  int
    NodeAddr  string
    NodeId  string
}

var NodeQuicPool *QuicPool

func init() {
	
	NodeQuicPool = NewQuicPool(false,  WithMaxConns(1000),
    WithMaxIdleTime(time.Minute * 10));
	
}

// QuicPool manages QUIC connections to remote nodes
type QuicPool struct {
    // Connections mapped by node ID
    conns map[string]*ConnectionInfo
    // Protect concurrent access to conns
    mu sync.Map
    
    // Configuration
    maxConns     int
    maxIdleTime  time.Duration
    maxAttempts  int
    pingInterval time.Duration
    
    // TLS config for connections
    tlsConfig *tls.Config
    // QUIC config
    quicConfig *quic.Config
    
    // Channel to signal shutdown
    done chan struct{}
    
    // Function to get node address from ID
    // nodeResolver func(string) (string, error)
}

// NewQuicPool creates a new connection pool
func NewQuicPool(insecure bool, opts ...Option) *QuicPool {
    pool := &QuicPool{
        conns:        make(map[string]*ConnectionInfo),
        maxConns:     1000,
        maxIdleTime:  time.Minute * 5,
        maxAttempts:  3,
        pingInterval: time.Second * 30,
        done:        make(chan struct{}),
        tlsConfig:  &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"mlayer-p2p"},
			VerifyPeerCertificate: func (rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error  {
				if insecure {
					return nil
				}
				for _, cert := range rawCerts {
					if(ValidMads[hex.EncodeToString(crypto.Keccak256Hash(cert))]  == nil) {
						return ErrInvalidCert
					}
				}
				return nil
			},
		},
        quicConfig:  &quic.Config{
            KeepAlivePeriod: time.Second * 10,
            MaxIdleTimeout:  time.Second * 30,
        },
    }
    
    // Apply options
    for _, opt := range opts {
        opt(pool)
    }
    
    // Start maintenance routines
    go pool.maintenance()
    
    return pool
}

// Option allows configuring the pool
type Option func(*QuicPool)

func WithMaxConns(n int) Option {
    return func(p *QuicPool) {
        p.maxConns = n
    }
}

func WithMaxIdleTime(d time.Duration) Option {
    return func(p *QuicPool) {
        p.maxIdleTime = d
    }
}
func  (p *QuicPool) AddConnection(ctx context.Context, remoteAddress string, nodeId string, conn quic.Connection) (*ConnectionInfo, error) {
    // Store connection info
    mut, _ := p.mu.LoadOrStore(remoteAddress, &sync.RWMutex{})
    mu := mut.(*sync.RWMutex)
    mu.RLock()
    defer mu.RUnlock()
    if info, exists := p.conns[remoteAddress]; exists && info.State == StateConnected {
        return info, nil
    }
   
    p.conns[remoteAddress] = &ConnectionInfo{
        Conn:     conn,
        State:    StateConnected,
        LastUsed: time.Now(),
        LastPing: time.Now(),
        NodeAddr: remoteAddress,
        NodeId: nodeId,
    }
    p.conns[nodeId] = p.conns[remoteAddress]
    return p.conns[remoteAddress], nil
}

// GetConnection returns an existing connection or creates a new one
func (p *QuicPool) GetConnection(ctx context.Context, remoteAddress string, nodedId string, reset bool) (quic.Connection, error) {
    // Try to get existing connection
    mut, _ := p.mu.LoadOrStore(remoteAddress, &sync.RWMutex{})
    mu := mut.(*sync.RWMutex)
    if !reset {
       
        mu.RLock()
        if info, exists := p.conns[remoteAddress]; exists && info.State == StateConnected {
            mu.RUnlock()
            info.LastUsed = time.Now()
            return info.Conn, nil
        }
        if nodedId != "" {
            if info, exists := p.conns[nodedId]; exists && info.State == StateConnected {
                if info.NodeAddr == remoteAddress {
                    mu.RUnlock()
                    info.LastUsed = time.Now()
                    info.NodeAddr = remoteAddress
                    return info.Conn, nil
                }            
            }
        }
        mu.RUnlock()
    
    
        // Need to create new connection
        mu.Lock()
        defer mu.Unlock()
        
        // Check again in case another goroutine created the connection
        if info, exists := p.conns[remoteAddress]; exists && info.State == StateConnected {
            info.LastUsed = time.Now()
            return info.Conn, nil
        }
    } else {
        mu.Lock()
        defer mu.Unlock()
    }
    // Create new connection
    conn, err := p.dialNode(ctx, remoteAddress)
    if err != nil {
        return nil, err
    }
    
    // Store connection info
    p.conns[remoteAddress] = &ConnectionInfo{
        Conn:     conn,
        State:    StateConnected,
        LastUsed: time.Now(),
        LastPing: time.Now(),
        NodeAddr: remoteAddress,
    }
    
    return conn, nil
}

// dialNode establishes a new QUIC connection
func (p *QuicPool) dialNode(ctx context.Context, addr string) (quic.Connection, error) {
    conn, err := quic.DialAddr(ctx, addr, p.tlsConfig, p.quicConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to dial QUIC: %w", err)
    }
    return conn, nil
}

// maintenance runs periodic tasks to keep the connection pool healthy
func (p *QuicPool) maintenance() {
    ticker := time.NewTicker(time.Second * 20)
    defer ticker.Stop()
    
    for {
        select {
        case <-p.done:
            return
        case <-ticker.C:
            p.removeStaleConnections()
            p.pingActiveConnections()
        }
    }
}

// removeStaleConnections removes idle and failed connections
func (p *QuicPool) removeStaleConnections() {
    // p.mu.Lock()
    // defer p.mu.Unlock()
    
    now := time.Now()
    for nodeID, info := range p.conns {
        func () {
            mut, _ := p.mu.LoadOrStore(info.NodeAddr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.Lock()
            defer mu.Unlock()
            // Remove if connection is too old
            if now.Sub(info.LastUsed) > p.maxIdleTime {
                info.Conn.CloseWithError(0, "idle timeout")
                delete(p.conns, nodeID)
                return
            }
            // Remove if connection is closed or errored
            if info.Conn.Context().Err() != nil {
                delete(p.conns, nodeID)
                return
            }
        }()
      
    }
}

// pingActiveConnections sends pings to check connection health
func (p *QuicPool) pingActiveConnections() {

    now := time.Now()
    for nodeID, info := range p.conns {
        func() {
            mut, _ := p.mu.LoadOrStore(info.NodeAddr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.Lock()
            defer mu.Unlock()
            // Skip recently pinged connections
            if now.Sub(info.LastPing) < p.pingInterval {
               return
            }

            // Open a stream for ping
            ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
            stream, err := info.Conn.OpenStreamSync(ctx)
            cancel()

            if err != nil {
                // Connection might be dead
                info.State = StateDisconnected
                info.Attempts++
                if info.Attempts > p.maxAttempts {
                    delete(p.conns, nodeID)
                }
                return
            }

            // Close stream immediately - we just wanted to check if we could open it
            stream.Close()
            info.LastPing = now
            info.State = StateConnected
            info.Attempts = 0
        }()
        
    }
}

// HandleNodeOnline marks a node as available and establishes connection if needed
func (p *QuicPool) HandleNodeOnline(nodeID string, addr string) {
    mut, _ := p.mu.LoadOrStore(addr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.Lock()
            defer mu.Unlock()
    info, exists := p.conns[nodeID]
    if !exists {
        // Node not in pool, add it
        ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
        defer cancel()
        
        if conn, err := p.dialNode(ctx, addr); err == nil {
            p.conns[nodeID] = &ConnectionInfo{
                Conn:     conn,
                State:    StateConnected,
                LastUsed: time.Now(),
                LastPing: time.Now(),
                NodeAddr: addr,
            }
        }
    } else if info.State == StateDisconnected {
        // Try to reconnect
        ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
        defer cancel()
        
        if conn, err := p.dialNode(ctx, addr); err == nil {
            info.Conn = conn
            info.State = StateConnected
            info.LastUsed = time.Now()
            info.LastPing = time.Now()
            info.Attempts = 0
        }
    }
}

// HandleNodeOffline marks a node as offline
func (p *QuicPool) HandleNodeOffline(nodeID string) {
   
    
    if info, exists := p.conns[nodeID]; exists {
       
        mut, _ := p.mu.LoadOrStore(info.NodeAddr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.Lock()
            defer mu.Unlock()
        info.State = StateDisconnected
        info.Conn.CloseWithError(0, "node offline")
        delete(p.conns, nodeID)
    }
}

// Close shuts down the connection pool
func (p *QuicPool) Close() error {
    close(p.done)
    
    for _, info := range p.conns {
         func () {
        mut, _ := p.mu.LoadOrStore(info.NodeAddr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.Lock()
            defer mu.Unlock()
        info.Conn.CloseWithError(0, "pool shutdown")
         }()
    }
    p.conns = make(map[string]*ConnectionInfo)
    
    return nil
}

// GetStats returns current pool statistics
func (p *QuicPool) GetStats() map[string]interface{} {
   
    
    stats := make(map[string]interface{})
    stats["total_connections"] = len(p.conns)
    
    active := 0
    disconnected := 0
    for _, info := range p.conns {
        func () {
        mut, _ := p.mu.LoadOrStore(info.NodeAddr, &sync.RWMutex{})
            mu := mut.(*sync.RWMutex)
            mu.RLock()
            defer mu.RUnlock()
        if info.State == StateConnected {
            active++
        } else {
            disconnected++
        }
    }()
    }
    
    stats["active_connections"] = active
    stats["disconnected_connections"] = disconnected
    
    return stats
}

