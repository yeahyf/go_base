package cache

var key = "zset1"

// func TestZadd(t *testing.T) {
// 	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")
// 	key := "zset1"
// 	var value float32
// 	prefix := "323abcedf_us_011_Tomson Handis_extends11111"
// 	var member string
// 	for i := 1; i < 10000; i++ {
// 		value = rand.Float32() * 10000000
// 		member = prefix + strconv.FormatInt(rand.Int63n(10000000), 10)
// 		err := p.ZAdd(&key, &member, value)
// 		if err != nil {
// 			t.Fail()
// 		}
// 	}
// }

// func TestZcard(t *testing.T) {
// 	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")
// 	card, err := p.Zcard(&key)
// 	if err == nil {
// 		t.Log(card)
// 	} else {
// 		t.Fail()
// 	}
// }

// func TestZcount(t *testing.T) {
// 	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")
// 	card, err := p.Zcount(&key, 100, 10000)
// 	if err == nil {
// 		t.Log(card)
// 	} else {
// 		t.Fail()
// 	}
// }

// func TestZrevange(t *testing.T) {
// 	p := NewRedisPool(1, 2, 30, "127.0.0.1:6379", "")
// 	card, err := p.Zrevrange(&key, 41, 51)
// 	if err == nil {
// 		for _, v := range card {
// 			t.Log(v)
// 		}
// 	} else {
// 		t.Fail()
// 	}
// }

// func TestZrank(t *testing.T) {
// 	p := NewRedisPool(1, 2, 30, "192.168.1.10:6379", "master")
// 	member := "aefdef2591373"
// 	newmem := "xxxxxxxxxxxxx"
// 	key := "01_1"
// 	err := p.Zupdatemember(&key, &newmem, &member)
// 	if err == nil {
// 		t.Log("OK")
// 	} else {
// 		fmt.Println(err)
// 		t.Fail()
// 	}
// }
