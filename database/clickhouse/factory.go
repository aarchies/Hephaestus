package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"gorm.io/gorm"
)

type ConnectionFactory struct {
	conn     driver.Conn
	db       *gorm.DB
	dataBase string
}
