package clickhouse

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	ck "gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Option struct {
	Hosts                []string
	Username             string
	Password             string
	DataBase             string
	MaxIdleConn          int
	MaxOpenConn          int
	BlockBufferSize      int
	MaxCompressionBuffer int
	IsDebug              bool
	LogMode              string
}

func NewOption(hosts []string, username, password, database string) *Option {
	return &Option{
		Hosts:                hosts,
		Username:             username,
		Password:             password,
		DataBase:             database,
		MaxIdleConn:          256,
		MaxOpenConn:          256,
		IsDebug:              false,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	}
}

func (m *Option) WithMaxIdleConn(c int) *Option {
	m.MaxIdleConn = c
	return m
}

func (m *Option) WithMaxOpenConn(c int) *Option {
	m.MaxOpenConn = c
	return m
}

func (m *Option) WithIsDebug(c bool) *Option {
	m.IsDebug = c
	return m
}

func (m *Option) WithBlockBufferSize(c int) *Option {
	m.BlockBufferSize = c
	return m
}

func (m *Option) WithMaxCompressionBuffer(c int) *Option {
	m.MaxCompressionBuffer = c
	return m
}

func (m *Option) Connect() *ConnectionFactory {
	option := &clickhouse.Options{
		Addr: m.Hosts,
		Auth: clickhouse.Auth{
			Database: m.DataBase,
			Username: m.Username,
			Password: m.Password,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		ConnOpenStrategy:     clickhouse.ConnOpenRoundRobin,
		MaxOpenConns:         m.MaxOpenConn,
		MaxIdleConns:         m.MaxIdleConn,
		DialTimeout:          time.Second * 30,
		Debug:                m.IsDebug,
		BlockBufferSize:      uint8(m.BlockBufferSize),
		MaxCompressionBuffer: m.MaxCompressionBuffer,
	}
	conn, err := clickhouse.Open(option)
	if err != nil {
		logrus.Fatalf("connecting clickhouse error! %s", err.Error())
	}
	if err := conn.Ping(context.Background()); err != nil {
		logrus.Fatalln(err.Error())
	}

	// gorm
	_logger := logger.New(logger.Writer(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Silent,
		Colorful:      true,
	})
	_gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(logger.Writer(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Silent,
			Colorful:      true,
		}),
	}

	switch m.LogMode {
	case "silent", "Silent":
		_gormConfig.Logger = _logger.LogMode(logger.Silent)
	case "error", "Error":
		_gormConfig.Logger = _logger.LogMode(logger.Error)
	case "warn", "Warn":
		_gormConfig.Logger = _logger.LogMode(logger.Warn)
	case "info", "Info":
		_gormConfig.Logger = _logger.LogMode(logger.Info)
	default:
		_gormConfig.Logger = _logger.LogMode(logger.Silent)
	}
	gormDb, err := gorm.Open(ck.New(ck.Config{
		Conn: clickhouse.OpenDB(&clickhouse.Options{
			Addr: m.Hosts,
			Auth: clickhouse.Auth{
				Database: m.DataBase,
				Username: m.Username,
				Password: m.Password,
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			ConnOpenStrategy:     clickhouse.ConnOpenRoundRobin,
			DialTimeout:          time.Second * 30,
			Debug:                m.IsDebug,
			BlockBufferSize:      uint8(m.BlockBufferSize),
			MaxCompressionBuffer: m.MaxCompressionBuffer,
		}),
	}), _gormConfig)

	if err != nil {
		logrus.Fatalln(err.Error())
	}

	logrus.Infof("clientHouse cluster connected successful! host:%s dataBase:[%s]\n", m.Hosts, m.DataBase)

	return &ConnectionFactory{
		conn:     conn,
		db:       gormDb,
		dataBase: m.DataBase,
	}
}

func (c *ConnectionFactory) Conn() driver.Conn {
	return c.conn
}

func (c *ConnectionFactory) DB() *gorm.DB {
	return c.db
}

func (c *ConnectionFactory) DataBase() string {
	return c.dataBase
}

func (c *ConnectionFactory) AsyncInsert(table string, data interface{}) error {
	batch, err := c.conn.PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %s.%s SETTINGS async_insert=1, wait_for_async_insert=1", c.DataBase(), table))
	if err != nil {
		return err
	}

	if err := batch.AppendStruct(data); err != nil {
		return err
	}

	return batch.Send()
}
