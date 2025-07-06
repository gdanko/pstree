//go:build !windows
// +build !windows

package util

import (
	"os"
	"strings"
)

func hasUnicodeSupport() bool {
	lang := os.Getenv("LANG")
	lcAll := os.Getenv("LC_ALL")
	return strings.Contains(strings.ToLower(lang), "utf-8") ||
		strings.Contains(strings.ToLower(lcAll), "utf-8")
}
