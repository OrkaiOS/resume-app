//go:build prod

package main

import "embed"

var hasFrontendFS = true

//go:embed static/*
var prodFrontendFS embed.FS
