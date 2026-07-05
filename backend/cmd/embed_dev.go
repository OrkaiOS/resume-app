//go:build !prod

package main

import "embed"

var hasFrontendFS = false

var prodFrontendFS embed.FS
