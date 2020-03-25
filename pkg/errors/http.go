package errors

import "io"
import "fmt"
import "strconv"
import "net/http"
import "encoding/json"
import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// converts a gRPC error code to a HTTP code
func HttpFromGrpc(grpcCode codes.Code) int {

	switch grpcCode {

	case codes.OK:
		return 0
	// OK is returned on success.

	case codes.Canceled:
		// Canceled indicates the operation was canceled (typically by the caller).
		return http.StatusGone

	case codes.InvalidArgument:
		// InvalidArgument indicates client specified an invalid argument.
		// Note that this differs from FailedPrecondition. It indicates arguments
		// that are problematic regardless of the state of the system
		// (e.g., a malformed file name).
		return http.StatusBadRequest

	case codes.DeadlineExceeded:
		// DeadlineExceeded means operation expired before completion.
		// For operations that change the state of the system, this error may be
		// returned even if the operation has completed successfully. For
		// example, a successful response from a server could have been delayed
		// long enough for the deadline to expire.
		return http.StatusRequestTimeout

	case codes.NotFound:
		// NotFound means some requested entity (e.g., file or directory) was
		// not found.
		return http.StatusNotFound

	case codes.AlreadyExists:
		// AlreadyExists means an attempt to create an entity failed because one
		// already exists.
		return http.StatusConflict

	case codes.PermissionDenied:
		// PermissionDenied indicates the caller does not have permission to
		// execute the specified operation. It must not be used for rejections
		// caused by exhausting some resource (use ResourceExhausted
		// instead for those errors). It must not be
		// used if the caller cannot be identified (use Unauthenticated
		// instead for those errors).
		return http.StatusForbidden

	case codes.ResourceExhausted:
		// ResourceExhausted indicates some resource has been exhausted, perhaps
		// a per-user quota, or perhaps the entire file system is out of space.
		return http.StatusInsufficientStorage

	case codes.FailedPrecondition:
		// FailedPrecondition indicates operation was rejected because the
		// system is not in a state required for the operation's execution.
		// For example, directory to be deleted may be non-empty, an rmdir
		// operation is applied to a non-directory, etc.
		//
		// A litmus test that may help a service implementor in deciding
		// between FailedPrecondition, Aborted, and Unavailable:
		//  (a) Use Unavailable if the client can retry just the failing call.
		//  (b) Use Aborted if the client should retry at a higher-level
		//      (e.g., restarting a read-modify-write sequence).
		//  (c) Use FailedPrecondition if the client should not retry until
		//      the system state has been explicitly fixed. E.g., if an "rmdir"
		//      fails because the directory is non-empty, FailedPrecondition
		//      should be returned since the client should not retry unless
		//      they have first fixed up the directory by deleting files from it.
		//  (d) Use FailedPrecondition if the client performs conditional
		//      REST Get/Update/Delete on a resource and the resource on the
		//      server does not match the condition. E.g., conflicting
		//      read-modify-write on the same resource.
		return http.StatusPreconditionFailed

	case codes.Aborted:
		// Aborted indicates the operation was aborted, typically due to a
		// concurrency issue like sequencer check failures, transaction aborts,
		// etc.
		//
		// See litmus test above for deciding between FailedPrecondition,
		// Aborted, and Unavailable.
		return http.StatusResetContent

	case codes.OutOfRange:
		// OutOfRange means operation was attempted past the valid range.
		// E.g., seeking or reading past end of file.
		//
		// Unlike InvalidArgument, this error indicates a problem that may
		// be fixed if the system state changes. For example, a 32-bit file
		// system will generate InvalidArgument if asked to read at an
		// offset that is not in the range [0,2^32-1], but it will generate
		// OutOfRange if asked to read from an offset past the current
		// file size.
		//
		// There is a fair bit of overlap between FailedPrecondition and
		// OutOfRange. We recommend using OutOfRange (the more specific
		// error) when it applies so that callers who are iterating through
		// a space can easily look for an OutOfRange error to detect when
		// they are done.
		return http.StatusRequestedRangeNotSatisfiable

	case codes.Unimplemented:
		// Unimplemented indicates operation is not implemented or not
		// supported/enabled in this service.
		return http.StatusNotImplemented

	case codes.Internal:
		// Internal errors. Means some invariants expected by underlying
		// system has been broken. If you see one of these errors,
		// something is very broken.
		return http.StatusInternalServerError

	case codes.Unavailable:
		// Unavailable indicates the service is currently unavailable.
		// This is a most likely a transient condition and may be corrected
		// by retrying with a backoff. Note that it is not always safe to retry
		// non-idempotent operations.
		//
		// See litmus test above for deciding between FailedPrecondition,
		// Aborted, and Unavailable.
		return http.StatusServiceUnavailable

	case codes.DataLoss:
		// DataLoss indicates unrecoverable data loss or corruption.
		return http.StatusTeapot

	case codes.Unauthenticated:
		// Unauthenticated indicates the request does not have valid
		// authentication credentials for the operation.
		return http.StatusUnauthorized

	case codes.Unknown:
		// Unknown error. An example of where this error may be returned is
		// if a Status value received from another address space belongs to
		// an error-space that is not known in this address space. Also
		// errors raised by APIs that do not return enough error information
		// may be converted to this error.
		return http.StatusUnprocessableEntity

	default:
		return http.StatusSeeOther
	}
}

// WithMessage annotates err with a new message.
// If err is nil, WithMessage returns nil.
func WithHttpStatus(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withHttpStatus{
		cause:  err,
		httpStatus:   code
	}
}

type withHttpStatus struct {
	cause  error
	httpStatus   int
}

func (w *withHttpStatus) Error() string {
	return strconv.Itoa(w.httpStatus) + " : " + http.StatusText(w.httpStatus) + " : " + w.cause.Error()
}

func (w *withHttpStatus) Status() int {
	return w.httpStatus
}

func (w *withHttpStatus) StatusText() string {
	return http.StatusText(w.httpStatus)
}

func (w *withHttpStatus) Cause() error {
	return w.cause
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withHttpStatus) Unwrap() error {
	return w.cause
}

func (w *withHttpStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, http.StatusText(w.httpStatus))
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
