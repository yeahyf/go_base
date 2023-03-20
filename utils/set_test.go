package utils

import (
	"testing"
)

func Test_Set(t *testing.T) {
	set := NewSet[string](10)
	set.Add("1")
	set.Add("2")
}

func Test_Set2(t *testing.T) {
	set := NewSet[int](10)
	set.Add(1)
	set.Add(2)
	for k := range set {
		t.Log(k)
	}
}

func Test_GetInt64FromBytes(t *testing.T) {
	bytes := []byte{'1', '2', '3', '8', '8', '2', '9', '9', '0'}
	t.Log(GetInt64FromBytes(bytes))
}
