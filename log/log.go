package log

import (
	"fmt"

	seelog "github.com/cihub/seelog"
)

var L seelog.LoggerInterface

func SetLogConf(logConfig *string) {
	fmt.Println("Start set log config ... ")
	//初始化全局变量Logger为seelog的禁用状态，主要为了防止Logger被多次初始化
	var err error
	L = seelog.Disabled
	L, err = seelog.LoggerFromConfigAsFile(*logConfig)
	if err != nil {
		fmt.Println(err)
	}
}
