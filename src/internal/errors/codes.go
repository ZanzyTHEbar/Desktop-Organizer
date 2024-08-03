package errsx

type DesktopCleanerErr int

const (
	// OK indicates the operation was successful.
	OK DesktopCleanerErr = iota
	InvalidData

	// Canceled indicates the operation was canceled (typically by the caller).
	//
	// DesktopCleaner will generate this error code when cancellation is requested.
	Canceled

	// Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	//
	// DesktopCleaner will generate this error code in the above two mentioned cases.
	Unknown

	// InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	//
	// DesktopCleaner will generate this error code if the request data cannot be parsed.
	InvalidArgument

	// DeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	//
	// DesktopCleaner will generate this error code when the deadline is exceeded.
	DeadlineExceeded

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	//
	// DesktopCleaner will not generate this error code.
	NotFound

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	//
	// DesktopCleaner will not generate this error code.
	AlreadyExists

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	//
	// DesktopCleaner will not generate this error code.
	PermissionDenied

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	//
	// DesktopCleaner will generate this error code in out-of-memory and server overload
	// situations, or when a message is larger than the configured maximum size.
	ResourceExhausted

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
	//      Get/Update/Delete on a resource and the resource on the
	//      server does not match the condition. E.g., conflicting
	//      read-modify-write on the same resource.
	//
	// DesktopCleaner will not generate this error code.
	FailedPrecondition

	// Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	Aborted

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
	//
	// DesktopCleaner will not generate this error code.
	OutOfRange

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	//
	// DesktopCleaner will generate this error code when an endpoint does not exist.
	Unimplemented

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	//
	// DesktopCleaner will generate this error code in several internal error conditions.
	Internal

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	//
	// DesktopCleaner will generate this error code in abrupt shutdown of a server process
	// or network connection.
	Unavailable

	// DataLoss indicates unrecoverable data loss or corruption.
	//
	// DesktopCleaner will not generate this error code.
	DataLoss

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	//
	// DesktopCleaner will generate this error code when the authentication metadata
	// is invalid or missing, and expects auth handlers to return errors with
	// this code when the auth token is not valid.
	Unauthenticated
)

// String returns the string representation of c.
func (codes DesktopCleanerErr) String() string {
	return codeNames[codes]
}

// HTTPStatus reports a suitable HTTP status code for an error, based on its code.
// If err is nil it reports 200. If it's not an *Error it reports 500.
func (codes DesktopCleanerErr) HTTPStatus() int {
	return codeStatus[codes]
}

func (codes DesktopCleanerErr) MarshalJSON() ([]byte, error) {
	s := codes.String()
	return []byte("\"" + s + "\""), nil
}

var codeNames = [...]string{
	OK:                 "ok",
	Canceled:           "canceled",
	Unknown:            "unknown",
	InvalidArgument:    "invalid_argument",
	DeadlineExceeded:   "deadline_exceeded",
	NotFound:           "not_found",
	AlreadyExists:      "already_exists",
	PermissionDenied:   "permission_denied",
	ResourceExhausted:  "resource_exhausted",
	FailedPrecondition: "failed_precondition",
	Aborted:            "aborted",
	OutOfRange:         "out_of_range",
	Unimplemented:      "unimplemented",
	Internal:           "internal",
	Unavailable:        "unavailable",
	DataLoss:           "data_loss",
	Unauthenticated:    "unauthenticated",
	InvalidData:        "invalid_data",
}

var codeStatus = [...]int{
	OK:                 200,
	Canceled:           499,
	Unknown:            500,
	InvalidArgument:    400,
	DeadlineExceeded:   504,
	NotFound:           404,
	AlreadyExists:      409,
	PermissionDenied:   403,
	ResourceExhausted:  429,
	FailedPrecondition: 400,
	Aborted:            409,
	OutOfRange:         400,
	Unimplemented:      501,
	Internal:           500,
	Unavailable:        503,
	DataLoss:           500,
	Unauthenticated:    401,
	InvalidData:        422,
}
