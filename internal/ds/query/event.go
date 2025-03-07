package query

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"

	"github.com/mlayerprotocol/go-mlayer/pkg/log"
)

var logger = &log.Logger

var ErrorNotFound = datastore.ErrNotFound
var ErrorKeyNotFound = badger.ErrKeyNotFound
var ErrorKeyExist = fmt.Errorf("key exists")

// var topicCounterCache = cache.NewShardedCache("topicCounter", cache.CacheOptions{NumShards: 10, TTL: 10 * time.Minute, CleanInterval: 12 * time.Minute })


func IsErrorNotFound(e error) bool {
	return e == datastore.ErrNotFound || e == badger.ErrKeyNotFound
}

func GetEventById(id string, modelType entities.EntityModel) (*entities.Event, error) {
	key := (&entities.Event{ID: id}).DataKey()
	
	value, err := stores.EventStore.Get(context.Background(), datastore.NewKey(key))
	
	if err != nil {
		return nil, err
	}
	event, err := entities.UnpackEvent(value, modelType)
	if err != nil {
		logger.Errorf("GetEventByIdError: %v", err)
		return nil, err
	}
	return event, err
}
func GetEventByIdTxn(id string, modelType entities.EntityModel, txn *datastore.Txn) (*entities.Event, error) {
	if txn == nil {
		return GetEventById(id, modelType)
	}
	key := (&entities.Event{ID: id}).DataKey()
	
	
	value, err := (*txn).Get(context.Background(), datastore.NewKey(key))
	if err != nil {
		return nil, err
	}
	event, err := entities.UnpackEvent(value, modelType)
	if err != nil {
		return nil, err
	}
	return event, err
}

func createEvent(event *entities.Event, tx *datastore.Txn, wbParam *badger.WriteBatch ) (err error) {
	defer utils.TrackExecutionTime(time.Now(), "DSQUERY::CreateEvent::")
	ds := stores.EventStore
	
	event.ID, err = event.GetId()
	if err != nil {
		return err
	}
	var txn datastore.Txn
	eventBytes := event.MsgPack()
	keys := event.GetKeys()
	// logger.Debugf("Creating Event: %+v", event)

	// txn, err := InitTx(ds, tx)
	// if tx == nil {
	// 	defer txn.Discard(context.Background())
	// }
	var wb *badger.WriteBatch = wbParam
	if tx == nil && wb == nil {
		wb = ds.DB.NewWriteBatch()
		defer wb.Cancel()
		logger.Infof("UseringWriteBatch")
	} 
	if tx != nil {
		logger.Infof("UsingTXN")
		txn = *tx
	}
	logger.Debugf("CreatingEventWithKey: %v", event)
	for _, key := range keys {
		logger.Debugf("SavingEventWithKey: %s, %v", key, event.ID)
		if key == event.BlockKey() {
			if event.Synced == nil || !*event.Synced {
				continue
			}
			if txn == nil {
				err = wb.Set([]byte(key), []byte{})
				return err
			} 
			if err := txn.Put(context.Background(), datastore.NewKey(key), []byte{}); err != nil {
				return err
			}
			continue
		}
		if key == event.DataKey() {
			// logger.Debugf("SavingEventId: %s, %v", key, event.ID)
			if txn == nil {
				err = wb.Set([]byte(key), eventBytes)
				
			} else {
				err = txn.Put(context.Background(), datastore.NewKey(key), eventBytes)
			}
				if  err != nil {
					return err
				}
			
		} else {
			if txn == nil {
				err = wb.Set([]byte(key), utils.UuidToBytes(event.ID))
			} else {
				err = txn.Put(context.Background(), datastore.NewKey(key), utils.UuidToBytes(event.ID))
			}
			if  err != nil {
				return err
			}
		}

	}
	if tx == nil && wbParam == nil {
		// if err = wb.Flush(); err != nil {
		// 	return (err)
		// }
		return wb.Flush()
	}
	return nil

}

func UpdateEvent(event *entities.Event, tx *datastore.Txn, wbParam *badger.WriteBatch, create bool) (err error) {
	wb := wbParam
	var txn datastore.Txn
	if tx == nil && wbParam == nil {
		wb = stores.EventStore.DB.NewWriteBatch()
	}
	if tx != nil {
		txn, err = InitTx(stores.EventStore, tx)
		if err != nil {
			return err
		}
	}
	
	if tx != nil {
		defer txn.Discard(context.Background())
	}
	_, err = stores.EventStore.Get(context.Background(), datastore.NewKey(event.DataKey()))
	// logger.Infof("ERRORGETINGEVENT: %v, %s", err, event.BlockKey())
	if create {
	
	  if IsErrorNotFound(err) {
		err = createEvent(event, tx, wb)
		if err != nil {
			return err
		}
		if tx == nil && wbParam == nil {
			
			return wb.Flush()
		}
	  } 
	}  else {
		if IsErrorNotFound(err) {
			return fmt.Errorf("event does not exist")
		  } 
	}
	if event.GetDataModelType() == entities.MessageModel {
		message := event.Payload.Data.(entities.Message)
		vecKey :=  datastore.NewKey(event.VectorKey(message.Topic))
		// var vec []byte
		// v, found := topicCounterCache.Get(vecKey.String())
		c := 0
		errGet := stores.EventStore.DB.Update(func(txn *badger.Txn) error {
			item , err := txn.Get(vecKey.Bytes())
			if err != nil {
				return err
			}
			item.Value(func(val []byte) error {
				c, err = strconv.Atoi(string(val))
				if err == nil {
					if c < int(event.Index) {
						txn.Set(vecKey.Bytes(), []byte(fmt.Sprint(event.Index)))
					}
				}
				return err
			})
			return nil
		})

		if errGet != nil && errGet != badger.ErrKeyNotFound{
			return errGet
		}


		// if !found {
		// 	if txn != nil {
		// 		vec , err = txn.Get(context.Background(), vecKey)
		// 	} else {
		// 		vec , err = stores.EventStore.Get(context.Background(), vecKey)
		// 	}
		
		// 	if err != nil {
		// 		if !IsErrorNotFound(err) {
		// 			return err
		// 		}
		// 	} else {
		// 		c, err = strconv.Atoi(string(vec))
		// 		if err != nil {
		// 			return err
		// 		}
		// 	}
		// }

		
	
		
		// if c < int(event.Index) {
		// 	if txn != nil {
		// 	txn.Put(context.Background(), vecKey, []byte(fmt.Sprint(event.Index)))
		// 	} else {
		// 		wb.Set(vecKey.Bytes(), []byte(fmt.Sprint(event.Index)) )
		// 	}
		// }
	}
	eventBytes := event.MsgPack()
	if txn == nil {
		err = wb.Set([]byte(event.DataKey()), eventBytes)
	} else {
		err = txn.Put(context.Background(), datastore.NewKey(event.DataKey()), eventBytes)
	}
	if err != nil {
		logger.Infof("PUTEVENT %v", err)
		return err
	}

	
	if event.Synced != nil && *event.Synced {
		blockKey := datastore.NewKey(event.BlockKey())
		if txn == nil {
			err = wb.Set(blockKey.Bytes(), []byte{})
		} else {
			err =  txn.Put(context.Background(), blockKey, []byte{})
		}
		
		// _, getError := txn.Get(context.Background(), blockKey)
		// if getError != nil {
		// 	if IsErrorNotFound(getError) {
		// 		// logger.Infof("CREATING: %s", blockKey.String())
		// 		err := txn.Put(context.Background(), blockKey, []byte{})
		// 		if err != nil {
		// 			return err
		// 		}
		// 	} else {
		// 		logger.Infof("GetPUTEVENT %v", getError)
		// 		return getError
		// 	}
		// }
		// if *event.Synced {
		// 	txn.Delete(context.Background(), datastore.NewKey(strings.Replace(event.BlockKey(), "/1/", "/0/", 1)))
		// }
	
		if tx == nil && wbParam == nil {
			logger.Infof("Flushing write batch")
			return wb.Flush()
			// if err = txn.Commit(context.Background()); err != nil {
			// 	return (err)
			// }
		}
	}
	logger.Debugf("UpdatedEventSuccessfully: %s, error: %s", event.ID, event.Error)
	if  event.IsValid != nil {
		logger.Debugf("UpdateEventIsValid: %s, %v", event.ID, *event.IsValid)
	}
	if  event.Synced != nil {
		logger.Debugf("UpdateEventSynced: %s, %v", event.ID, *event.Synced)
	}
	return err
}

func GetEventFromPath(ePath *entities.EventPath) (*entities.Event, error) {

	if ePath == nil || len(ePath.ID) == 0 {
		return nil, nil
	}
	event, err := GetEventById(ePath.ID, ePath.Model)

	if err != nil {
		return nil, err
	}
	return event, nil
}
func  IncrementCounterByKey(key string, delta uint64, tx *datastore.Txn) error {
	ds := stores.NetworkStatsStore
	txn, err := InitTx(ds, tx)
	if err != nil {
		return err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	count := new(big.Int).SetUint64(delta)
	if value, err := txn.Get(context.Background(), datastore.NewKey(key)); err != nil {
		if !IsErrorNotFound(err) {
			return err
		}
	} else {
		count = new(big.Int).Add(new(big.Int).SetBytes(value), new(big.Int).SetUint64(delta))
	}
	// logger.Infof("IncrementingCounterForApplication: %s, %s", key,  new(big.Int).SetBytes(count.Bytes()))
	err = txn.Put(context.Background(), datastore.NewKey( key), count.Bytes())
	if err != nil {
		return err
	}
	if tx == nil {
		return txn.Commit(context.Background())
	}
	return nil
}


func IncrementCounters(cycle uint64, validator entities.PublicKeyString, app string, tx *datastore.Txn) (err error) {
	ds := stores.NetworkStatsStore
	txn, err := InitTx(ds, tx)
	if tx == nil {
		defer txn.Discard(context.Background())
	}
	var count *big.Int
 	delta := int64(1);
	keys := []datastore.Key{
		// datastore.NewKey(entities.CycleCounterKey(cycle, nil, nil, nil)),
		datastore.NewKey(entities.NetworkCounterKey(nil)),
	}
	
	keys = append(keys, datastore.NewKey(entities.CycleCounterKey(cycle, &validator, utils.FalsePtr(), nil)))
	if len(app) > 0 {
		// keys = append(keys, datastore.NewKey(entities.CycleCounterKey(cycle, &validator, utils.FalsePtr(), &app)))
		keys = append(keys, datastore.NewKey(entities.NetworkCounterKey(&app))) // appcount in its lifetime
	} else {
		keys = append(keys, datastore.NewKey(entities.CycleApplicationKey(cycle, app))) //
		
	}
	for _, key := range keys {

		if value, err := txn.Get(context.Background(), key); err != nil {
			if !IsErrorNotFound(err) {
				return err
			}
			count = big.NewInt(int64(delta))
		} else {
			count = new(big.Int).Add(new(big.Int).SetBytes(value), big.NewInt(delta))
		}
		// logger.Infof("IncrementingCounterForApplication: %s, %s", key,  new(big.Int).SetBytes(count.Bytes()))
		err = txn.Put(context.Background(), key, count.Bytes())
		if err != nil {
			return err
		}
	}
	if tx == nil {
		return txn.Commit(context.Background())
	}
	return nil
}

func GetCycleCounts(cycle uint64, validator entities.PublicKeyString, claimed *bool, app *string, limit *entities.QueryLimit) ([]models.EventCounter, error) {
	rsl, err := stores.NetworkStatsStore.Query(context.Background(), query.Query{
		Prefix: entities.CycleCounterKey(cycle, &validator, claimed, app),
		Limit:  limit.Limit,
		Offset: limit.Offset,
	})
	if err != nil {
		if !IsErrorNotFound(err) {
			return nil, err
		}
		return []models.EventCounter{}, nil
	}
	counts := []models.EventCounter{}
	for {
		entry, ok := <-rsl.Next()
		if !ok {
			break
		}
		parts := strings.Split(entry.Key, "/")
		cy := cycle
		count := new(big.Int).SetBytes(entry.Value).Uint64()
		counts = append(counts, models.EventCounter{
			Cycle:     &cy,
			Validator: validator,
			Application:    parts[4],
			Count:     &count,
		})
	}
	return counts, err
}

func GetNetworkCounts(app *string, limit *entities.QueryLimit) ([]models.EventCounter, error) {
	counts := []models.EventCounter{}
	if app == nil {
		rsl, err := stores.NetworkStatsStore.Get(context.Background(), datastore.NewKey(entities.NetworkCounterKey(app)))
		if err != nil && !IsErrorNotFound(err) {
			return counts, err
		}
		if len(rsl) > 0 {
			count := new(big.Int).SetBytes(rsl).Uint64()
			counts = append(counts, models.EventCounter{
				Count:  &count,
			})
		}
		return counts, nil
	}
	rsl, err := stores.EventStore.Query(context.Background(), query.Query{
		Prefix: entities.NetworkCounterKey(app),
		Limit:  limit.Limit,
		Offset: limit.Offset,
	})
	if err != nil {
		if !IsErrorNotFound(err) {
			return nil, err
		}
		return counts, nil
	}
	
	for {
		entry, ok := <-rsl.Next()
		if !ok {
			break
		}
		parts := strings.Split(entry.Key, "/")
		var subn string = ""
		if len(parts) == 2 {
			subn = parts[1]
		}
		count := new(big.Int).SetBytes(entry.Value).Uint64()
		counts = append(counts, models.EventCounter{
			Application: subn,
			Count:  &count,
		})
	}
	return counts, err
}

func GetDependentEvents(event *entities.Event) (*[]entities.Event, error) {

	data := []entities.Event{}
	// err := db.SqlDb.Where(
	// 	&models.AuthorizationEvent{Event: entities.Event{PreviousEvent: *entities.NewEventPath(entities.AuthModel, event.Hash)}},
	// ).Or(&models.AuthorizationEvent{Event: entities.Event{AuthEvent: *entities.NewEventPath(entities.AuthModel, event.Hash)}},
	// // ).Or("? LIKE ANY (associations)", fmt.Sprintf("%%%s%%", event.Hash)
	// ).Find(&data).Error
	// if err != nil {
	// 	return nil, err
	// }
	prevEvent, _ := GetEventFromPath(&(event.PreviousEvent))
	if prevEvent != nil &&  (prevEvent.Synced == nil ||  !*prevEvent.Synced ) {
		data = append(data, *prevEvent)
	}
	authEvent, _ := GetEventFromPath(&(event.AuthEvent))
	if authEvent != nil &&  (authEvent.Synced == nil ||  !*authEvent.Synced ) {
		data = append(data, *authEvent)
	}
	return &data, nil
}

func GetStateBytesFromEventPath(path *entities.EventPath) ([]byte, error) {
	dsStores := stores.StateStore
	if path.Model == entities.MessageModel {
		dsStores = stores.MessageStore
	}
	stateKey := fmt.Sprintf(entities.DataKey, path.Model, path.ID)
	logger.Infof("STATEKEEYYY %s", stateKey)
	return dsStores.Get(context.Background(), datastore.NewKey(stateKey))

}

func GetStateFromEventPath(path *entities.EventPath) (any, error) {
	b, err := GetStateBytesFromEventPath(path)
	if err != nil {
		return nil, err
	}
	state := entities.GetStateModelFromEntityType(path.Model)
	err = encoder.MsgPackUnpackStruct(b, &state)
	return state, err
}
