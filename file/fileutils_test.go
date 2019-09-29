package file

import (
	"fmt"
	"testing"
)

func TestCopyFile(t *testing.T) {
	_, err := CopyFile("/Users/yeahyf/pub/res/1/xyzd.png", "/Users/yeahyf/pub_bak/res/2/xyzd.png")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestReadLine(t *testing.T) {
	err := ReadLine("/Users/yeahyf/go/src/gobase/file/fileutils.go", print)
	if err != nil {
		t.Fail()
	}
}

func print(line *string) {
	fmt.Println(*line)
}
