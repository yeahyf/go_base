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

func Test_CloseAction(t *testing.T) {
	CloseAction(nil)
}

func Test_GetBytesForInt64(t *testing.T) {
	num := uint64(123456789)
	result := GetBytesForInt64(num)
	if len(result) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(result))
	}
}

func Test_Contain(t *testing.T) {
	set := NewSet[string](10)
	set.Add("1")
	set.Add("2")

	if !set.Contain("1") {
		t.Error("Expected set to contain '1'")
	}
	if !set.Contain("2") {
		t.Error("Expected set to contain '2'")
	}
	if set.Contain("3") {
		t.Error("Expected set not to contain '3'")
	}

	intSet := NewSet[int](10)
	intSet.Add(1)
	intSet.Add(2)

	if !intSet.Contain(1) {
		t.Error("Expected set to contain 1")
	}
	if intSet.Contain(3) {
		t.Error("Expected set not to contain 3")
	}
}
