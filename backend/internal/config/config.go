package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Auth     AuthConfig
	Parser   ParserConfig
	Minio    MinioConfig
}

type ServerConfig struct {
	Port  int    `env:"SERVER_PORT" envDefault:"8080"`
	Mode  string `env:"SERVER_MODE" envDefault:"debug"`
}

type DatabaseConfig struct {
	Host        string `env:"DB_HOST" envDefault:"postgres"`
	Port        int    `env:"DB_PORT" envDefault:"5432"`
	Name        string `env:"DB_NAME" envDefault:"media_parser"`
	User        string `env:"DB_USER" envDefault:"media_parser_user"`
	Password    string `env:"DB_PASSWORD" envDefault:"media_parser_password"`
	MaxConns    int    `env:"DB_MAX_CONNS" envDefault:"20"`
	MinConns    int    `env:"DB_MIN_CONNS" envDefault:"5"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"redis"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:"media_parser_password"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	CacheTTL int    `env:"CACHE_TTL_HOURS" envDefault:"24"`
}

type RabbitMQConfig struct {
	Host     string `env:"RABBIT_HOST" envDefault:"rabbitmq"`
	Port     int    `env:"RABBIT_PORT" envDefault:"5672"`
	User     string `env:"RABBIT_USER" envDefault:"media_parser_user"`
	Password string `env:"RABBIT_PASSWORD" envDefault:"media_parser_password"`
	VHost    string `env:"RABBIT_VHOST" envDefault:"/media_parser"`
	Queue    string `env:"RABBIT_QUEUE" envDefault:"parse_queue"`
	Prefetch int    `env:"QUEUE_PREFETCH" envDefault:"10"`
}

type AuthConfig struct {
	Tokens []string `env:"AUTH_TOKENS" envSeparator:"," envDefault:""`
}

type ParserConfig struct {
	BrowserHeadless        bool   `env:"BROWSER_HEADLESS" envDefault:"true"`
	PageLoadTimeout        int    `env:"PAGE_LOAD_TIMEOUT" envDefault:"60"`
	CacheTTLHours          int    `env:"CACHE_TTL_HOURS" envDefault:"24"`
	ParserWorkers          int    `env:"PARSER_WORKERS" envDefault:"15"`
	HTTPWorkers            int    `env:"HTTP_WORKERS" envDefault:"50"`
	DownloadWorkers        int    `env:"DOWNLOAD_WORKERS" envDefault:"5"`
	MaxConcurrentPerHost   int    `env:"MAX_CONCURRENT_PER_SOURCE" envDefault:"3"`
	RequestTimeout         int    `env:"REQUEST_TIMEOUT" envDefault:"30"`
	// YouTube parser config
	YouTubeEnabled     bool `env:"YOUTUBE_ENABLED" envDefault:"true"`
	YouTubeTimeout     int  `env:"YOUTUBE_TIMEOUT" envDefault:"30"` // секунды
	// Robots.txt config
	IgnoreRobotsTxt bool   `env:"IGNORE_ROBOTS_TXT" envDefault:"true"`
	ParserUserAgent string `env:"PARSER_USER_AGENT" envDefault:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"`
}

type MinioConfig struct {
	Host     string `env:"MINIO_HOST" envDefault:"minio"`
	Port     int    `env:"MINIO_PORT" envDefault:"9000"`
	User     string `env:"MINIO_USER" envDefault:"minioadmin"`
	Password string `env:"MINIO_PASSWORD" envDefault:"minioadmin"`
	Bucket   string `env:"MINIO_BUCKET" envDefault:"media-parser"`
	UseSSL   bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func (c *RedisConfig) CacheTTLHours() time.Duration {
	return time.Duration(c.CacheTTL) * time.Hour
}

func (c *ParserConfig) PageLoadTimeoutSeconds() time.Duration {
	return time.Duration(c.PageLoadTimeout) * time.Second
}

func (c *ParserConfig) RequestTimeoutSeconds() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		c.Host, c.Port, c.Name, c.User, c.Password,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *RabbitMQConfig) URL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.VHost,
	)
}

