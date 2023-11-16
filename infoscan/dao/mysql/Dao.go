package mysql

import (
	"GScan/infoscan/dao"
	"GScan/infoscan/dao/base"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func NewDB(host string, port int, username string, password string, db string) *base.DAO {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             5 * time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Error,    // 日志级别
			IgnoreRecordNotFoundError: true,            // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,           // 禁用彩色打印
		},
	)
	mysqldb, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       username + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + db + "?charset=utf8mb4&parseTime=True&loc=Local", // DSN data source name
		DefaultStringSize:         512,                                                                                                                             // string 类型字段的默认长度
		DisableDatetimePrecision:  true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalln("数据库配置错误", err.Error())
	}
	err = mysqldb.AutoMigrate(&dao.Page{}, &dao.WebTree{}, dao.Job{}, dao.ProcessResult{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	return &base.DAO{Db: mysqldb, Mutex: sync.Mutex{}}

}
