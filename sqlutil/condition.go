package sqlutil

type Condition interface {
	SQL() (string, []interface{})
	SetPrefix(Type) Condition
}
