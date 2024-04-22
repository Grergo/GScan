package base

import (
	"GScan/infoscan/dao"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
)

type DAO struct {
	Db *gorm.DB
	sync.Mutex
}

func (D *DAO) InsertPages(page []*dao.Page) {
	//D.Mutex.Lock()
	D.Db.Create(page)
	//D.Mutex.Unlock()
}

func (D *DAO) SelectPagesByMap(kv map[string]interface{}) ([]dao.Page, error) {
	var pages []dao.Page
	if err := D.Db.Where(kv).Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}

func (D *DAO) UpdatePage(pages []*dao.Page) {
	//D.Mutex.Lock()
	//D.Db.Save(page)
	D.Db.Clauses(clause.OnConflict{UpdateAll: true, Columns: []clause.Column{{Name: "id"}}}).Create(pages)
	//D.Mutex.Unlock()

}

func (D *DAO) DeleteById(ID int64) {
	D.Db.Where("ID = ?", ID).Delete(&dao.Page{})
}

func (D *DAO) AddResult(result *dao.ProcessResult) {
	D.Mutex.Lock()
	D.Db.Create(result)
	D.Mutex.Unlock()
}
func (D *DAO) GetResult(jobid uint) []dao.ProcessResult {
	D.Mutex.Lock()
	var rr []dao.ProcessResult
	D.Db.Where(&dao.ProcessResult{
		JobID: jobid,
	}).Find(&rr)
	D.Mutex.Unlock()
	return rr
}

func (D *DAO) WebTreeAddBySubID(jobID uint, FPIDs []uint, subID uint) {
	D.Mutex.Lock()
	var rs dao.WebTree
	if errors.Is(D.Db.Where(&dao.WebTree{JobID: jobID, PageID: subID}).First(&rs).Error, gorm.ErrRecordNotFound) {
		rs.JobID = jobID
		rs.PageID = subID
		rs.FiD = FPIDs
		D.Db.Create(&rs)
	} else {
		rs.FiD = FPIDs
		D.Db.Save(&rs)
	}
	D.Mutex.Unlock()
}

/*
func (D *DAO) WebTreeAdd(jobID uint, pageID uint, fID []uint) {
	D.Mutex.Lock()
	defer D.Mutex.Unlock()
	var rs dao.WebTree
	if errors.Is(D.Db.Where(&dao.WebTree{JobID: jobID, PageID: pageID}).First(&rs).Error, gorm.ErrRecordNotFound) {
		rs.JobID = jobID
		rs.PageID = pageID
		rs.FiD = fID
		D.Db.Create(&rs)
	} else {
		rs.FiD = fID
		D.Db.Save(&rs)
	}
}
*/

func (D *DAO) WebTreeAdd(trees []dao.WebTree) {
	D.Db.Clauses(clause.OnConflict{UpdateAll: true, Columns: []clause.Column{{Name: "id"}}}).Create(trees)
}

func (D *DAO) WebTreeGet(jobID uint, id uint) ([]uint, error) {
	D.Mutex.Lock()
	defer D.Mutex.Unlock()
	var res dao.WebTree
	err := D.Db.Where(dao.WebTree{
		JobID:  jobID,
		PageID: id,
	}).First(&res).Error
	if err != nil {
		return nil, err
	}

	return res.FiD, nil
}
func (D *DAO) WebPageLink(jobID uint, id uint) [][]uint {
	var res [][]uint
	getf(D, jobID, id, res)
	return res
}

func (D *DAO) AddJob(name string) *dao.Job {
	D.Mutex.Lock()
	job := dao.Job{
		Name: name,
	}
	D.Db.Create(&job)
	D.Mutex.Unlock()
	return &job
}

func getf(s *DAO, jobID uint, ID uint, res [][]uint) {
	if v, err := s.WebTreeGet(jobID, ID); err == nil {
		for _, vs := range v {
			res = append(res, append(res[len(res)-1], vs))
			getf(s, jobID, vs, res)
		}
	}
}

func (D *DAO) GetOnePages(page *dao.Page) *dao.Page {
	D.Mutex.Lock()
	var rp *dao.Page
	D.Db.Where(page).First(&rp)
	D.Mutex.Unlock()
	return rp
}

func (D *DAO) Getjobs() []*dao.Job {
	D.Mutex.Lock()
	var jobs []*dao.Job
	D.Db.Find(&jobs)
	D.Mutex.Unlock()
	return jobs
}

func (D *DAO) WebTreeGetAll(jobID uint) ([]*dao.WebTree, error) {
	D.Mutex.Lock()
	defer D.Mutex.Unlock()
	var res []*dao.WebTree
	err := D.Db.Where(dao.WebTree{
		JobID: jobID,
	}).Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (D *DAO) GetAllPages(page *dao.Page) []*dao.Page {
	D.Mutex.Lock()
	var rp []*dao.Page
	D.Db.Select("ID", "URL").Where(page).Find(&rp)
	D.Mutex.Unlock()
	return rp
}
