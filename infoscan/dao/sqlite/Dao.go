package sqlite

import (
	"GScan/infoscan/dao"
	"GScan/infoscan/dao/base"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

func NewDB(dbfile string) *base.DAO {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             5 * time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Error,    // 日志级别
			IgnoreRecordNotFoundError: true,            // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,           // 禁用彩色打印
		},
	)
	db, err := gorm.Open(sqlite.Open(dbfile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = db.AutoMigrate(&dao.Page{}, &dao.WebTree{}, dao.Job{}, dao.ProcessResult{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	return &base.DAO{Db: db, Mutex: sync.Mutex{}}
}
