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
