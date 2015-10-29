package main

import (
	ftp4go "github.com/shenshouer/ftp4go"
	"os"
	"log"
)

var(
	downloadFileName 	= "DockerToolbox-1.8.2a.pkg1"
	BASE_FTP_PATH 		= "/home/bob/"					// base data path in ftp server
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	ftpClient := ftp4go.NewFTP(0) // 1 for debugging

	//connect
	_, err := ftpClient.Connect("172.8.4.101", ftp4go.DefaultFtpPort, "")
	if err != nil {
		log.Println("The connection failed. err:", err)
		os.Exit(1)
	}
	defer ftpClient.Quit()

	_, err = ftpClient.Login("bob", "p@ssw0rd", "")
	if err != nil {
		log.Println("The login failed err:", err)
		os.Exit(1)
	}

	//Print the current working directory
	var cwd string
	cwd, err = ftpClient.Pwd()
	if err != nil {
		log.Println("The Pwd command failed. err:",err)
		os.Exit(1)
	}
	log.Println("The current folder is", cwd)


	// get the remote file size
	size, err := ftpClient.Size("/home/bob/"+downloadFileName)
	if err != nil {
		log.Println("The Size command failed. err:", err)
		//os.Exit(1)
	}
	log.Println("size ", size)

	// start resume file download
	if err = ftpClient.DownloadResumeFile("/home/bob/"+downloadFileName, "/Users/goyoo/ftptest/"+downloadFileName, false); err != nil{
		panic(err)
	}

}