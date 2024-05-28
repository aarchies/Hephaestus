package sqlitec

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var c *gorm.DB

func Open(dataBase string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dataBase), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logrus.Fatalln("无法连接到数据库:" + err.Error())
	}
	c = db
	return db
}

func Migrate(model interface{}) {
	if err := c.AutoMigrate(&model); err != nil {
		logrus.Fatalln("自动迁移失败:" + err.Error())
	}
}

func DB() *gorm.DB {
	return c
}
