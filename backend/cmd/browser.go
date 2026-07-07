package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// openCommandFor returns the OS-specific command and args to open url in the
// default browser. ok is false on unsupported operating systems.
func openCommandFor(goos, url string) (cmd string, args []string, ok bool) {
	switch goos {
	case "darwin":
		return "open", []string{url}, true
	case "linux":
		return "xdg-open", []string{url}, true
	case "windows":
		return "cmd", []string{"/c", "start", url}, true
	}
	return "", nil, false
}

// shouldOpenBrowser reports whether the current environment has a GUI for the
// browser to open into. env is injected for testability — no global state.
func shouldOpenBrowser(goos string, env func(string) string) bool {
	if env("SSH_TTY") != "" || env("SSH_CONNECTION") != "" {
		return false
	}
	if env("DOCKER_CONTAINER") != "" || env("DOCKER_HOST") != "" {
		return false
	}
	if goos == "linux" && env("DISPLAY") == "" && env("WAYLAND_DISPLAY") == "" {
		return false
	}
	return true
}

// @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
// @orkai:decision exec.Command.Start + Process.Release so the browser detaches from the server process group; server keeps running even if the browser is closed.
func openBrowser(url string) error {
	cmd, args, ok := openCommandFor(runtime.GOOS, url)
	if !ok || !shouldOpenBrowser(runtime.GOOS, os.Getenv) {
		return nil
	}
	c := exec.Command(cmd, args...)
	if err := c.Start(); err != nil {
		return fmt.Errorf("cmd.openBrowser: %w", err)
	}
	_ = c.Process.Release()
	return nil
}

// waitAndOpenBrowser polls /health until the server responds (max 5s, 100ms
// interval), then opens the browser. Intended to run in a goroutine so it
// never blocks server startup.
func waitAndOpenBrowser(port, url string) {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get("http://localhost:" + port + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				_ = openBrowser(url)
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = openBrowser(url)
}
