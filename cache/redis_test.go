package cache

import (
	"testing"
)

var p *RedisPool

func initPool() {
	p = NewRedisPoolByDB(1, 2, 30, "127.0.0.1:6379", "", 0)
}

func TestList(t *testing.T) {
	initPool()
	key := "web:uv"
	data := []string{"www.sina.com", "www.baidu.com"}
	err := p.LPush(key, data...)
	if err != nil {
		t.Fail()
	}

	d, err := p.Pop(key, "LPOP")
	if err != nil {
		t.Fail()
	} else {
		t.Log(d)
	}
}

func TestZRem(t *testing.T) {
	initPool()
	key := "web:uv"
	value, err := p.ZRem(key, "www.sina.com", "www.baidu.com")
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestZRangeByScore(t *testing.T) {
	initPool()
	key := "salary"
	value, err := p.ZRangeByScore(key, 3000, 4000)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestZRangeByScoreWithScore(t *testing.T) {
	initPool()
	key := "salary"
	f, s, err := p.ZRangeByScoreWithScore(key, 3000, 4000)
	if err != nil {
		t.Fail()
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestLPush(t *testing.T) {
	initPool()
	key := "list"
	err := p.LPush(key, "11", "12", "13", "14", "15", "16", "17", "18", "19", "20")
	if err != nil {
		t.Fail()
	}
}

func TestLPop(t *testing.T) {
	initPool()
	key := "list"
	value, err := p.LPop(key)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestLMPop(t *testing.T) {
	initPool()
	key := "list"
	value, err := p.LMPop(key, 0, 3)
	if err != nil {
		t.Fail()
	} else {
		t.Log(value)
	}
}

func TestZAdd(t *testing.T) {
	initPool()
	setList := "list"
	err := p.ZAdd(setList, "f1", 1)
	if err != nil {
		t.Fail()
	}
}

func TestZMAdd(t *testing.T) {
	initPool()
	setList := "list"
	field := []string{"f2", "f3", "f4"}
	value := []float64{12.2, 433.5, 89.9}
	err := p.ZMAdd(setList, field, value)
	if err != nil {
		t.Fail()
	}
}

func TestZRange(t *testing.T) {
	initPool()
	setList := "list"
	r, err := p.ZRange(setList, 0, 3)
	if err != nil {
		t.Fail()
	} else {
		t.Log(r)
	}
}

func TestZRevRange(t *testing.T) {
	initPool()
	setList := "list"
	r, err := p.ZRevRange(setList, 0, 3)
	if err != nil {
		t.Fail()
	} else {
		t.Log(r)
	}
}

func TestZRangeWithScore(t *testing.T) {
	initPool()
	setList := "list"
	f, s, err := p.ZRangeWithScore(setList, 0, 100)
	if err != nil {
		t.Fail()
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestZRevRangeWithScore(t *testing.T) {
	initPool()
	setList := "list"
	f, s, err := p.ZRevRangeWithScore(setList, 0, 100)
	if err != nil {
		t.Fail()
	} else {
		t.Log(f)
		t.Log(s)
	}
}

func TestZCard(t *testing.T) {
	initPool()
	setList := "list"
	r, err := p.ZCard(setList)
	if err != nil {
		t.Fail()
	} else {
		t.Log(r)
	}
}

func TestZRemRangeByRank(t *testing.T) {
	initPool()
	setList := "list"
	err := p.ZRemRangeByRank(setList, 0, 1)
	if err != nil {
		t.Fail()
	}
}

func TestHSet(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HSet(setList, "key1", "value1")
	if err != nil {
		t.Fail()
	}
}

func TestHMSet(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HMSet(setList, "key2", "value2", "key3", "value3", "key4", "value4")
	if err != nil {
		t.Fail()
	}
}

func TestHMGet(t *testing.T) {
	initPool()
	setList := "set"
	result, err := p.HMGet(setList, []string{"key1", "key2", "key3", "key4", "key5", "key6"})
	if err != nil {
		t.Fail()
	} else {
		t.Log(result)
	}
}

func TestHLen(t *testing.T) {
	initPool()
	setList := "set"
	result, err := p.HLen(setList)
	if err != nil {
		t.Fail()
	} else {
		t.Log(result)
	}
}

func TestHMDel(t *testing.T) {
	initPool()
	setList := "set"
	err := p.HDel(setList, "key2", "key1", "key3")
	if err != nil {
		t.Fail()
	}
}

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
