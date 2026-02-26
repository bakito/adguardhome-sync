package utils

import "fmt"

func PtrToString[I any](i *I) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%v", i)
}

func PtrEquals[I comparable](a, b *I) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
