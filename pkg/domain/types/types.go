package types

type Schema string

func (x Schema) ToQuery() string {
	return "data.msg." + string(x)
}
