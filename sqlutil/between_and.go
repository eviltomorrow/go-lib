package sqlutil

import "fmt"

func WithBetweenAnd(expr string, value1, value2 interface{}) Condition {
	return &betweenAnd{prefix: AND, expr: expr, values: []interface{}{value1, value2}}
}

type betweenAnd struct {
	prefix Type
	expr   string
	values []interface{}
}

func (c *betweenAnd) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s%s BETWEEN ? AND ?", c.prefix.String(), c.expr), c.values
}

func (c *betweenAnd) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
