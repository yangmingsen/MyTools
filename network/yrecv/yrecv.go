package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//单文件变量
//======================================================================
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

//多文件变量
//======================================================================
type CFileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
}

type ResponseInfo struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`

	//用于自定义通信规则时使用 其他不用是 传 OK
	Status string `json:"status"`
}

var mutilFilePort string
var filePort string

const msgCloseFlag = "\r"

func init() {
	mutilFilePort = "9949"
	filePort = "8848"
}

//public变量
//======================================================================
var wg sync.WaitGroup

//公共函数
//**********************************************************************

func parseCFileInfoToJsonStr(info CFileInfo) string {
	bytes, _ := json.Marshal(info)
	return string(bytes)
}

func parseStrToCFileInfo(str string) CFileInfo {
	return parseByteToCFileInfo([]byte(str))
}

func parseByteToCFileInfo(bytes []byte) CFileInfo {
	var result CFileInfo

	json.Unmarshal(bytes, &result)
	return result
}

func parseResponseToJsonStr(res ResponseInfo) string {
	bytes, _ := json.Marshal(res)
	return string(bytes)
}

func parseByteToResponseInfo(bytes []byte) ResponseInfo {
	var result ResponseInfo

	json.Unmarshal(bytes, &result)
	return result
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

func showSingleBar(current *int64, total int64) {
	defer wg.Done()

	last := int64(0)
	avgShow := -1

	//显示条选择
	if total < SizeB {
		avgShow = B
	} else if total < SizeKB {
		avgShow = KB
	} else if total < SizeMB {
		avgShow = MB
	} else if total < SizeGB {
		avgShow = GB
	}

	for {
		var bar = ""
		//获取本次长度
		tmpLen := *current - last
		//保存上次结果
		last = *current

		percent := (float64(*current) / float64(total)) * float64(100)
		for i := 0; i < int(percent)/2; i++ {
			bar += "#"
		}
		fmt.Printf("\r总进度[%-50s]", bar)

		a := float64(*current)
		b := float64(total)

		switch avgShow {
		case B:
			fmt.Printf("%.2f%%, %.0fB => %.0fB,", (a/b)*100, a, b)
		case KB:
			fmt.Printf("%.2f%%, %.2fKB => %.2fKB,", (a/b)*100, a/1024, b/1024)
		case MB:
			fmt.Printf("%.2f%%, %.2fMB => %.2fMB,", (a/b)*100, a/1024/1024, b/1024/1024)
		case GB:
			fmt.Printf("%.2f%%, %.2fGB => %.2fGB,", (a/b)*100, a/1024/1024/1024, b/1024/1024/1024)
		}

		d := float64(tmpLen)
		if d < float64(SizeB) {
			fmt.Printf(" 平均(%.2fB/s)", d)
		} else if d < float64(SizeKB) {
			fmt.Printf(" 平均(%.2fKB/s)", d/1024)
		} else if d < float64(SizeMB) {
			fmt.Printf(" 平均(%.2fMB/s)", d/1024/1024)
		}

		if *current >= total {
			fmt.Println("接收完毕")
			break
		}
		time.Sleep(1 * time.Second)

	}

}

//单文件函数
//***********************************************************************
func recvFile(conn net.Conn, fileName string, fileSize int64) {
	defer conn.Close()

	current := int64(0)

	wg.Add(1)
	go showSingleBar(&current, fileSize)

	file, err0 := os.Create(fileName)
	if err0 != nil {
		fmt.Println("os.Create(fileName) err0:", err0)
		return
	}
	defer file.Close()

	buf := make([]byte, 4096)
	for {
		n, _ := conn.Read(buf)
		current += int64(n)

		file.Write(buf[:n])

		if current == fileSize {
			//fmt.Println("接收文件 ", fileName, " 完成")
			break
		}

	}

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
	fmt.Println("DServer Successful running in  " + ipStr + ":8849")

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

func handleRecvFile(listener net.Listener) {
	defer wg.Done()

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
				recvFile(conn, fileName, recvFileSizeInt64)
			}()

		} else {
			conn.Write([]byte("no"))
		}

	}

}

//多文件函数
//***********************************************************************

func doBindServer(ip string, port string) (bindIp net.Listener) {
	bindIp, err0 := net.Listen("tcp", ip+":"+port)
	if err0 == nil {
		return bindIp
	} else {
		fmt.Println("ip: ", ip, " Can't bind, [", err0, "]")
		return nil
	}
}

func doCreateServer(port string) net.Listener {
	listenIp := getLocalIpv4List()
	ipLen := len(listenIp)
	if ipLen == 0 {
		fmt.Println("Not available Ip Address to bind")
		return nil
	}

	//var bindIp net.Listener
	//var err0 error

	for _, theIp := range listenIp {
		//fmt.Println("ip=" + theIp)

		bindServer := doBindServer(theIp, port)
		if bindServer != nil {
			return bindServer
		}
		//bindIp, err0 = net.Listen("tcp", theIp+":"+port)
		//if err0 == nil {
		//	break
		//} else {
		//	fmt.Println("ip: ", theIp, " Can't bind, Go next!")
		//}
	}
	return nil

	//if err0 != nil {
	//	panic(err0)
	//	return nil
	//}
	//ipStr := bindIp.Addr().String()
	//fmt.Println("server Successful runing in ", ipStr)
	//
	//return bindIp
}

//发送数据
func writeMsg(conn net.Conn, msg string) {
	sendMsg := []byte(msg + msgCloseFlag)
	_, err := conn.Write(sendMsg)
	if err != nil {
		fmt.Printf("writeMsg【%s】失败\n", sendMsg)
		panic(err)
	}
	//fmt.Println("发送数据【"+msg+"】长度=", wLen)
}

func replaceSeparator(path string) string {
	const separator = os.PathSeparator
	var spa string
	if separator == '\\' {
		spa = "\\"
	} else {
		spa = "/"
	}

	var resPath = ""
	var tmpPath []string
	if strings.Contains(path, "\\") {
		tmpPath = strings.Split(path, "\\")
	} else {
		tmpPath = strings.Split(path, "/")
	}

	for i := 0; i < len(tmpPath); i++ {
		if i+1 == len(tmpPath) {
			resPath += tmpPath[i]
		} else {
			resPath += (tmpPath[i] + spa)
		}
	}

	return resPath
}

// exists returns whether the given file or directory exists or not
func fileExists(path string) (bool, os.FileInfo, error) {
	f, err := os.Stat(path)
	if err == nil {
		return true, f, nil
	}
	if os.IsNotExist(err) {
		return false, nil, nil
	}
	return true, f, err
}

// 处理目录建立
func doDirHandler(conn net.Conn) {
	defer wg.Done()

	rcvMsg := readMsg(conn)
	response := ResponseInfo{Ok: true, Message: "收到目录数据", Status: "Ok"}
	writeMsg(conn, parseResponseToJsonStr(response))

	pathArr := strings.Split(rcvMsg, "\n")
	for _, path := range pathArr {
		// os.Mkdir("abc", os.ModePerm)              //创建目录
		// os.MkdirAll("dir1/dir2/dir3", os.ModePerm)   //创建多级目录
		err := os.MkdirAll(replaceSeparator(path), os.ModePerm)
		if err != nil {
			response.Ok = false
			response.Message = err.Error()
			writeMsg(conn, parseResponseToJsonStr(response))
			fmt.Println("创建目录失败【"+path+"】", err)
		}
	}

	response.Ok = true
	response.Message = "目录建立完毕...."
	writeMsg(conn, parseResponseToJsonStr(response))

	//关闭连接
	conn.Close()

	fmt.Println("目录建立完毕....")
}

//
func doFileHandler(conn net.Conn) {
	defer wg.Done()

	for {
		//接收文件传输标志
		//可能是 s 或 c
		// 如果是c 表示结束传输
		rcvMsg := readMsg(conn)
		if rcvMsg == "c" {
			fmt.Println("Client传输完毕")
			break
		}

		//接收文件信息 fileInfo
		fileInfo := parseStrToCFileInfo(readMsg(conn))

		//转换path
		filePath := replaceSeparator(fileInfo.Path)

		//响应检查信息
		var response = ResponseInfo{Ok: true, Message: "", Status: "Ok"}
		exists, finfo, _ := fileExists(filePath)
		if exists { //如果存在
			if finfo.Size() != fileInfo.Size {
				//如果文件大小不一致 也得重传
				err := os.Remove(filePath) //删除错误文件
				if err != nil {
					log.Fatal(err)
				}

				serverSize := strconv.FormatInt(finfo.Size(), 10)
				clientSize := strconv.FormatInt(fileInfo.Size, 10)
				response.Message = "文件【" + filePath + "】:" + "远程Server文件【" + serverSize + "】与本地文件【" + clientSize + "】大小不一致准备重传. "
				response.Status = "diffSize"

				writeMsg(conn, parseResponseToJsonStr(response))

			} else { //否则 存在不传

				response.Status = "Exist"
				response.Message = "远程文件【" + filePath + "】存在, 不传."

				writeMsg(conn, parseResponseToJsonStr(response))
				continue
			}

		} else {
			// 可以传送
			writeMsg(conn, parseResponseToJsonStr(response))
		}

		//建立文件
		file, err0 := os.Create(filePath)
		defer file.Close()
		if err0 != nil {
			response.Ok = false
			response.Message = "接收方创建文件【" + filePath + "】失败:" + err0.Error()
			response.Status = "Ok"

			writeMsg(conn, parseResponseToJsonStr(response))
			continue
		} else {
			response.Ok = true
			response.Message = "Ok"
			writeMsg(conn, parseResponseToJsonStr(response))
		}

		if fileInfo.Size != 0 {
			current := int64(0)

			buf := make([]byte, 4096)
			for {
				n, _ := conn.Read(buf)
				current += int64(n)
				file.Write(buf[:n])

				if current >= fileInfo.Size {
					break
				}
			}
		}
		//关闭文件
		file.Close()
		fmt.Println("文件【"+filePath+"】接收完毕，大小【", fileInfo.Size, "B】")

		response.Ok = true
		response.Message = "Ok"
		writeMsg(conn, parseResponseToJsonStr(response))

	}
	conn.Close()

}

//以 \r 结尾
func readMsg(conn net.Conn) string {
	var rcvBytes []byte

	tmpByte := make([]byte, 1)
	var cnt int = 0

	for {
		_, err := conn.Read(tmpByte)
		if err != nil {
			panic(err)
			break
		}
		if tmpByte[0] == byte(13) {
			break
		}

		rcvBytes = append(rcvBytes, tmpByte[0])
		cnt++
	}

	return string(rcvBytes[:cnt])

}

func doAcceptMultiFileTranServer(netS net.Listener) {
	fmt.Println("MServer Successful running in ", netS.Addr().String())
	for {
		conn, err := netS.Accept()
		if err != nil {
			fmt.Println("listener.Accept() err1:", err)
			continue
		}
		rMsg := readMsg(conn)
		fmt.Println("接收到Client：" + conn.RemoteAddr().String() + " 请求类型：" + rMsg)
		if rMsg == "d" {

			wg.Add(1)
			go doDirHandler(conn)
		} else if rMsg == "f" {

			wg.Add(1)
			go doFileHandler(conn)
		} else {
			fmt.Println("======非法数据传入========")
			break
		}

	}
}

func doMultiFileTranServer() {
	//多文件传输建立port
	netS := doCreateServer(mutilFilePort)
	doAcceptMultiFileTranServer(netS)
}

func doAutoBindServer() {
	singleServer := doCreateServer("8848")
	defer singleServer.Close()
	fmt.Println("SServer Successful running in ", singleServer.Addr().String())
	if singleServer == nil {
		os.Exit(-1)
	}
	wg.Add(1)
	go handleRecvFile(singleServer)

	//do udp fun
	wg.Add(1)
	go func() {
		defer wg.Done()
		ipStr := singleServer.Addr().String()
		listenUDP(ipStr)
	}()

	doMultiFileTranServer()
}

func doSpecificBindServer(ip string) {
	singleServer := doBindServer(ip, "8848")
	defer singleServer.Close()
	fmt.Println("SServer Successful running in ", singleServer.Addr().String())
	if singleServer == nil {
		os.Exit(-1)
	}
	wg.Add(1)
	go handleRecvFile(singleServer)

	//do udp fun
	wg.Add(1)
	go func() {
		defer wg.Done()
		listenUDP(ip)
	}()

	MServer := doBindServer(ip, "9949")
	doAcceptMultiFileTranServer(MServer)

}

func doHelp() {
	fmt.Println("args must be => yrecv")
	fmt.Println("args must be => yrecv -b 本地监听ip")
}

func main() {
	args := os.Args
	argLen := len(args)

	if argLen == 1 {
		doAutoBindServer()
	} else if argLen > 1 {
		doWhat := args[1]
		if doWhat == "-b" {
			if argLen == 3 {
				ip := args[2]
				doSpecificBindServer(ip)
			} else {
				doHelp()
			}

		} else if doWhat == "-h" {
			doHelp()
		} else {
			doHelp()
		}
	}

	wg.Wait()

}
