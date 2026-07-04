// Package middleware contains cross-cutting HTTP middleware for the
// resume-app API: panic recovery, request logging, CORS, and Prometheus
// metrics. There is no authentication or authorization middleware —
// resume-app is a local-first single-user tool.
package middleware
