package app

// config pulled from environment variables
type Config struct {
	MQHost     string `env:"MQHOST"`
	MQPort     string `env:"MQPORT" `
	Strategy   string `env:"STRATEGY"`
	MQUser     string `env:"MQUSER"`
	MQPassword string `env:"MQPASSWORD"`
}
