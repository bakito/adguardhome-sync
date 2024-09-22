package utils

import "fmt"

func Ptr[I interface{}](i I) *I {
	return &i
}

func PtrToString[I interface{}](i *I) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%v", i)
}

func PtrEquals[I comparable](a *I, b *I) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
