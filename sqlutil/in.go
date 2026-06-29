package sqlutil

import (
	"fmt"
	"strings"
)

func WithIn(expr string, values []interface{}) Condition {
	return &in{prefix: AND, expr: expr, values: values}
}

type in struct {
	prefix Type
	expr   string
	values []interface{}
}

func (c *in) SQL() (string, []interface{}) {
	args := make([]string, 0, len(c.values))
	for i := 0; i < len(c.values); i++ {
		args = append(args, "?")
	}
	return fmt.Sprintf("%s%s IN (%s)", c.prefix.String(), c.expr, strings.Join(args, ",")), c.values
}

func (c *in) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
