package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Query builder tests
// ---------------------------------------------------------------------------

func TestNewQuery_NoTable(t *testing.T) {
	err := NewQuery(testDB).QueryOne(func(row *sql.Row) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "table is invalid") {
		t.Errorf("expected 'table is invalid' error, got %v", err)
	}

	err = NewQuery(testDB).Query(func(rows *sql.Rows) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "table is invalid") {
		t.Errorf("expected 'table is invalid' error, got %v", err)
	}
}

func TestNewQueryOne(t *testing.T) {
	truncateTestTable(t)
	id := insertTestUser(t, "alice", 30, "alice@test.com", 95.5, 1)

	err := NewQuery(testDB).
		Columns([]string{"name", "age"}).
		Table(testTable).
		Where(WithEq("id", id)).
		QueryOne(func(row *sql.Row) error {
			var name string
			var age int
			if err := row.Scan(&name, &age); err != nil {
				return err
			}
			if name != "alice" || age != 30 {
				t.Errorf("got name=%s age=%d, want alice 30", name, age)
			}
			return nil
		})
	if err != nil {
		t.Fatalf("QueryOne failed: %v", err)
	}
}

func TestNewQueryOne_NoRows(t *testing.T) {
	truncateTestTable(t)
	err := NewQuery(testDB).
		Table(testTable).
		Where(WithEq("id", 9999)).
		QueryOne(func(row *sql.Row) error {
			return row.Scan()
		})
	if err == nil {
		t.Fatal("expected error for no rows, got nil")
	}
}

func TestNewQueryOne_RowCallbackError(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "bob", 25, "bob@test.com", 80, 1)

	err := NewQuery(testDB).
		Table(testTable).
		Where(WithEq("name", "bob")).
		QueryOne(func(row *sql.Row) error {
			return fmt.Errorf("callback error")
		})
	if err == nil || err.Error() != "callback error" {
		t.Errorf("expected 'callback error', got %v", err)
	}
}

func TestNewQuery_All(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, "a@t.com", 50, 1)
	insertTestUser(t, "b", 30, "b@t.com", 60, 1)
	insertTestUser(t, "c", 40, "c@t.com", 70, 0)

	count := 0
	err := NewQuery(testDB).
		Table(testTable).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestNewQuery_WithConditions(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, nil, 50, 1)
	insertTestUser(t, "b", 30, "b@t.com", 60, 1)
	insertTestUser(t, "c", 40, "c@t.com", 70, 0)

	var names []string
	err := NewQuery(testDB).
		Columns([]string{"name"}).
		Table(testTable).
		Where(
			WithGt("age", 25),
			WithEq("status", 1),
		).
		OrderBy(ASC("name")).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				var n string
				if err := rows.Scan(&n); err != nil {
					return err
				}
				names = append(names, n)
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(names) != 1 || names[0] != "b" {
		t.Errorf("names = %v, want [b]", names)
	}
}

func TestNewQuery_WithLimit(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, nil, 50, 1)
	insertTestUser(t, "b", 30, nil, 60, 1)
	insertTestUser(t, "c", 40, nil, 70, 1)

	count := 0
	err := NewQuery(testDB).
		Table(testTable).
		OrderBy(ASC("id")).
		Limit(1, 2).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestNewQuery_GroupBy(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, nil, 50, 1)
	insertTestUser(t, "b", 20, nil, 60, 1)
	insertTestUser(t, "c", 30, nil, 70, 1)

	count := 0
	err := NewQuery(testDB).
		Columns([]string{"age", "COUNT(*) as cnt"}).
		Table(testTable).
		GroupBy([]string{"age"}).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestNewQuery_ChainedWhere(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, nil, 50, 1)
	insertTestUser(t, "b", 30, nil, 60, 0)

	count := 0
	err := NewQuery(testDB).
		Table(testTable).
		Where(WithGt("age", 25)).
		Where(WithEq("status", 1)).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestNewQuery_RowsCallbackError(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "x", 10, nil, 0, 1)

	err := NewQuery(testDB).
		Table(testTable).
		Query(func(rows *sql.Rows) error {
			return fmt.Errorf("rows callback error")
		})
	if err == nil || err.Error() != "rows callback error" {
		t.Errorf("expected 'rows callback error', got %v", err)
	}
}

func TestNewQuery_Query(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "q", 10, nil, 0, 1)

	count := 0
	err := NewQuery(testDB).
		Table(testTable).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestNewQuery_ParenthesesCondition(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 20, nil, 50, 1)
	insertTestUser(t, "b", 35, nil, 60, 1)
	insertTestUser(t, "c", 40, nil, 70, 0)

	count := 0
	err := NewQuery(testDB).
		Table(testTable).
		Where(
			WithEq("status", 1),
			WithParentheses(
				WithEq("age", 20),
				WithGt("age", 30),
			).SetPrefix(OR),
		).
		Query(func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		})
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// ---------------------------------------------------------------------------
// Insert builder tests
// ---------------------------------------------------------------------------

func TestNewInsert_NoTable(t *testing.T) {
	_, err := NewInsert(testDB).Insert(map[string]interface{}{"name": "x"})
	if err == nil || !strings.Contains(err.Error(), "table is invalid") {
		t.Errorf("expected 'table is invalid' error, got %v", err)
	}
}

func TestNewInsert_EmptyValue(t *testing.T) {
	_, err := NewInsert(testDB).Table(testTable).Insert(map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "value is invalid") {
		t.Errorf("expected 'value is invalid' error, got %v", err)
	}
}

func TestNewInsert(t *testing.T) {
	truncateTestTable(t)
	n, err := NewInsert(testDB).Table(testTable).Insert(map[string]interface{}{
		"name": "dave", "age": 25, "email": "dave@t.com", "score": 88.5, "status": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}
}

func TestNewInsertBatch(t *testing.T) {
	truncateTestTable(t)
	n, err := NewInsert(testDB).Table(testTable).InsertBatch(
		[]string{"name", "age", "status"},
		[]map[string]interface{}{
			{"name": "eve", "age": 20, "status": 1},
			{"name": "fay", "age": 22, "status": 0},
		},
	)
	if err != nil {
		t.Fatalf("InsertBatch failed: %v", err)
	}
	if n != 2 {
		t.Errorf("RowsAffected = %d, want 2", n)
	}
}

func TestNewInsertBatch_NoColumns(t *testing.T) {
	_, err := NewInsert(testDB).Table(testTable).InsertBatch([]string{}, []map[string]interface{}{{"a": 1}})
	if err == nil || !strings.Contains(err.Error(), "columns is invalid") {
		t.Errorf("expected 'columns is invalid' error, got %v", err)
	}
}

func TestNewInsertBatch_NoValues(t *testing.T) {
	_, err := NewInsert(testDB).Table(testTable).InsertBatch([]string{"name"}, []map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "values is invalid") {
		t.Errorf("expected 'values is invalid' error, got %v", err)
	}
}

func TestNewInsertBatch_MissingColumn(t *testing.T) {
	_, err := NewInsert(testDB).Table(testTable).InsertBatch(
		[]string{"name", "missing"},
		[]map[string]interface{}{{"name": "x"}},
	)
	if err == nil || !strings.Contains(err.Error(), "not found column") {
		t.Errorf("expected 'not found column' error, got %v", err)
	}
}

func TestNewInsertBatchCtx(t *testing.T) {
	truncateTestTable(t)
	ctx := context.Background()
	n, err := NewInsert(testDB).Table(testTable).InsertBatchCtx(ctx,
		[]string{"name", "age"},
		[]map[string]interface{}{
			{"name": "g", "age": 18},
		},
	)
	if err != nil {
		t.Fatalf("InsertBatchCtx failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}
}

// ---------------------------------------------------------------------------
// Update builder tests
// ---------------------------------------------------------------------------

func TestNewUpdate_NoTable(t *testing.T) {
	_, err := NewUpdate(testDB).Field(map[string]interface{}{"name": "x"}).Update()
	if err == nil || !strings.Contains(err.Error(), "table is invalid") {
		t.Errorf("expected 'table is invalid' error, got %v", err)
	}
}

func TestNewUpdate_NoFields(t *testing.T) {
	_, err := NewUpdate(testDB).Table(testTable).Update()
	if err == nil || !strings.Contains(err.Error(), "fields is invalid") {
		t.Errorf("expected 'fields is invalid' error, got %v", err)
	}
}

func TestNewUpdate(t *testing.T) {
	truncateTestTable(t)
	id := insertTestUser(t, "update_me", 10, "old@t.com", 50, 0)

	n, err := NewUpdate(testDB).
		Table(testTable).
		Field(map[string]interface{}{"name": "updated", "age": 99}).
		Where(WithEq("id", id)).
		Update()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}

	var name string
	var age int
	testDB.QueryRow(fmt.Sprintf("SELECT name, age FROM %s WHERE id = ?", testTable), id).Scan(&name, &age)
	if name != "updated" || age != 99 {
		t.Errorf("got name=%s age=%d, want updated 99", name, age)
	}
}

// ---------------------------------------------------------------------------
// Delete builder tests
// ---------------------------------------------------------------------------

func TestNewDelete_NoTable(t *testing.T) {
	_, err := NewDelete(testDB).Delete()
	if err == nil || !strings.Contains(err.Error(), "table is invalid") {
		t.Errorf("expected 'table is invalid' error, got %v", err)
	}
}

func TestNewDelete_All(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "d1", 1, nil, 0, 1)
	insertTestUser(t, "d2", 2, nil, 0, 1)
	n, err := NewDelete(testDB).Table(testTable).Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if n != 2 {
		t.Errorf("RowsAffected = %d, want 2", n)
	}
}

func TestNewDelete_WithWhere(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "keep", 10, nil, 0, 1)
	insertTestUser(t, "del", 20, nil, 0, 0)

	n, err := NewDelete(testDB).
		Table(testTable).
		Where(WithEq("status", 0)).
		Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}

	var count int
	testDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", testTable)).Scan(&count)
	if count != 1 {
		t.Errorf("remaining = %d, want 1", count)
	}
}

func TestNewDelete_ChainedWhere(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 10, nil, 0, 0)
	insertTestUser(t, "b", 20, nil, 0, 1)

	n, err := NewDelete(testDB).
		Table(testTable).
		Where(WithEq("status", 0)).
		Where(WithGt("age", 15)).
		Delete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if n != 0 {
		t.Errorf("RowsAffected = %d, want 0 (no row matches both conditions)", n)
	}
}

// ---------------------------------------------------------------------------
// Raw API: TableWithInsertOne / TableWithInsertMany
// ---------------------------------------------------------------------------

func TestTableWithInsertOne(t *testing.T) {
	truncateTestTable(t)
	ctx := context.Background()
	n, err := TableWithInsertOne(ctx, testDB, testTable, map[string]interface{}{
		"name": "raw1", "age": 11, "status": 1,
	})
	if err != nil {
		t.Fatalf("TableWithInsertOne failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}
}

func TestTableWithInsertOne_EmptyTable(t *testing.T) {
	_, err := TableWithInsertOne(context.Background(), testDB, "", map[string]interface{}{"a": 1})
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithInsertOne_EmptyValue(t *testing.T) {
	_, err := TableWithInsertOne(context.Background(), testDB, testTable, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("expected 'invalid value' error, got %v", err)
	}
}

func TestTableWithInsertMany(t *testing.T) {
	truncateTestTable(t)
	ctx := context.Background()
	n, err := TableWithInsertMany(ctx, testDB, testTable,
		[]string{"name", "age", "status"},
		[]map[string]interface{}{
			{"name": "m1", "age": 1, "status": 1},
			{"name": "m2", "age": 2, "status": 0},
		},
	)
	if err != nil {
		t.Fatalf("TableWithInsertMany failed: %v", err)
	}
	if n != 2 {
		t.Errorf("RowsAffected = %d, want 2", n)
	}
}

func TestTableWithInsertMany_EmptyTable(t *testing.T) {
	_, err := TableWithInsertMany(context.Background(), testDB, "", []string{"a"}, []map[string]interface{}{{"a": 1}})
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithInsertMany_EmptyColumns(t *testing.T) {
	_, err := TableWithInsertMany(context.Background(), testDB, testTable, []string{}, []map[string]interface{}{{"a": 1}})
	if err == nil || !strings.Contains(err.Error(), "invalid columns") {
		t.Errorf("expected 'invalid columns' error, got %v", err)
	}
}

func TestTableWithInsertMany_EmptyValues(t *testing.T) {
	_, err := TableWithInsertMany(context.Background(), testDB, testTable, []string{"a"}, []map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "invalid values") {
		t.Errorf("expected 'invalid values' error, got %v", err)
	}
}

func TestTableWithInsertMany_MissingColumn(t *testing.T) {
	_, err := TableWithInsertMany(context.Background(), testDB, testTable,
		[]string{"name", "missing"},
		[]map[string]interface{}{{"name": "x"}},
	)
	if err == nil || !strings.Contains(err.Error(), "not found column") {
		t.Errorf("expected 'not found column' error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Raw API: TableWithSelectOne / TableWithSelectMany / TableWithSelectRange
// ---------------------------------------------------------------------------

func TestTableWithSelectOne(t *testing.T) {
	truncateTestTable(t)
	id := insertTestUser(t, "select_one", 50, "sel@t.com", 100, 1)
	ctx := context.Background()

	err := TableWithSelectOne(ctx, testDB, testTable,
		[]string{"name", "age"},
		map[string]interface{}{"id": id},
		func(row *sql.Row) error {
			var name string
			var age int
			if err := row.Scan(&name, &age); err != nil {
				return err
			}
			if name != "select_one" || age != 50 {
				t.Errorf("got %s/%d, want select_one/50", name, age)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("TableWithSelectOne failed: %v", err)
	}
}

func TestTableWithSelectOne_EmptyTable(t *testing.T) {
	err := TableWithSelectOne(context.Background(), testDB, "", []string{"a"}, map[string]interface{}{"a": 1}, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithSelectOne_EmptyColumn(t *testing.T) {
	err := TableWithSelectOne(context.Background(), testDB, testTable, []string{}, map[string]interface{}{"a": 1}, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid column") {
		t.Errorf("expected 'invalid column' error, got %v", err)
	}
}

func TestTableWithSelectOne_EmptyWhere(t *testing.T) {
	err := TableWithSelectOne(context.Background(), testDB, testTable, []string{"a"}, map[string]interface{}{}, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid where") {
		t.Errorf("expected 'invalid where' error, got %v", err)
	}
}

func TestTableWithSelectMany(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "sma", 10, nil, 10, 1)
	insertTestUser(t, "smb", 20, nil, 20, 0)
	ctx := context.Background()

	var names []string
	err := TableWithSelectMany(ctx, testDB, testTable,
		[]string{"name"},
		map[string]interface{}{"status": 1},
		map[string]string{"name": "ASC"},
		func(rows *sql.Rows) error {
			for rows.Next() {
				var n string
				if err := rows.Scan(&n); err != nil {
					return err
				}
				names = append(names, n)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("TableWithSelectMany failed: %v", err)
	}
	if len(names) != 1 || names[0] != "sma" {
		t.Errorf("names = %v, want [sma]", names)
	}
}

func TestTableWithSelectMany_EmptyTable(t *testing.T) {
	err := TableWithSelectMany(context.Background(), testDB, "", []string{"a"}, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithSelectMany_EmptyColumn(t *testing.T) {
	err := TableWithSelectMany(context.Background(), testDB, testTable, []string{}, nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid column") {
		t.Errorf("expected 'invalid column' error, got %v", err)
	}
}

func TestTableWithSelectMany_NoWhere(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "a", 1, nil, 0, 1)
	insertTestUser(t, "b", 2, nil, 0, 1)
	ctx := context.Background()

	count := 0
	err := TableWithSelectMany(ctx, testDB, testTable,
		[]string{"id"}, nil, nil,
		func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("TableWithSelectMany(no where) failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestTableWithSelectRange(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "r1", 1, nil, 0, 1)
	insertTestUser(t, "r2", 2, nil, 0, 1)
	insertTestUser(t, "r3", 3, nil, 0, 1)
	ctx := context.Background()

	names := make([]string, 0)
	err := TableWithSelectRange(ctx, testDB, testTable,
		[]string{"name"},
		map[string]interface{}{"status": 1},
		map[string]string{"id": "ASC"},
		1, 2,
		func(rows *sql.Rows) error {
			for rows.Next() {
				var n string
				if err := rows.Scan(&n); err != nil {
					return err
				}
				names = append(names, n)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("TableWithSelectRange failed: %v", err)
	}
	if len(names) != 2 || names[0] != "r2" || names[1] != "r3" {
		t.Errorf("names = %v, want [r2 r3]", names)
	}
}

func TestTableWithSelectRange_EmptyTable(t *testing.T) {
	err := TableWithSelectRange(context.Background(), testDB, "", []string{"a"}, nil, nil, 0, 10, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithSelectRange_EmptyColumn(t *testing.T) {
	err := TableWithSelectRange(context.Background(), testDB, testTable, []string{}, nil, nil, 0, 10, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid column") {
		t.Errorf("expected 'invalid column' error, got %v", err)
	}
}

func TestTableWithSelectRange_NoWhere(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "z1", 1, nil, 0, 1)
	insertTestUser(t, "z2", 2, nil, 0, 1)
	ctx := context.Background()

	count := 0
	err := TableWithSelectRange(ctx, testDB, testTable,
		[]string{"id"}, nil, nil, 0, 10,
		func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("TableWithSelectRange(no where) failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// ---------------------------------------------------------------------------
// Raw API: TableWithCount
// ---------------------------------------------------------------------------

func TestTableWithCount(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "c1", 1, nil, 0, 1)
	insertTestUser(t, "c2", 2, nil, 0, 0)
	insertTestUser(t, "c3", 3, nil, 0, 1)
	ctx := context.Background()

	total, err := TableWithCount(ctx, testDB, testTable, nil)
	if err != nil {
		t.Fatalf("TableWithCount(no where) failed: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}

	active, err := TableWithCount(ctx, testDB, testTable, map[string]interface{}{"status": 1})
	if err != nil {
		t.Fatalf("TableWithCount(with where) failed: %v", err)
	}
	if active != 2 {
		t.Errorf("active = %d, want 2", active)
	}
}

func TestTableWithCount_EmptyTable(t *testing.T) {
	_, err := TableWithCount(context.Background(), testDB, "", nil)
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Raw API: TableWithUpdate / TableWithDelete
// ---------------------------------------------------------------------------

func TestTableWithUpdate(t *testing.T) {
	truncateTestTable(t)
	id := insertTestUser(t, "raw_update", 5, "old@raw.com", 10, 0)
	ctx := context.Background()

	n, err := TableWithUpdate(ctx, testDB, testTable,
		map[string]interface{}{"name": "updated_raw", "score": 99.9},
		map[string]interface{}{"id": id},
	)
	if err != nil {
		t.Fatalf("TableWithUpdate failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}
}

func TestTableWithUpdate_EmptyTable(t *testing.T) {
	_, err := TableWithUpdate(context.Background(), testDB, "", map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2})
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithUpdate_EmptyValue(t *testing.T) {
	_, err := TableWithUpdate(context.Background(), testDB, testTable, map[string]interface{}{}, map[string]interface{}{"id": 1})
	if err == nil || !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("expected 'invalid value' error, got %v", err)
	}
}

func TestTableWithUpdate_EmptyWhere(t *testing.T) {
	_, err := TableWithUpdate(context.Background(), testDB, testTable, map[string]interface{}{"a": 1}, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "invalid where") {
		t.Errorf("expected 'invalid where' error, got %v", err)
	}
}

func TestTableWithDelete(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "del1", 1, nil, 0, 1)
	insertTestUser(t, "del2", 2, nil, 0, 0)
	ctx := context.Background()

	n, err := TableWithDelete(ctx, testDB, testTable, map[string]interface{}{"status": 0})
	if err != nil {
		t.Fatalf("TableWithDelete failed: %v", err)
	}
	if n != 1 {
		t.Errorf("RowsAffected = %d, want 1", n)
	}
}

func TestTableWithDelete_EmptyTable(t *testing.T) {
	_, err := TableWithDelete(context.Background(), testDB, "", map[string]interface{}{"a": 1})
	if err == nil || !strings.Contains(err.Error(), "invalid table") {
		t.Errorf("expected 'invalid table' error, got %v", err)
	}
}

func TestTableWithDelete_EmptyWhere(t *testing.T) {
	_, err := TableWithDelete(context.Background(), testDB, testTable, map[string]interface{}{})
	if err == nil || !strings.Contains(err.Error(), "invalid where") {
		t.Errorf("expected 'invalid where' error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Raw API: ExecQueryOne / ExecQueryMany
// ---------------------------------------------------------------------------

func TestExecQueryOne(t *testing.T) {
	truncateTestTable(t)
	id := insertTestUser(t, "exec1", 77, "exec@t.com", 50, 1)
	ctx := context.Background()

	err := ExecQueryOne(ctx, testDB,
		fmt.Sprintf("SELECT name FROM %s WHERE id = ?", testTable),
		[]interface{}{id},
		func(row *sql.Row) error {
			var name string
			return row.Scan(&name)
		},
	)
	if err != nil {
		t.Fatalf("ExecQueryOne failed: %v", err)
	}
}

func TestExecQueryMany(t *testing.T) {
	truncateTestTable(t)
	insertTestUser(t, "em1", 1, nil, 0, 1)
	insertTestUser(t, "em2", 2, nil, 0, 1)
	ctx := context.Background()

	count := 0
	err := ExecQueryMany(ctx, testDB,
		fmt.Sprintf("SELECT id FROM %s WHERE status = ?", testTable),
		[]interface{}{1},
		func(rows *sql.Rows) error {
			for rows.Next() {
				count++
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("ExecQueryMany failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// ---------------------------------------------------------------------------
// Migration utility tests (pure unit, no DB needed)
// ---------------------------------------------------------------------------

func TestFileChecksum(t *testing.T) {
	got := fileChecksum([]byte("hello"))
	if got != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("fileChecksum('hello') = %s, want known sha256", got)
	}

	got2 := fileChecksum([]byte(""))
	if got2 != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("fileChecksum('') = %s, want empty sha256", got2)
	}
}

func TestSplitStatements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single statement",
			input: "SELECT 1",
			want:  []string{"SELECT 1"},
		},
		{
			name:  "multiple statements",
			input: "SELECT 1; SELECT 2",
			want:  []string{"SELECT 1", "SELECT 2"},
		},
		{
			name:  "with trailing semicolon",
			input: "SELECT 1;",
			want:  []string{"SELECT 1"},
		},
		{
			name:  "with comments",
			input: "-- comment\nSELECT 1;\n# another comment\nSELECT 2",
			want:  []string{"-- comment\nSELECT 1", "# another comment\nSELECT 2"},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "only whitespace",
			input: "  ;  ;  ",
			want:  nil,
		},
		{
			name:  "comment only block",
			input: "-- just a comment",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitStatements(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d; got %v", len(got), len(tt.want), got)
				return
			}
			for i := range got {
				if strings.TrimSpace(got[i]) != strings.TrimSpace(tt.want[i]) {
					t.Errorf("stmts[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSplitStatements_LeadingComment(t *testing.T) {
	input := "-- header comment\nSELECT 1;\n-- middle comment\nSELECT 2"
	got := splitStatements(input)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2; got %v", len(got), got)
	}
	if !strings.Contains(got[0], "SELECT 1") {
		t.Errorf("stmts[0] = %q, should contain SELECT 1", got[0])
	}
}

// ---------------------------------------------------------------------------
// Migration integration test
// ---------------------------------------------------------------------------

func TestMigrate(t *testing.T) {
	dir := t.TempDir()

	file1 := filepath.Join(dir, "001_init.sql")
	if err := os.WriteFile(file1, []byte("CREATE TABLE IF NOT EXISTS migrate_test_1 (id INT);"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}

	file2 := filepath.Join(dir, "002_add.sql")
	if err := os.WriteFile(file2, []byte("CREATE TABLE IF NOT EXISTS migrate_test_2 (id INT);"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}

	ctx := context.Background()
	defer func() {
		testDB.Exec("DROP TABLE IF EXISTS migrate_test_1")
		testDB.Exec("DROP TABLE IF EXISTS migrate_test_2")
		testDB.Exec("DROP TABLE IF EXISTS schema_migrations")
	}()

	if err := Migrate(ctx, testDSN, dir); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if count != 2 {
		t.Errorf("migration records = %d, want 2", count)
	}

	var exists int
	testDB.QueryRow("SELECT COUNT(*) FROM migrate_test_1").Scan(&exists)
	if exists != 0 {
		t.Errorf("migrate_test_1 exists = %d, want 0", exists)
	}
	testDB.QueryRow("SELECT COUNT(*) FROM migrate_test_2").Scan(&exists)
	if exists != 0 {
		t.Errorf("migrate_test_2 exists = %d, want 0", exists)
	}

	if err := Migrate(ctx, testDSN, dir); err != nil {
		t.Fatalf("Migrate (second run) failed: %v", err)
	}
}

func TestMigrate_WithNonSQLFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("not sql"), 0o644); err != nil {
		t.Fatalf("write note.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "init.sql"), []byte("SELECT 1"), 0o644); err != nil {
		t.Fatalf("write init.sql: %v", err)
	}

	ctx := context.Background()
	defer func() {
		testDB.Exec("DROP TABLE IF EXISTS schema_migrations")
	}()

	if err := Migrate(ctx, testDSN, dir); err != nil {
		t.Fatalf("Migrate with non-sql files failed: %v", err)
	}
}

func TestMigrate_InvalidDir(t *testing.T) {
	ctx := context.Background()
	err := Migrate(ctx, testDSN, "/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid dir, got nil")
	}
}

func TestStripDBName(t *testing.T) {
	tests := []struct {
		dsn      string
		wantBase string
		wantName string
	}{
		{
			dsn:      "root:root@tcp(127.0.0.1:3306)/personal_auth?charset=utf8mb4",
			wantBase: "root:root@tcp(127.0.0.1:3306)/?charset=utf8mb4",
			wantName: "personal_auth",
		},
		{
			dsn:      "root:root@tcp(127.0.0.1:3306)/personal_auth",
			wantBase: "root:root@tcp(127.0.0.1:3306)/",
			wantName: "personal_auth",
		},
		{
			dsn:      "root:root@tcp(127.0.0.1:3306)/",
			wantBase: "root:root@tcp(127.0.0.1:3306)/",
			wantName: "",
		},
		{
			dsn:      "no-slash-dsn",
			wantBase: "no-slash-dsn",
			wantName: "",
		},
	}
	for _, tt := range tests {
		base, name := stripDBName(tt.dsn)
		if base != tt.wantBase {
			t.Errorf("stripDBName(%q) base = %q, want %q", tt.dsn, base, tt.wantBase)
		}
		if name != tt.wantName {
			t.Errorf("stripDBName(%q) name = %q, want %q", tt.dsn, name, tt.wantName)
		}
	}
}
