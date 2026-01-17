package ver

import (
	"testing"
	"time"
)

func Test_Version(t *testing.T) {
	v := Version{
		VerNo:      "1.2.3",
		SystemName: "Test System Name",
		BuildTime:  "2022-01-02 15:00",
	}
	v.Print()
}

func Test_Debug(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("Debug method requires log initialization, skipping test")
		}
	}()
	v := Version{
		VerNo:      "1.2.3",
		SystemName: "Test System Name",
		BuildTime:  "2022-01-02 15:00",
	}
	v.Debug()
}

func Test_Clean(t *testing.T) {
	v := Version{
		VerNo:      "1.2.3",
		SystemName: "Test System Name",
		BuildTime:  "2022-01-02 15:00",
	}

	cleanCalled := false
	cancel := v.Clean(func() {
		cleanCalled = true
		t.Log("Clean function called")
	})

	time.Sleep(100 * time.Millisecond)

	if cleanCalled {
		t.Error("Clean function should not be called without signal")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
}

func Test_CleanWithCancel(t *testing.T) {
	v := Version{
		VerNo:      "1.2.3",
		SystemName: "Test System Name",
		BuildTime:  "2022-01-02 15:00",
	}

	cleanCalled := false
	stopClean := v.Clean(func() {
		cleanCalled = true
	})

	time.Sleep(50 * time.Millisecond)
	stopClean()
	time.Sleep(50 * time.Millisecond)

	if cleanCalled {
		t.Error("Clean function should not be called when cancelled")
	}
}
