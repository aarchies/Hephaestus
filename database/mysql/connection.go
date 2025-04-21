package mysql

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Option struct {
	Hosts       []string
	Port        int
	Username    string
	Password    string
	DataBase    string
	MaxIdleConn int
	MaxOpenConn int
	LogMode     string
	Config      string
}

func NewOption(hosts []string, port int, username, password, database string) *Option {
	return &Option{
		Hosts:       hosts,
		Port:        port,
		Username:    username,
		Password:    password,
		DataBase:    database,
		MaxIdleConn: 10,
		MaxOpenConn: 100,
		LogMode:     "info",
		Config:      "",
	}
}

func (m *Option) Connect() *gorm.DB {
	_logger := logger.New(logger.Writer(log.New(os.Stdout, "\r\n", log.LstdFlags)), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	option := &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true,
		Logger: _logger,
	}
	switch m.LogMode {
	case "silent", "Silent":
		option.Logger = _logger.LogMode(logger.Silent)
	case "error", "Error":
		option.Logger = _logger.LogMode(logger.Error)
	case "warn", "Warn":
		option.Logger = _logger.LogMode(logger.Warn)
	case "info", "Info":
		option.Logger = _logger.LogMode(logger.Info)
	}

	if db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   // 根据版本自动配置
	}), option); err != nil {
		return nil
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConn)
		sqlDB.SetMaxOpenConns(m.MaxOpenConn)
		return db
	}
}

func (m *Option) WithConfig(config string) *Option {
	m.Config = strings.TrimSpace(config)
	return m
}

func (m *Option) WithLogMode(mode string) *Option {
	m.LogMode = strings.TrimSpace(mode)
	return m
}

func (m *Option) WithMaxIdleConn(c int) *Option {
	m.MaxIdleConn = c
	return m
}

func (m *Option) WithMaxOpenConn(c int) *Option {
	m.MaxOpenConn = c
	return m
}

func (m *Option) Dsn() string {
	hosts := strings.Join(m.Hosts, ",")
	return m.Username + ":" + m.Password + "@tcp(" + hosts + ":" + strconv.Itoa(m.Port) + ")/" + m.DataBase + "?" + m.Config
}
