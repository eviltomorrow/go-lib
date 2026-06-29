package sqlutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/eviltomorrow/go-lib/db/mysql"
)

var UpdateTimeout = 10 * time.Second

func NewUpdate(exec db.Exec) Update {
	return &update{
		exec:    exec,
		builder: strings.Builder{},
	}
}

type Update interface {
	Update() (int64, error)
	UpdateCtx(ctx context.Context) (int64, error)

	UpdateField
	UpdateTable
	UpdateWhere
}

type UpdateField interface {
	Field(map[string]interface{}) Update
}

type UpdateTable interface {
	Table(name string) Update
}

type UpdateWhere interface {
	Where(...Condition) Update
}

type update struct {
	exec db.Exec

	builder strings.Builder

	fields     map[string]interface{}
	table      string
	conditions []Condition
}

func (h *update) Field(fields map[string]interface{}) Update {
	h.fields = fields
	return h
}

func (h *update) Table(name string) Update {
	h.table = name
	return h
}

func (h *update) Where(conditions ...Condition) Update {
	h.conditions = append(h.conditions, conditions...)
	return h
}

func (h *update) UpdateCtx(ctx context.Context) (int64, error) {
	if h.table == "" {
		return 0, fmt.Errorf("table is invalid")
	}
	if len(h.fields) == 0 {
		return 0, fmt.Errorf("fields is invalid")
	}

	table := h.table
	fields := make([]string, 0, len(h.fields))
	args := make([]interface{}, 0, len(h.fields)+16)

	for k, v := range h.fields {
		fields = append(fields, fmt.Sprintf("%s = ?", k))
		args = append(args, v)
	}

	var where strings.Builder
	for _, condition := range h.conditions {
		c, arg := condition.SQL()
		if _, err := where.WriteString(c); err != nil {
			return 0, err
		}
		args = append(args, arg...)
	}

	sql := fmt.Sprintf("UPDATE %s SET %s", table, strings.Join(fields, ","))
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

func (h *update) Update() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), UpdateTimeout)
	defer cancel()

	return h.UpdateCtx(ctx)
}
