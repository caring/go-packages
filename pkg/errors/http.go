package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
)

// HTTPFromGrpc converts a gRPC error code to a HTTP code
func HTTPFromGrpc(grpcCode codes.Code) int {

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

// ToHTTP writes the error to the http response.
func ToHTTP(in error, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	// If the error was of *a specific type?* we can find
	// a specific satus code to write to the header
	if err, ok := FromGrpcError(in).(*withGrpcStatus); ok {
		w.WriteHeader(HTTPFromGrpc(err.grpcCode))
		return json.NewEncoder(w).Encode(err.grpcCode)
	}
	// If not it's an arbitary error value so use 500.
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(in)
}

// FromHTTP reads an error from the http response.
// this assumes that errors are coming in as JSON
// formatted the way that can be marshaled into a
// WithHttpStatus error type
func FromHTTP(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return nil
	}
	defer resp.Body.Close()
	var err error
	if decErr := json.NewDecoder(resp.Body).Decode(&err); decErr != nil {
		return WithHTTPStatus(decErr, resp.StatusCode)
	}
	return WithHTTPStatus(New("Unknown error"), resp.StatusCode)
}

// WithHTTPStatus annotates an error with an http code
func WithHTTPStatus(err error, code int) error {
	if err == nil {
		return nil
	}
	return &withhttpCode{
		cause:    err,
		httpCode: code,
	}
}

type withhttpCode struct {
	cause    error
	httpCode int
}

func (w *withhttpCode) Error() string {
	return strconv.Itoa(w.httpCode) + " : " + http.StatusText(w.httpCode) + " : " + w.cause.Error()
}

func (w *withhttpCode) ErrorCode() int {
	return w.httpCode
}

func (w *withhttpCode) String() string {
	return http.StatusText(w.httpCode)
}

func (w *withhttpCode) Cause() error {
	return w.cause
}

func (w *withhttpCode) Unwrap() error {
	return w.cause
}

func (w *withhttpCode) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, http.StatusText(w.httpCode))
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
