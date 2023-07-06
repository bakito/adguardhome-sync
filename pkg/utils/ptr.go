package utils

func Ptr[I interface{}](i I) *I {
	return &i
}
