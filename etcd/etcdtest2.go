package main


import (
	"flag"
	"strings"
	"github.com/coreos/go-etcd/etcd"
	"runtime"
	"time"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
)

type (
	closeFunc func()
	Result struct {
		Count				int64
		APR					float64
		RequestPerSencode	float64
		TotalTime			int64
		Action				string
		KeySize				int
	}
)

var(
	etcdHosts	string 		= "http://127.0.0.1:2379"
	keySize		int			= 64
	action		string 		= Action_set
	counts		int			= 10

	etcd_client	*etcd.Client
	base_key	string		= "/foo"
	startTime   time.Time
	endTime	    time.Time
	requestTimeChan chan int64
	tmpRequestTimeChan chan int64
	stop chan bool
)

const(
	Action_set = "set"
	Action_get = "get"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&etcdHosts, "etcdHosts", etcdHosts, `all the server url in etcd cluster, eg: "http://10.12.1.85:2379,http://10.12.1.86:2379" default: http://127.0.0.1:2379`)
	flag.IntVar(&counts, "counts", counts, `并发进程数,默认为10`)
	flag.IntVar(&keySize, "keySize", keySize, `the length of the value of key for request, default: 64`)
	flag.StringVar(&action, "action", action, `set which action you want to test, eg: set or get. default: set`)
	flag.Parse()

	if counts == 0{
		counts = 10
	}

	requestTimeChan = make(chan int64, counts)
	tmpRequestTimeChan = make(chan int64, counts)
	stop = make(chan bool, counts)

	machines := strings.Split(etcdHosts, ",")
	etcd_client = etcd.NewClient(machines)
}

func main() {
	defer func() {
		if r := recover(); r != nil{
			exsit()
		}
	}()

	value := generateValue(keySize)
	switch action {
	case Action_get:
		fmt.Println(fmt.Sprintf("递归读取%s下所有key", base_key))
	case Action_set:
		fmt.Println(fmt.Sprintf("往%s目录下写值为%s的随机key",base_key, value))
	}

	startTime = time.Now()

	for i:=0; i< counts; i++{
		switch action {
		case Action_get:
			go MultiKeyGet(base_key, requestTimeChan, stop, tmpRequestTimeChan)
		case Action_set:
			go MultiKeySet(base_key, value, requestTimeChan, stop, tmpRequestTimeChan)
		}
	}

	// 每秒输出统计信息
	go func(timeChan chan int64){
		fmt.Println("Action\tKeySize\tcount\ttotal\t\tavg per request\trequest/Sencod")
		for{
			time.Sleep(1 * time.Second)
			var total int64
			for i:=0; i<counts; i++{
				tmp := <- timeChan
				total += tmp
			}
			sencodTime := time.Now()
			rs := &Result{
				Count: total,
				Action: action,
				KeySize: keySize,
				TotalTime: sencodTime.UnixNano() - startTime.UnixNano(),
			}
			rs.APR = float64(rs.TotalTime)/float64(rs.Count)
			rs.RequestPerSencode = float64(total)/float64(sencodTime.Unix()-startTime.Unix())
			fmt.Println(rs)
		}
	}(tmpRequestTimeChan)

	handleSignal(exsit)

}

func MultiKeySet(key, value string, requestTimeChan chan int64, stop chan bool, tmpRequestTimeChan chan int64){
	var requestTime int64 = 0
	timeout := make(chan bool, 1)
	go func() {
		for{
			time.Sleep(1 * time.Second)
			timeout <- true
		}
	}()
	for{
		requestTime++
		select {
		case <- timeout:
			tmpRequestTimeChan <- requestTime
		case flag := <-stop://停止
			if flag {
				requestTimeChan <- requestTime
				break
			}
		default:
			realkey := fmt.Sprintf("%s/%d_%d", key, time.Now().Unix(), rand.Int63n(time.Now().Unix()))
			if _, err := etcd_client.Set(realkey, value, 0); err != nil {
				panic(fmt.Errorf("Set key error: %v", err))
			}
		}
	}
}

func MultiKeyGet(key string,requestTimeChan chan int64, stop chan bool, tmpRequestTimeChan chan int64){
	var requestTime int64 = 0
	timeout := make(chan bool, 1)
	go func() {
		for{
			time.Sleep(1 * time.Second)
			timeout <- true
		}
	}()
	for{
		requestTime++
		select {
		case <- timeout:
			tmpRequestTimeChan <- requestTime
		case flag := <- stop: //停止
			if flag{
				requestTimeChan <- requestTime
				break
			}
		default:
			if _, err := etcd_client.Get(key, false, true); err != nil{
				panic(fmt.Errorf("Get key error: %v", err))
			}
		}
	}
}

// 生成value
func generateValue(keySize int) string{
	value := []rune{}
	for i := 0; i < keySize; i++{
		value = append(value, 'a')
	}
	return string(value[:])
}

func exsit(){
	endTime = time.Now()
	for i:=0; i<counts; i++{
		stop <- true
	}

	var count int64 = 0
	for i:=0; i < counts; i++{
		tmp := <- requestTimeChan
		count += tmp
	}

	rs := &Result{
		Count: count,
		Action: "Set",
		KeySize: keySize,
		TotalTime: endTime.UnixNano() - startTime.UnixNano(),
	}
	rs.APR = float64(rs.TotalTime)/float64(rs.Count)
	rs.RequestPerSencode = float64(count)/float64(endTime.Unix()-startTime.Unix())

	fmt.Println("start time",startTime.Format("2006-01-02 15:04:05.999999999 -0700 MST"))
	fmt.Println("end time",endTime.Format("2006-01-02 15:04:05.999999999 -0700 MST"))
	fmt.Println(rs)
}


func handleSignal(closeF closeFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	for sig := range c {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			closeF()
			return
		}
	}
}


func (r *Result) Total() string{
	return fmt.Sprintf("%dns", r.TotalTime)
}

func (r *Result) AvgPerRequest() string{
	return fmt.Sprintf("%.5fns", r.APR)
}

func (r *Result) String() string {
	return fmt.Sprintf(`%s	%d	%d	%s	%s	%.5f`,
		r.Action, r.KeySize, r.Count, r.Total(), r.AvgPerRequest(), r.RequestPerSencode)
}