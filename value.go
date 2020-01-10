package goBolt

type Value interface{}

type ValueConverter interface {
	// ConvertValue converts a value to a driver Value.
	ConvertValue(v interface{}) (Value, error)
}

type Valuer interface {
	// Value returns a driver Value.
	// Value must not panic.
	Value() (Value, error)
}
