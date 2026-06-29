package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func validConfig() Config {
	return Config{
		DSN:                "root:root@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		MinOpen:            3,
		MaxOpen:            10,
		MaxLifetime:        300 * time.Second,
		ConnectTimeout:     10 * time.Second,
		StartupRetryTimes:  3,
		StartupRetryPeriod: 3 * time.Second,
	}
}

func TestVerifyConfig_Valid(t *testing.T) {
	cfg := validConfig()
	assert.Nil(t, cfg.VerifyConfig())
}

func TestVerifyConfig_EmptyDSN(t *testing.T) {
	cfg := validConfig()
	cfg.DSN = ""
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "dsn")
}

func TestVerifyConfig_MinOpenBelowLimit(t *testing.T) {
	cfg := validConfig()
	cfg.MinOpen = 1
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "min_open")
}

func TestVerifyConfig_MaxOpenAboveLimit(t *testing.T) {
	cfg := validConfig()
	cfg.MaxOpen = 20
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "max_open")
}

func TestVerifyConfig_MinOpenGreaterThanMaxOpen(t *testing.T) {
	cfg := validConfig()
	cfg.MinOpen = 10
	cfg.MaxOpen = 5
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "min_open should be less than")
}

func TestVerifyConfig_ZeroMaxLifetime(t *testing.T) {
	cfg := validConfig()
	cfg.MaxLifetime = 0
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "max_lifetime")
}

func TestVerifyConfig_ZeroConnectTimeout(t *testing.T) {
	cfg := validConfig()
	cfg.ConnectTimeout = 0
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "connect_timeout")
}

func TestVerifyConfig_ZeroStartupRetryTimes(t *testing.T) {
	cfg := validConfig()
	cfg.StartupRetryTimes = 0
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "startup_retry_times")
}

func TestVerifyConfig_ZeroStartupRetryPeriod(t *testing.T) {
	cfg := validConfig()
	cfg.StartupRetryPeriod = 0
	err := cfg.VerifyConfig()
	assert.Contains(t, err.Error(), "startup_retry_period")
}

func TestVerifyConfig_SetsDefaults(t *testing.T) {
	cfg := Config{
		DSN:                validConfig().DSN,
		MaxLifetime:        300 * time.Second,
		ConnectTimeout:     10 * time.Second,
		StartupRetryTimes:  3,
		StartupRetryPeriod: 3 * time.Second,
	}
	assert.Nil(t, cfg.VerifyConfig())
	assert.Equal(t, DEFAULT_MYSQL_MIN_OPEN, cfg.MinOpen)
	assert.Equal(t, DEFAULT_MYSQL_MAX_OPEN, cfg.MaxOpen)
}

func TestString(t *testing.T) {
	cfg := validConfig()
	s := cfg.String()
	assert.Contains(t, s, "root:root")
}
