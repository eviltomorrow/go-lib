package sqlutil

import (
	"fmt"
)

func WithNotNull(expr string) Condition {
	return &notNull{prefix: AND, expr: expr}
}

type notNull struct {
	prefix Type
	expr   string
}

func (c *notNull) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s%s IS NOT NULL", c.prefix.String(), c.expr), nil
}

func (c *notNull) SetPrefix(prefix Type) Condition {
	c.prefix = prefix
	return c
}
