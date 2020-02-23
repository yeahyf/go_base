package strutil

import (
	"fmt"
	"testing"
)

func TestSortString(t *testing.T ){
	str := make([]*string,5)
	s1 := "Akjfewla"
	s2 := "ekfej"
	s3 := "3elkfje"
	s4 := "WWdlkefje"
	s5 := "kdkdlkwejf"
	str[0] = &s1
	str[1] = &s2
	str[2] = &s3
	str[3] = &s4
	str[4] = &s5

	SortString(str)
	for _,v := range str{
		fmt.Printf("%s\n",*v)
	}
}

func TestSortStr(t *testing.T) {
	str := make([]string,5)
	s1 := "Akjfewla"
	s2 := "ekfej"
	s3 := "3elkfje"
	s4 := "WWdlkefje"
	s5 := "kdkdlkwejf"
	str[0] = s1
	str[1] = s2
	str[2] = s3
	str[3] = s4
	str[4] = s5

	SortStr(str)
	for _,v := range str{
		fmt.Printf("%s\n",v)
	}
}
