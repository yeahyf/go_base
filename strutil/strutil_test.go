package strutil

import (
	"bytes"
	"compress/gzip"
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

func TestBytes2str(t *testing.T) {
	b := []byte("hello world")
	s := Bytes2str(b)
	if s != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", s)
	}
}

func TestBytes2Str(t *testing.T) {
	b := []byte("hello world")
	s := Bytes2Str(b)
	if s == nil {
		t.Error("Expected non-nil string pointer")
	}
	if *s != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", *s)
	}
}

func TestStr2bytes(t *testing.T) {
	s := "hello world"
	b := Str2bytes(s)
	if string(b) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(b))
	}
}

func TestString2bytes(t *testing.T) {
	s := "hello world"
	b := String2bytes(&s)
	if string(b) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(b))
	}
}

func TestSortString(t *testing.T) {
	list := []string{"banana", "apple", "cherry"}
	SortString(list)
	if list[0] != "apple" || list[1] != "banana" || list[2] != "cherry" {
		t.Errorf("Expected ['apple', 'banana', 'cherry'], got %v", list)
	}
}

func TestIsBase64String(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"SGVsbG8gV29ybGQ=", true},
		{"aGVsbG8=", true},
		{"invalid@#", false},
		{"", false},
		{"a", false},
		{"ab", false},
		{"abc", false},
		{"abcd", true},
	}

	for _, test := range tests {
		result := IsBase64String(&test.input)
		if result != test.expected {
			t.Errorf("IsBase64String(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestConvertBytes(t *testing.T) {
	normalStr := "hello world"
	result := ConvertBytes(&normalStr)
	if string(result) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(result))
	}

	base64Str := "SGVsbG8gV29ybGQ="
	result = ConvertBytes(&base64Str)
	if string(result) != "SGVsbG8gV29ybGQ=" {
		t.Errorf("Expected 'SGVsbG8gV29ybGQ=', got '%s'", string(result))
	}
}

func TestConvertString(t *testing.T) {
	normalBytes := []byte("hello world")
	result := ConvertString(normalBytes)
	if result == nil || *result != "hello world" {
		t.Errorf("Expected 'hello world', got '%v'", result)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("hello world"))
	gz.Close()
	gzippedBytes := buf.Bytes()

	result = ConvertString(gzippedBytes)
	if result == nil {
		t.Error("Expected non-nil string pointer")
	}
}

func TestGunzip(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("hello world"))
	gz.Close()
	gzippedBytes := buf.Bytes()

	result, err := Gunzip(gzippedBytes)
	if err != nil {
		t.Errorf("Gunzip failed: %v", err)
	}
	if string(result) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(result))
	}

	_, err = Gunzip([]byte("not gzipped"))
	if err == nil {
		t.Error("Expected error for non-gzipped data")
	}
}
