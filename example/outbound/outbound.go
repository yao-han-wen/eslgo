package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yao-han-wen/eslgo"
)

func main() {
	// Start listening, this is a blocking function
	s, err := eslgo.NewOutboundServer(":8084", handleConnection)
	if err != nil {
		return
	}
	go func() {
		if err := s.Serve(); err != nil {
			log.Println(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}

func handleConnection(ctx context.Context, conn *eslgo.Connection) {
	defer conn.Close()

	log.Println("Got connection!")

	//do somethings
	for {
		select {
		case <-ctx.Done():
			return
		case <-conn.CloseNotify():
			return
		}
	}
}
