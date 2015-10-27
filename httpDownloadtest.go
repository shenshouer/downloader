package main

import (
	"net/http"
	"log"
	"strings"
	"fmt"
	"os"
)

type (
	ByteSize float64

	// 下载文件信息
	DownloadFile struct {
		FileName 		string
		ContentLength 	float64
		Sizer			string
		AcceptRanges	bool
		OriginUrl		string
	}
)

const (
	B  ByteSize = iota
	KB 		    = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB

	max_thread = 10 // 下载的最大线程数
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)
//	url := "http://ftp.ksu.edu.tw/pub/CentOS/7/isos/x86_64/CentOS-7-x86_64-DVD-1503-01.iso"
	url := "https://d1opms6zj7jotq.cloudfront.net/idea/ideaIU-14.1.5.tar.gz"
	downlaod_file, err := download_file_info(url)
	if err != nil{
		log.Fatal(err)
	}

	log.Println(downlaod_file)

	err = downloading(downlaod_file)
	if err != nil{
		log.Fatal(err)
	}
}

// 获取下载文件信息
func download_file_info(url string) (download_file *DownloadFile, err error) {
	head_resp, err := http.Head(url)

	if err != nil{
		return nil, err
	}

	download_file = &DownloadFile{}
	download_file.OriginUrl = url

	if ranges := head_resp.Header.Get("Accept-Ranges"); len(strings.Trim(ranges, " ")) > 0{
		log.Println("==>", ranges)
		download_file.AcceptRanges = true;
	}
	download_file.ContentLength = float64(head_resp.ContentLength)
	sections := strings.Split(url, "/")
	download_file.FileName = sections[len(sections) -1 ]

	if download_file.ContentLength > TB {
		download_file.Sizer = fmt.Sprintf("%.2f%s", download_file.ContentLength / TB, "TB")
	}else if download_file.ContentLength > GB {
		download_file.Sizer = fmt.Sprintf("%.2f%s", download_file.ContentLength / GB, "GB")
	}else if download_file.ContentLength > MB {
		download_file.Sizer = fmt.Sprintf("%.2f%s", download_file.ContentLength / MB, "MB")
	}else if download_file.ContentLength > KB {
		download_file.Sizer = fmt.Sprintf("%.2f%s", download_file.ContentLength / KB, "KB")
	}else{
		download_file.Sizer = fmt.Sprintf("%.2f%s", download_file.ContentLength , "B")
	}
	return
}

// 开始下载
func downloading(download_file *DownloadFile)(err error) {
	tmpFile, err := os.OpenFile("./"+download_file.FileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil{
		log.Fatal(err)
		return err
	}

	stat, err := tmpFile.Stat()
	if err != nil{
		log.Fatal(err)
		return err
	}
	tmpFile.Seek(stat.Size(), 0)

	return nil
}
