package utils

import "encoding/json"

func Clone[I interface{}](in I, out I) I {
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, out)
	return out
}
