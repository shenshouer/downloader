package main

// 多线程下载文件
// 限制线程数,不限制每个线程的下载块大小
import (
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"strings"
	"os"
	"log"
)

var wg sync.WaitGroup

func main() {
	url := "http://download.firefox.com.cn/releases/firefox/40.0/zh-CN/Firefox-latest.dmg"
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
	limit := 10
	len_sub := length / limit
	diff := length % limit

	for i := 0; i < limit ; i++ {
		wg.Add(1)

		min := len_sub * i
		max := len_sub * (i + 1)

		if (i == limit - 1) {
			max += diff
		}

		go func(min int, max int, i int) {
			client := &http.Client {}
			req, _ := http.NewRequest("GET", url, nil)
			range_header := "bytes=" + strconv.Itoa(min) +"-" + strconv.Itoa(max-1)
			req.Header.Add("Range", range_header)
			resp,_ := client.Do(req)
			defer resp.Body.Close()
			reader, _ := ioutil.ReadAll(resp.Body)
			log.Println("min", min, "max", max, "i", i)
			tmpFile.WriteAt(reader, int64(min))
			wg.Done()

		}(min, max, i)
	}
	wg.Wait()
}