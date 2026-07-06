// Package orkai wraps HTTP calls to the orkai MCP API for onboarding entity
// creation (category, standards, skill) and entity linking.
//
// The client communicates with a local orkai daemon. The MCP URL and auth
// token are configured via environment variables (ORKAI_MCP_URL defaults to
// http://localhost:18787/mcp and ORKAI_MCP_TOKEN is required).
package orkai
