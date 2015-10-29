package main

import (
	"crypto/sha256"
//	"crypto/tls"
	"fmt"
	"io"
	"os"

	"github.com/dutchcoders/goftp"
)

func main() {
	var err error
	var ftp *goftp.FTP

	// For debug messages: goftp.ConnectDbg("ftp.server.com:21")
	if ftp, err = goftp.Connect("172.8.4.101:21"); err != nil {
		panic(err)
	}

	defer ftp.Close()

//	config := tls.Config{
//		InsecureSkipVerify: true,
//		ClientAuth:         tls.RequestClientCert,
//	}
//
//	if err = ftp.AuthTLS(config); err != nil {
//		panic(err)
//	}

	if err = ftp.Login("bob", "p@ssw0rd"); err != nil {
		panic(err)
	}

	if err = ftp.Cwd("/home/bob"); err != nil {
		panic(err)
	}

	var curpath string
	if curpath, err = ftp.Pwd(); err != nil {
		panic(err)
	}

	fmt.Printf("Current path: %s\n", curpath)

	var files []string
	if files, err = ftp.List(""); err != nil {
		panic(err)
	}

	fmt.Println(files)

//	var file *os.File
//	if file, err = os.Open("/tmp/test.txt"); err != nil {
//		panic(err)
//	}
//
//	if err := ftp.Stor("/test.txt", file); err != nil {
//		panic(err)
//	}

	err = ftp.Walk("/home/bob", func(path string, info os.FileMode, err error) error {
		_, err = ftp.Retr(path, func(r io.Reader) error {
			var hasher = sha256.New()
			if _, err = io.Copy(hasher, r); err != nil {
				return err
			}

			hash := fmt.Sprintf("%s %x", path, sha256.Sum256(nil))
			fmt.Println(hash)

			return err
		})

		return nil
	})

	fmt.Println(err)
}