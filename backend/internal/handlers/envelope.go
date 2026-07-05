package handlers

// Standard machine-readable error codes.
//
// NOT_FOUND, VALIDATION_ERROR, and CONFLICT come from the API Contract
// Standard. INTERNAL is a resume-app addition for 500 responses: the
// standard's code list is oriented to PENTA auth flows (which do not apply
// to this local-first single-user tool) and does not define a 500 code.
const (
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeConflict   = "CONFLICT"
	ErrCodeInternal   = "INTERNAL"
)

// Envelope is the standard API response shape:
// {"data": ..., "error": null} on success, or
// {"data": null, "error": {"code": "...", "message": "...", "details": {}}} on error.
type Envelope struct {
	Data  any        `json:"data"`
	Error *ErrorBody `json:"error"`
}

// ErrorBody is the error payload carried inside the standard envelope.
// Details holds field-level errors for validation issues, e.g.
// {"field": "email", "issue": "invalid format"}; it is an empty object for
// non-validation errors.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

// Success builds an envelope for a successful response.
func Success(data any) Envelope {
	return Envelope{Data: data, Error: nil}
}

// Failure builds an envelope for an error response with no field-level details.
func Failure(code, message string) Envelope {
	return Envelope{Data: nil, Error: &ErrorBody{Code: code, Message: message, Details: map[string]any{}}}
}

// FailureWithDetails builds an error envelope carrying field-level details.
// details is typically a map like {"field": "email", "issue": "invalid format"}.
func FailureWithDetails(code, message string, details any) Envelope {
	if details == nil {
		details = map[string]any{}
	}
	return Envelope{Data: nil, Error: &ErrorBody{Code: code, Message: message, Details: details}}
}
