// Пакет для работы с файлом конфига.
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

// MustLoad - инициализирует данные из конфиг файла. Путь к файлу берет из
// переменной окружения CONFIG_PATH_SF, пароль для доступа к БД - из переменной
// окружения DB_PASSWD. Если не удается, то завершает приложение с ошибкой.
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH_SF")
	if configPath == "" {
		log.Fatal("CONFIG_PATH_SF is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot read config file: %s, %s", configPath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("cannot decode config file: %s, %s", configPath, err)
	}

	cfg.StoragePasswd = os.Getenv("DB_PASSWD")

	return &cfg
}
