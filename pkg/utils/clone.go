package utils

import "encoding/json"

func Clone[I interface{}](in I, out I) I {
	b, _ := json.Marshal(in)
	_ = json.Unmarshal(b, out)
	return out
}

func JsonEquals(a interface{}, b interface{}) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return string(ja) == string(jb)
}
