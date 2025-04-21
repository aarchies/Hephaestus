package config

type Config struct {
	Hosts       []string `yaml:"hosts"`
	UserName    string   `yaml:"username"`
	PassWord    string   `yaml:"password"`
	DataBase    string   `yaml:"dataBase"`
	MaxIdleConn int      `yaml:"max-idle-conn"` // 空闲中的最大连接数
	MaxOpenConn int      `yaml:"max-open-conn"` // 打开到数据库的最大连接数
	IsDebug     bool     `yaml:"is_debug"`
	LogMode     string   `yaml:"log-mode"`
}
