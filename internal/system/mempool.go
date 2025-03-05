package system

import (
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
)

// mempoolWorker manages batch processing of key-value data
var Mempool *mempoolWorker

type mempoolWorker struct {
	db           *ds.Datastore
	batchChannel *chan *entities.KeyByteValue
	interval     time.Duration
	ttl time.Duration
	buffer map[string]*[]byte
	// mu sync.Mutex
}

// NewmempoolWorker initializes and returns a new mempoolWorker
func NewMempoolWorker(db *ds.Datastore, interval time.Duration, ttl time.Duration, channel *chan *entities.KeyByteValue) (*mempoolWorker, error) {
	return &mempoolWorker{
		db:           db,
		batchChannel: channel,
		interval:     interval,
		ttl: ttl,
		buffer: map[string]*[]byte{},
	}, nil
}

// Start listening for batches and processing them at intervals
func (mw *mempoolWorker) Start() {
	go mw.storeBatches()
}

func (mw *mempoolWorker) Write(kv *entities.KeyByteValue) {
	*mw.batchChannel <- kv
}



// storeBatches listens for batches and writes them to BadgerDB at intervals
func (mw *mempoolWorker) storeBatches() {
	ticker := time.NewTicker(mw.interval)
	defer ticker.Stop()

	// var batchBuffer map[string]*[]byte

	for {
		select {
		case kv, ok := <- *mw.batchChannel:
			if !ok {
				// Channel closed, flush remaining data
				if len(mw.buffer) > 100 {
					mw.buffer[kv.Key] = &kv.Value
					mw.writeToBadger()
				} 
				return
			}
			mw.buffer[kv.Key] = &kv.Value

		case <-ticker.C:
			if len(mw.buffer) > 0 {
				mw.writeToBadger()
				// batchBuffer = nil
			}
		}
	}
}

// writeToBadger writes a batch of key-value pairs to BadgerDB
func (mw *mempoolWorker) writeToBadger() {
	// mw.mu.Lock()
	// defer mw.mu.Unlock()
	
	err := mw.db.DB.Update(func(txn *badger.Txn) error {
		for k, v := range mw.buffer {
			fmt.Printf("WritingKeyToMempool %s", k)
			var e *badger.Entry
			if mw.ttl > 0 {
				e = badger.NewEntry([]byte(k), *v).WithTTL(mw.ttl)
			} else {
				e = badger.NewEntry([]byte(k), *v)
			}
			if err := txn.SetEntry(e); err != nil {
				return err
			}
			delete(mw.buffer, k)
		}
	
		return nil
	})

	if err != nil {
		log.Println("Failed to store batch:", err)
	} else {
		
		log.Println("Stored batch of", len(mw.buffer), "items")
	}
}



// writeToBadger writes a batch of key-value pairs to BadgerDB
func (mw *mempoolWorker) GetData(key string) (v *[]byte, err error) {
	if v, ok := mw.buffer[key]; ok {
		return v, nil
	}
	err = mw.db.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil  {
			// if err != badger.ErrKeyNotFound {
			// 	logger.Errorf("KeyNotFound")
			// 	return err
			// }
			return err
		}
		return item.Value(func(val []byte) error {
			// kv = &entities.KeyByteValue{Key: key, Value: val}
			v = &val
			return nil
		})
	})

	if err != nil {
		log.Println("Failed to store batch:", err)
		return nil, err
	}
	return v, err
}