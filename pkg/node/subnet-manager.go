package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/service"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
)

// SubnetManager manages the goroutines for each subnet
type SubnetManager struct {
	mu       sync.Mutex
	subnets  map[string]chan *entities.Event       // Map of subnet channels
	wg       sync.WaitGroup              // WaitGroup to track active goroutines
	timeouts map[string]context.CancelFunc // Map of cancel functions for subnets
	Context *context.Context
}

// NewSubnetManager creates a new SubnetManager
func NewSubnetManager(c *context.Context) *SubnetManager {
	return &SubnetManager{
		subnets:  make(map[string]chan *entities.Event),
		timeouts: make(map[string]context.CancelFunc),
		Context: c,
	}
}

// HandleEvent handles an incoming event and manages the subnet goroutine
func (sm *SubnetManager) HandleEvent(event *entities.Event) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if the subnet goroutine already exists
	ch, exists := sm.subnets[event.Subnet]
	if !exists {
		// Create a new channel and goroutine for the subnet
		ch = make(chan *entities.Event, 1000)
		subn := event.Subnet
		if subn == "" {
			subn = "subn"
		}
		sm.subnets[event.Subnet] = ch

		// Create a cancellable context for the goroutine
		ctx, cancel := context.WithCancel(*sm.Context)
		sm.timeouts[event.Subnet] = cancel

		// Start the goroutine
		sm.wg.Add(1)
		go sm.processSubnet(ctx, event.Subnet, ch)
	}

	// Send the event to the subnet's channel
	ch <- event
}

// processSubnet processes events for a specific subnet
func (sm *SubnetManager) processSubnet(ctx context.Context, subnet string, ch chan *entities.Event) {
	defer sm.wg.Done()

	fmt.Printf("Started goroutine for subnet: %s\n", subnet)
	timer := time.NewTimer(5 * time.Second) // Timeout duration

	for {
		select {
		case <-ctx.Done(): // Subnet goroutine cancelled
			fmt.Printf("Stopping goroutine for subnet: %s\n", subnet)
			return
		case event := <-ch: // Process incoming event
			modelType := event.GetDataModelType()
		
			logger.Debugf("StartedProcessingEvent \"%s\" in Subnet: %s", event.ID, event.Subnet)
			cfg, ok := (*sm.Context).Value(constants.ConfigKey).(*configs.MainConfiguration)
			if !ok {
				logger.Errorf("unable to get config from context")
				return
			}
			if event.Validator != entities.PublicKeyString(hex.EncodeToString(cfg.PublicKeyEDD)) {
				isValidator, err := chain.NetworkInfo.IsValidator(string(event.Validator))
				if err != nil {
					logger.Error(err)
					return
				}
				if !isValidator {
					logger.Error(fmt.Errorf("not signed by a validator"))
					return
				}
				event.Broadcasted = true
				service.SaveEvent(modelType, entities.Event{}, event, nil, nil)
			}

			// service.SaveEvent(modelType, entities.Event{}, event, nil, nil)
			switch modelType {
				case entities.SubnetModel:
					service.HandleNewPubSubSubnetEvent(event, &ctx)
				case entities.AuthModel:
					service.HandleNewPubSubAuthEvent(event, &ctx)
				case entities.TopicModel:
					service.HandleNewPubSubTopicEvent(event, &ctx)
				case entities.SubscriptionModel:
					service.HandleNewPubSubSubscriptionEvent(event, &ctx)
				case entities.MessageModel:
					service.HandleNewPubSubMessageEvent(event, &ctx)
			}
			go func() {
				syncedBlockMutex.Lock()
				defer syncedBlockMutex.Unlock()
				if chain.NetworkInfo.Synced {
					lastSynced, err := ds.GetLastSyncedBlock(sm.Context)
					eventBlock := new(big.Int).SetUint64(event.BlockNumber)
					if err == nil && lastSynced.Cmp(eventBlock) == -1 {
						ds.SetLastSyncedBlock(sm.Context, eventBlock)
					}
				}
			}()
			
			if !timer.Stop() {
				<-timer.C // Drain the timer channel if necessary
			}
			timer.Reset(5 * time.Second) // Reset the timeout timer
		case <-timer.C: // Timeout, no events received
			fmt.Printf("No events for 5 seconds. Stopping subnet: %s\n", subnet)
			sm.stopSubnet(subnet)
			return
		}
	}
}

// stopSubnet stops the goroutine and cleans up resources for a subnet
func (sm *SubnetManager) stopSubnet(subnet string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if cancel, exists := sm.timeouts[subnet]; exists {
		cancel() // Cancel the context
		delete(sm.timeouts, subnet)
	}

	if ch, exists := sm.subnets[subnet]; exists {
		close(ch) // Close the channel
		delete(sm.subnets, subnet)
	}
}
