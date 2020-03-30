package errors

import (
	"io"
	"fmt"
	"net/http"
	"encoding/json"
)
import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// converts a HTTP error code to a gRPC code
func GrpcFromHttp(httpCode int) codes.Code {

	switch httpCode {

	case http.StatusOK:
		// OK is returned on success.
		return codes.OK

	case http.StatusGone:
		// Canceled indicates the operation was canceled (typically by the caller).
		return codes.Canceled

	case http.StatusBadRequest:
		// InvalidArgument indicates client specified an invalid argument.
		// Note that this differs from FailedPrecondition. It indicates arguments
		// that are problematic regardless of the state of the system
		// (e.g., a malformed file name).
		return codes.InvalidArgument

	case http.StatusRequestTimeout:
		// DeadlineExceeded means operation expired before completion.
		// For operations that change the state of the system, this error may be
		// returned even if the operation has completed successfully. For
		// example, a successful response from a server could have been delayed
		// long enough for the deadline to expire.
		return codes.DeadlineExceeded

	case http.StatusNotFound:
		// NotFound means some requested entity (e.g., file or directory) was
		// not found.
		return codes.NotFound

	case http.StatusConflict:
		// AlreadyExists means an attempt to create an entity failed because one
		// already exists.
		return codes.AlreadyExists

	case http.StatusForbidden:
		// PermissionDenied indicates the caller does not have permission to
		// execute the specified operation. It must not be used for rejections
		// caused by exhausting some resource (use ResourceExhausted
		// instead for those errors). It must not be
		// used if the caller cannot be identified (use Unauthenticated
		// instead for those errors).
		return codes.PermissionDenied

	case http.StatusInsufficientStorage:
		// ResourceExhausted indicates some resource has been exhausted, perhaps
		// a per-user quota, or perhaps the entire file system is out of space.
		return codes.ResourceExhausted

	case http.StatusPreconditionFailed:
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
		return codes.FailedPrecondition

	case http.StatusResetContent:
		// Aborted indicates the operation was aborted, typically due to a
		// concurrency issue like sequencer check failures, transaction aborts,
		// etc.
		//
		// See litmus test above for deciding between FailedPrecondition,
		// Aborted, and Unavailable.
		return codes.Aborted

	case http.StatusRequestedRangeNotSatisfiable:
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
		return codes.OutOfRange

	case http.StatusNotImplemented:
		// Unimplemented indicates operation is not implemented or not
		// supported/enabled in this service.
		return codes.Unimplemented

	case http.StatusInternalServerError:
		// Internal errors. Means some invariants expected by underlying
		// system has been broken. If you see one of these errors,
		// something is very broken.
		return codes.Internal

	case http.StatusServiceUnavailable:
		// Unavailable indicates the service is currently unavailable.
		// This is a most likely a transient condition and may be corrected
		// by retrying with a backoff. Note that it is not always safe to retry
		// non-idempotent operations.
		//
		// See litmus test above for deciding between FailedPrecondition,
		// Aborted, and Unavailable.
		return codes.Unavailable

	case http.StatusTeapot:
		// DataLoss indicates unrecoverable data loss or corruption.
		return codes.DataLoss

	case http.StatusUnauthorized:
		// Unauthenticated indicates the request does not have valid
		// authentication credentials for the operation.
		return codes.Unauthenticated

	case http.StatusUnprocessableEntity:
		return codes.Unknown

	}

	return codes.Unknown
}

// FromGrpcError takes a grpc error passed in, and gets the status
// of it to create a WithGrpcStatus error type
func FromGrpcError(origErr error) error {
	if origErr == nil {
		return nil
	}
	// sanity check this is actually a GRPCError
	st, valid := status.FromError(origErr); 
	if valid {
		if st.Code() == codes.OK {
				return nil
			}
			for _, detail := range st.Details() {
				switch t := detail.(type) {
				case *errdetails.DebugInfo:
					if err := json.Unmarshal([]byte(t.Detail), origErr); err == nil {
						return nil
					}
				}
			}
			err := &withGrpcStatus{
				cause:      origErr,
				grpcCode: st.Code(),
				grpcStatus: st,
			}
			return &withStack{
				err,
				callers(),
			}
	}
	return nil
}

// WithGrpcStatus annotates err with the grpc code and a status.
// If err is nil, WithMessage returns nil.
func WithGrpcStatus(err error, code codes.Code) error {
	if err == nil {
		return nil
	}
	return &withGrpcStatus{
		cause:      err,
		grpcCode: code,
		grpcStatus: status.New(code, err.Error()),
	}
}

type withGrpcStatus struct {
	cause      error
	grpcCode codes.Code
	grpcStatus *status.Status
}

func (w *withGrpcStatus) Error() string {
	return fmt.Sprintf("rpc error: code = %s desc = %s", w.grpcCode, w.grpcStatus.Message())
}

func (w *withGrpcStatus) ErrorCode() uint32 {
	return uint32(w.grpcCode)
}

func (w *withGrpcStatus) String() string {
	return w.grpcStatus.Message()
}

func (w *withGrpcStatus) Cause() error {
	return w.cause
}

func (w *withGrpcStatus) Unwrap() error {
	return w.cause
}

func (w *withGrpcStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", w.Cause())
			io.WriteString(s, w.String())
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, w.Error())
	}
}
