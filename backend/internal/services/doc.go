// Package services contains the business logic of resume-app.
//
// Services hold the application's use cases: they orchestrate stores,
// the LLM provider, the PDF pipeline, and the RAG layer. Services never
// touch HTTP directly and accept typed domain structs, not Gin context.
package services
