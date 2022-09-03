package ver

import (
	"testing"
)

func Test_Version(t *testing.T) {
	v := Version{
		VerNo:      "1.2.3",
		SystemName: "Test System Name",
		BuildTime:  "2022-01-02 15:00",
	}
	v.Print()
}
