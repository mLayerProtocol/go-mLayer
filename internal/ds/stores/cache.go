package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/mlayerprotocol/go-mlayer/pkg/core/cache"
)

var SystemCache *cache.ShardedCache
var NodesTopicCache *cache.ShardedCache


type SystemCacheKeyPrefix string


const (
	AccountConnectedKey SystemCacheKeyPrefix = "acct"
)

func (c SystemCacheKeyPrefix) NewKey(key string) string {
	return fmt.Sprintf("%s/%s", c, key)
}



func InitCaches(ctx *context.Context) (_caches []*cache.ShardedCache) {
	SystemCache = cache.NewShardedCache("system", cache.CacheOptions{NumShards: 16, TTL:  5 * time.Minute, CleanInterval: 10*time.Minute})
	NodesTopicCache = cache.NewShardedCache("node", cache.CacheOptions{NumShards: 1024, TTL:  5 * time.Minute, CleanInterval: 10*time.Minute})
	_caches = append(_caches, SystemCache)
	return _caches
}
