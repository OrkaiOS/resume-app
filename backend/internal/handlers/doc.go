// Package handlers contains HTTP handlers for the resume-app API.
//
// Each handler parses and validates HTTP requests via Gin, calls the
// matching service, and returns a typed response using the standard
// envelope: {"data": ..., "error": null} or
// {"data": null, "error": {"code": "...", "message": "..."}}.
package handlers
