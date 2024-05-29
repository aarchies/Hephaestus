package clickhousec

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	Gc "gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

var (
	conn            driver.Conn
	db              *gorm.DB
	currentDataBase string
)

func Connect(hosts []string, username string, password string, dataBase string, isDebug bool) driver.Conn {

	fmt.Printf("Load clienthouse config Host:%s username:%s pwd:%s dataBase:%s \n", hosts, username, password, dataBase)

	config := &clickhouse.Options{
		Addr: hosts,
		Auth: clickhouse.Auth{
			Database: "default",
			Username: username,
			Password: password,
		},
		Debug: isDebug,
		Debugf: func(format string, v ...any) {
			logrus.Debugf(format+"\n", v...)
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		ConnOpenStrategy: clickhouse.ConnOpenRoundRobin,
	}
	dB, err := clickhouse.Open(config)
	if err != nil {
		logrus.Fatalf("连接clickhouse失败! %s", err.Error())
	}

	if err := dB.Ping(context.Background()); err != nil {
		logrus.Fatalln(err.Error())
	}
	conn = dB

	if d, err := gorm.Open(Gc.New(Gc.Config{
		Conn: clickhouse.OpenDB(config),
	})); err != nil {
		logrus.Fatalln(err.Error())
	} else {
		db = d
	}

	currentDataBase = dataBase
	return conn
}

func DB() driver.Conn {
	return conn
}

func GormDB() *gorm.DB {
	return db
}

func DataBase() string {
	return currentDataBase
}
