package cache

import (
	"GScan/infoscan/dao"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/patrickmn/go-cache"
	"sync"
)

type PagesCacheDB struct {
	Urls     *hashset.Set
	Data     *cache.Cache
	DAO      dao.IDAO
	SignChan chan string
	Lock     *sync.RWMutex
}

type WebTreeCacheDB struct {
	PIDS *hashset.Set
	Data *cache.Cache
	DAO  dao.IDAO
	Lock *sync.RWMutex
}

type WebTreeData struct {
	JobID uint
	FIDS  []uint
}
