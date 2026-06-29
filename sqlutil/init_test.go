package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	dbmysql "github.com/eviltomorrow/go-lib/db/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	testDB     *sql.DB
	testDSN    string
	testDBName = "sqlutil_test"
	testTable  = "sqlutil_test_users"
)

func TestMain(m *testing.M) {
	dsn := "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
	rootDB, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to mysql: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := rootDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to drop test database: %v\n", err)
		os.Exit(1)
	}
	if _, err := rootDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE `%s`", testDBName)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test database: %v\n", err)
		os.Exit(1)
	}
	rootDB.Close()

	testDSN = fmt.Sprintf("root:root@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", testDBName)
	testDB, err = sql.Open("mysql", testDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test db: %v\n", err)
		os.Exit(1)
	}
	testDB.SetMaxOpenConns(5)
	testDB.SetMaxIdleConns(2)
	dbmysql.DB = testDB

	createTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		age INT NOT NULL DEFAULT 0,
		email VARCHAR(255),
		score DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		status INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		modify_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`, testTable)
	if _, err := testDB.ExecContext(ctx, createTableSQL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test table: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testDB.Exec("DROP TABLE IF EXISTS " + testTable)
	testDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", testDBName))
	testDB.Close()
	dbmysql.DB = nil

	os.Exit(code)
}

func truncateTestTable(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", testTable))
	if err != nil {
		t.Fatalf("failed to truncate test table: %v", err)
	}
}

func insertTestUser(t *testing.T, name string, age int, email interface{}, score float64, status int) int64 {
	t.Helper()
	result, err := testDB.Exec(
		fmt.Sprintf("INSERT INTO %s (name, age, email, score, status) VALUES (?, ?, ?, ?, ?)", testTable),
		name, age, email, score, status,
	)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}
