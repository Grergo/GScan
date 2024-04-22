package sqlite

import (
	"GScan/infoscan/dao"
	"GScan/infoscan/dao/base"
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func NewDB(dbfile string, logPath string) *base.DAO {
	logfile, _ := os.OpenFile(filepath.Join(logPath, fmt.Sprintf("%s_sql.log", time.Now().Format("2006-01-02 15-04-05"))), os.O_CREATE|os.O_RDWR, 0755)
	newLogger := logger.New(
		log.New(logfile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢 SQL 阈值
			LogLevel:                  logger.Error,           // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,                   // 禁用彩色打印
		},
	)
	db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{
		Logger:          newLogger,
		CreateBatchSize: 1000,
	})
	// 设置WAL模式
	db.Exec("PRAGMA journal_mode=WAL;")
	// 关闭写同步
	db.Exec("PRAGMA synchronous = OFF;")
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = db.AutoMigrate(&dao.Page{}, &dao.WebTree{}, dao.Job{}, dao.ProcessResult{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	return &base.DAO{Db: db, Mutex: sync.Mutex{}}
}
