package sqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	db "github.com/eviltomorrow/go-lib/db/mysql"
)

var QueryTimeout = 10 * time.Second

func NewQuery(exec db.Exec) Query {
	return &query{
		exec:    exec,
		builder: strings.Builder{},
		args:    make([]interface{}, 0, 16),
	}
}

type Query interface {
	QueryOne(f func(row *sql.Row) error) error
	Query(f func(rows *sql.Rows) error) error
	QueryOneCtx(ctx context.Context, f func(row *sql.Row) error) error
	QueryCtx(ctx context.Context, f func(rows *sql.Rows) error) error

	QueryColumn
	QueryTable
	QueryWhere
	QueryGroupBy
	QueryOrderBy
	QueryLimit
}

type QueryColumn interface {
	Columns(columns []string) Query
}

type QueryTable interface {
	Table(name string) Query
}

type QueryWhere interface {
	Where(...Condition) Query
}

type QueryGroupBy interface {
	GroupBy([]string) Query
}

type QueryOrderBy interface {
	OrderBy(...order) Query
}

type QueryLimit interface {
	Limit(int64, int64) Query
}

type query struct {
	exec db.Exec

	builder strings.Builder
	args    []interface{}

	table      string
	columns    []string
	conditions []Condition
	groupby    []string
	orderby    []string
	limit      []int64
}

func (h *query) QueryOne(f func(row *sql.Row) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeout)
	defer cancel()

	return h.QueryOneCtx(ctx, f)
}

func (h *query) QueryOneCtx(ctx context.Context, f func(row *sql.Row) error) error {
	if h.table == "" {
		return fmt.Errorf("table is invalid")
	}

	column := "*"
	if len(h.columns) != 0 {
		column = strings.Join(h.columns, ", ")
	}

	table := h.table

	var (
		where strings.Builder
		args  []interface{}
	)
	for _, condition := range h.conditions {
		c, arg := condition.SQL()
		if _, err := where.WriteString(c); err != nil {
			return err
		}
		args = append(args, arg...)
	}

	groupby := ""
	if len(h.groupby) != 0 {
		groupby = strings.Join(h.groupby, ",")
	}

	orderby := ""
	if len(h.orderby) != 0 {
		orderby = strings.Join(h.orderby, ",")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", column, table)
	if where.Len() != 0 {
		w := where.String()

		w = strings.TrimPrefix(w, AND.String())
		w = strings.TrimPrefix(w, OR.String())
		sql = fmt.Sprintf("%s WHERE %s", sql, w)
	}
	if groupby != "" {
		sql = fmt.Sprintf("%s GROUP BY %s", sql, groupby)
	}
	if orderby != "" {
		sql = fmt.Sprintf("%s ORDER BY %s", sql, orderby)
	}
	if len(h.limit) == 2 {
		sql = fmt.Sprintf("%s LIMIT ?, ?", sql)
		args = append(args, h.limit[0], h.limit[1])
	}

	row := h.exec.QueryRowContext(ctx, sql, args...)
	if row.Err() != nil {
		return row.Err()
	}

	if err := f(row); err != nil {
		return err
	}
	return nil
}

func (h *query) QueryCtx(ctx context.Context, f func(row *sql.Rows) error) error {
	if h.table == "" {
		return fmt.Errorf("table is invalid")
	}

	column := "*"
	if len(h.columns) != 0 {
		column = strings.Join(h.columns, ", ")
	}

	table := h.table

	var (
		where strings.Builder
		args  []interface{}
	)
	where.Grow(256)

	for _, condition := range h.conditions {
		c, arg := condition.SQL()
		if _, err := where.WriteString(c); err != nil {
			return err
		}
		args = append(args, arg...)
	}

	groupby := ""
	if len(h.groupby) != 0 {
		groupby = strings.Join(h.groupby, ",")
	}

	orderby := ""
	if len(h.orderby) != 0 {
		orderby = strings.Join(h.orderby, ",")
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", column, table)
	if where.Len() != 0 {
		w := where.String()

		w = strings.TrimPrefix(w, AND.String())
		w = strings.TrimPrefix(w, OR.String())
		sql = fmt.Sprintf("%s WHERE %s", sql, w)
	}
	if groupby != "" {
		sql = fmt.Sprintf("%s GROUP BY %s", sql, groupby)
	}
	if orderby != "" {
		sql = fmt.Sprintf("%s ORDER BY %s", sql, orderby)
	}
	if len(h.limit) == 2 {
		sql = fmt.Sprintf("%s LIMIT ?, ?", sql)
		args = append(args, h.limit[0], h.limit[1])
	}

	rows, err := h.exec.QueryContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := f(rows); err != nil {
		return err
	}
	return nil
}

func (h *query) Query(f func(row *sql.Rows) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), QueryTimeout)
	defer cancel()

	return h.QueryCtx(ctx, f)
}

func (h *query) Columns(columns []string) Query {
	h.columns = columns
	return h
}

func (h *query) Table(name string) Query {
	h.table = name
	return h
}

func (h *query) Where(conditions ...Condition) Query {
	h.conditions = append(h.conditions, conditions...)
	return h
}

func (h *query) GroupBy(gropuBy []string) Query {
	h.groupby = gropuBy
	return h
}

func (h *query) OrderBy(orderBy ...order) Query {
	for _, o := range orderBy {
		text, args := o.SQL()
		h.orderby = append(h.orderby, text)
		h.args = append(h.args, args...)
	}
	return h
}

func (h *query) Limit(offset int64, count int64) Query {
	h.limit = []int64{offset, count}
	return h
}
