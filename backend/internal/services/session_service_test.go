package services

import (
	"testing"
)

func TestNewSessionService(t *testing.T) {
	t.Parallel()
	svc := NewSessionService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil SessionService")
	}
}
