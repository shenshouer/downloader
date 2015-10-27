// 多线程下载
// 限制线程数
// 限制单个线程下载数据的大小
package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"strings"
	"os"
	"log"
	"fmt"
//	"time"
	"bytes"
	"io"
)

var wg sync.WaitGroup

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

	limit 		= 10		// 最大线程数
	PartSize 	= 1 * MB	// 每个线程下载的最大数据块
)

type (
	ByteSize float64
	PartDownloadInfo struct {
		Url			string
		Index 		int
		OffsetMin 	int
		OffsetMax 	int
	}
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)
//	url := "http://download.firefox.com.cn/releases/firefox/40.0/zh-CN/Firefox-latest.dmg"
//	url := "https://d1opms6zj7jotq.cloudfront.net/idea/ideaIU-14.1.5.tar.gz"
	url := "http://ftp.ksu.edu.tw/pub/CentOS/7/isos/x86_64/CentOS-7-x86_64-DVD-1503-01.iso"
	res, _ := http.Head(url)

	sections := strings.Split(url, "/")
	fileName := sections[len(sections) -1 ]
	tmpFile, err := os.OpenFile("./"+fileName, os.O_RDWR|os.O_CREATE, 0666)
	defer tmpFile.Close()
	if err != nil{
		log.Fatal(err)
	}

	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0])
	log.Println("==>>", length, "Header", maps)
	len_sub := length / PartSize
	diff := length % PartSize

	// 对象缓存池
	partDownloadInfosChan := make(chan PartDownloadInfo, len_sub)
	// 实际需要的线程数量
	goroutineNum := limit
	if len_sub < limit{
		goroutineNum = len_sub
	}


	log.Println(fmt.Sprintf("初始化%d个线程池,共需%d个线程下载", goroutineNum, len_sub))

	// 创建容量为limit的线程池
	for j := 0; j < goroutineNum; j++{
		go download(partDownloadInfosChan, tmpFile)
	}

	for i := 0; i < len_sub ; i++ {
		wg.Add(1)

		max := PartSize * (i + 1)
		if (i == len_sub - 1) {
			max += diff
		}

		partDownloadInfo := PartDownloadInfo{
			Url: url,
			Index: i,
			OffsetMin: PartSize * i,
			OffsetMax : max,
		}
		partDownloadInfosChan <- partDownloadInfo
	}
	wg.Wait()
}

func download(partDownloadInfosChan <- chan PartDownloadInfo, downloadFile *os.File){
	for partDownloadInfo := range partDownloadInfosChan {

		client := &http.Client {}
		req, err := http.NewRequest("GET", partDownloadInfo.Url, nil)
		if err != nil{
			log.Fatal(err)
		}
		range_header := "bytes=" + strconv.Itoa(partDownloadInfo.OffsetMin) +"-" + strconv.Itoa(partDownloadInfo.OffsetMax-1)
		req.Header.Add("Range", range_header)
		log.Printf("正在下载第%d块大小为%.3fMB文件,请求头为 %s \n",partDownloadInfo.Index, float64(partDownloadInfo.OffsetMax-partDownloadInfo.OffsetMin)/MB, range_header)
		resp,_ := client.Do(req)
		defer resp.Body.Close()

		log.Println("=========>> ",resp.ContentLength, resp.StatusCode, resp.Proto)
		actual_part_size := partDownloadInfo.OffsetMax-partDownloadInfo.OffsetMin
		if resp.ContentLength == int64(actual_part_size) {
			reader, _ := ioutil.ReadAll(resp.Body)
			downloadFile.WriteAt(reader, int64(partDownloadInfo.OffsetMin))
		}else{
			buf := bytes.NewBuffer(make([]byte, actual_part_size))
			n, err := buf.ReadFrom(resp.Body)
			if err != nil && err != io.EOF{
				log.Fatal(err)
			}

			log.Printf("已经读取了%n个字节\n", n)
			downloadFile.WriteAt(buf.Bytes(), int64(partDownloadInfo.OffsetMin))
		}

		wg.Done()
	}
}
