//go:build windows
// +build windows

package util

import (
	"golang.org/x/sys/windows"
)

func hasUnicodeSupport() bool {
	const CP_UTF8 = 65001
	outCP, err := windows.GetConsoleOutputCP()
	return err == nil && outCP == CP_UTF8
}
