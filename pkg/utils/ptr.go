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
