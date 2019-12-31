package strutil

import "unsafe"

///处理字符串与[]byte数组的转换

//byte数组转化为字符串，返回字符串
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//byte数组转化为字符串，返回字符串引用
func Bytes2Str(b []byte) *string {
	return (*string)(unsafe.Pointer(&b))
}

//字符串转化为byte数组
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}      // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

//字符串转化为byte数组
func String2bytes(s *string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(s)) // 获取s的起始地址开始后的两个 uintptr 指针
	h := [3]uintptr{x[0], x[1], x[1]}     // 构造三个指针数组
	return *(*[]byte)(unsafe.Pointer(&h))
}

func SortString(list []*string) {
	length := len(list)
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++ {
			if CompareString(list[i], list[j]) {
				temp := list[i]
				list[i] = list[j]
				list[j] = temp
			}
		}
	}
}

// 返回true str1>=str2, 返回false str1<str2
func CompareString(str1, str2 *string) bool {
	len1 := len(*str1)
	len2 := len(*str2)
	length := len1
	if len1 > len2 {
		length = len2
	}
	for i := 0; i < length; i++ {
		if (*str1)[i] > (*str2)[i] {
			return true
		}
		if (*str1)[i] < (*str2)[i] {
			return false
		}
	}
	if len1 >= len2 {
		return true
	} else {
		return false
	}
}
