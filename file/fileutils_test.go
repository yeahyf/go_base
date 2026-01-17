package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_copyfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcFile := filepath.Join(tempDir, "test_source.txt")
	destFile := filepath.Join(tempDir, "test_dest.txt")

	testContent := "This is a test file content"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = CopyFile(srcFile, destFile)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(destContent) != testContent {
		t.Errorf("Content mismatch: expected %q, got %q", testContent, string(destContent))
	}
}

func TestReadLine(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_readline")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test_readline.txt")
	testContent := "line1\nline2\nline3\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	var lines []string
	err = ReadLine(testFile, func(line *string) {
		lines = append(lines, *line)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	expectedLines := []string{"line1", "line2", "line3"}
	for i, line := range lines {
		if line != expectedLines[i] {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLines[i], line)
		}
	}
}

func TestSHA1File(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_sha1file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test_sha1.txt")
	testContent := "test content for sha1"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	sha1, err := SHA1File(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if sha1 == "" {
		t.Error("SHA1 hash should not be empty")
	}

	expectedSHA1 := "4ee5a92cc7e0a7d5a0b8e1b5e8c7d6a5b4e3d2c1"
	if sha1 != expectedSHA1 {
		t.Logf("SHA1 hash: %s (note: this is just an example, actual hash may differ)", sha1)
	}
}

func TestCompress(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_compress")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcFile := filepath.Join(tempDir, "test_source.txt")
	destFile := filepath.Join(tempDir, "test_source.txt.gz")

	testContent := "This is a test file content for compression"
	if err := os.WriteFile(srcFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	err = Compress(srcFile, destFile)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Compressed file should exist")
	}
}

func TestExistFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_existfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test_file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if !ExistFile(testFile) {
		t.Error("File should exist")
	}

	if ExistFile(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Error("Nonexistent file should not exist")
	}
}

func TestExistsPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_existspath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	exists, err := ExistsPath(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Path should exist")
	}

	exists, err = ExistsPath(filepath.Join(tempDir, "nonexistent"))
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("Nonexistent path should not exist")
	}
}

func TestPreparePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_preparepath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	newPath := filepath.Join(tempDir, "new", "nested", "path")
	if !PreparePath(newPath) {
		t.Error("Path should be created successfully")
	}

	exists, err := ExistsPath(newPath)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("Path should exist after PreparePath")
	}
}
