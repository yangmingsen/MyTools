package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

//解析ipv4为数组格式
func parseUdpFormat(ip string) []byte {
	res := make([]byte, 4)

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

func doDetect(detectCondition string, listenIP string) {
	nr2 := parseUdpFormat(detectCondition)
	for x := 2; x < 255; x++ {
		var host = byte(x)
		socket, err1 := net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.IPv4(nr2[0], nr2[1], nr2[2], host),
			Port: 8849, //对应yrecv
		})
		if err1 != nil {
			fmt.Println("连接探测服务端[", nr2, "]失败，err:", err1)
			continue
		}
		defer socket.Close()
		_, err2 := socket.Write([]byte(listenIP)) //发送主机信息

		if err2 != nil {
			fmt.Println("发送探测数据失败，err:", err2)
		}
	}

}

func main() {
	args := os.Args

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
			break
		} else {
			fmt.Println("ip: ", theIp, " Can't bind, Go next!")
		}
	}

	listenIp1 := bindIp.Addr().String()

	detectIp := args[1]
	reallyListenIP := strings.Split(listenIp1, ":")[0]

	nr := parseUdpFormat(reallyListenIP)

	listen, err0 := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(nr[0], nr[1], nr[2], nr[3]),
		Port: 8850, //port + 1 => 8849
	})
	if err0 != nil {
		fmt.Println("UDP建立失败")
		panic(err0)
		os.Exit(-1)
	}
	fmt.Println("detect recv server running in " + listenIp1)

	defer listen.Close()

	//do udp fun
	wg.Add(1)
	go func() {
		defer wg.Done()
		doDetect(detectIp, reallyListenIP) //detectip
	}()

	for {
		var data [64]byte
		n, _, err := listen.ReadFromUDP(data[:]) // 接收数据
		if err != nil {
			fmt.Println("read udp failed, err:", err)
			continue
		}
		fmt.Println("find: " + string(data[:n]))

	}

	wg.Wait()

}
