package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func recvFile(conn net.Conn, fileName string) {
	file, err0 := os.Create(fileName)
	if err0 != nil {
		fmt.Println("os.Create(fileName) err0:", err0)
		return
	}
	defer file.Close()

	buf := make([]byte, 4096)
	for {
		n, _ := conn.Read(buf)
		if n == 0 {
			fmt.Println("接收文件 ", fileName, " 完成")
			return
		}

		file.Write(buf[:n])
	}

	defer conn.Close()
}

func getLocalIpv4() string {
	addrs, err := net.InterfaceAddrs() //获取所有ip地址, 包含ipv4,ipv6
	if err != nil {
		panic(err)
	}

	fmt.Println(addrs)

	ipv4Addr := addrs[1].String()
	split := strings.Split(ipv4Addr, "/")

	return split[0]
}

func main() {
	listenIp := getLocalIpv4()
	listenIp += ":8848"

	fmt.Println("server runing in ", listenIp)

	//监听ip
	listener, err0 := net.Listen("tcp", listenIp)
	if err0 != nil {
		fmt.Println(" net.Listen err0:", err0)
		return
	}
	defer listener.Close()

	handleRecvFile(listener)

}

const (
	SizeB  int64 = 1024
	SizeKB int64 = 1048576
	SizeMB int64 = 1073741824
	SizeGB int64 = 1099511627776
)

var wg sync.WaitGroup

func handleRecvFile(listener net.Listener) {

	for {
		//阻塞监听client connection
		conn, err1 := listener.Accept()
		if err1 != nil {
			fmt.Println("listener.Accept() err1:", err1)
			return
		}

		//获取client 发送的文件名
		buf := make([]byte, 4096)
		n, err2 := conn.Read(buf)
		if err2 != nil {
			fmt.Println("conn.Read(buf) err2:", err2)
			return
		}

		fileInfo := string(buf[:n])
		split := strings.Split(fileInfo, "+")
		fileName := split[0]
		fileSize := split[1]

		//将string转换为64位int
		recvFileSizeInt64, _ := strconv.ParseInt(fileSize, 10, 64)

		addr := conn.RemoteAddr().String() //获得远程ip的格式为 [ip]:[port]
		splitAddr := strings.Split(addr, ":")

		//提示信息
		fmt.Print("您是否（Y/N）愿意接受对方（" + splitAddr[0] + "）向您发送文件: " + fileName + " 大小为: ")
		if recvFileSizeInt64 < SizeB {
			fmt.Printf("%.0fB\n", float64(recvFileSizeInt64))
		} else if recvFileSizeInt64 < SizeKB {
			fmt.Printf("%.2fKB\n", float64(recvFileSizeInt64)/1024)
		} else if recvFileSizeInt64 < SizeMB {
			fmt.Printf("%.2fMB\n", float64(recvFileSizeInt64)/1024/1024)
		} else if recvFileSizeInt64 < SizeGB {
			fmt.Printf("%.2fGB\n", float64(recvFileSizeInt64)/1024/1024/1024)
		}

		//用户选择
		var ch string
		fmt.Scan(&ch)
		ch = strings.ToLower(ch)

		if ch == "y" {
			//同意接受文件
			conn.Write([]byte("yes"))

			wg.Add(1)
			//接收文件
			go func() {
				defer wg.Done()
				recvFile(conn, fileName)
			}()

		} else {
			conn.Write([]byte("no"))
		}

	}

	wg.Wait()

}

//弃用
func main1() {

	listenIp := getLocalIpv4()
	listenIp += ":8848"

	fmt.Println("server runing in ", listenIp)

	//监听ip
	listener, err0 := net.Listen("tcp", listenIp)
	if err0 != nil {
		fmt.Println(" net.Listen err0:", err0)
		return
	}
	defer listener.Close()

	//阻塞监听client connection
	conn, err1 := listener.Accept()
	if err1 != nil {
		fmt.Println("listener.Accept() err1:", err1)
		return
	}
	defer conn.Close()

	//获取client 发送的文件名
	buf := make([]byte, 4096)
	n, err2 := conn.Read(buf)
	if err2 != nil {
		fmt.Println("conn.Read(buf) err2:", err2)
		return
	}

	//获取
	fileName := string(buf[:n])
	conn.Write([]byte("ok"))

	//接收文件
	recvFile(conn, fileName)

}
