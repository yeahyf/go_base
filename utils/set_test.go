package utils

import (
	"testing"
)

func Test_Set(t *testing.T) {
	set := NewStringSet(10)
	set.Add("foo")
	set.Add("foo")
	set.Add("foo")
	set.Add("foo")
	if set.Len() != 1 {
		t.Failed()
	}
	if !set.Contains("foo") {
		t.Failed()
	}
}

func Test_Set2(t *testing.T) {
	set := NewStringSet(10)
	set.Add("foo")
	set.Add("lkasjdf")
	set.Add("测")
	set.Add("clef")
	if set.Len() == 4 {
		t.Log("OK")
	} else {
		t.Failed()
	}
	if !set.Contains("foo") {
		t.Failed()
	}
}
