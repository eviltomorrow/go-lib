package sqlutil

import (
	"fmt"
)

func WithLt(expr string, value interface{}) Condition {
	return &lt{prefix: AND, expr: expr, value: value}
}

type lt struct {
	prefix Type
	expr   string
	value  interface{}
}

func (c *lt) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s%s < ?", c.prefix.String(), c.expr), []interface{}{c.value}
}

func (c *lt) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
