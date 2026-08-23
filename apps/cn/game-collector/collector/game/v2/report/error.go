package report

// ErrorKind classifies collector v2 failures for reporting and retry decisions.
type ErrorKind string

const (
	ErrorRateLimited ErrorKind = "rate_limited"
	ErrorBlocked     ErrorKind = "blocked"
	ErrorUpstream    ErrorKind = "upstream_error"
	ErrorDecode      ErrorKind = "decode_error"
	ErrorValidation  ErrorKind = "validation_error"
	ErrorStorage     ErrorKind = "storage_error"
	ErrorCanceled    ErrorKind = "canceled"
	ErrorUnknown     ErrorKind = "unknown"
)

// ErrorInfo stores a sanitized error summary. It must not include secrets or raw payloads.
type ErrorInfo struct {
	Kind    ErrorKind `json:"kind"`
	Message string    `json:"message"`
}
