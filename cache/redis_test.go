package cache

import (
	"testing"
)

//var key = "zset1"

func TestSetGetDelete(t *testing.T) {
	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")

	key := "01_1"
	value := "aslkdjfalsdfkj"

	err := p.SetValue(&key, &value, 0)
	if err != nil {
		t.Fail()
	}
	var v *string
	v, err = p.GetValue(&key)
	if err != nil {
		t.Fail()
	}
	if v!=nil && *v != value {
		t.Fail()
	}

	var result int
	result, err = p.DeleteValue(&key)
	if err != nil {
		t.Fail()
	}
	if result != 1 {
		t.Fail()
	}
}


func TestMultiSetGetDelete(t *testing.T) {
	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")

	key := "01_2"
	value := "aslkdjfalsdfkj"

	err := p.SetValueForDBIdx (&key, &value, 0,3)
	if err != nil {
		t.Fail()
	}
	var v *string
	v, err = p.GetValueForDBIdx(&key,3)
	if err != nil {
		t.Fail()
	}
	if v!= nil && *v != value {
		t.Fail()
	}

	var result int
	result, err = p.DeleteValueForDBIdx(&key,3)
	if err != nil {
		t.Fail()
	}
	if result != 1 {
		t.Fail()
	}
}
