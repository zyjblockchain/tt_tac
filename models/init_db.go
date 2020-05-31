package models

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
)

var DB *gorm.DB

func InitDB(dsn string) {
	log.Infof("数据库dsn: %s", dsn)
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 设置数据库的日志级别
	if gin.Mode() == gin.ReleaseMode {
		db.LogMode(false)
	} else {
		db.LogMode(true)
	}

	DB = db
	autoCreateTable()
	log.Infof("数据库连接成功")
}

// autoCreateTable 自动建表
func autoCreateTable() {
	DB.AutoMigrate(&CollectionTx{}, &Kv{}, &TacOrder{}, &TxTransfer{}, &User{}, &FlashChangeOrder{}, &SendTransfer{})
}
