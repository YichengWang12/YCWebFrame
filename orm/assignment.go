package orm

type Assignable interface {
	assign()
}

type Assignment struct {
	col string
	val Expression
}

func Assign(col string, val any) Assignment {
	v, ok := val.(Expression)
	if !ok {
		v = value{val: val}
	}
	return Assignment{
		col: col,
		val: v,
	}
}

func (Assignment) assign() {}
