package logging

import "go.uber.org/zap"

// Field provides a wrapping struct to be used internally to log indexable fields.
type Field struct {
	field zap.Field
}

// NewStringField creates an string field used for log indexing.
func NewStringField(k, v string) Field {
	f := Field{}
	s := zap.String(k, v)
	f.field = s

	return f
}

// NewInt64Field creates an int64 field used for log indexing.
func NewInt64Field(k string, v int64) Field {
	f := Field{}
	i := zap.Int64(k, v)
	f.field = i

	return f
}

// NewFloat64Field creates an float64 field used for log indexing.
func NewFloat64Field(k string, v float64) Field {
	f := Field{}
	fl := zap.Float64(k, v)
	f.field = fl

	return f
}

// NewBoolField creates a bool field used for log indexing.
func NewBoolField(k string, v bool) Field {
	f := Field{}
	b := zap.Bool(k, v)
	f.field = b

	return f
}

// NewAnyField takes a key and an arbitrary value and chooses the
// best way to represent them as a field, falling back to a reflection-based
// approach only if necessary.
func NewAnyField(k string, v interface{}) Field {
	f := Field{}
	a := zap.Any(k, v)
	f.field = a

	return f
}

// NewStringsField creates an array of strings field for log indexing
func NewStringsField(k string, vs []string) Field {
	f := Field{}
	ss := zap.Strings(k, vs)
	f.field = ss

	return f
}

// NewInt64sField creates an array of int64s field for log indexing
func NewInt64sField(k string, vs []int64) Field {
	f := Field{}
	is := zap.Int64s(k, vs)
	f.field = is

	return f
}

// NewFloat64sField creates an array of int64s field for log indexing
func NewFloat64sField(k string, vs []float64) Field {
	f := Field{}
	fs := zap.Float64s(k, vs)
	f.field = fs

	return f
}

// NewBoolsField creates an array of bools field used for log indexing.
func NewBoolsField(k string, vs []bool) Field {
	f := Field{}
	bs := zap.Bools(k, vs)
	f.field = bs

	return f
}
