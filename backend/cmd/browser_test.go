package main

import (
	"testing"
)

func TestOpenCommandFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		goos     string
		url      string
		wantCmd  string
		wantArgs []string
		wantOK   bool
	}{
		{name: "darwin open", goos: "darwin", url: "http://localhost:8080", wantCmd: "open", wantArgs: []string{"http://localhost:8080"}, wantOK: true},
		{name: "linux xdg-open", goos: "linux", url: "http://localhost:8080", wantCmd: "xdg-open", wantArgs: []string{"http://localhost:8080"}, wantOK: true},
		{name: "windows start", goos: "windows", url: "http://localhost:8080", wantCmd: "cmd", wantArgs: []string{"/c", "start", "http://localhost:8080"}, wantOK: true},
		{name: "unsupported os", goos: "freebsd", url: "http://localhost:8080", wantCmd: "", wantArgs: nil, wantOK: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, args, ok := openCommandFor(tt.goos, tt.url)
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !equalStrings(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestShouldOpenBrowser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		goos string
		env  map[string]string
		want bool
	}{
		{name: "darwin clean env opens", goos: "darwin", env: nil, want: true},
		{name: "darwin SSH_TTY blocks", goos: "darwin", env: map[string]string{"SSH_TTY": "/dev/ttys000"}, want: false},
		{name: "darwin SSH_CONNECTION blocks", goos: "darwin", env: map[string]string{"SSH_CONNECTION": "1.2.3.4 22 5.6.7.8 50000"}, want: false},
		{name: "linux DISPLAY set opens", goos: "linux", env: map[string]string{"DISPLAY": ":0"}, want: true},
		{name: "linux WAYLAND_DISPLAY set opens", goos: "linux", env: map[string]string{"WAYLAND_DISPLAY": "wayland-0"}, want: true},
		{name: "linux no DISPLAY no WAYLAND blocks", goos: "linux", env: nil, want: false},
		{name: "linux SSH_TTY blocks even with DISPLAY", goos: "linux", env: map[string]string{"DISPLAY": ":0", "SSH_TTY": "/dev/ttys000"}, want: false},
		{name: "darwin DOCKER_CONTAINER blocks", goos: "darwin", env: map[string]string{"DOCKER_CONTAINER": "true"}, want: false},
		{name: "linux DOCKER_HOST blocks", goos: "linux", env: map[string]string{"DOCKER_HOST": "tcp://localhost:2375", "DISPLAY": ":0"}, want: false},
		{name: "windows clean env opens", goos: "windows", env: nil, want: true},
		{name: "windows SSH_TTY blocks", goos: "windows", env: map[string]string{"SSH_TTY": "/dev/ttys000"}, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			envFunc := func(k string) string {
				if tt.env == nil {
					return ""
				}
				return tt.env[k]
			}
			got := shouldOpenBrowser(tt.goos, envFunc)
			if got != tt.want {
				t.Errorf("shouldOpenBrowser(%q, env) = %v, want %v", tt.goos, got, tt.want)
			}
		})
	}
}
