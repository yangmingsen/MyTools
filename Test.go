package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	SizeB  int64 = 1024
	SizeKB int64 = 1048576
	SizeMB int64 = 1073741824
	SizeGB int64 = 1099511627776
)

func getClientIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", nil
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}

	return "", errors.New("Can not find the client ip address!")

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
	fmt.Println(getLocalIpv4())
}

func main1() {
	var a string

	var sizeInt64 int64
	sizeInt64 = int64(SizeKB * 4)

	var hitInfo string

	if sizeInt64 < SizeB {
		hitInfo += strconv.FormatInt(int64(sizeInt64), 10) + "B"
	} else if sizeInt64 < SizeKB {
		hitInfo += strconv.FormatFloat(float64(sizeInt64)/1024, 'E', -1, 64) + "KB"
	} else if sizeInt64 < SizeMB {
		sizeFloat64 := float64(sizeInt64)
		sizeFloat64 = sizeFloat64 / 1024 / 1024
		fmt.Printf("%.2fMB", sizeFloat64)
		//sizeF64 := float64(recvFileSizeInt64)
		// :=strconv.FormatFloat(float64(recvFileSizeInt64)/1024/1024,'E',-1,64)+"MB"
	} else if sizeInt64 < SizeGB {
		hitInfo += strconv.FormatFloat(float64(sizeInt64)/1024/1024/1024, 'E', -1, 64) + "GB"
	}

	if a == "y" {
		fmt.Println("yes")
	}
	fmt.Println(a)
}
