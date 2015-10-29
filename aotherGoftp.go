package main

import (
	"github.com/secsy/goftp"
	"log"
)

func main(){
	log.SetFlags(log.Flags()|log.Lshortfile)

	ftp, err := goftp.Dial("172.8.4.101:21")
	if err != nil{
		log.Fatal(err)
	}


	fileInfo, err := ftp.Stat("/home/bob/DockerToolbox-1.8.2a.pkg")
	if err != nil{
		log.Fatal(err)
	}

	log.Println(fileInfo)
}