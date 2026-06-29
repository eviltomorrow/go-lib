package sqlutil

import "fmt"

type order interface {
	SQL() (string, []interface{})
}

type orderHandler struct {
	suffix string
	expr   string
}

func (c *orderHandler) SQL() (string, []interface{}) {
	return fmt.Sprintf("%s %s", c.expr, c.suffix), nil
}

func DESC(name string) order {
	return &orderHandler{suffix: "DESC", expr: name}
}

func ASC(name string) order {
	return &orderHandler{suffix: "ASC", expr: name}
}
