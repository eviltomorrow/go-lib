package sqlutil

import (
	"fmt"
	"strings"
)

func WithNotIn(expr string, values []interface{}) Condition {
	return &notIn{prefix: AND, expr: expr, values: values}
}

type notIn struct {
	prefix Type
	expr   string
	values []interface{}
}

func (c *notIn) SQL() (string, []interface{}) {
	args := make([]string, 0, len(c.values))
	for i := 0; i < len(c.values); i++ {
		args = append(args, "?")
	}
	return fmt.Sprintf("%s%s NOT IN (%s)", c.prefix.String(), c.expr, strings.Join(args, ",")), c.values
}

func (c *notIn) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
