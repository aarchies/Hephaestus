package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	"log"
)

var connect driver.Conn

func Register(hosts []string, username string, password string, isDebug bool) driver.Conn {

	fmt.Printf("Load clienthouse config Host:%s\n", hosts)

	con, err := clickhouse.Open(&clickhouse.Options{
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
		MaxIdleConns:     128,
		ConnOpenStrategy: clickhouse.ConnOpenRoundRobin,
	})
	if err != nil {
		logrus.Fatalf("连接clickhouse失败! %s", err.Error())
	}

	if err := con.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	connect = con
	return connect
}

func GetClickhouseClient() driver.Conn {
	return connect
}
