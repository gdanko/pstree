package util

// HasUnicodeSupport is implemented per platform.
func HasUnicodeSupport() bool {
	return hasUnicodeSupport()
}
