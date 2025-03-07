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
 
const MAX_APP_IDLE_TIME = 10 * time.Second
// EventProcessor manages the goroutines for each app
type EventProcessor struct {
	mu       sync.Mutex
	// apps  map[string]chan *entities.Event       // Map of app channels
	apps sync.Map     // Map of app channels
	wg       sync.WaitGroup              // WaitGroup to track active goroutines
	timeouts map[string]context.CancelFunc // Map of cancel functions for apps
	Context *context.Context
	config *configs.MainConfiguration
	responseMap sync.Map 
}
 


// NewEventProcessor creates a new EventProcessor
func NewEventProcessor(c *context.Context) *EventProcessor {
	cfg := (*c).Value(constants.ConfigKey).(*configs.MainConfiguration)
	return &EventProcessor{
		// apps:  make(map[string]chan *entities.Event),
		timeouts: make(map[string]context.CancelFunc),
		Context: c,
		config: cfg,
	}
}

// HandleEvent handles an incoming event and manages the app goroutine
func (ep *EventProcessor) HandleEvent(event *entities.Event, responseChannel *chan *entities.EventProcessorResponse) {
	// ep.mu.Lock()
	// defer ep.mu.Unlock()
	if responseChannel != nil {
		ep.responseMap.Store(event.ID, responseChannel)
	}
	// Check if the app goroutine already exists
	// ch, exists := ep.apps[event.Application]
	var ch *chan *entities.Event 
	// logger.Debugf("Handling event: %+v\n, %v", event, exists)
	 value, ok := ep.apps.Load(event.Application);
	 if ok {
		// fmt.Println("Value for 'foo':", value) // Output: Value for 'foo': 42
		ch = value.(*(chan *entities.Event))
	}
	if !ok {
		logger.Debug("CreatingChannel...")
		// Create a new channel and goroutine for the app
		chm := make(chan *entities.Event, 1000)
		ch = &chm
		subn := event.Application
		if subn == "" {
			subn = "subn"
		}
		// ep.apps[subn] = ch
		 ep.apps.Store(subn, ch)
	
		// Create a cancellable context for the goroutine
		ctx, cancel := context.WithCancel(*ep.Context)
		ep.timeouts[event.Application] = cancel

		// Start the goroutine
		ep.wg.Add(1)
		modelType := entities.GetModelTypeFromEventType(constants.EventType(event.EventType))
		app := event.Payload.Application
		switch modelType {
			case entities.AuthModel:
				app =  event.Payload.Data.(entities.Authorization).Application
			case entities.TopicModel:
				app =  event.Payload.Data.(entities.Topic).Application
			case entities.SubscriptionModel:
				app =  event.Payload.Data.(entities.Subscription).Application
			case entities.MessageModel:
				app =  event.Payload.Data.(entities.Message).Application
		}
		logger.Debugf("Started goroutine for app: %s, %s\n", app, modelType)
		go ep.processApplication(ctx, app, ch)
	}

	// Send the event to the app's channel
	*ch <- event
}

// processApplication processes events for a specific app
func (ep *EventProcessor) processApplication(ctx context.Context, app string, eventChannel *chan *entities.Event) {
	defer ep.wg.Done()
	var resp *entities.EventProcessorResponse = &entities.EventProcessorResponse{}
	

	
	timer := time.NewTimer(MAX_APP_IDLE_TIME) // Timeout duration
	
	for {
		select {
		case <-ctx.Done(): // Application goroutine cancelled
			logger.Debugf("Stopping goroutine for app: %s\n", app)
			return
		case event := <-*eventChannel: // Process incoming event
			// modelType := event.GetDataModelType()
			responseChannel, ok := ep.responseMap.LoadAndDelete(event.ID)
			if ok {
				defer func() {
						respCh := responseChannel.(*chan *entities.EventProcessorResponse)
						*respCh  <- resp
				}()
			}
			logger.Debugf("StartedProcessingEvent \"%s\" in Application: %s", event.ID, event.Application)
			cfg, ok := (*ep.Context).Value(constants.ConfigKey).(*configs.MainConfiguration)
			if !ok {
				logger.Errorf("unable to get config from context")
				return
			}
			if event.Validator != entities.PublicKeyString(hex.EncodeToString(cfg.PublicKeyEDD)) {
				isValidator, err := chain.NetworkInfo.IsValidator(string(event.Validator))
				if err != nil {
					logger.Error("processApplication/IsValidator: ", err)
					return
				}
				if !isValidator {
					logger.Error(fmt.Errorf("not signed by a validator"))
					return
				}
				event.Broadcasted = true
				// service.SaveEvent(modelType, entities.Event{}, event, nil, nil)
			}

			respP, err := service.HandleNewPubSubEvent(*event,
				ep.Context)
			if err != nil {
				logger.Errorf("PUBSUBEERRORHANLING  %s, %+v, %v",  event.ID, event.Payload, err)
				resp.Error = err.Error()
			} else {
				resp = respP
			}
			
			
			
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
			timer.Reset(MAX_APP_IDLE_TIME)
		case <-timer.C: // Timeout, no events received
			logger.Debugf("No events for 10 seconds. Stopping app: %s\n", app)
			ep.stopApplication(app)
			return
		}
	}
	
}

// stopApplication stops the goroutine and cleans up resources for a app
func (ep *EventProcessor) stopApplication(app string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if cancel, exists := ep.timeouts[app]; exists {
		cancel() // Cancel the context
		delete(ep.timeouts, app)
	}
	
	if value, ok := ep.apps.Load(app); ok {
		chm := value.(*chan *entities.Event)
		close(*chm) // Close the channel
		ep.apps.Delete(app)
		// delete(ep.apps, app)
	}
}
