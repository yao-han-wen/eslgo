package main

import (
	"log"
	"time"

	"github.com/yao-han-wen/eslgo"
)

func main() {
	client, err := eslgo.NewInboundSocket("192.168.101.97:8021", "ClueCon", eslgo.WithCmdTimeout(5))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	go func() {
		// eventChan, err := client.SendEventCommand("plain ALL")
		eventChan, err := client.SendEventCommand("event xml ALL")
		// eventChan, err := client.SendEventCommand("event json ALL")
		if err != nil {
			log.Println("SendEventCommand error", err)
			return
		}

		jobUuid, err := client.SendBgApiCommand("status")
		if err != nil {
			log.Println("SendBgApiCommand error", err)
			return
		}
		log.Println("bgapi 指令结果：" + jobUuid)

		rsApi, err := client.SendApiCommand("status")
		if err != nil {
			log.Println("SendApiCommand error", err)
			return
		}
		log.Println("api 指令结果：" + rsApi)

		for e := range eventChan {
			log.Println(e)
		}
	}()

	//手动测试关闭
	go func() {
		time.Sleep(time.Second * 5)
		client.Close()
	}()

	//关闭通知
	closeNotify := <-client.CloseNotify()
	log.Println("closeNotify", closeNotify)

}
