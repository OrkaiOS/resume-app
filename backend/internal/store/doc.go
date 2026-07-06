// Package store contains the data-access layer of resume-app.
//
// Stores persist and retrieve domain models against a local SQLite
// database. Each store is accessed through a small interface so handlers
// and services can be tested with fakes (dependency inversion).
//
// Resources: profiles, opportunities, resumes, cover_letters, artifacts,
// user_settings.
package store
