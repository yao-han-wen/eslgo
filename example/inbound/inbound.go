package main

import (
	"log"
	"time"

	"github.com/yao-han-wen/eslgo"
)

func main() {
	client, err := eslgo.NewInboundSocket("192.168.101.97:8021", eslgo.WithConnectPassword("ClueCon"))
	if err != nil {
		log.Println(err)
		return
	}
	defer client.Close()

	go func() {
		// eventChan, err := client.SendEventCommand("plain ALL")
		err = client.SendEventCommand("event xml ALL")
		// eventChan, err := client.SendEventCommand("event json ALL")
		if err != nil {
			log.Println("SendEventCommand error:", err)
			return
		}

		jobUUID, err := client.SendBgApiCommand("status")
		if err != nil {
			log.Println("SendBgApiCommand error:", err)
			return
		}
		log.Println("bgapi result, jobUUID:" + jobUUID)

		rsApi, err := client.SendApiCommand("status")
		if err != nil {
			log.Println("SendApiCommand error:", err)
			return
		}
		log.Println("api result:" + rsApi)

		eventChan := client.GetEventChan()
		for resp := range eventChan {
			e, err := resp.ToEvent()
			if err != nil {
				log.Println("ToEvent error:", err)
				return
			}
			log.Println("event:", e)
		}
	}()

	//手动测试关闭
	go func() {
		time.Sleep(time.Second * 5)
		client.Close()
		log.Println("手动关闭")
	}()

	//关闭通知
	closeNotify := <-client.CloseNotify()
	log.Println("closeNotify", closeNotify)

}
