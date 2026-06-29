package sqlutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/eviltomorrow/go-lib/zlog"
	"go.uber.org/zap"
)

const defaultMigrationTable = "schema_migrations"

type migrationRecord struct {
	Filename string
	Checksum string
}

func fileChecksum(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func splitStatements(data string) []string {
	var stmts []string
	for _, block := range strings.Split(data, ";") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		hasSQL := false
		for _, line := range strings.Split(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			hasSQL = true
			break
		}
		if hasSQL {
			stmts = append(stmts, block)
		}
	}
	return stmts
}

func stripDBName(dsn string) (string, string) {
	slashIdx := strings.LastIndex(dsn, "/")
	if slashIdx == -1 {
		return dsn, ""
	}
	qIdx := strings.Index(dsn[slashIdx:], "?")
	var dbname, params string
	if qIdx == -1 {
		dbname = dsn[slashIdx+1:]
	} else {
		dbname = dsn[slashIdx+1 : slashIdx+qIdx]
		params = dsn[slashIdx+qIdx:]
	}
	baseDSN := dsn[:slashIdx+1] + params
	return baseDSN, dbname
}

func applySQL(ctx context.Context, db *sql.DB, data []byte) error {
	stmts := splitStatements(string(data))
	if len(stmts) == 0 {
		return nil
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx failure: %w", err)
	}
	for i, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute statement %d failure: %w", i+1, err)
		}
	}
	return tx.Commit()
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	ddl := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		filename   VARCHAR(255) NOT NULL,
		checksum   VARCHAR(64)  NOT NULL,
		applied_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (filename)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`, defaultMigrationTable)
	_, err := db.ExecContext(ctx, ddl)
	return err
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[string]string, error) {
	rows, err := db.QueryContext(ctx,
		fmt.Sprintf("SELECT filename, checksum FROM %s ORDER BY filename", defaultMigrationTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var r migrationRecord
		if err := rows.Scan(&r.Filename, &r.Checksum); err != nil {
			return nil, err
		}
		result[r.Filename] = r.Checksum
	}
	return result, rows.Err()
}

func upsertMigration(ctx context.Context, db *sql.DB, filename, checksum string) error {
	_, err := db.ExecContext(ctx,
		fmt.Sprintf("REPLACE INTO %s (filename, checksum, applied_at) VALUES (?, ?, ?)", defaultMigrationTable),
		filename, checksum, time.Now())
	return err
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Second)
	return db, nil
}

func Migrate(ctx context.Context, dsn string, sqlDir string) error {
	baseDSN, dbname := stripDBName(dsn)

	baseDB, err := openDB(baseDSN)
	if err != nil {
		return fmt.Errorf("open base connection failure: %w", err)
	}
	defer baseDB.Close()

	if err := baseDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping base connection failure: %w", err)
	}

	if _, err := baseDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbname)); err != nil {
		return fmt.Errorf("create database failure: %w", err)
	}
	baseDB.Close()

	db, err := openDB(dsn)
	if err != nil {
		return fmt.Errorf("open database connection failure: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database connection failure: %w", err)
	}

	if err := ensureMigrationTable(ctx, db); err != nil {
		return fmt.Errorf("ensure migration table failure: %w", err)
	}

	entries, err := os.ReadDir(sqlDir)
	if err != nil {
		return fmt.Errorf("read sql dir failure: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	if len(files) == 0 {
		zlog.Warn("no sql files found", zap.String("dir", sqlDir))
		return nil
	}

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return fmt.Errorf("query applied migrations failure: %w", err)
	}

	for _, filename := range files {
		data, err := os.ReadFile(filepath.Join(sqlDir, filename))
		if err != nil {
			return fmt.Errorf("read file %s failure: %w", filename, err)
		}

		cs := fileChecksum(data)
		if oldCS, ok := applied[filename]; ok && oldCS == cs {
			zlog.Debug("migration skipped", zap.String("file", filename))
			continue
		}

		if err := applySQL(ctx, db, data); err != nil {
			return fmt.Errorf("apply migration %s failure: %w", filename, err)
		}

		if err := upsertMigration(ctx, db, filename, cs); err != nil {
			return fmt.Errorf("record migration %s failure: %w", filename, err)
		}

		zlog.Info("migration applied", zap.String("file", filename))
	}

	for filename := range applied {
		found := false
		for _, f := range files {
			if f == filename {
				found = true
				break
			}
		}
		if !found {
			zlog.Warn("orphan migration record", zap.String("file", filename))
		}
	}

	return nil
}
