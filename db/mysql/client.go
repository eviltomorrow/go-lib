package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eviltomorrow/go-lib/zlog"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var DB *sql.DB

const QueryPerLimit int64 = 50

func InitMySQL(c *Config) (func() error, error) {
	client, err := tryConnect(c)
	if err != nil {
		return nil, err
	}
	DB = client

	return func() error {
		if DB == nil {
			return nil
		}

		if err := DB.Close(); err != nil {
			return err
		}
		DB = nil
		return nil
	}, nil
}

func tryConnect(c *Config) (*sql.DB, error) {
	i := 1
	for {
		pool, err := buildMySQL(c)
		if err == nil {
			return pool, nil
		}
		zlog.Error("connect to mysql failure", zap.Error(err))
		i++
		if i > c.StartupRetryTimes {
			return nil, err
		}

		time.Sleep(time.Duration(c.StartupRetryPeriod))
	}
}

func buildMySQL(c *Config) (*sql.DB, error) {
	if c.DSN == "" {
		return nil, fmt.Errorf("MySQL: no DSN set")
	}
	pool, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, err
	}
	pool.SetConnMaxLifetime(time.Duration(c.MaxLifetime))
	pool.SetMaxOpenConns(c.MaxOpen)
	pool.SetMaxIdleConns(c.MinOpen)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.ConnectTimeout))
	defer cancel()

	if err = pool.PingContext(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

// Exec exec mysql
type Exec interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
