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
 
const MAX_SUBNET_IDLE_TIME = 10 * time.Second
// EventProcessor manages the goroutines for each subnet
type EventProcessor struct {
	mu       sync.Mutex
	subnets  map[string]chan *entities.Event       // Map of subnet channels
	wg       sync.WaitGroup              // WaitGroup to track active goroutines
	timeouts map[string]context.CancelFunc // Map of cancel functions for subnets
	Context *context.Context
	config *configs.MainConfiguration
}

// NewEventProcessor creates a new EventProcessor
func NewEventProcessor(c *context.Context) *EventProcessor {
	cfg := (*c).Value(constants.ConfigKey).(*configs.MainConfiguration)
	return &EventProcessor{
		subnets:  make(map[string]chan *entities.Event),
		timeouts: make(map[string]context.CancelFunc),
		Context: c,
		config: cfg,
	}
}

// HandleEvent handles an incoming event and manages the subnet goroutine
func (ep *EventProcessor) HandleEvent(event *entities.Event) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	// Check if the subnet goroutine already exists
	ch, exists := ep.subnets[event.Subnet]
	if !exists {
		// Create a new channel and goroutine for the subnet
		ch = make(chan *entities.Event, 1000)
		subn := event.Subnet
		if subn == "" {
			subn = "subn"
		}
		ep.subnets[subn] = ch

		// Create a cancellable context for the goroutine
		ctx, cancel := context.WithCancel(*ep.Context)
		ep.timeouts[event.Subnet] = cancel

		// Start the goroutine
		ep.wg.Add(1)
		ep.processSubnet(ctx, event.Subnet, ch)
	}

	// Send the event to the subnet's channel
	ch <- event
}

// processSubnet processes events for a specific subnet
func (ep *EventProcessor) processSubnet(ctx context.Context, subnet string, ch chan *entities.Event) {
	defer ep.wg.Done()

	logger.Debug("Started goroutine for subnet: %s\n", subnet)
	timer := time.NewTimer(MAX_SUBNET_IDLE_TIME) // Timeout duration

	for {
		select {
		case <-ctx.Done(): // Subnet goroutine cancelled
			logger.Debugf("Stopping goroutine for subnet: %s\n", subnet)
			return
		case event := <-ch: // Process incoming event
			modelType := event.GetDataModelType()
			
			logger.Debugf("StartedProcessingEvent \"%s\" in Subnet: %s", event.ID, event.Subnet)
			cfg, ok := (*ep.Context).Value(constants.ConfigKey).(*configs.MainConfiguration)
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

	
			service.HandleNewPubSubEvent(*event, ep.Context)
			go func() {
				syncedBlockMutex.Lock()
				defer syncedBlockMutex.Unlock()
				if chain.NetworkInfo.Synced {
					lastSynced, err := ds.GetLastSyncedBlock(ep.Context)
					eventBlock := new(big.Int).SetUint64(event.BlockNumber)
					if err == nil && lastSynced.Cmp(eventBlock) == -1 {
						ds.SetLastSyncedBlock(ep.Context, eventBlock)
					}
				}
			}()
			
			if !timer.Stop() {
				<-timer.C // Drain the timer channel if necessary
			}
			timer.Reset(MAX_SUBNET_IDLE_TIME)
		case <-timer.C: // Timeout, no events received
			logger.Debugf("No events for 10 seconds. Stopping subnet: %s\n", subnet)
			ep.stopSubnet(subnet)
			return
		}
	}
}

// stopSubnet stops the goroutine and cleans up resources for a subnet
func (ep *EventProcessor) stopSubnet(subnet string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if cancel, exists := ep.timeouts[subnet]; exists {
		cancel() // Cancel the context
		delete(ep.timeouts, subnet)
	}

	if ch, exists := ep.subnets[subnet]; exists {
		close(ch) // Close the channel
		delete(ep.subnets, subnet)
	}
}
