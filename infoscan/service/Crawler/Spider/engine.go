package Spider

import (
	"GScan/infoscan/dao"
	"GScan/infoscan/service/Crawler/Processor"
	"GScan/pkg/logger"
	"errors"
	"gorm.io/gorm"
	"net/url"
	"strings"
)

func (s *Spider) Processor(page *dao.Page, body []byte) {
	page.Title = Processor.Gettitle(body)
	parse, err := url.Parse(page.URL)
	if err != nil {
		logger.PF(logger.LERROR, "<Spider>[%s]PageID:%d URL错误,%s", page.ID, s.Host, err.Error())
		return
	}
	if parse.Host != s.Host {
		page.External = true
		s.PageCacheDB.Update(page.URL, page)
		//s.DAO.UpdatePage(page) // TODO: fix 1  保存结果，是否可以暂时保存到内存中？
		s.CallbackFunc(page, body)
		return
	} else {
		page.External = false
	}
	if page.Status != "Success" {
		if (!page.External) && strings.Contains(strings.ToLower(page.Error), "timeout") { //内链页面重试
			if (page.ErrorNum - 1) < s.config.Retry {
				s.AddUrlbypage([]*dao.Page{page})
			}
		}
		if !strings.HasPrefix(page.Error, "not text") {
			logger.PF(logger.LWARN, "<Spider>[%s]%s访问出错(%d),%s", s.Host, page.URL, page.ErrorNum, page.Error)
		}
		//s.DAO.UpdatePage(page) //TODO: fix 2 保存结果，是否可以保存到内存中？
		s.PageCacheDB.Update(page.URL, page)
		return
	}
	urls := Processor.Findurl(body, page.URL)
	if len(urls[0]) > 0 || len(urls[1]) > 0 {
		logger.PF(logger.LINFO, "<Spider>[%s]%s发现内链%d个，外链%d个", s.Host, page.URL, len(urls[0]), len(urls[1]))
	}
	for _, u := range urls[1] {
		page.ExtURLList = append(page.ExtURLList, u.String())
	}
	//s.DAO.UpdatePage(page) //TODO: fix 3
	s.PageCacheDB.Update(page.URL, page)
	npages := s.AddUrlbyURL(append(urls[0], urls[1]...))
	var pids []uint
	for _, p := range npages {
		pids = append(pids, p.ID)
	}
	//s.DAO.WebTreeAdd(s.JobID, page.ID, pids) // TODO:fix 4 全程在内存中操作，完成之后保存到数据库
	s.WebTreeCacheDB.Add(s.JobID, page.ID, pids)
	s.AddUrlbypage(npages)
}

func (s *Spider) AddUrlbypage(URL []*dao.Page) {
	for _, v := range URL {
		s.scheduler.Submit(v)
	}
}

func (s *Spider) AddUrlbyURL(URL []*url.URL) []*dao.Page {
	var pages []*dao.Page
	pages, err := s.AddNewPage(URL)
	if err != nil {
		//todo
	}
	logger.PF(logger.LDEBUG, "<Spider>[%s]添加新URL %d 个", s.Host, len(pages))
	return pages
}

func (s *Spider) AddNewPage(urls []*url.URL) ([]*dao.Page, error) {
	//todo 完善异常处理
	var pgs []*dao.Page
	for _, surl := range urls {
		urlstr := surl.String()
		if len(urlstr) <= len(surl.Scheme)+3 {
			logger.PF(logger.LERROR, "<Spider>发现异常连接%s", urlstr)
			continue
		}
		strurl := surl.String()[len(surl.Scheme)+3:]
		if ok := s.BloomFilter.TestString(strurl); !ok {
			s.BloomFilter.AddString(strurl)
			pg := dao.PagePool.Get().(*dao.Page)
			pg.JobID = s.JobID
			pg.Status = "未访问"
			pg.Model = gorm.Model{}
			pg.ID = 0
			pg.URL = surl.String()
			pg.Title = ""
			pg.Error = ""
			pg.ErrorNum = 0
			pg.Code = 0
			pg.Type = ""
			pg.Length = -1
			pgs = append(pgs, pg)
		}
	}
	if len(pgs) > 0 {
		s.DAO.InsertPages(pgs)
		s.PageCacheDB.Add(pgs)
		return pgs, nil
	} else {
		return nil, errors.New("没有新页面")
	}
}
