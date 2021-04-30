package logging

import "go.uber.org/zap"

type DataField interface {
	getField() zap.Field
}

// Field is a typed and structured log entry string key and value pair
type Field struct {
	field zap.Field
}

func (f Field) getField() zap.Field {
	return f.field
}

// String constructs a field with a string value
func String(k, v string) Field {
	f := Field{}
	s := zap.String(k, v)
	f.field = s

	return f
}

// Int64 constructs a field with a int64 value
func Int64(k string, v int64) Field {
	f := Field{}
	i := zap.Int64(k, v)
	f.field = i

	return f
}

// Float64 constructs a field with a float64 value
func Float64(k string, v float64) Field {
	f := Field{}
	fl := zap.Float64(k, v)
	f.field = fl

	return f
}

// Bool constructs a field with a bool value
func Bool(k string, v bool) Field {
	f := Field{}
	b := zap.Bool(k, v)
	f.field = b

	return f
}

// Any takes a key and an arbitrary value and chooses the
// best way to represent them as a field, falling back to a reflection-based
// approach only if necessary.
func Any(k string, v interface{}) Field {
	f := Field{}
	a := zap.Any(k, v)
	f.field = a

	return f
}

// Strings constructs a field with a slice of strings value
func Strings(k string, vs []string) Field {
	f := Field{}
	ss := zap.Strings(k, vs)
	f.field = ss

	return f
}

// Int64s constructs a field with a slice of int64s value
func Int64s(k string, vs []int64) Field {
	f := Field{}
	is := zap.Int64s(k, vs)
	f.field = is

	return f
}

// Float64s constructs a field with a slice of float64s value
func Float64s(k string, vs []float64) Field {
	f := Field{}
	fs := zap.Float64s(k, vs)
	f.field = fs

	return f
}

// Bools constructs a field with a slice of bools value
func Bools(k string, vs []bool) Field {
	f := Field{}
	bs := zap.Bools(k, vs)
	f.field = bs

	return f
}
