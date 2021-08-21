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

func isIncludeIP(ip string) bool {
	excludeList := [...]string{"127", "local"}
	for _, ex := range excludeList {
		if strings.HasPrefix(ip, ex) == true {
			return true
		}
	}
	return false
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

			if isIncludeIP(ip) == false {
				sprIP := strings.Split(ip, "/")
				res = append(res, sprIP[0])
			}
		}
	}
	//fmt.Println(res)
	return res
}

//解析ipv4为数组格式
func parseUdpFormat(ip string) []byte {
	res := make([]byte, 4)
	//splitStr := strings.Split(ip,":")
	//ipStr := splitStr[0]
	//port  := splitStr[1]
	//intPort, _ := strconv.Atoi(port)

	p := strings.Split(ip, ".")
	a, _ := strconv.Atoi(p[0])
	b, _ := strconv.Atoi(p[1])
	c, _ := strconv.Atoi(p[2])
	d, _ := strconv.Atoi(p[3])

	res[0] = byte(a)
	res[1] = byte(b)
	res[2] = byte(c)
	res[3] = byte(d)

	return res
}

//udp fun
func listenUDP(ip string) {
	splitStr := strings.Split(ip, ":")
	ipStr := splitStr[0]
	//port  := splitStr[1];
	//intPort, _ := strconv.Atoi(port)
	nr := parseUdpFormat(ipStr)

	listen, err0 := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(nr[0], nr[1], nr[2], nr[3]),
		Port: 8849, //port + 1 => 8849
	})
	if err0 != nil {
		fmt.Println("UDP建立失败")
		panic(err0)
		os.Exit(-1)
	}
	fmt.Println("detect server running in " + ipStr + ":8849")

	defer listen.Close()

	//获取hostname
	hostname, _ := os.Hostname()
	var sendInfo = ipStr + "/" + hostname

	for {
		var data [32]byte
		n, addr, err := listen.ReadFromUDP(data[:])
		if err != nil {
			fmt.Println("read udp failed, err:", err)
			continue
		}

		fmt.Printf("Recv detect request data:%v addr:%v len:%v\n", string(data[:n]), addr, n)
		reallyRemoteIP := string(data[:n])
		//recvIpStr := addr.String()
		//sprArr := strings.Split(recvIpStr, ":")
		nr2 := parseUdpFormat(reallyRemoteIP)

		socket, err1 := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.IPv4(nr2[0], nr2[1], nr2[2], nr2[3]),
			Port: 8850, //探测服务器
		})
		if err1 != nil {
			fmt.Println("连接探测服务端[", reallyRemoteIP, "]失败，err:", err)
			continue
		}
		defer socket.Close()
		_, err2 := socket.Write([]byte(sendInfo)) //发送主机信息
		if err2 != nil {
			fmt.Println("响应探测数据失败，err:", err)
			continue
		}

	}

}

func main() {
	//用于参数传递
	//args := os.Args

	listenIp := getLocalIpv4List()
	ipLen := len(listenIp)
	if ipLen == 0 {
		fmt.Println("Not available Ip Address to bind")
		return
	}

	var bindIp net.Listener
	var err0 error
	for _, theIp := range listenIp {
		//fmt.Println("ip=" + theIp)
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
		ipStr := bindIp.Addr().String()
		fmt.Println("server Successful runing in ", ipStr)

		//do udp fun
		wg.Add(1)
		//接收文件
		go func() {
			defer wg.Done()
			listenUDP(ipStr)
		}()
	}

	defer bindIp.Close()

	handleRecvFile(bindIp)

	wg.Wait()
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

}
