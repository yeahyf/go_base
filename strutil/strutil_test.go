package strutil

import (
	"testing"
)

func Test_StrByOXR(t *testing.T) {
	source := "testing"
	key := "abc"
	str := StrByXOR([]byte(source), []byte(key))
	str = StrByXOR(str, []byte(key))
	if string(str) == source {
		t.Log("ok")
	}
}
