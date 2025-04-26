package utils

import (
	"os"
	"strings"
)

// ReadNormalizedFile read a file as string and replace all windows line endings with \n.
func ReadNormalizedFile(name string) (string, error) {
	s, err := ReadFile(name)
	return strings.ReplaceAll(s, "\r\n", "\n"), err
}

// ReadFile read a file as string.
func ReadFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}

// NormalizeLineEndings replace all windows line endings with \n.
func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
