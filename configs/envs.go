package configs

type Config struct {
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL,required"`
	Env         string `env:"ENV" envDefault:"development"`
}

func initConfig() *Config {
	return &Config{}
}