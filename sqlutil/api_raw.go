package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	db "github.com/eviltomorrow/go-lib/db/mysql"
)

func TableWithDelete(ctx context.Context, exec db.Exec, table string, where map[string]interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("invalid table")
	}

	if len(where) == 0 {
		return 0, fmt.Errorf("invalid where")
	}

	fields := make([]string, 0, len(where))
	args := make([]interface{}, 0, len(where))
	for k, v := range where {
		fields = append(fields, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}
	_sql := fmt.Sprintf(`delete from %s where %s`, table, strings.Join(fields, " and "))

	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TableWithUpdate(ctx context.Context, exec db.Exec, table string, value, where map[string]interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("invalid table")
	}

	if len(value) == 0 {
		return 0, fmt.Errorf("invalid value")
	}

	if len(where) == 0 {
		return 0, fmt.Errorf("invalid where")
	}

	fieldsValue := make([]string, 0, len(value))
	argsValue := make([]interface{}, 0, len(value)+len(where))
	for k, v := range value {
		fieldsValue = append(fieldsValue, fmt.Sprintf("%s = ?", k))
		argsValue = append(argsValue, v)
	}
	fieldsValue = append(fieldsValue, "modify_timestamp = now()")

	fieldsWhere := make([]string, 0, len(where))
	argsWhere := make([]interface{}, 0, len(where))
	for k, v := range where {
		fieldsWhere = append(fieldsWhere, fmt.Sprintf("%s = ?", k))
		argsWhere = append(argsWhere, v)
	}

	_sql := fmt.Sprintf(`update %s set %s where %s`, table, strings.Join(fieldsValue, ", "), strings.Join(fieldsWhere, " and "))
	result, err := exec.ExecContext(ctx, _sql, append(argsValue, argsWhere...)...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TableWithInsertMany(ctx context.Context, exec db.Exec, table string, columns []string, values []map[string]interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("invalid table")
	}
	if len(columns) == 0 {
		return 0, fmt.Errorf("invalid columns")
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("invalid values")
	}

	fields := make([]string, 0, len(columns))
	for range columns {
		fields = append(fields, "?")
	}
	field := fmt.Sprintf("(%s)", strings.Join(fields, ", "))

	args := make([]interface{}, 0, len(columns)*len(values))
	fields = make([]string, 0, len(values))
	for _, value := range values {
		for _, column := range columns {
			arg, ok := value[column]
			if !ok {
				return 0, fmt.Errorf("not found column with [%s]", column)
			}
			args = append(args, arg)
		}
		fields = append(fields, field)
	}

	_sql := fmt.Sprintf("insert into %s (%s) values %s", table, strings.Join(columns, ","), strings.Join(fields, ","))
	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TableWithInsertOne(ctx context.Context, exec db.Exec, table string, value map[string]interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("invalid table")
	}

	if len(value) == 0 {
		return 0, fmt.Errorf("invalid value")
	}

	fields1 := make([]string, 0, len(value))
	fields2 := make([]string, 0, len(value))
	args := make([]interface{}, 0, len(value))

	for k, v := range value {
		fields1 = append(fields1, k)
		fields2 = append(fields2, "?")
		args = append(args, v)
	}
	_sql := fmt.Sprintf(`insert into %s(%s) values (%s)`, table, strings.Join(fields1, ", "), strings.Join(fields2, ", "))

	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func TableWithSelectOne(ctx context.Context, exec db.Exec, table string, column []string, where map[string]interface{}, f func(row *sql.Row) error) error {
	if table == "" {
		return fmt.Errorf("invalid table")
	}

	if len(column) == 0 {
		return fmt.Errorf("invalid column")
	}

	if len(where) == 0 {
		return fmt.Errorf("invalid where")
	}

	fields := make([]string, 0, len(where))
	args := make([]interface{}, 0, len(where))
	for k, v := range where {
		fields = append(fields, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	_sql := fmt.Sprintf(`select %s from %s where %s`, strings.Join(column, ", "), table, strings.Join(fields, " and "))
	row := exec.QueryRowContext(ctx, _sql, args...)
	if row.Err() != nil {
		return row.Err()
	}

	if err := f(row); err != nil {
		return err
	}
	return nil
}

func TableWithSelectMany(ctx context.Context, exec db.Exec, table string, column []string, where map[string]interface{}, order map[string]string, f func(row *sql.Rows) error) error {
	if table == "" {
		return fmt.Errorf("invalid table")
	}

	if len(column) == 0 {
		return fmt.Errorf("invalid column")
	}

	_sql := fmt.Sprintf(`select %s from %s`, strings.Join(column, ", "), table)
	args := make([]interface{}, 0, len(where))

	if len(where) != 0 {
		fields := make([]string, 0, len(where))

		for k, v := range where {
			fields = append(fields, fmt.Sprintf("%s = ?", k))
			args = append(args, v)
		}
		_sql = fmt.Sprintf("%s where %s", _sql, strings.Join(fields, " and "))
	}

	if len(order) != 0 {
		fields := make([]string, 0, len(order))

		for k, v := range order {
			fields = append(fields, fmt.Sprintf("%s %s", k, v))
		}
		_sql = fmt.Sprintf("%s order by %s", _sql, strings.Join(fields, ", "))
	}
	rows, err := exec.QueryContext(ctx, _sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := f(rows); err != nil {
		return err
	}
	return rows.Err()
}

func TableWithSelectRange(ctx context.Context, exec db.Exec, table string, column []string, where map[string]interface{}, order map[string]string, offset, limit int64, f func(row *sql.Rows) error) error {
	if table == "" {
		return fmt.Errorf("invalid table")
	}

	if len(column) == 0 {
		return fmt.Errorf("invalid column")
	}

	args := make([]interface{}, 0, len(where))
	_sql := fmt.Sprintf(`select %s from %s`, strings.Join(column, ", "), table)

	if len(where) != 0 {
		fields := make([]string, 0, len(where))

		for k, v := range where {
			fields = append(fields, fmt.Sprintf("%s = ?", k))
			args = append(args, v)
		}
		_sql = fmt.Sprintf("%s where %s", _sql, strings.Join(fields, " and "))
	}

	if len(order) != 0 {
		fields := make([]string, 0, len(order))
		for k, v := range order {
			fields = append(fields, fmt.Sprintf("%s %s", k, v))
		}
		_sql = fmt.Sprintf("%s order by %s", _sql, strings.Join(fields, ", "))
	}

	_sql = fmt.Sprintf("%s limit ?, ?", _sql)
	rows, err := exec.QueryContext(ctx, _sql, append(args, offset, limit)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := f(rows); err != nil {
		return err
	}
	return nil
}

func TableWithCount(ctx context.Context, exec db.Exec, table string, where map[string]interface{}) (int64, error) {
	if table == "" {
		return 0, fmt.Errorf("invalid table")
	}

	_sql := fmt.Sprintf(`select count(1) as count from %s`, table)
	args := make([]interface{}, 0, len(where))

	if len(where) != 0 {
		fields := make([]string, 0, len(where))

		for k, v := range where {
			fields = append(fields, fmt.Sprintf("%s = ?", k))
			args = append(args, v)
		}
		_sql = fmt.Sprintf("%s where %s", _sql, strings.Join(fields, " and "))
	}

	row := exec.QueryRowContext(ctx, _sql, args...)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	if row.Err() != nil {
		return 0, row.Err()
	}

	return count, nil
}

func ExecQueryMany(ctx context.Context, exec db.Exec, sql string, args []interface{}, f func(row *sql.Rows) error) error {
	rows, err := exec.QueryContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := f(rows); err != nil {
		return err
	}
	return rows.Err()
}

func ExecQueryOne(ctx context.Context, exec db.Exec, sql string, args []interface{}, f func(row *sql.Row) error) error {
	row := exec.QueryRowContext(ctx, sql, args...)
	if row.Err() != nil {
		return row.Err()
	}

	if err := f(row); err != nil {
		return err
	}
	return row.Err()
}
