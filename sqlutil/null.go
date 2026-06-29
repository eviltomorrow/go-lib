package sqlutil

import (
	"fmt"
)

func WithNull(expr string) Condition {
	return &null{prefix: AND, expr: expr}
}

type null struct {
	prefix Type
	expr   string
}

func (c *null) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s%s IS NULL", c.prefix.String(), c.expr), nil
}

func (c *null) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
