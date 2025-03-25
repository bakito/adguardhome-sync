package utils

import (
	"bytes"
	"encoding/json"
)

func Clone[I any](in, out I) I {
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, out)
	return out
}

func JSONEquals(a, b any) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return bytes.Equal(ja, jb)
}
