package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"downloader/tcp/echo/protocol"
	"io"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", "127.0.0.1:8989")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	echoProtocol := &protocol.EchoProtocol{}

	// ping <--> pong
	go func() {
		for{
			// write
			if _, err := conn.Write(protocol.NewEchoPacket([]byte("hello"), false).Serialize()); err != nil{
				if neterr, ok := err.(net.Error); ok && neterr.Timeout(){
					fmt.Println("=====>> ", err)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
	for {
		//check connect
		conn.SetReadDeadline(time.Now())
		zero := make([]byte,0);
		if _, err := conn.Read(zero); err == io.EOF {
			fmt.Println("Connect close")
			conn.Close()
			conn = nil
			panic("Connect close!")
		} else {
			conn.SetReadDeadline(time.Time{})
		}

		// read
		p, err := echoProtocol.ReadPacket(conn)
		if err == nil {
			echoPacket := p.(*protocol.EchoPacket)
			fmt.Printf("Server reply:[%v] [%v]\n", echoPacket.GetLength(), string(echoPacket.GetBody()))
		}else {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout(){
				fmt.Println("=====>> ", err)
			}
		}
	}

	conn.Close()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}