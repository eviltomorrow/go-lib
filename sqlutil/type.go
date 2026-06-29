package sqlutil

type Type int

func (t Type) String() string {
	switch t {
	case AND:
		return " AND "
	case OR:
		return " OR "
	default:
		return " AND "
	}
}

const (
	AND Type = iota
	OR
)
