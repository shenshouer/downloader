package main

import (
	"github.com/valyala/gorpc"
	"log"
	"downloader/rpc/utils"
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(comm.Person{})
	gob.Register(comm.Student{})
}

func main(){
	log.SetFlags(log.Flags()|log.Lshortfile)

	c := &gorpc.Client{
		// TCP address of the server.
//		Addr: "127.0.0.1:12345",
		Addr: "10.10.18.97:12345",
	}
	c.Start()

	// All client methods issuing RPCs are thread-safe and goroutine-safe,
	// i.e. it is safe to call them from multiple concurrently running goroutines.
	resp, err := c.Call("foobar")
	if err != nil {
		log.Fatalf("Error when sending request to server: %s", err)
	}
	log.Println(resp)

	for i:=0; true; i++{
		time.Sleep(2 * time.Second)
		person := comm.Person{
			No:		i,
			Name: 	"张三",
			Age: 	14,
		}

		resp, err = c.Call(person)
		if err != nil {
			log.Fatalf("Error when sending request to server: %s", err)
		}

		log.Println(resp)
	}
}