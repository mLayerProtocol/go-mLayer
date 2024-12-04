package query

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-datastore"
	"github.com/mlayerprotocol/go-mlayer/common/constants"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/chain"
	"github.com/mlayerprotocol/go-mlayer/internal/ds/stores"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
)

type NewStateParam struct {
	OldIDKey  string
	DataKey   string
	EventHash string
	Data      []byte
	RefKey    *string
	OldRefKey *string
}
type CreateStateParam struct {
	ID           string
	ModelType    entities.EntityModel
	IDKey        string
	DataKey      string
	RefKey       *string
	Keys         []string
	Data         []byte
	EventHash    string
	RestKeyValue []byte
}
type QueryLimit struct {
	Limit  int
	Offset int
}

func InitTx(ds *ds.Datastore, tx *datastore.Txn) (datastore.Txn, error) {
	if tx != nil {
		return *tx, nil
	}
	return ds.NewTransaction(context.Background(), false)
}

var DefaultQueryLimit = &QueryLimit{Limit: 100, Offset: 0}

func EntityKey(model entities.EntityModel, id string) string {
	return fmt.Sprintf("%s/id/%s", model, id)
}

func EntityDataKey(model entities.EntityModel, eventHash string) string {
	return fmt.Sprintf("data/%s/%s", model, eventHash)
}

func GetStateById(did string, modelType entities.EntityModel) ([]byte, error) {
	var stateData []byte
	err := stores.StateStore.DB.View(func(txn *badger.Txn) error {
		key := EntityKey(modelType, did)
		logger.Infof("IDKEY: %s", key)
		item, err := txn.Get(datastore.NewKey(key).Bytes())
		if err != nil {
			logger.Infof("GettateByIdError: %s, %s, %v", modelType, did, err)
			return err
		}
		return item.Value(func(val []byte) error {

			if val == nil {
				return datastore.ErrNotFound
			}
			logger.Infof("IDKEYVALUE: %s", string(val))
			value, err := txn.Get(datastore.NewKey(EntityDataKey(modelType, string(val))).Bytes())
			if err != nil {
				return err
			}
			return value.Value(func(val []byte) error {
				if val == nil {
					return datastore.ErrNotFound
				}
				stateData = val
				return nil
			})
		})
	})
	return stateData, err
}
func SaveHistoricState(model entities.EntityModel, id string, state []byte) error {
	ds := stores.StateStore
	if model == entities.MessageModel {
		ds = stores.MessageStore
	}
	return ds.Put(context.Background(), datastore.NewKey(fmt.Sprintf("data/%s/%s", model, id)), state)
}
func CreateState(newState CreateStateParam, tx *datastore.Txn) (err error) {
	ds := stores.StateStore

	if newState.ID == "" {
		return fmt.Errorf("CreateStateParam ID is required")
	}

	stateBytes := newState.Data
	keys := newState.Keys
	txn, err := InitTx(ds, tx)
	if err != nil {
		return err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}

	// return txn.Set(key.Bytes(), value)
	for _, key := range keys {
		logger.Infof("NewStateKey: %v, %v", key, newState.EventHash)
		if strings.EqualFold(key, newState.DataKey) {
			if err := txn.Put(context.Background(), datastore.NewKey(key), stateBytes); err != nil {
				return err
			}
			continue
		}
		if strings.EqualFold(key, newState.IDKey) {
			if err := txn.Put(context.Background(), datastore.NewKey(key), []byte(newState.EventHash)); err != nil {
				return err
			}
			continue
		}

		if newState.RefKey != nil && *newState.RefKey != "" && strings.EqualFold(key, *newState.RefKey) {
			val, err := txn.Get(context.Background(), datastore.NewKey(*newState.RefKey))
			if err != nil && !IsErrorNotFound(err) {
				return err
			}
			if len(val) > 0 {
				logger.Debugf("VALLLL: %s", string(val))
				return fmt.Errorf("\"%s\" ref already exists", newState.ModelType)
			}
			if err := txn.Put(context.Background(), datastore.NewKey(*newState.RefKey), []byte(newState.ID)); err != nil {
				logger.Errorf("error updateing subnet ref: %v", err)
				return err
			}
			continue
		}
		if newState.RestKeyValue == nil {
			newState.RestKeyValue = []byte(newState.ID)
		}
		if err := txn.Put(context.Background(), datastore.NewKey(key), newState.RestKeyValue); err != nil {
			return err
		}
	}
	if tx == nil {
		err = txn.Commit(context.Background())
	}

	if err != nil {
		return err
	}
	return nil
}

func UpdateState(id string, newState NewStateParam, tx *datastore.Txn) error {
	ds := stores.StateStore
	txn, err := InitTx(ds, tx)
	if err != nil {
		return err
	}
	if tx == nil {
		defer txn.Discard(context.Background())
	}

	// state.ID, err = entities.GetId(state)
	// if err != nil {
	// 	return  nil, err
	// }
	stateBytes := newState.Data

	// oldState, err := GetSubnetStateById(id)

	if err != nil {
		return err
	}

	if err := txn.Put(context.Background(), datastore.NewKey(newState.DataKey), stateBytes); err != nil {
		logger.Errorf("error updateing state key: %v", err)
		return err
	}

	if err := txn.Put(context.Background(), datastore.NewKey(newState.OldIDKey), []byte(newState.EventHash)); err != nil {
		logger.Errorf("error updateing state key: %v", err)
		return err
	}

	if newState.RefKey != nil {
		err := txn.Delete(context.Background(), datastore.NewKey(*newState.OldRefKey))
		if err != nil {
			return err
		}
		if err := txn.Put(context.Background(), datastore.NewKey(*newState.RefKey), []byte(id)); err != nil {
			logger.Errorf("error updateing subnet ref: %v", err)
			return err
		}
	}

	if tx == nil {
		if err := txn.Commit(context.Background()); err != nil {
			return err
		}
	}

	logger.Infof("SubnetKey: %s", newState.EventHash)

	return nil
}

func RefExists(entityType entities.EntityModel, ref string, subnet string) (bool, error) {
	ds := stores.StateStore
	refKey := fmt.Sprintf("%s|ref|%s", entityType, ref)
	if subnet != "" {
		refKey = fmt.Sprintf("%s|ref|%s|%s", entityType, subnet, ref)
	}
	value, err := ds.Get(context.Background(), datastore.NewKey(refKey))
	if err != nil {
		if err == datastore.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if value == nil {
		return false, nil
	}
	return len(value) > 0, err
}

func GetStateFromEntityPath(ePath *entities.EntityPath) ([]byte, error) {
	if ePath == nil || len(ePath.Hash) == 0 {
		return nil, nil
	}
	return GetStateById(ePath.Hash, ePath.Model)

}
func GetNumAccounts() (uint64, error) {
	countKey := datastore.NewKey("numAcc")
	if countByte, err := stores.SystemStore.Get(context.Background(), countKey); err != nil {
		if !IsErrorNotFound(err) {
			return 0, err
		}
		return 0, nil
	} else {
		return new(big.Int).SetBytes(countByte).Uint64(), nil
	}
}
func UpdateAccountCounter(account string) (uint64, error) {
	txn, err := InitTx(stores.SystemStore, nil)
	if err != nil {
		return 0, err
	}
	defer txn.Discard(context.Background())
	countKey := datastore.NewKey("numAcc")
	count := uint64(0)
	if countByte, err := txn.Get(context.Background(), countKey); err != nil {
		if !IsErrorNotFound(err) {
			return 0, err
		}
	} else {
		count = new(big.Int).SetBytes(countByte).Uint64()
	}

	accountKey := datastore.NewKey(fmt.Sprintf("acc/%s", account))
	_, err = stores.NetworkStatsStore.Get(context.Background(), accountKey)
	if IsErrorNotFound(err) {
		err := txn.Put(context.Background(), countKey, new(big.Int).SetUint64(count+1).Bytes())
		if err != nil {
			return count, err
		}
		err = stores.NetworkStatsStore.Put(context.Background(), accountKey, []byte{})
		if err != nil {
			return 0, err
		}
		err = txn.Commit(context.Background())
		if err != nil {
			return count, err
		}
		return count + 1, nil
	} else {
		return count, err
	}

}

func createArchiveDir(cfg *configs.MainConfiguration) (string, error) {
	dir := filepath.Join(cfg.DataDir, "archive")
	if strings.HasPrefix(cfg.DataDir, "./") && !strings.HasPrefix(dir, "./") {
		dir = "./" + dir
	}
	if strings.HasPrefix(cfg.DataDir, "../") && !strings.HasPrefix(dir, "../") {
		dir = "../" + dir
	}
	return dir, os.MkdirAll(dir, os.ModePerm)
}

type ExportData struct {
	Block []byte   `json:"bl"`
	Data  [][]byte `json:"d"`
}

func ArchiveEvents(cfg *configs.MainConfiguration) error {
	openFile := map[uint64]*os.File{}
	lastIndexKey := datastore.NewKey("arc/id")
	// for block := fromBlock; block <= toBlock; block++ {

	// events,  err := stores.EventStore.Query(context.Background(), query.Query{
	// 	Prefix:  where.BlockKey(),
	// })
	// if err != nil  {
	// 	if IsErrorNotFound(err) {
	// 		return []byte{}, nil
	// 	}
	// 	return nil, err
	// }
	defer func () {
		for _, v := range openFile {
			v.Close()
		}
	}()
	currentBlock := chain.NetworkInfo.CurrentBlock
	lastIndexed := "/bl"
	startBlock := 0
	if lastIndexedBytes, err := stores.EventStore.Get(context.Background(), lastIndexKey); err != nil {
		if !IsErrorNotFound(err) {
			return err
		}
		
	} else {
		logger.Infof("LASTINDEXEDKEY: %s", string(lastIndexedBytes))
		parts := strings.Split(string(lastIndexedBytes), "/")
		// lastIndexed = parts[len(parts)-2]
		lastIndexed = string(lastIndexedBytes)
		startBlock, err = strconv.Atoi(parts[2])
		if err != nil {
			return err
		}
	}
	
	logger.Debugf("Checking block archive: %d, %d", startBlock, currentBlock.Uint64())
	// var results [][]byte
	// var rsl map[uint64][][]byte
	_, eee := stores.EventStore.Get(context.Background(), datastore.NewKey("/bl/00000000000018507458/001732783208380/c8d5c3e9-7129-5d8a-5be5-eb5181785ecc"))
	logger.Infof("NOTFOUND: %v", IsErrorNotFound(eee))
	for i := uint64(startBlock); i < currentBlock.Uint64() - 5; i++ {
		
		
		keysFound := 0
		err := stores.EventStore.DB.View(func(txn *badger.Txn) error {

			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 5000
			opts.PrefetchValues = true // Fetch both keys and values
			it := txn.NewIterator(opts)
			defer it.Close()

			// Seek to a specific key
			
			// if lastIndexed != "" {
			// 	it.Seek([]byte(lastIndexed))
			// }

			// lastIndexed = (&entities.Event{BlockNumber: i}).BlockKey()
		
			where := entities.Event{BlockNumber: i}
		
			// where.Synced = utils.TruePtr()
			// Iterate from the target key onward
			
			bytesWritten := int64(0)
			
			lastKey := where.BlockKey()
			// if i >= 18508719 {
			// 	return nil
			//  }
			seekKey := lastIndexed
			if i > uint64(startBlock) {
				seekKey = where.BlockKey()
			}
			//  logger.Infof("LASTINDEX1  %d, %s, Where: %s", i, seekKey, where.BlockKey())
			for  it.Seek([]byte(seekKey)); it.ValidForPrefix([]byte(where.BlockKey())); it.Next() {
			// for it.Valid() {
				logger.Infof("LASTINDEX_2 %s, %s", lastIndexed, it.Item().Key())
					
		
				item := it.Item()
				// if len(item.Key()) > 0 {
				// 	continue
				// }
				k := item.Key()
				if string(k) == lastIndexed {
					logger.Infof("ALREADYARCHIVED: %s, %s", k, lastIndexed)
					continue
				}
				key := string(item.Key())
			
				err := item.Value(func(v []byte) error {
				
					id := fmt.Sprintf("id/%s", key[strings.LastIndex(key, "/")+1:])

					if len(id) > 0 {
						eventData, err := stores.EventStore.Get(context.Background(), datastore.NewKey(id))

						
						if err != nil {
							logger.Errorf("ArchiveErrorWhenGettingEventData: %e", err)
							return err
						}
						event, err := entities.UnpackEventGeneric(eventData)
						if err != nil {
							logger.Errorf("EVENTNOTFOUND %v", err)
							return err
						}
					
						// if rsl[event.Cycle] == nil {
						// 	rsl[event.Cycle]  = [][]byte{}
						// }
						// rsl[event.Cycle]  = append(rsl[event.Cycle], eventData)
						eventData = append(eventData, []byte{':', '|'}...)
						compressed, err := utils.CompressToGzip(eventData)
						keysFound ++
						if openFile[event.Cycle] == nil || bytesWritten >= constants.MAX_SYNC_FILE_SIZE {
							
							cycleDir := filepath.Join(cfg.ArchiveDir, fmt.Sprint(event.Cycle))
							
							if err := os.MkdirAll(cycleDir, os.ModePerm); err != nil {
								panic(err)
							}
							if err != nil {
								// Path does not exist
								return err
							}
							files, err := utils.ListFilesInDir(cycleDir)
							if err != nil {
								return err
							}
							fileName := ""
							if len(files) > 0 {
							
								fileName = files[len(files)-1]
								
								fileInfo, err := os.Stat(filepath.Join(cycleDir, fileName))
								
								if err != nil && !os.IsNotExist(err) {
									return err
								}
								if err == nil {
									
									bytesWritten = fileInfo.Size()
								}
								if os.IsNotExist(err) || fileInfo.Size() >= constants.MAX_SYNC_FILE_SIZE {
									// If file exists and exceeds the size limit, create a new file
									bytesWritten = 0
									fileName = fmt.Sprintf("%s.dat", utils.IntMilliToTimestampString(int64(time.Now().UnixMilli())))
									if err == nil && openFile[event.Cycle] != nil {
										openFile[event.Cycle].Close()
									}
								} else {
									bytesWritten = fileInfo.Size()
									
								}
							} else {
								bytesWritten = 0
								fileName = fmt.Sprintf("%s.dat", utils.IntMilliToTimestampString(int64(time.Now().UnixMilli())))
							}
							
			
							logger.Infof("THECYCLE: %d, %s", event.Cycle, lastIndexed)
							logger.Infof("COMPRESED: %d, file: %s", len(compressed), filepath.Join(cycleDir, fileName))
							openFile[event.Cycle], err = os.OpenFile(filepath.Join(cycleDir, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
							if err != nil {
								logger.Errorf("OpenArchiveFileError: %v", err)
								return err
							}
							
						}
						
						_, err = openFile[event.Cycle].Write(compressed)
						if err != nil {
							logger.Errorf("WriteArchiveFileError: %v", err)
							return err
						}
						bytesWritten += int64(len(compressed))

						// = append(results, eventData)
					}
					return nil
				})
				if err != nil {
					logger.Errorf("ERRORRRRRRR: %v", err)
					return err
				}
				lastIndexed = string(k)
				lastKey = string(k)
			}
			// logger.Debugf("Archived block: %d; KeysFound: %d, lastKey: %s, Progress %d%", i, keysFound, lastKey, i * 100 / (currentBlock.Uint64()-1) )
			err := stores.EventStore.Put(context.Background(), lastIndexKey, []byte(lastKey))
			if err != nil {
				return err
			}

			// if len(results) > 0 {
			// 	utils.CompressToGzip(results[])
			// }
			return nil
		})
		if err != nil {
			logger.Errorf("ERRORRRRRR: %v", err)
			return err
		}
		
	}

	return nil

	// for cy, v := range rsl {
	// 	cycleDir := filepath.Join( cfg.ArchiveDir, fmt.Sprint(cy))
	// 	if err := os.MkdirAll(cycleDir, os.ModePerm); err != nil {
	// 		panic(err)
	// 	}
	// if err := os.MkdirAll(cycleDir, os.ModePerm); err != nil {
	// 			panic(err)
	// 		}

	// 			files, err := utils.ListFilesInDir(cycleDir)
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			fileName := ""
	// 			if len(files) > 0 {
	// 				fileName = files[len(files)-1]
	// 				fileInfo, err := os.Stat(fileName)
	// 				if err != nil && !os.IsNotExist(err) {
	// 					return nil, err
	// 				}
	// 				if err == nil && fileInfo.Size() >= constants.MAX_SYNC_FILE_SIZE {
	// 					// If file exists and exceeds the size limit, create a new file
	// 					fileName = fmt.Sprintf("%s.dat", utils.IntMilliToTimestampString(int64(time.Now().UnixMilli())))
	// 				}
	// 			} else {
	// 				fileName = fmt.Sprintf("%s.dat", utils.IntMilliToTimestampString(int64(time.Now().UnixMilli())))
	// 			}

	// 			file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			defer file.Close()
	// 			for _, d := range v {
	// 				d = append(d, []byte(":|")...)
	// 				compressed, err := utils.CompressToGzip(d)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 			_, err = file.Write(compressed)
	// 			}
	// 			if err != nil {
	// 				return err
	// 			}
	// 			results = append(results, eventData)

	// }

	// for {
	// 	val, ok := <-events.Next()
	// 	if !ok {
	// 		fmt.Println("Channel closed")
	// 		break
	// 	}
	// 	id := fmt.Sprintf("id/%s", val.Key[strings.LastIndex(val.Key, "/")+1:])

	// 	if len(id) > 0 {
	// 		 event, err := stores.EventStore.Get(context.Background(), datastore.NewKey(id))

	// 		 logger.Infof("DATAKEY: %d", len(event))
	// 		 if err != nil  {
	// 			return nil, err
	// 		 }
	// 		 results = append(results, event)
	// 	}

	// }
	// data := ExportData{Data: results, Block: new(big.Int).SetUint64(where.BlockNumber).Bytes()}
	// s, err := encoder.MsgPackStruct(data)
	// if err != nil {
	// 	return nil, err
	// }
	// return s, nil
}
