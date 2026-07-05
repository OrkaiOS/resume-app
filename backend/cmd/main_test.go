package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantServe bool
		wantExit  int
	}{
		{name: "no args defaults to serve", args: nil, wantServe: true, wantExit: 0},
		{name: "empty args defaults to serve", args: []string{}, wantServe: true, wantExit: 0},
		{name: "serve starts server", args: []string{"serve"}, wantServe: true, wantExit: 0},
		{name: "help prints usage", args: []string{"help"}, wantServe: false, wantExit: 0},
		{name: "--help prints usage", args: []string{"--help"}, wantServe: false, wantExit: 0},
		{name: "-h prints usage", args: []string{"-h"}, wantServe: false, wantExit: 0},
		{name: "unknown command errors", args: []string{"foo"}, wantServe: false, wantExit: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			serve, exit := resolveCommand(tt.args)
			if serve != tt.wantServe {
				t.Errorf("serve = %v, want %v", serve, tt.wantServe)
			}
			if exit != tt.wantExit {
				t.Errorf("exit = %v, want %v", exit, tt.wantExit)
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printUsageTo(&buf)
	got := buf.String()

	if !strings.Contains(got, "orkai-resume") {
		t.Error("usage should mention orkai-resume")
	}
	if !strings.Contains(got, "serve") {
		t.Error("usage should list serve command")
	}
	if !strings.Contains(got, "help") {
		t.Error("usage should list help command")
	}
}
