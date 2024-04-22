package cache

import (
	"GScan/infoscan/dao"
	"GScan/pkg/logger"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

func (c *PagesCacheDB) Add(pages []*dao.Page) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	for _, page := range pages {
		if c.Urls.Contains(page.URL) {
			continue
		}
		c.Urls.Add(page.URL)
		err := c.Data.Add(page.URL, page, cache.DefaultExpiration)
		if err != nil {
			logger.PF(logger.LERROR, "<CacheProcessor>[%s]%s  :%s", "PagesCacheDB", "Add", err)
		}
	}
}

func (c *PagesCacheDB) Update(url string, page *dao.Page) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.Urls.Contains(url) {
		err := c.Data.Replace(url, page, cache.DefaultExpiration)
		if err != nil {
			logger.PF(logger.LERROR, "<CacheProcessor>[%s]%s  :%s", "PagesCacheDB", "update", err)
		}
	} else {
		c.Urls.Add(url)
		err := c.Data.Add(url, page, cache.DefaultExpiration)
		if err != nil {
			logger.PF(logger.LERROR, "<CacheProcessor>[%s]%s  :%s", "PagesCacheDB", "update", err)
		}
	}
}

func (c *PagesCacheDB) Remove(key string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Urls.Remove(key)
	c.Data.Delete(key)
}

/*
清理已经更新过的Page,将其生成「」dao.page 然后存入数据库
*/
func (c *PagesCacheDB) cleanup() (int, time.Duration) {
	count := 0
	var pages []dao.Page
	var pagelist []*dao.Page
	var keysToRemove []string
	c.Lock.RLock()
	for _, v := range c.Urls.Values() {
		page, found := c.Data.Get(v.(string))
		if (found && page.(*dao.Page).Status != "") || (found && page.(*dao.Page).Status != "未访问") {
			pages = append(pages, *page.(*dao.Page))
			keysToRemove = append(keysToRemove, v.(string))
			count++
		}
	}
	c.Lock.RUnlock()

	for _, v := range keysToRemove {
		c.Remove(v)
	}
	startTime := time.Now()

	for _, v := range pages {
		pagelist = append(pagelist, &v)
	}

	c.DAO.UpdatePage(pagelist)
	endTime := time.Now()
	return count, endTime.Sub(startTime)
}

func (c *PagesCacheDB) Dump() {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	var pages []*dao.Page
	for _, v := range c.Data.Items() {
		//c.DAO.UpdatePage(v.Object.(*dao.Page))
		pages = append(pages, v.Object.(*dao.Page))
	}
	c.DAO.UpdatePage(pages)
	logger.PF(logger.LINFO, "<CacheProcessor>[%s]%s  :%d", "PagesCacheDB", "Dump", len(c.Data.Items()))
}

func (w *WebTreeCacheDB) Add(jobid, pageid uint, pids []uint) {
	w.Lock.Lock()
	defer w.Lock.Unlock()
	for _, pid := range pids {
		if !w.PIDS.Contains(pid) {
			w.PIDS.Add(pid)
			w.Data.Add(strconv.Itoa(int(pid)), WebTreeData{JobID: jobid, FIDS: []uint{pageid}}, cache.DefaultExpiration)
		} else {
			data, found := w.Data.Get(strconv.Itoa(int(pid)))
			if found {
				w.Data.Replace(strconv.Itoa(int(pid)), WebTreeData{
					JobID: data.(WebTreeData).JobID,
					FIDS:  append(data.(WebTreeData).FIDS, pageid),
				}, cache.DefaultExpiration)
			} else {
				w.Data.Add(strconv.Itoa(int(pid)), WebTreeData{JobID: jobid, FIDS: []uint{pageid}}, cache.DefaultExpiration)
			}
		}
	}
}

func (w *WebTreeCacheDB) Dump() {
	w.Lock.RLock()
	defer w.Lock.RUnlock()
	logger.PF(logger.LINFO, "<CacheProcessor>[%s]%s  :%d", "WebTreeCacheDB", "Start Dump", len(w.PIDS.Values()))
	var trees []dao.WebTree
	for key, v := range w.Data.Items() {
		ukey, _ := strconv.ParseUint(key, 10, 64)
		//w.DAO.WebTreeAdd(v.Object.(WebTreeData).JobID, uint(ukey), v.Object.(WebTreeData).FIDS)
		trees = append(trees, dao.WebTree{
			JobID:  v.Object.(WebTreeData).JobID,
			PageID: uint(ukey),
			FiD:    v.Object.(WebTreeData).FIDS,
		})
	}
	w.DAO.WebTreeAdd(trees)
	logger.PF(logger.LINFO, "<CacheProcessor>[%s]%s  :%d", "WebTreeCacheDB", "Dump", len(w.PIDS.Values()))
}

func (c *PagesCacheDB) Clean() {
	for {
		select {
		case <-c.SignChan:
			return
		default:
			count, elapsedTime := c.cleanup()
			logger.PF(logger.LINFO, "<CacheProcessor>[%s]%s  :%d,用时:%s", "PagesCacheDB", "保存到数据库数量", count, elapsedTime)
			time.Sleep(20 * time.Second)
		}
	}
}

func (c *PagesCacheDB) PROCESSEND(notices <-chan string) {
	for {
		select {
		case <-notices:
			c.Dump()
			return
		default:
			continue
		}
	}
}
func (w *WebTreeCacheDB) PROCESSEND(notices <-chan string) {
	for {
		select {
		case <-notices:
			w.Dump()
			return
		default:
			continue
		}
	}
}
