package mysql

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	DEFAULT_MYSQL_MIN_OPEN       = 3
	DEFAULT_MYSQL_MAX_OPEN       = 10
	DEFAULT_MYSQL_MIN_OPEN_LIMIT = DEFAULT_MYSQL_MIN_OPEN
	DEFAULT_MYSQL_MAX_OPEN_LIMIT = DEFAULT_MYSQL_MAX_OPEN
)

type Config struct {
	StartupRetryPeriod time.Duration `json:"startup_retry_period" toml:"-" mapstructure:"-"`
	StartupRetryTimes  int           `json:"startup_retry_times" toml:"-" mapstructure:"-"`
	ConnectTimeout     time.Duration `json:"connect_timeout" toml:"-" mapstructure:"-"`

	DSN         string        `json:"dsn" toml:"dsn" mapstructure:"dsn"`
	MinOpen     int           `json:"min_open" toml:"min_open" mapstructure:"min_open"`
	MaxOpen     int           `json:"max_open" toml:"max_open" mapstructure:"max_open"`
	MaxLifetime time.Duration `json:"max_lifetime" toml:"-" mapstructure:"-"`
}

func (m *Config) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(m)
	return string(buf)
}

func (c *Config) VerifyConfig() error {
	if c.DSN == "" {
		return fmt.Errorf("mysql.dsn has no value")
	}

	if c.MinOpen == 0 {
		c.MinOpen = DEFAULT_MYSQL_MIN_OPEN
	}
	if c.MaxOpen == 0 {
		c.MaxOpen = DEFAULT_MYSQL_MAX_OPEN
	}
	if c.MinOpen < DEFAULT_MYSQL_MIN_OPEN_LIMIT {
		return fmt.Errorf("mysql.min_open must be greater than %d", DEFAULT_MYSQL_MIN_OPEN_LIMIT)
	}
	if c.MaxOpen > DEFAULT_MYSQL_MAX_OPEN_LIMIT {
		return fmt.Errorf("mysql.max_open must be less than %d", DEFAULT_MYSQL_MAX_OPEN_LIMIT)
	}
	if c.MinOpen > c.MaxOpen {
		return fmt.Errorf("mysql.min_open should be less than mysql.max-open")
	}
	if c.MaxLifetime <= 0 {
		return fmt.Errorf("mysql.max_lifetime has no value")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("mysql.connect_timeout has no value")
	}
	if c.StartupRetryTimes <= 0 {
		return fmt.Errorf("mysql.startup_retry_times has no value")
	}
	if c.StartupRetryPeriod <= 0 {
		return fmt.Errorf("mysql.startup_retry_period has no value")
	}
	return nil
}
