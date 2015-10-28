package main

import (
	"github.com/coreos/go-etcd/etcd"
	"log"
	"time"
	"fmt"
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	// 停止监听信号通道
	stopWatchChan := make(chan bool)
	// 监听数据接收通道,无缓存,实时数据
	receiverChan := make(chan *etcd.Response)

	machines := []string{"http://127.0.0.1:2379"}
	client := etcd.NewClient(machines)
	defer func() {
		client.Close()
	}()

	// 设置key
	go func() {
		for i := 0; i < 20; i++ {
			time.Sleep(2 * time.Second)
			key := fmt.Sprintf("/watch/child_%d/key_%d", i % 10, i)
//			log.Printf("Set or update key:%s, value:%s\n", key, time.Now().String())
			if _, err := client.Set(key, time.Now().String(), 0); err != nil{
				log.Printf("[ERROR] %v \n", err)
			}
		}

		// 停止监控
		stopWatchChan <- true
	}()

	// 接收
	go func(){
		for{
			resp := <- receiverChan
			log.Println(resp.Action, "resp.Node.Key:", resp.Node.Key, "resp.Node.Value:", resp.Node.Value)
		}
	}()

	// 启动监听
	if _, err := client.Watch("/watch", 0, true, receiverChan, stopWatchChan); err != nil{
		log.Fatal(err)
	}

	// 停止监控
	stopWatchChan <- true
}
