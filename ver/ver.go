package ver

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yeahyf/go_base/log"
)

// Version 系统版本信息
type Version struct {
	SystemName string //系统名称
	VerNo      string //系统版本号
	BuildTime  string //构建时间
}

func (v Version) Print() {
	fmt.Println("============================================")
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("["+v.SystemName+"]", "Start ...")
	fmt.Println(v.BuildTime, "System Build! Ver:", v.VerNo)
	fmt.Println("pid:", os.Getpid())
}

func (v Version) Debug() {
	if log.IsDebug() {
		log.Debug(time.Now().Format("2006-01-02 15:04:05"))
		log.Debug("["+v.SystemName+"]", "Start ...")
		log.Debug(v.BuildTime, " System Build! Ver: ", v.VerNo)
	}
}

func (v Version) Clean(clear func()) {
	notify := make(chan os.Signal, 1)
	signal.Notify(notify, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-notify
		clear()
	}()
}
