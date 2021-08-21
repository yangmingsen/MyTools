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

// 获取单个ip地址，具有固定性。 容易出错
func getLocalIpv4() string {
	addrs, err := net.InterfaceAddrs() //获取所有ip地址, 包含ipv4,ipv6
	if err != nil {
		panic(err)
	}

	//fmt.Println(addrs)
	for _, addr := range addrs {
		fmt.Println(addr)
	}

	ipv4Addr := addrs[1].String()
	split := strings.Split(ipv4Addr, "/")

	return split[0]
}

// 获取所有ipv4绑定列表
func getLocalIpv4List() []string {
	addrs, err := net.InterfaceAddrs() //获取所有ip地址, 包含ipv4,ipv6
	if err != nil {
		panic(err)
	}

	//addrsLen := len(addrs);

	res := make([]string, 0)

	//fmt.Println(addrs)
	for _, addr := range addrs {
		//fmt.Println(addr)
		ip := addr.String()
		contains := strings.Contains(ip, ".")
		if contains {
			res = append(res, strings.Split(ip, "/")[0])
		}
	}
	fmt.Println(res)
	return res
}

func main() {
	listenIp := getLocalIpv4List()

	var bindIp net.Listener
	var err0 error
	for _, theIp := range listenIp {
		fmt.Println("ip=" + theIp)
		bindIp, err0 = net.Listen("tcp", theIp+":8848")
		if err0 == nil {
			goto goThe
		} else {
			fmt.Println("ip: ", theIp, " Can't bind, Go next!")
		}
	}

goThe:
	{
		if err0 != nil {
			panic(err0)
			return
		}
		fmt.Println("server Successful runing in ", bindIp.Addr().String())
	}

	defer bindIp.Close()

	handleRecvFile(bindIp)
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
