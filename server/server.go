package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yeahyf/go_base/log"
)

//GracefulServer 定义Server
type GracefulServer struct {
	Server           *http.Server
	shutdownFinished chan struct{}
}

//ListenAndServe 启动服务
func (s *GracefulServer) listenAndServe() (err error) {
	if s.shutdownFinished == nil {
		s.shutdownFinished = make(chan struct{})
	}
	err = s.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	} else if err != nil {
		err = fmt.Errorf("unexpected error from ListenAndServe: %w", err)
		return
	}
	log.Debug("waiting for shutdown finishing ...")
	<-s.shutdownFinished
	log.Debug("server shutdown finished")
	return
}

//WaitForExitingSignal 等待退出的信号
func (s *GracefulServer) waitForExitingSignal(timeout time.Duration) {
	var waiter = make(chan os.Signal, 1) // buffered channel
	signal.Notify(waiter, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	<-waiter

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := s.Server.Shutdown(ctx)
	if err != nil {
		log.Errorf("shutting down: %s", err)
	} else {
		fmt.Println("http server shutdown successfully")
		close(s.shutdownFinished)
	}
}

func StartServer(port int, mux *http.ServeMux) {
	server := &GracefulServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
	//异步监听退出信号
	go server.waitForExitingSignal(10 * time.Second)

	fmt.Printf("http server listening on port %d...\n", port)
	//开始监听
	err := server.listenAndServe()
	if err != nil {
		err = fmt.Errorf("unexpected error from ListenAndServe: %s", err)
	}
	fmt.Println("system main goroutine exited.")
	return
}
