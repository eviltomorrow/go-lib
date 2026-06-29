package sqlutil

import (
	"testing"
)

func TestTypeString(t *testing.T) {
	if got := AND.String(); got != " AND " {
		t.Errorf("AND.String() = %q, want \" AND \"", got)
	}
	if got := OR.String(); got != " OR " {
		t.Errorf("OR.String() = %q, want \" OR \"", got)
	}
	var t2 Type = 99
	if got := t2.String(); got != " AND " {
		t.Errorf("default Type.String() = %q, want \" AND \"", got)
	}
}

func TestConditionEq(t *testing.T) {
	c := WithEq("name", "alice")
	sql, args := c.SQL()
	if sql != " AND name = ?" {
		t.Errorf("SQL() = %q, want \" AND name = ?\"", sql)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Errorf("args = %v, want [alice]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR name = ?" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR name = ?\"", sql)
	}
}

func TestConditionGt(t *testing.T) {
	c := WithGt("age", 18)
	sql, args := c.SQL()
	if sql != " AND age > ?" {
		t.Errorf("SQL() = %q, want \" AND age > ?\"", sql)
	}
	if len(args) != 1 || args[0] != 18 {
		t.Errorf("args = %v, want [18]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR age > ?" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR age > ?\"", sql)
	}
}

func TestConditionLt(t *testing.T) {
	c := WithLt("age", 65)
	sql, args := c.SQL()
	if sql != " AND age < ?" {
		t.Errorf("SQL() = %q, want \" AND age < ?\"", sql)
	}
	if len(args) != 1 || args[0] != 65 {
		t.Errorf("args = %v, want [65]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR age < ?" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR age < ?\"", sql)
	}
}

func TestConditionNotEq(t *testing.T) {
	c := WithNotEq("status", 0)
	sql, args := c.SQL()
	if sql != " AND status != ?" {
		t.Errorf("SQL() = %q, want \" AND status != ?\"", sql)
	}
	if len(args) != 1 || args[0] != 0 {
		t.Errorf("args = %v, want [0]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR status != ?" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR status != ?\"", sql)
	}
}

func TestConditionIn(t *testing.T) {
	c := WithIn("status", []interface{}{1, 2, 3})
	sql, args := c.SQL()
	if sql != " AND status IN (?,?,?)" {
		t.Errorf("SQL() = %q, want \" AND status IN (?,?,?)\"", sql)
	}
	if len(args) != 3 {
		t.Errorf("len(args) = %d, want 3", len(args))
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR status IN (?,?,?)" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR status IN (?,?,?)\"", sql)
	}
}

func TestConditionIn_Single(t *testing.T) {
	c := WithIn("id", []interface{}{1})
	sql, _ := c.SQL()
	if sql != " AND id IN (?)" {
		t.Errorf("SQL() = %q, want \" AND id IN (?)\"", sql)
	}
}

func TestConditionIn_Empty(t *testing.T) {
	c := WithIn("id", []interface{}{})
	sql, args := c.SQL()
	if sql != " AND id IN ()" {
		t.Errorf("SQL() = %q, want \" AND id IN ()\"", sql)
	}
	if len(args) != 0 {
		t.Errorf("len(args) = %d, want 0", len(args))
	}
}

func TestConditionNotIn(t *testing.T) {
	c := WithNotIn("status", []interface{}{0})
	sql, args := c.SQL()
	if sql != " AND status NOT IN (?)" {
		t.Errorf("SQL() = %q, want \" AND status NOT IN (?)\"", sql)
	}
	if len(args) != 1 || args[0] != 0 {
		t.Errorf("args = %v, want [0]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR status NOT IN (?)" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR status NOT IN (?)\"", sql)
	}
}

func TestConditionNull(t *testing.T) {
	c := WithNull("email")
	sql, args := c.SQL()
	if sql != " AND email IS NULL" {
		t.Errorf("SQL() = %q, want \" AND email IS NULL\"", sql)
	}
	if args != nil {
		t.Errorf("args = %v, want nil", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR email IS NULL" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR email IS NULL\"", sql)
	}
}

func TestConditionNotNull(t *testing.T) {
	c := WithNotNull("email")
	sql, args := c.SQL()
	if sql != " AND email IS NOT NULL" {
		t.Errorf("SQL() = %q, want \" AND email IS NOT NULL\"", sql)
	}
	if args != nil {
		t.Errorf("args = %v, want nil", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR email IS NOT NULL" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR email IS NOT NULL\"", sql)
	}
}

func TestConditionBetweenAnd(t *testing.T) {
	c := WithBetweenAnd("age", 18, 60)
	sql, args := c.SQL()
	if sql != " AND age BETWEEN ? AND ?" {
		t.Errorf("SQL() = %q, want \" AND age BETWEEN ? AND ?\"", sql)
	}
	if len(args) != 2 || args[0] != 18 || args[1] != 60 {
		t.Errorf("args = %v, want [18, 60]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	if sql != " OR age BETWEEN ? AND ?" {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want \" OR age BETWEEN ? AND ?\"", sql)
	}
}

func TestConditionParentheses(t *testing.T) {
	c := WithParentheses(
		WithEq("age", 18),
		WithGt("score", 90),
	)
	sql, args := c.SQL()
	want := " AND (age = ? AND score > ?)"
	if sql != want {
		t.Errorf("SQL() = %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != 18 || args[1] != 90 {
		t.Errorf("args = %v, want [18, 90]", args)
	}

	c.SetPrefix(OR)
	sql, _ = c.SQL()
	want = " OR (age = ? AND score > ?)"
	if sql != want {
		t.Errorf("after SetPrefix(OR), SQL() = %q, want %q", sql, want)
	}
}

func TestConditionParentheses_Nested(t *testing.T) {
	c := WithParentheses(
		WithEq("a", 1),
		WithParentheses(
			WithEq("b", 2),
			WithEq("c", 3),
		).SetPrefix(OR),
	)
	sql, args := c.SQL()
	want := " AND (a = ? OR (b = ? AND c = ?))"
	if sql != want {
		t.Errorf("SQL() = %q, want %q", sql, want)
	}
	if len(args) != 3 {
		t.Errorf("len(args) = %d, want 3", len(args))
	}
}

func TestOrderBy(t *testing.T) {
	a := ASC("name")
	sql, args := a.SQL()
	if sql != "name ASC" {
		t.Errorf("ASC SQL() = %q, want \"name ASC\"", sql)
	}
	if args != nil {
		t.Errorf("ASC args = %v, want nil", args)
	}

	d := DESC("age")
	sql, args = d.SQL()
	if sql != "age DESC" {
		t.Errorf("DESC SQL() = %q, want \"age DESC\"", sql)
	}
	if args != nil {
		t.Errorf("DESC args = %v, want nil", args)
	}
}
