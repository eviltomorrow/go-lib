package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBName = "test_db_mysql"
	rootDSN    = "root:root@tcp(127.0.0.1:3306)/"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MySQL not available, skipping integration tests: %v\n", err)
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rootDB.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "MySQL not reachable, skipping integration tests: %v\n", err)
		rootDB.Close()
		os.Exit(0)
	}

	rootDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName))
	if _, err := rootDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE `%s`", testDBName)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test database: %v\n", err)
		rootDB.Close()
		os.Exit(1)
	}
	rootDB.Close()

	testDSN := fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", testDBName)
	testDB, err = sql.Open("mysql", testDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test db: %v\n", err)
		os.Exit(1)
	}
	testDB.SetMaxOpenConns(5)
	testDB.SetMaxIdleConns(2)
	DB = testDB

	code := m.Run()

	DB = nil
	testDB.Close()

	rootDB2, _ := sql.Open("mysql", rootDSN)
	rootDB2.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName))
	rootDB2.Close()

	os.Exit(code)
}

func TestBuildMySQL_Valid(t *testing.T) {
	cfg := Config{
		DSN:                fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", testDBName),
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     5 * time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	pool, err := buildMySQL(&cfg)
	require.Nil(t, err)
	require.NotNil(t, pool)
	assert.NoError(t, pool.Ping())
	pool.Close()
}

func TestBuildMySQL_EmptyDSN(t *testing.T) {
	cfg := Config{
		DSN:                "",
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	pool, err := buildMySQL(&cfg)
	assert.NotNil(t, err)
	assert.Nil(t, pool)
}

func TestBuildMySQL_InvalidHost(t *testing.T) {
	cfg := Config{
		DSN:                "root:root@tcp(192.0.2.1:3306)/test?connectTimeout=1s",
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	pool, err := buildMySQL(&cfg)
	assert.NotNil(t, err)
	assert.Nil(t, pool)
}

func TestBuildMySQL_InvalidCredentials(t *testing.T) {
	cfg := Config{
		DSN:                fmt.Sprintf("wrong:wrong@tcp(127.0.0.1:3306)/%s", testDBName),
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     3 * time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	pool, err := buildMySQL(&cfg)
	assert.NotNil(t, err)
	assert.Nil(t, pool)
}

func TestTryConnect_RetryThenFail(t *testing.T) {
	cfg := Config{
		DSN:                "root:root@tcp(192.0.2.1:3306)/test?connectTimeout=1s",
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     time.Second,
		StartupRetryTimes:  2,
		StartupRetryPeriod: 10 * time.Millisecond,
	}

	pool, err := tryConnect(&cfg)
	assert.NotNil(t, err)
	assert.Nil(t, pool)
}

func TestInitMySQL_Success(t *testing.T) {
	origDB := DB
	t.Cleanup(func() { DB = origDB })

	cfg := Config{
		DSN:                fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", testDBName),
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     5 * time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	cleanup, err := InitMySQL(&cfg)
	require.Nil(t, err)
	require.NotNil(t, DB)
	assert.NoError(t, DB.Ping())

	DB2 := DB
	err = cleanup()
	assert.Nil(t, err)
	assert.Nil(t, DB)
	assert.ErrorContains(t, DB2.Ping(), "closed")
}

func TestInitMySQL_Failure(t *testing.T) {
	origDB := DB
	t.Cleanup(func() { DB = origDB })

	DB = nil

	cfg := Config{
		DSN:                "root:root@tcp(192.0.2.1:3306)/test?connectTimeout=1s",
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: 10 * time.Millisecond,
	}

	_, err := InitMySQL(&cfg)
	assert.NotNil(t, err)
	assert.Nil(t, DB)
}

func TestDB_QueryWorks(t *testing.T) {
	result, err := DB.ExecContext(context.Background(), "SELECT 1")
	require.Nil(t, err)
	affected, err := result.RowsAffected()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), affected)
}

func TestDB_CreateTableAndInsert(t *testing.T) {
	ctx := context.Background()
	tableName := "test_mysql_driver"

	_, err := DB.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, tableName))
	require.Nil(t, err)
	t.Cleanup(func() { DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)) })

	result, err := DB.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (name) VALUES (?)", tableName), "hello")
	require.Nil(t, err)
	id, err := result.LastInsertId()
	assert.Nil(t, err)
	assert.True(t, id > 0)

	var name string
	err = DB.QueryRowContext(ctx, fmt.Sprintf("SELECT name FROM %s WHERE id = ?", tableName), id).Scan(&name)
	assert.Nil(t, err)
	assert.Equal(t, "hello", name)
}

func TestInitMySQL_CloseTwice(t *testing.T) {
	origDB := DB
	t.Cleanup(func() { DB = origDB })

	cfg := Config{
		DSN:                fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", testDBName),
		MaxLifetime:        300 * time.Second,
		MaxOpen:            5,
		MinOpen:            3,
		ConnectTimeout:     5 * time.Second,
		StartupRetryTimes:  1,
		StartupRetryPeriod: time.Second,
	}

	cleanup, err := InitMySQL(&cfg)
	require.Nil(t, err)

	cleanup()
	err = cleanup()
	assert.Nil(t, err)
}

func TestExecInterface(t *testing.T) {
	var e Exec = DB
	assert.NotNil(t, e)
}

func TestQueryPerLimitConstant(t *testing.T) {
	assert.Equal(t, int64(50), QueryPerLimit)
}
