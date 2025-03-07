package entities

import (
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/spaolacci/murmur3"
)
var eventModelsAsByte = [][]byte{}

var validSubscriptionFilter =  map[string]bool{
	"id":   true,
    "acct": true,
    "sub":  true,
    "top":  true,
    "rol":  true,
    "app": true,
    "s":    true,
    "cy":   true,
    "blk":  true,
    "ep":   true,
    "r":    true,}

func init() {
	for _, t := range EntityModels {
		eventModelsAsByte = append(eventModelsAsByte, []byte(t))
	} 
}



type SocketSubscriptionId struct {
	Conn *websocket.Conn
	Id string
}
type WsClientLog struct {
	counter map[*websocket.Conn]map[uint64]int
	Clients map[uint64][]*SocketSubscriptionId
	mutex  *sync.Mutex
	cfg *configs.MainConfiguration
}

func NewWsClientLog() WsClientLog {
	return WsClientLog{Clients:map[uint64][]*SocketSubscriptionId{}, mutex: &sync.Mutex{}, counter: make(map[*websocket.Conn]map[uint64]int) }
}

// register a client and return the topics the client is interested in
func (c *WsClientLog) RegisterClient(subscription *ClientWsSubscription) (topics [] string) {
	
	keys := []uint64{}
	for app, val := range subscription.Filter {
		// keys = append(keys, key)
		for _, _type := range val {
			if _type == "*" {
				for _, t := range eventModelsAsByte {
					logger.Debugf("REGISTERING: %s", app, string(t) )
					keys = append(keys, murmur3.Sum64(append(utils.UuidToBytes(app), t...)))
				}
				
			} else {
				logger.Debugf("REGISTERING: %s", app, string(_type) )
				keys = append(keys, murmur3.Sum64(append(utils.UuidToBytes(app), []byte(_type)...)))
				if len(_type) > 10 {
					topics = append(topics, _type)
				}
			}
		}

	}
	// process channel
	c.mutex.Lock()
	for _, key := range keys {
		// wsClients[key] = make(map[*websocket.Conn]*ClientWsSubscription)
		// wsClients[key][subscription.Conn] = subscription
		// for _, val2 := range subscription.Filter[key] {
		// 	wsClients[val2][subscription.Conn] = subscription
		// }
		if c.counter[subscription.Conn] == nil {
			c.counter[subscription.Conn] = make(map[uint64]int)
		}
		if c.Clients[key] == nil {
			c.Clients[key] = []*SocketSubscriptionId{}
		}
		
		c.counter[subscription.Conn][key] = len(c.Clients[key])
		
		c.Clients[key] = append(c.Clients[key], &SocketSubscriptionId{Conn: subscription.Conn, Id: subscription.Id})
		// wsClients[key][subscription.Conn] = subscription.Account
	}
	c.mutex.Unlock()
	
	return topics
}

func (c *WsClientLog) RegisterClientV2(subscription *ClientWsSubscriptionV2) (topics [] string) {
	
	keys := []uint64{}
	for _type, _filter := range subscription.Filter {
		// keys = append(keys, key)
		
		if !slices.Contains(EntityModels, EntityModel(_type)) {
			continue
		}
		
		subKey :=[]byte(_type)
		
		for v := range validSubscriptionFilter {
			if  len(_filter[v]) > 0 {
				keys = append(keys, murmur3.Sum64(append(subKey, append([]byte(v), []byte(fmt.Sprint(_filter[v]))...)...)))
				if v == "top" && (_type == string(MessageModel) || _type == string(SubscriptionModel))  || v == "id" && _type == string(TopicModel)   {
					topics = append(topics, _filter[v])
				}
				
			}
		}
		
		// if len(_filter["acct"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["acct"])...)
		// }
		// if len(_filter["sub"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["sub"])...)
		// }
		// if len(_filter["top"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["top"])...)
		// }
		// if len(_filter["rol"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["rol"])...)
		// }
		// if len(_filter["app"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["app"])...)
		// }
		// // if len(_filter["ref"]) > 0 {
		// // 	subKey = append(subKey, []byte(_filter["ref"])...)
		// // }

		// keys = append(keys, murmur3.Sum64(subKey))
		// for key, val  := range _filter {
		// 	subKey = append(subKey, []byte(key)...)
		// 	subKey = append(subKey, []byte(val)...)
		// //	for filterKey, filterValue := range _type {
		// 		// if _type == "*" {
		// 			for _, t := range eventModelsAsByte {
		// 				logger.Debugf("REGISTERING: %s", app, string(t) )
						
		// 				// for filterKey, filterValue := range _type {
		// 				// 	if filterKey == "acct" {
		// 				// 		subKey = append(subKey, []byte(filterValue))
		// 				// 	}
		// 				// 	if filterKey == "acct" {
		// 				// 		subKey = append(subKey, []byte(filterValue))
		// 				// 	}
		// 				// }
		// 				if len(_type["acct"]) > 0 {
		// 					subKey = append(subKey, []byte(_type["acct"])...)
		// 				}
		// 				if len(_type["sub"]) > 0 {
		// 					subKey = append(subKey, []byte(_type["sub"])...)
		// 				}
		// 				if len(_type["top"]) > 0 {
		// 					subKey = append(subKey, []byte(_type["top"])...)
		// 				}
		// 				if len(_type["rol"]) > 0 {
		// 					subKey = append(subKey, []byte(_type["rol"])...)
		// 				}
		// 				keys = append(keys, murmur3.Sum64(append(utils.UuidToBytes(app), t...)))
		// 			}		
		// 		// } else {
		// 		// 	logger.Debugf("REGISTERING: %s", app, string(_type) )
		// 		// 	keys = append(keys, murmur3.Sum64(append(utils.UuidToBytes(app), []byte(_type)...)))
		// 		// 	if len(_type) > 10 {
		// 		// 		topics = append(topics, _type)
		// 		// 	}
		// 		// }
		// //	}
		// }

	}
	// process channel
	c.mutex.Lock()
	for _, key := range keys {
		// wsClients[key] = make(map[*websocket.Conn]*ClientWsSubscription)
		// wsClients[key][subscription.Conn] = subscription
		// for _, val2 := range subscription.Filter[key] {
		// 	wsClients[val2][subscription.Conn] = subscription
		// }
		if c.counter[subscription.Conn] == nil {
			c.counter[subscription.Conn] = make(map[uint64]int)
		}
		if c.Clients[key] == nil {
			c.Clients[key] = []*SocketSubscriptionId{}
		}
		
		c.counter[subscription.Conn][key] = len(c.Clients[key])
		
		c.Clients[key] = append(c.Clients[key], &SocketSubscriptionId{Conn: subscription.Conn, Id: subscription.Id})
		// wsClients[key][subscription.Conn] = subscription.Account
	}
	c.mutex.Unlock()
	
	return topics
}

func (c *WsClientLog) RemoveClient(conn *websocket.Conn) {
	c.mutex.Lock()
	for key, v := range c.counter[conn] {
		if c.Clients[key] == nil || len(c.Clients[key]) == 0 {
			continue
		}
		n := len(c.Clients[key])
		c.Clients[key][v] = c.Clients[key][n-1]
		c.Clients[key] = c.Clients[key][:n-1]
	}
	c.mutex.Unlock()
}

func (c *WsClientLog) GetClients(appId string, entittyModel string ) []*SocketSubscriptionId {
	
	hash := murmur3.Sum64(append(utils.UuidToBytes(appId), []byte(entittyModel)...))
	if c.Clients[hash] == nil {
		return []*SocketSubscriptionId{}
	}
	return c.Clients[hash]
}

func (c *WsClientLog) GetClientsV2( entittyModel string , entity interface{}) (clients []*SocketSubscriptionId) {
	temp := map[string]interface{}{}
	d, _ := json.Marshal(entity)
	json.Unmarshal(d, &temp)
	keys := getSubscriptionKeys(entittyModel, temp)
	for _, k := range keys {
		clients = append(clients, c.Clients[k]...)
	}
	// hash := murmur3.Sum64(append([]byte(entityModel)...))
	// if c.Clients[hash] == nil {
	// 	return []*SocketSubscriptionId{}
	// }
	return clients
}

func getSubscriptionKeys( entittyModel string , entity map[string]interface{}) []uint64 {
	// rootKey :=[]byte(modelType)
	// subKey := []byte(modelType)
	keys := []uint64{}
	
	prefix := []byte(entittyModel)
	for k, v := range entity {
		if k =="_v" {
			continue
		}
		if validSubscriptionFilter[k] && len(fmt.Sprint(v)) > 0 {
			keys = append(keys, 
				murmur3.Sum64(append(prefix, append([]byte(k), 
				[]byte(fmt.Sprint(v))...)...)))
		}
	}
		// if len(entity["acct"]) > 0 {
		// 	subKey = append(subKey, []byte(entity["acct"])...)
		// 	rootKey = subKey
		// 	keys = append(keys, subKey)
		// }
		// if len(entity["sub"]) > 0 {
		// 	keys = append(keys, subKey)
		// 	subKey = append(subKey, []byte(entity["sub"])...)
		// 	keys = append(keys, append(rootKey, []byte(entity["sub"])...))
		// 	keys = append(keys, subKey)
		// }
		// if len(entity["top"]) > 0 {
		// 	subKey = append(subKey, []byte(entity["top"])...)
		// 	keys = append(keys, subKey)
		// 	keys = append(keys, append(rootKey, []byte(entity["top"])...))
		// }
		// if len(entity["rol"]) > 0 {
		// 	subKey = append(subKey, []byte(entity["rol"])...)
		// }
		// if len(entity["app"]) > 0 {
		// 	subKey = append(subKey, []byte(entity["app"])...)
		// }
		// if len(_filter["ref"]) > 0 {
		// 	subKey = append(subKey, []byte(_filter["ref"])...)
		// }

		// keys = append(keys, murmur3.Sum64(subKey))
		return keys
}