package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Структура конфига
type Config struct {
	RSSFeeds      []string      `yaml:"rss"`
	RequestPeriod time.Duration `yaml:"request_period"`
	StoragePath   string        `yaml:"storage_path"`
	StorageUser   string        `yaml:"storage_user"`
	StoragePasswd string        `yaml:"storage_passwd"`
	HTTPServer    `yaml:"http_server"`
}
type HTTPServer struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// MustLoad - инициализирует данные из конфига. Если не удается, то завершает приложение с ошибкой.
func MustLoad() *Config {
	// configPath := os.Getenv("CONFIG_PATH")
	configPath := "./config/config.yaml"

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// Проверяем, существует ли файл.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	// Читаем файл конфига и декодируем в структуру.
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot read config file: %s, %s", configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("cannot decode config file: %s, %s", configPath, err)
	}

	// Читаем пароль к БД из переменных окружения.
	cfg.StoragePasswd = os.Getenv("DB_PASSWD")

	return &cfg
}
