package log

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testConfigFile string

func setupTestConfig(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigFile = filepath.Join(tmpDir, "test_log_config.json")

	config := Config{
		Logs: []LevelConfig{
			{
				Filename:   filepath.Join(tmpDir, "debug.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelDebug,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "info.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelInfo,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "error.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelError,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "warn.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelWarn,
				Type:       0,
				Rotation:   1,
			},
		},
		Level: LevelDebug,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(testConfigFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	SetLogConf(&testConfigFile)
}

func TestBasicLogOutput(t *testing.T) {
	setupTestConfig(t)

	tests := []struct {
		name     string
		logFunc  func(...interface{})
		logLevel string
		message  string
	}{
		{"Debug log", Debug, LevelDebug, "This is a debug message"},
		{"Info log", Info, LevelInfo, "This is an info message"},
		{"Error log", Error, LevelError, "This is an error message"},
		{"Warn log", Warn, LevelWarn, "This is a warn message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.logLevel)
			tt.logFunc(tt.message)
		})
	}
}

func TestFormattedLogOutput(t *testing.T) {
	setupTestConfig(t)

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		logLevel string
		format   string
		args     []interface{}
	}{
		{"Debugf log", Debugf, LevelDebug, "Debug message: %s, count: %d", []interface{}{"test", 42}},
		{"Infof log", Infof, LevelInfo, "Info message: %s, count: %d", []interface{}{"test", 42}},
		{"Errorf log", Errorf, LevelError, "Error message: %s, count: %d", []interface{}{"test", 42}},
		{"Warnf log", Warnf, LevelWarn, "Warn message: %s, count: %d", []interface{}{"test", 42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLevel(tt.logLevel)
			tt.logFunc(tt.format, tt.args...)
		})
	}
}

func TestLogLevelSwitching(t *testing.T) {
	setupTestConfig(t)

	levels := []string{LevelDebug, LevelInfo, LevelWarn, LevelError}

	for _, level := range levels {
		t.Run("Set level to "+level, func(t *testing.T) {
			SetLevel(level)
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestIsDebug(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)
	if !IsDebug() {
		t.Error("Expected IsDebug() to return true when level is debug")
	}

	SetLevel(LevelInfo)
	if IsDebug() {
		t.Error("Expected IsDebug() to return false when level is info")
	}

	SetLevel(LevelError)
	if IsDebug() {
		t.Error("Expected IsDebug() to return false when level is error")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelError)

	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Debug("This debug message should not appear")
	Info("This info message should not appear")
	Warn("This warn message should not appear")
	Error("This error message should appear")

	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)
}

func TestMultipleArguments(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debug("arg1", "arg2", "arg3", 123, true)
	Info("info", "with", "multiple", "arguments")
	Error("error", 42, "occurred")
	Warn("warning", "message", "with", "data", 3.14)
}

func TestComplexFormattedMessage(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debugf("Complex format: %s %d %v %f", "string", 42, true, 3.14159)
	Infof("User %s performed action %s at %v", "john", "login", time.Now())
	Errorf("Error occurred: %v", os.ErrNotExist)
	Warnf("Warning: %s, retries: %d", "connection timeout", 3)
}

func TestLogWithNilArguments(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debug(nil)
	Info(nil)
	Error(nil)
	Warn(nil)
}

func TestLogWithEmptyString(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debug("")
	Info("")
	Error("")
	Warn("")
}

func TestLogWithSpecialCharacters(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debug("Special chars: \n\t\r")
	Info("Unicode: 你好世界 🌍")
	Error("Quotes: \"single\" 'double'")
	Warn("Backslash: \\")
}

func TestLogWithLargeMessage(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	largeMessage := strings.Repeat("This is a large message. ", 1000)
	Debug(largeMessage)
	Info(largeMessage)
}

func TestLogWithStruct(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	type TestStruct struct {
		Name  string
		Value int
	}

	s := TestStruct{Name: "test", Value: 42}
	Debugf("Struct: %+v", s)
	Infof("Struct: %#v", s)
}

func TestPerformanceOptimization(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelError)

	iterations := 10000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		Debugf("This should not be logged: %d", i)
	}
	elapsed := time.Since(start)

	t.Logf("Logged %d disabled debug messages in %v (%.2f ns/op)",
		iterations, elapsed, float64(elapsed.Nanoseconds())/float64(iterations))

	if elapsed > time.Second {
		t.Errorf("Performance issue: logging %d disabled messages took too long: %v", iterations, elapsed)
	}
}

func TestConcurrentLogging(t *testing.T) {
	setupTestConfig(t)

	SetLevel(LevelDebug)

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Debugf("Goroutine %d, message %d", id, j)
				Infof("Goroutine %d, info %d", id, j)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLogConfigReload(t *testing.T) {
	t.Skip("Skipping long-running config reload test")
}

func TestLogWithPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	setupTestConfig(t)

	SetLevel(LevelDebug)

	Debug("Before panic")
}

func BenchmarkDebugLog(b *testing.B) {
	tmpDir := b.TempDir()
	benchmarkConfigFile := filepath.Join(tmpDir, "benchmark_log_config.json")

	config := Config{
		Logs: []LevelConfig{
			{
				Filename:   filepath.Join(tmpDir, "debug.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelDebug,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "info.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelInfo,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "error.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelError,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "warn.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelWarn,
				Type:       0,
				Rotation:   1,
			},
		},
		Level: LevelDebug,
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(benchmarkConfigFile, data, 0644)
	SetLogConf(&benchmarkConfigFile)
	SetLevel(LevelDebug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debug("Benchmark message")
	}
}

func BenchmarkDebugfLog(b *testing.B) {
	tmpDir := b.TempDir()
	benchmarkConfigFile := filepath.Join(tmpDir, "benchmark_log_config.json")

	config := Config{
		Logs: []LevelConfig{
			{
				Filename:   filepath.Join(tmpDir, "debug.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelDebug,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "info.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelInfo,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "error.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelError,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "warn.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelWarn,
				Type:       0,
				Rotation:   1,
			},
		},
		Level: LevelDebug,
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(benchmarkConfigFile, data, 0644)
	SetLogConf(&benchmarkConfigFile)
	SetLevel(LevelDebug)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debugf("Benchmark message: %d", i)
	}
}

func BenchmarkDisabledDebugLog(b *testing.B) {
	tmpDir := b.TempDir()
	benchmarkConfigFile := filepath.Join(tmpDir, "benchmark_log_config.json")

	config := Config{
		Logs: []LevelConfig{
			{
				Filename:   filepath.Join(tmpDir, "debug.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelDebug,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "info.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelInfo,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "error.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelError,
				Type:       0,
				Rotation:   1,
			},
			{
				Filename:   filepath.Join(tmpDir, "warn.log"),
				MaxSize:    10,
				MaxBackups: 3,
				MaxAge:     7,
				Compress:   false,
				Name:       LevelWarn,
				Type:       0,
				Rotation:   1,
			},
		},
		Level: LevelDebug,
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(benchmarkConfigFile, data, 0644)
	SetLogConf(&benchmarkConfigFile)
	SetLevel(LevelError)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Debugf("This should not be logged: %d", i)
	}
}
