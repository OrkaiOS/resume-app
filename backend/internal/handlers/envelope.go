package handlers

// Envelope is the standard API response shape:
// {"data": ..., "error": null} on success, or
// {"data": null, "error": {"code": "...", "message": "..."}} on error.
type Envelope struct {
	Data  any        `json:"data"`
	Error *ErrorBody `json:"error"`
}

// ErrorBody is the error payload carried inside the standard envelope.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success builds an envelope for a successful response.
func Success(data any) Envelope {
	return Envelope{Data: data, Error: nil}
}

// Failure builds an envelope for an error response.
func Failure(code, message string) Envelope {
	return Envelope{Data: nil, Error: &ErrorBody{Code: code, Message: message}}
}
