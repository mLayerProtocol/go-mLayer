package p2p

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/go-datastore/query"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/ds"
)

var StateDhtSyncer DhtSyncer
type DhtPrefix string

const (
    SnetDhtPrefix  DhtPrefix = "app"
    SnetRefDhtPrefix  DhtPrefix = "snetRef"
    ValDhtPrefix  DhtPrefix = "val"
)
type DhtSyncer struct{
    store *ds.Datastore
    context *context.Context
}

// func init () {
//     StateDhtSyncer = NewDhtSyncer(stores.StateStore, context.Background())
// }


func NewDhtSyncer(store *ds.Datastore, ctx context.Context) DhtSyncer {
    dht := DhtSyncer{store, &ctx}
    // go dht.sync()
    return dht
}


func (v *DhtSyncer) AdaKey(dhtKeyPrefix DhtPrefix, dataKey string, dataValue []byte) error {
  return  v.store.Set(*v.context,  ds.Key(fmt.Sprintf("/dht/%s/%s", dhtKeyPrefix, dataKey)), dataValue, true)
}   

func (v *DhtSyncer) GetDhtValue(dhtKeyPrefix DhtPrefix, dataKey string) ([]byte, error) {
    dhtKey := fmt.Sprintf("ml/%s/%s", dhtKeyPrefix,  dataKey)
    data, err := idht.GetValue(*v.context, dhtKey, nil)
    if err == nil {
     return nil, err
    }
    return data, err
  } 

func (v *DhtSyncer) Sync() {
    for {
        if Initialized {
            notSynced,  err := v.store.Query(*v.context, query.Query{
                Prefix: "/dht",
            })
            if err != nil {
                continue
            }
           
            for result := range notSynced.Next() {
            parts := strings.Split(result.Entry.Key, "/")
            dhtKey := fmt.Sprintf("ml/%s/%s", parts[len(parts)-2],  parts[len(parts)-1])
            err := idht.PutValue(*v.context, dhtKey, result.Entry.Value, nil)
            if err == nil {
                v.store.Delete(*v.context, ds.Key(result.Entry.Key) )
            }
            }
        } else {
            logger.Infof("P2pNotInitialized...")
        }
        time.Sleep(1 * time.Second)
    }
}
