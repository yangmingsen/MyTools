package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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
var sizeChoose int //显示条选项

var wg sync.WaitGroup

var sendFileSize int64 //发送文件大小(B)

func sendProgressBar() {
	defer wg.Done()

	var c float64
	c = a
	for {
		d := a - c //获取上次 c-a之间的长度
		c = a

		if d < float64(SizeB) {
			fmt.Printf("平均(%.2fB/s), ", d)
		} else if d < float64(SizeKB) {
			fmt.Printf("平均(%.2fKB/s), ", d/1024)
		} else if d < float64(SizeMB) {
			fmt.Printf("平均(%.2fMB/s), ", d/1024/1024)
		}

		switch sizeChoose {
		case B:
			fmt.Printf("完成度 :%.2f%%, %.0fB => %.0fB  \n", (a/b)*100, a, b)
		case KB:
			fmt.Printf("完成度:%.2f%%, %.2fKB => %.2fKB \n", (a/b)*100, a/1024, b/1024)
		case MB:
			fmt.Printf("完成度:%.2f%%, %.2fMB => %.2fMB \n", (a/b)*100, a/1024/1024, b/1024/1024)
		case GB:
			fmt.Printf("完成度:%.2f%%, %.2fGB => %.2fGB \n", (a/b)*100, a/1024/1024/1024, b/1024/1024/1024)
		}

		if a >= b {
			break
		}

		time.Sleep(time.Second)
	}

}

func sendFile(conn net.Conn, filePath string) {

	file, err0 := os.Open(filePath)

	go sendProgressBar()

	if err0 != nil {
		fmt.Println("os.Open err0:", err0)
		return
	}
	defer file.Close()

	var nowSize int64
	nowSize = 0 //当前下载位置

	buf := make([]byte, 4096)
	for {
		n, err1 := file.Read(buf)
		nowSize += int64(n)
		a = float64(nowSize) //更新a值

		if err1 != nil {
			if err1 == io.EOF {
				//fmt.Println("send file ok!")
			} else {
				fmt.Println("file.Read err1:", err1)
			}

			return
		}

		_, err2 := conn.Write(buf[:n])
		if err2 != nil {
			fmt.Println("conn.Write err2:", err2)
			return
		}
	}
}

func main() {
	args := os.Args

	if len(args) != 3 {
		fmt.Println("args format must be => ydect sendIp resource")
		return
	}

	filePath := args[2]
	targetIp := args[1] + ":8848"

	//filePath := "G:\\resource\\codeblocks-17.12mingw-setup.exe"
	//filePath := "G:\\resource\\QQ_9.0.2.23475_setup.exe"
	//targetIp := "192.168.21.143:8848"

	//提取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("os.Stat err:", err)
		return
	}
	fileName := fileInfo.Name()    //获取文件名字
	sendFileSize = fileInfo.Size() //获取文件大小

	//设置a,b初始化值
	b = float64(sendFileSize)
	a = float64(0)

	//显示条选择
	if sendFileSize < SizeB {
		sizeChoose = B
	} else if sendFileSize < SizeKB {
		sizeChoose = KB
	} else if sendFileSize < SizeMB {
		sizeChoose = MB
	} else if sendFileSize < SizeGB {
		sizeChoose = GB
	}

	//向服务器发起请求
	conn, err1 := net.Dial("tcp", targetIp)
	if err1 != nil {
		fmt.Println("net.Dial err: ", err)
		return
	}
	defer conn.Close()

	//向目标服务器发送文件名
	_, err2 := conn.Write([]byte(fileName + "+" + strconv.FormatInt(int64(sendFileSize), 10)))
	if err2 != nil {
		fmt.Println("conn.Write err:", err)
		return
	}

	//提示用户发送的文件正在等待对方同意接受
	fmt.Println("您发送的文件正在等待对方接受,请您稍等....")

	//recv remote server 'ok' str.
	buf := make([]byte, 16)
	n, err3 := conn.Read(buf)
	if err3 != nil {
		fmt.Println("conn.Read err:", err3)
		return
	}

	if "yes" == string(buf[:n]) {
		fmt.Println("对方同意接受您的文件,正在发送中....")
		wg.Add(1) //一个 go 等待
		sendFile(conn, filePath)
		wg.Wait()
	} else {
		fmt.Println("对方拒绝了您的发送文件")
	}

}
