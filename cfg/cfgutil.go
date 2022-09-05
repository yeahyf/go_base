package cfg

import (
	"flag"
	"os"
	"path"

	"github.com/yeahyf/go_base/file"
	"github.com/yeahyf/go_base/log"
)

func LoadCfg() string {
	cfgPath := flag.String("cfg", "./conf", "please set the conf path")
	flag.Parse()

	pathInfo, err := os.Stat(*cfgPath)
	if err != nil || !pathInfo.IsDir() {
		flag.PrintDefaults()
		panic("conf path does not exist")
	}

	cfgFile := path.Join(*cfgPath, "cfg.properties")
	if !file.ExistFile(cfgFile) {
		flag.PrintDefaults()
		panic("cfg.properties does not exist")
	}

	logFile := path.Join(*cfgPath, "zap.json")
	if !file.ExistFile(logFile) {
		flag.PrintDefaults()
		panic("zap.json does not exist")
	}

	// 加载日志模块
	log.SetLogConf(&logFile)
	// 加载系统参数配置文件
	Load(&cfgFile)
	return *cfgPath
}
