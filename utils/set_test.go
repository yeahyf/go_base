package utils

import (
	"testing"
)

func Test_Set(t *testing.T) {
	set := NewSet[string](10)
	set.Add("1")
	set.Add("2")
	if set.Contains("1") {
		t.Log("OK")
	}
}

func Test_Set2(t *testing.T) {
	set := NewSet[int](10)
	set.Add(1)
	set.Add(2)
	for k := range set {
		t.Log(k)
	}
}
