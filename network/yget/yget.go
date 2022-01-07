package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"
)

const (
	SizeB  int64 = 1024
	SizeKB int64 = 1048576
	SizeMB int64 = 1073741824
	SizeGB int64 = 1099511627776
)

const (
	B  = 1
	KB = 2
	MB = 3
	GB = 4
)

var a, b float64
var sizeChoose int

var wg sync.WaitGroup

func downloadProgressBar() {
	defer wg.Done()
	var c float64
	c = a
	for {
		d := a - c //获取上次 c-a之间的长度
		c = a

		if d < float64(SizeB) {
			fmt.Printf("\r平均(%.2fB/s), ", d)
		} else if d < float64(SizeKB) {
			fmt.Printf("\r平均(%.2fKB/s), ", d/1024)
		} else if d < float64(SizeMB) {
			fmt.Printf("\r平均(%.2fMB/s), ", d/1024/1024)
		}

		switch sizeChoose {
		case B:
			fmt.Printf("完成度 :%.2f%%, %dB => %dB", (a/b)*100, a, b)
		case KB:
			fmt.Printf("完成度:%.2f%%, %.2fKB => %.2fKB", (a/b)*100, a/1024, b/1024)
		case MB:
			fmt.Printf("完成度:%.2f%%, %.2fMB => %.2fMB", (a/b)*100, a/1024/1024, b/1024/1024)
		case GB:
			fmt.Printf("完成度:%.2f%%, %.2fGB => %.2fGB", (a/b)*100, a/1024/1024/1024, b/1024/1024/1024)
		}

		if a >= b {
			break
		}

		time.Sleep(time.Second)
	}

}

var c chan os.Signal

func httpGet(url string) (result string, err error) {
	splits := strings.Split(url, "/")
	nameStr := splits[len(splits)-1]
	file, err0 := os.Create(nameStr)
	if err0 != nil {
		err = err0
		return
	}

	defer file.Close()

	resp, err1 := http.Get(url)
	if err1 != nil {
		err = err1
		return
	}
	defer resp.Body.Close()

	contentLen := resp.ContentLength //获取下载文件大小

	if contentLen < SizeB {
		sizeChoose = B
	} else if contentLen < SizeKB {
		sizeChoose = KB
	} else if contentLen < SizeMB {
		sizeChoose = MB
	} else if contentLen < SizeGB {
		sizeChoose = GB
	}

	wg.Add(1)

	buf := make([]byte, 20480) //每次往缓存区 取20KB <=(20480/1024)<= 20480B
	var nowSize int64
	nowSize = 0 //当前下载位置

	go downloadProgressBar() //加载进度条

	//go func() {
	//	defer wg.Done()
	//
	//	LOOP:
	//	for {
	//		if nowSize == contentLen {break}
	//
	//		select {
	//		case s := <-c:
	//			fmt.Println("Process | get", s)
	//
	//			f1 = false
	//
	//			file.Close()
	//
	//			time.Sleep(6*time.Second)
	//
	//			os.Remove(nameStr)
	//			break LOOP
	//		default:
	//		}
	//	}
	//
	//}()

	b = float64(contentLen)
	for {
		n, err2 := resp.Body.Read(buf)

		nowSize += int64(n)
		a = float64(nowSize) //更新a值

		file.Write(buf[:n]) //向文件中写入数据

		if n == 0 {
			fmt.Println("get ok")
			break
		}
		if err2 != nil && err2 != io.EOF {
			err = err2
			return
		}
	}

	wg.Wait()

	return
}

func main() {

	args := os.Args

	if len(args) > 2 {
		fmt.Println("args to many, please reduce it")
		return
	}
	url := args[1]

	//url := "https://dldir1.qq.com/qqfile/qq/PCQQ9.1.6/25827/QQ9.1.6.25827.exe"
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	httpGet(url)

	return
}
