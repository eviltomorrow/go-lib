package sqlutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/eviltomorrow/go-lib/db/mysql"
)

var DeleteTimeout = 10 * time.Second

func NewDelete(exec db.Exec) Delete {
	return &delete{
		exec:    exec,
		builder: strings.Builder{},
	}
}

type Delete interface {
	Delete() (int64, error)
	DeleteCtx(ctx context.Context) (int64, error)

	DeleteTable
	DeleteWhere
}

type DeleteTable interface {
	Table(name string) Delete
}

type DeleteWhere interface {
	Where(...Condition) Delete
}

type delete struct {
	exec db.Exec

	builder strings.Builder
	args    []interface{}

	table      string
	conditions []Condition
}

func (h *delete) Delete() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DeleteTimeout)
	defer cancel()

	return h.DeleteCtx(ctx)
}

func (h *delete) DeleteCtx(ctx context.Context) (int64, error) {
	if h.table == "" {
		return 0, fmt.Errorf("table is invalid")
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
			return 0, err
		}
		args = append(args, arg...)
	}

	sql := fmt.Sprintf("DELETE FROM %s", table)
	if where.Len() != 0 {
		w := where.String()

		w = strings.TrimPrefix(w, AND.String())
		w = strings.TrimPrefix(w, OR.String())
		sql = fmt.Sprintf("%s WHERE %s", sql, w)
	}

	result, err := h.exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (h *delete) Table(name string) Delete {
	h.table = name
	return h
}

func (h *delete) Where(conditions ...Condition) Delete {
	h.conditions = append(h.conditions, conditions...)
	return h
}
