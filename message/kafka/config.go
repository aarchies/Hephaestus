package kafka

type Config struct {
	Adders   []string
	Username string
	Password string
	ClientId string
}

func NewConfig(adders []string, clientId string, username string, password string) Config {
	return Config{
		Adders:   adders,
		Username: username,
		Password: password,
		ClientId: clientId,
	}
}
