package cfg

import (
	"bytes"
	"strings"
	"testing"
)

func init() {
	initTestLogger()
}

func TestNewProperties(t *testing.T) {
	p := NewProperties()
	if p == nil {
		t.Fatal("NewProperties returned nil")
	}
	if p.values == nil {
		t.Fatal("NewProperties values map is nil")
	}
}

func TestProperties_Load(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "simple key-value pairs",
			input: "key1=value1\nkey2=value2\n",
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "with spaces around separator",
			input: "key1 = value1\nkey2 = value2\n",
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "with colon separator",
			input: "key1:value1\nkey2:value2\n",
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "with comments",
			input: "# This is a comment\n! Another comment\nkey1=value1\n",
			want: map[string]string{
				"key1": "value1",
			},
			wantErr: false,
		},
		{
			name:  "with empty lines",
			input: "\n\nkey1=value1\n\nkey2=value2\n\n",
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name:  "complex keys with dots",
			input: "app.name=MyApp\napp.version=1.0\n",
			want: map[string]string{
				"app.name":    "MyApp",
				"app.version": "1.0",
			},
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "only comments and empty lines",
			input:   "# comment\n\n! comment\n\n",
			want:    map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProperties()
			err := p.Load(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for k, v := range tt.want {
				if got := p.Get(k); got != v {
					t.Errorf("Get(%q) = %q, want %q", k, got, v)
				}
			}
		})
	}
}

func TestProperties_Get(t *testing.T) {
	p := NewProperties()
	p.values["key1"] = "value1"
	p.values["key2"] = "value2"

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing key", "key1", "value1"},
		{"another existing key", "key2", "value2"},
		{"non-existing key", "key3", ""},
		{"empty key", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.Get(tt.key); got != tt.want {
				t.Errorf("Get() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProperties_Get_Nil(t *testing.T) {
	var p *Properties
	if got := p.Get("key"); got != "" {
		t.Errorf("Get() on nil Properties = %q, want empty string", got)
	}
}

func TestGetString(t *testing.T) {
	testData := "string.key=hello world\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing string key", "string.key", "hello world"},
		{"non-existing key", "non.existing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetString(tt.key); got != tt.want {
				t.Errorf("GetString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	testData := "int.key=42\nint.negative=-10\nint.zero=0\nint.invalid=notanumber\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want int
	}{
		{"positive integer", "int.key", 42},
		{"negative integer", "int.negative", -10},
		{"zero", "int.zero", 0},
		{"invalid integer", "int.invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInt(tt.key); got != tt.want {
				t.Errorf("GetInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetInt64(t *testing.T) {
	testData := "int64.key=9223372036854775807\nint64.negative=-9223372036854775808\nint64.zero=0\nint64.invalid=notanumber\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want int64
	}{
		{"max int64", "int64.key", 9223372036854775807},
		{"min int64", "int64.negative", -9223372036854775808},
		{"zero", "int64.zero", 0},
		{"invalid int64", "int64.invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInt64(tt.key); got != tt.want {
				t.Errorf("GetInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	testData := "bool.true=true\nbool.false=false\nbool.one=1\nbool.zero=0\nbool.t=T\nbool.f=F\nbool.invalid=notabool\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"true value", "bool.true", true},
		{"false value", "bool.false", false},
		{"1 value", "bool.one", true},
		{"0 value", "bool.zero", false},
		{"T value", "bool.t", true},
		{"F value", "bool.f", false},
		{"invalid bool", "bool.invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBool(tt.key); got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntArray(t *testing.T) {
	testData := "int.array=1,2,3,4,5\nint.empty=\nint.single=42\nint.invalid=notanumber,2,3\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want []int
	}{
		{"multiple integers", "int.array", []int{1, 2, 3, 4, 5}},
		{"empty array", "int.empty", []int{0}},
		{"single integer", "int.single", []int{42}},
		{"invalid array", "int.invalid", []int{0, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIntArray(tt.key)
			if len(got) != len(tt.want) {
				t.Errorf("GetIntArray() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetIntArray()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGetStringArray(t *testing.T) {
	testData := "string.array=apple,banana,orange\nstring.empty=\nstring.single=hello\n"
	p = NewProperties()
	p.Load(strings.NewReader(testData))

	tests := []struct {
		name string
		key  string
		want []string
	}{
		{"multiple strings", "string.array", []string{"apple", "banana", "orange"}},
		{"empty array", "string.empty", []string{""}},
		{"single string", "string.single", []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringArray(tt.key)
			if len(got) != len(tt.want) {
				t.Errorf("GetStringArray() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetStringArray()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestProperties_Load_BytesReader(t *testing.T) {
	input := []byte("key1=value1\nkey2=value2\n")
	p := NewProperties()
	err := p.Load(bytes.NewReader(input))
	if err != nil {
		t.Errorf("Load() with bytes.Reader error = %v", err)
	}
	if got := p.Get("key1"); got != "value1" {
		t.Errorf("Get() = %q, want value1", got)
	}
}

func TestProperties_Load_WithSpecialCharacters(t *testing.T) {
	input := "key1=value with spaces\nkey2=value_with_underscores\nkey3=value-with-dashes\n"
	p := NewProperties()
	err := p.Load(strings.NewReader(input))
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}
	tests := []struct {
		key  string
		want string
	}{
		{"key1", "value with spaces"},
		{"key2", "value_with_underscores"},
		{"key3", "value-with-dashes"},
	}
	for _, tt := range tests {
		if got := p.Get(tt.key); got != tt.want {
			t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}
