package config

import (
	"errors"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
	"time"
)

type Config struct {
	Log        LogConfig        `yaml:"log"`
	MySQl      MySqlConfig      `yaml:"mysql"`
	Redis      RedisConfig      `yaml:"redis"`
	HttpServer HttpServerConfig `yaml:"http_server"`
	Auth       AuthConfig       `yaml:"auth"`
	SQLRunner  SQLRunnerConfig  `yaml:"sql_runner"`
	Gemini     GeminiConfig     `yaml:"gemini"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"ACCESS_TOKEN_TTL" env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"REFRESH_TOKEN_TTL" env-default:"720h"`
	SignKey         string        `yaml:"sign_key" env:"SIGN_KEY" env-default:""`
}

type LogConfig struct {
	Path   string `yaml:"path" env:"LOG_PATH" env-default:"logs"`
	Level  string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
	Format string `yaml:"format" env:"LOG_FORMAT" env-default:"json"`
}

type MySqlConfig struct {
	Addr           string `yaml:"addr" env:"MYSQL_ADDR" env-default:"localhost:3306"`
	User           string `yaml:"user" env:"MYSQL_USER" env-default:"root"`
	Password       string `yaml:"password" env:"MYSQL_PASSWORD" env-default:""`
	Schema         string `yaml:"schema" env:"MYSQL_SCHEMA" env-default:"public"`
	ConnectTimeout int    `yaml:"connect_timeout" env:"DB_CONNECT_TIMEOUT" env-default:"10"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
}

type HttpServerConfig struct {
	Addr            string        `yaml:"addr" env:"HTTP_ADDR" env-default:"localhost:8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	HandleTimeout   time.Duration `json:"handle_timeout" env:"HTTP_HANDE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
	Log             HttpLog       `yaml:"log"`
}

type HttpLog struct {
	MaxRequestContentLen   int      `yaml:"max_request_content_len" env:"HTTP_LOG_MAX_REQUEST_CONTENT_LEN" env-default:"2048"`
	MaxResponseContentLen  int      `yaml:"max_response_content_len" env:"HTTP_LOG_MAX_RESPONSE_CONTENT_LEN" env-default:"2048"`
	RequestLoggingContent  []string `yaml:"request_logging_content" env:"HTTP_LOG_REQUEST_LOGGING_CONTENT" env-default:""`
	ResponseLoggingContent []string `yaml:"response_logging_content" env:"HTTP_LOG_RESPONSE_LOGGING_CONTENT" env-default:""`
}

type SQLRunnerConfig struct {
	MaxRows int `yaml:"max_rows" env:"SQL_MAX_ROWS" env-default:"1000"`
}

type GeminiConfig struct {
	Model  string `yaml:"model" env:"GEMINI_MODEL" env-default:"gemini-3-flash-preview"`
	ApiKey string `yaml:"api_key" env:"GEMINI_API_KEY"`
}

func ReadConfig(path string, dotenv ...string) (*Config, error) {
	if err := godotenv.Load(dotenv...); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	cfg := new(Config)
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
