package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type CFileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
}

var mutilFilePort string
var filePort string
var wg sync.WaitGroup

const msgCloseFlag = "\r"

func init() {
	mutilFilePort = "9949"
	filePort = "8848"
}

//排除ip
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
	return res
}

func doCreateServer(port string) net.Listener {
	listenIp := getLocalIpv4List()
	ipLen := len(listenIp)
	if ipLen == 0 {
		fmt.Println("Not available Ip Address to bind")
		return nil
	}

	var bindIp net.Listener
	var err0 error

	for _, theIp := range listenIp {
		//fmt.Println("ip=" + theIp)
		bindIp, err0 = net.Listen("tcp", theIp+":"+port)
		if err0 == nil {
			break
		} else {
			fmt.Println("ip: ", theIp, " Can't bind, Go next!")
		}
	}

	if err0 != nil {
		panic(err0)
		return nil
	}
	ipStr := bindIp.Addr().String()
	fmt.Println("server Successful runing in ", ipStr)

	return bindIp
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
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func parseToJsonStr(info CFileInfo) string {
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

// 处理目录建立
func doDirHandler(conn net.Conn) {
	defer wg.Done()

	rcvMsg := readMsg(conn)

	writeMsg(conn, "收到目录数据")

	pathArr := strings.Split(rcvMsg, "\n")
	for _, path := range pathArr {
		// os.Mkdir("abc", os.ModePerm)              //创建目录
		// os.MkdirAll("dir1/dir2/dir3", os.ModePerm)   //创建多级目录
		err := os.MkdirAll(replaceSeparator(path), os.ModePerm)
		if err != nil {
			fmt.Println("创建目录失败【"+path+"】", err)
		}
	}
	fmt.Println("目录建立完毕....")
}

//
func doFileHandler(conn net.Conn) {
	defer wg.Done()

	for {
		//接收文件传输标志
		rcvMsg := readMsg(conn)
		if rcvMsg == "c" {
			fmt.Println("Client传输完毕")
			break
		}

		//接收文件信息
		fileInfo := parseStrToCFileInfo(readMsg(conn))
		//响应收到信息
		exists, _ := fileExists(fileInfo.Path)
		if exists { //存在不传
			writeMsg(conn, "no")
			continue
		} else {
			writeMsg(conn, "ok")
		}

		//转换path
		filePath := replaceSeparator(fileInfo.Path)
		file, err0 := os.Create(filePath)
		defer file.Close()
		if err0 != nil {
			fmt.Println("创建本地文件["+filePath+"]失败:", err0)
			writeMsg(conn, "continue")
			continue
		} else {
			writeMsg(conn, "next")
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
		fmt.Println("文件【"+filePath+"】接收完毕，大小【", fileInfo.Size, "】")

		writeMsg(conn, "ok2")

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

func doMultiFileTranServer() {
	//多文件传输建立port
	netS := doCreateServer(mutilFilePort)
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

func main() {
	doMultiFileTranServer()
	wg.Wait()
}
