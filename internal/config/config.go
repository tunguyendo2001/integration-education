package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App         AppConfig         `yaml:"app"`
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Storage     StorageConfig     `yaml:"storage"`
	ExternalAPI ExternalAPIConfig `yaml:"external_api"`
	Workers     WorkersConfig     `yaml:"workers"`
	SyncWindow  SyncWindowConfig  `yaml:"sync_window"` // Deprecated
	Logging     LoggingConfig     `yaml:"logging"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host               string        `yaml:"host"`
	Port               int           `yaml:"port"`
	User               string        `yaml:"user"`
	Password           string        `yaml:"password"`
	Name               string        `yaml:"name"`
	Charset            string        `yaml:"charset"`
	ParseTime          bool          `yaml:"parse_time"`
	Loc                string        `yaml:"loc"`
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnectionLifetime time.Duration `yaml:"connection_lifetime"`
}

type RedisConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	PoolSize       int    `yaml:"pool_size"`
	IngestionQueue string `yaml:"ingestion_queue"`
	SyncQueue      string `yaml:"sync_queue"`
	DLQSuffix      string `yaml:"dlq_suffix"`
}

type StorageConfig struct {
	S3 S3Config `yaml:"s3"`
}

type S3Config struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	UseSSL    bool   `yaml:"use_ssl"`
}

type ExternalAPIConfig struct {
	EducationDept EducationDeptConfig `yaml:"education_dept"`
}

type EducationDeptConfig struct {
	BaseURL                    string        `yaml:"base_url"`
	AuthEndpoint               string        `yaml:"auth_endpoint"`
	GradesEndpoint             string        `yaml:"grades_endpoint"`
	SemesterPermissionEndpoint string        `yaml:"semester_permission_endpoint"`
	Username                   string        `yaml:"username"`
	Password                   string        `yaml:"password"`
	TokenExpires               time.Duration `yaml:"token_expires"`
	Timeout                    time.Duration `yaml:"timeout"`
	BatchSize                  int           `yaml:"batch_size"`
	RetryAttempts              int           `yaml:"retry_attempts"`
	RetryDelay                 time.Duration `yaml:"retry_delay"`
}

type WorkersConfig struct {
	Ingestion IngestionWorkerConfig `yaml:"ingestion"`
	Sync      SyncWorkerConfig      `yaml:"sync"`
	Pull      PullWorkerConfig      `yaml:"pull"`
}

type IngestionWorkerConfig struct {
	Count     int `yaml:"count"`
	BatchSize int `yaml:"batch_size"`
}

type SyncWorkerConfig struct {
	Count     int `yaml:"count"`
	BatchSize int `yaml:"batch_size"`
}

type PullWorkerConfig struct {
	Count      int           `yaml:"count"`
	Interval   time.Duration `yaml:"interval"`
	RunOnStart bool          `yaml:"run_on_start"`
}

// Deprecated: Use SemesterPermissionEndpoint instead
type SyncWindowConfig struct {
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
	Timezone  string `yaml:"timezone"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// MySQL DSN format: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.Charset, c.Database.ParseTime, c.Database.Loc)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
