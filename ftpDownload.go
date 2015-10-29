//TODO 多线程下载
//TODO 下载失败消息通知
//TODO 下载成功消息通知
//TODO 下载完成后处理
package main

import (
	"downloader/ftp"
	"log"
	"os"
	"fmt"
//	"io/ioutil"
	"io"
	"bufio"
	"time"
)

type(
	ByteSize float64

	FTPInfo struct {
		client 		*ftp.ServerConn
		isLogin		bool
		FtpServer	string
		FtpUser		string
		FtpPassword	string
	}
)

const (
	BASE_FTP_PATH 	= "/home/bob/"				// ftp服务器根路径
	BASE_DATA_PATH 	= "/Users/goyoo/ftptest/"	// 本地存储路径

	B  ByteSize 	= iota
	KB 		    	= 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB

	max_thread 		= 10 						// 下载的最大线程数
	Block_Size 		= 1 * KB					// 每个线程中间将下载文件分块,每块数据大小
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)
	// 下载信息
	downloadFileName 	:= "Firefox-latest.dmg"
	ftpServer 			:= "172.8.4.101:21"
	ftpUser 			:= "bob"
	ftpPassword 		:= "p@ssw0rd"

	fi := &FTPInfo{
		FtpServer:ftpServer,
		FtpUser:ftpUser,
		FtpPassword:ftpPassword,
	}

	if err := fi.Dowload(downloadFileName); err != nil{
		log.Fatal(err)
	}
}

// 登陆FTP服务器
func (f *FTPInfo) Login () (err error) {
	log.Println(fmt.Sprintf("[Info] try to connect ftp server[%s]", f.FtpServer))
	if f.client, err = ftp.Connect(f.FtpServer); err != nil{
		log.Printf("[ERROR] connecting ftp server error: %v \n", err)
		return
	}
	log.Println("[Info] connected success!")
	log.Printf("[Info] try to login ftp server[%s] with password %s \n", f.FtpServer, f.FtpPassword)
	if err = f.client.Login(f.FtpUser, f.FtpPassword); err != nil{
		log.Printf("[ERROR] Login to ftp server error: %v\n", err)
		return
	}
	log.Println("[Info] login success!")
	f.isLogin = true
	return
}

// 下载文件
func (f *FTPInfo) Dowload(fileName string) (err error) {
	if !f.isLogin{
		if err = f.Login(); err != nil{
			return
		}
	}
	defer f.client.Quit()
	var size int64
	size, err = f.client.Size(BASE_FTP_PATH+fileName)
	if err != nil{
		log.Println("FTP Server 不支持 size操作", err)
	}
	log.Printf("下载文件大小为%d \n", size)

	var tmpFile *os.File
	log.Println("[Info] try to create the temp file to receive the download file!")
	if tmpFile, err = os.OpenFile(BASE_DATA_PATH+fileName, os.O_RDWR|os.O_CREATE, 0666); err != nil{
		log.Printf("[ERROR] create the temp file error %v \n", err)
		return
	}else{
		log.Println("[Info] create temp file success!")
		downloadFile := BASE_FTP_PATH+fileName
		log.Printf("[Info] try to download file %s \n", downloadFile)
		defer tmpFile.Close()
		//检查本地文件状态
		var size int64 = 0
		if size, err = checkFile(tmpFile); err != nil{
			log.Println("[Warn] check local file error, will restart to download all bytes of the file")
		}

		var reader io.ReadCloser
		if reader, err = f.client.RetrFrom(downloadFile, uint64(size)); err != nil{
			log.Printf("[ERROR] download file error %v \n", err)
			return
		}else{
			defer reader.Close()
			log.Println("[Info] start download ...")
//			log.Println("发送终止下载命令...")
//			if err := f.client.Abort(); err != nil{
//				log.Println("终止失败", err)
//			}

			bufReader := bufio.NewReaderSize(reader, Block_Size)
			s := make([]byte, Block_Size)
			var n int
			for{
				time.Sleep(1*time.Second)
				n, err = bufReader.Read(s)
				if _, err1 := tmpFile.Write(s[:n]); err1 != nil{
					log.Println("文件写入失败", err1)
					return err1
				}
				tmpFile.Sync()

				if err != nil{
					if err != io.EOF{
						log.Println("读取数据出错", err)
						break
					}
					log.Println("1文件下载完成")
				}
				log.Println("2文件下载完成")
			}

//			var databytes []byte
//			if databytes, err = ioutil.ReadAll(reader); err != nil{
//				log.Printf("[ERROR] occured error when downloading file. err:%v\n", err)
//				if databytes == nil && len(databytes) == 0{
//					return
//				}
//			}
//			log.Println("[Info] download success!")
//
//			var n int
//			if n, err = tmpFile.WriteAt(databytes, size); err != nil{
//				log.Printf("[ERROR] occured error when saving the download file, err:%v\n", err)
//				return
//			}
//			tmpFile.Sync()
//
//			log.Printf("[Info] save file success! file path:%s, file size:%d",tmpFile.Name(), n)
		}
	}
	return
}

// 检查文件,支持断点续传
func checkFile(file *os.File) (size int64, err error) {
	log.Printf("[Info] check local file [%s] stat\n", file.Name())
	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil{
		log.Printf("[ERROR] check local file[%s] error:%v", file.Name(), err)
		return
	}
	size = stat.Size()
	log.Printf("[Info] checked success. the size of %s is %d", file.Name(), size)
	return
}