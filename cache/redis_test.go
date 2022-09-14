package cache

import (
	"testing"
)

//var key = "zset1"

func BenchmarkKey(b *testing.B) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)
	for i := 0; i < b.N; i++ {
		key := "01_1"
		value := "aslkdjfalsdfkj"

		err := p.SetValue(key, value, 0)
		if err != nil {
			b.Fail()
		}
	}

}

func TestSingle(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	key := "01_1"
	value := "aslkdjfalsdfkj"

	err := p.SetValue(key, value, 0)
	if err != nil {
		t.Fail()
	}
	var v string
	v, err = p.GetValue(key)
	if err != nil {
		t.Fail()
	}
	if v != value {
		t.Fail()
	}

	var result int
	result, err = p.DeleteValue(key)
	if err != nil {
		t.Fail()
	}
	if result != 1 {
		t.Fail()
	}
	p.CloseRedisPool()
}

func TestMulti(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	keys := []string{"01_1", "01_2", "01_3", "01_4", "01_5"}
	targets := []string{"01_1", "v_1", "01_2", "v_2", "01_3", "v_3", "01_4", "v_4", "01_5", "v_5"}
	s := make([]interface{}, 0, len(targets))
	for _, v := range targets {
		s = append(s, v)
	}
	err := p.MSetValue(s)
	if err != nil {
		t.Fail()
	}
	p.MSetExpire(keys, 60)

	p.CloseRedisPool()
}

func TestMultiWithExpire(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	//keys := []string{"01_1", "01_2", "01_3", "01_4", "01_5"}
	targets := []string{"01_1", "v_1", "01_2", "v_2", "01_3", "v_3", "01_4", "v_4", "01_5", "v_5"}
	s := make([]interface{}, 0, len(targets))
	for _, v := range targets {
		s = append(s, v)
	}
	err := p.MSetValueWithExpire(s, 60)
	if err != nil {
		t.Fail()
	}
	p.CloseRedisPool()
}

func TestMultiGetWithExpire(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	slice := make([]interface{}, 0, 3)
	slice = append(slice, "123")
	slice = append(slice, "456")
	slice = append(slice, "789")

	result, err := p.MGetValue(slice)
	if err != nil {
		t.Fail()
	}
	t.Log(result)
	p.CloseRedisPool()
}

func TestDeleteValues(t *testing.T) {
	p := NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 10)

	keys := make([]string, 0, 3)
	keys = append(keys, "a")
	keys = append(keys, "b")
	keys = append(keys, "c")

	_, err := p.DeleteValues(keys)

	if err != nil {
		t.Fail()
	}

	p.CloseRedisPool()
}
