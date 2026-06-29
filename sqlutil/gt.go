package sqlutil

import (
	"fmt"
)

func WithGt(expr string, value interface{}) Condition {
	return &gt{prefix: AND, expr: expr, value: value}
}

type gt struct {
	prefix Type
	expr   string
	value  interface{}
}

func (c *gt) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s%s > ?", c.prefix.String(), c.expr), []interface{}{c.value}
}

func (c *gt) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
