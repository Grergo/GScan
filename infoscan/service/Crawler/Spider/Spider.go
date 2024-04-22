package Spider

import (
	"GScan/infoscan/cache"
	"GScan/infoscan/config"
	"GScan/infoscan/dao"
	"GScan/infoscan/service/Crawler/Processor"
	"GScan/pkg"
	"GScan/pkg/bloom"
	"GScan/pkg/logger"
	"context"
	"github.com/emirpasic/gods/sets/hashset"
	gocache "github.com/patrickmn/go-cache"
	"net/url"
	"sync"
)

type Spider struct {
	MainURL        *url.URL
	Host           string
	JobID          uint
	Reqer          Requester
	DataProcessor  Processor.IProcessor
	BloomFilter    *bloom.Filter
	DAO            dao.IDAO
	scheduler      *pkg.QueueScheduler[*dao.Page]
	CallbackFunc   func(page *dao.Page, body []byte)
	config         *config.Spider
	PageCacheDB    *cache.PagesCacheDB
	WebTreeCacheDB *cache.WebTreeCacheDB
	NoticeChan     chan string
	//UP              *URLPoll
	//*Limiter
}

//type Limiter struct {
//}

func NewSpider(config *config.Spider, jobid uint, db dao.IDAO) *Spider {
	s := &Spider{
		JobID:     jobid,
		DAO:       db,
		scheduler: &pkg.QueueScheduler[*dao.Page]{},
		config:    config,
		PageCacheDB: &cache.PagesCacheDB{
			Urls:     hashset.New(),
			Data:     gocache.New(gocache.NoExpiration, gocache.NoExpiration),
			DAO:      db,
			Lock:     new(sync.RWMutex),
			SignChan: make(chan string),
		},
		WebTreeCacheDB: &cache.WebTreeCacheDB{
			PIDS: hashset.New(),
			Data: gocache.New(gocache.NoExpiration, gocache.NoExpiration),
			DAO:  db,
			Lock: new(sync.RWMutex),
		},
	}
	s.scheduler.Init()
	s.scheduler.Run()
	return s
}
func (s *Spider) Run(ctx context.Context, wg *sync.WaitGroup) {
	logger.PF(logger.LINFO, "<Spider>[%s]开始运行", s.Host)
	s.runCacheJob(s.NoticeChan)
	s.runWK(ctx, wg, s.config.Threads)
	s.dump()
	close(s.PageCacheDB.SignChan)
	logger.PF(logger.LINFO, "<Spider>[%s]结束", s.Host)
}
func (s *Spider) SetReqer(r Requester) *Spider {
	s.Reqer = r
	return s
}
func (s *Spider) SetMainUrl(murl *url.URL) *Spider {
	s.MainURL = murl
	s.Host = s.MainURL.Host
	pages := s.AddUrlbyURL([]*url.URL{murl})
	s.AddUrlbypage(pages)
	return s
}
func (s *Spider) SetCallbackFunc(f func(page *dao.Page, body []byte)) *Spider {
	s.CallbackFunc = f
	return s
}
func (s *Spider) SetProcessor(processor Processor.IProcessor) *Spider {
	s.DataProcessor = processor
	return s
}

func (s *Spider) SetFilter(Filter *bloom.Filter) *Spider {
	s.BloomFilter = Filter
	return s
}
func (s *Spider) SetNoticeChan(notices chan string) *Spider {
	s.NoticeChan = notices
	return s
}

func (s *Spider) runCacheJob(notices chan string) {
	go s.PageCacheDB.Clean()
	go s.PageCacheDB.PROCESSEND(notices)
	go s.WebTreeCacheDB.PROCESSEND(notices)
}

func (s *Spider) dump() {
	s.PageCacheDB.Dump()
	s.WebTreeCacheDB.Dump()
}
