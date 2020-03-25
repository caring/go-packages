// +build go1.13

package errors

import (
	stderrs "errors"
)

// Is reports whether any error in err's chain matches target.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// It should be used in preference to simple equality checks:
//
// 	if errors.Is(err, os.ErrExist)
//
// is preferable to
//
// 	if err == os.ErrExist
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func Is(err, target error) bool { return stderrs.Is(err, target) }

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// The form...
//
// 	var perr *os.PathError
// 	if errors.As(err, &perr) {
// 		fmt.Println(perr.Path)
// 	}
//
// ... is preferable to
//
// 	if perr, ok := err.(*os.PathError); ok {
// 		fmt.Println(perr.Path)
// 	}
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(err error, target interface{}) bool { return stderrs.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return stderrs.Unwrap(err)
}
