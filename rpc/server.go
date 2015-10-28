package main

import (
	"github.com/valyala/gorpc"
	"log"
	"downloader/rpc/utils"
	"encoding/gob"
)

func init() {
	gob.Register(comm.Person{})
	gob.Register(comm.Student{})
}

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	s := gorpc.Server{
		// Accept clients on this TCP address.
		Addr: ":12345",
		// Echo handler - just return back the message we received from the client
		Handler: func(clientAddr string, request interface{}) interface{} {
			log.Printf("Obtained request %+v from the client %s\n", request, clientAddr)

			switch request.(type) {
			case string :
				return request
			case comm.Person:
				person, _ := request.(comm.Person)
				return comm.Student{
					Person:person,
					Grade: "高三",
					Class: "3班",
					School:"临湘二中",
				}
			}

			return "ERROR"
		},
	}

	if err := s.Serve(); err != nil {
		log.Fatalf("Cannot start rpc server: %s", err)
	}
}
