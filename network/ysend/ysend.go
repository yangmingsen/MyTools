package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//单文件远程端口 8848
//多文件远程端口 9949

//单文件变量
//==================================================================
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
var sizeChoose int     //显示条选项
var sendFileSize int64 //发送文件大小(B)

//多文件传输变量
//=================================================================
type CFileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
}

type ResponseInfo struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

var remoteIp string
var remotePort string
var goroutineNum int
var totalFileNum int32
var finishFileNum int32
var taskNum chan CFileInfo

const msgCloseFlag = "\r"

//公共变量 函数
//=================================================================
var wg sync.WaitGroup

func init() {
	remotePort = "9949"
	goroutineNum = 1
	totalFileNum = 0
	finishFileNum = 0
}

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

func parseStrToResponseInfo(str string) ResponseInfo {
	return parseByteToResponseInfo([]byte(str))
}

func parseByteToResponseInfo(bytes []byte) ResponseInfo {
	var result ResponseInfo

	json.Unmarshal(bytes, &result)
	return result
}

//单文件函数
//******************************************************************
func sendProgressBar() {
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
			fmt.Printf("完成度 :%.2f%%, %.0fB => %.0fB", (a/b)*100, a, b)
		case KB:
			fmt.Printf("完成度:%.2f%%, %.2fKB => %.2fKB", (a/b)*100, a/1024, b/1024)
		case MB:
			fmt.Printf("完成度:%.2f%%, %.2fMB => %.2fMB", (a/b)*100, a/1024/1024, b/1024/1024)
		case GB:
			fmt.Printf("完成度:%.2f%%, %.2fGB => %.2fGB", (a/b)*100, a/1024/1024/1024, b/1024/1024/1024)
		}

		if a >= b {
			fmt.Println()
			break
		}

		time.Sleep(time.Second)
	}

}

func sendFile(conn net.Conn, filePath string) {

	file, err0 := os.Open(filePath)
	fStat, _ := file.Stat()

	go sendProgressBar()

	if err0 != nil {
		fmt.Println("os.Open err0:", err0)
		return
	}
	defer file.Close()

	var nowSize int64
	nowSize = 0 //当前下载位置

	buf := make([]byte, 4096)

	//test

	for {
		n, err1 := file.Read(buf)

		nowSize += int64(n)
		a = float64(nowSize) //更新a值

		_, err2 := conn.Write(buf[:n])

		if err1 != nil {
			if err1 == io.EOF {
				fmt.Println("send file ok! FileSize=", fStat.Size(), "   NowSize=", nowSize)
			} else {
				fmt.Println("file.Read err1:", err1)
			}

			break
		}

		if err2 != nil {
			fmt.Println("conn.Write err2:", err2)
			break
		}
	}

}

// filePath 文件路径 格式为 some.zip or some1.txt
//targetIp 远程ip 格式为 ip:port => 192.168.25.72:8848
func parseSingleFileInfo(filePath string, targetIp string) {
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
	} else {
		fmt.Println("对方拒绝了您的发送文件")
	}
}

//多文件函数
//******************************************************************
//获取文件列表
func getFileList(dirpath string) ([]CFileInfo, error) {
	var fileList []CFileInfo
	dirErr := filepath.Walk(dirpath,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}

			var fileInfo = CFileInfo{Name: f.Name(), IsDir: false, Size: f.Size(), Path: path}
			if f.IsDir() {
				fileInfo.IsDir = true
			}
			fileList = append(fileList, fileInfo)
			return nil
		})
	return fileList, dirErr
}

//发送数据
func writeMsg(conn net.Conn, msg string) bool {
	sendMsg := []byte(msg + msgCloseFlag) // 13
	_, err := conn.Write(sendMsg)
	if err != nil {
		fmt.Printf("writeMsg【%s】失败\n", sendMsg)
		panic(err)
		return false
	}
	//fmt.Println("发送数据【"+msg+"】长度=", wLen)
	return true
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

//同步文件夹
func syncDir(dirList []CFileInfo) {

	dSize := len(dirList)
	var sendMsgStr = ""
	for i := 0; i < dSize; i++ {
		if i+1 == dSize {
			sendMsgStr += dirList[i].Path
		} else {
			sendMsgStr += dirList[i].Path + "\n"
		}
	}

	//向服务器发起请求
	connectIP := (remoteIp + ":" + remotePort)
	conn, err1 := net.Dial("tcp", connectIP)
	if err1 != nil {
		fmt.Println("远程服务连接【", connectIP, "】失败")
		panic(err1)
		os.Exit(-1)
	}

	writeMsg(conn, "d")

	ok := writeMsg(conn, sendMsgStr)
	if ok == false {
		fmt.Println("发送目录数据失败【退出】")
		panic(err1)
		os.Exit(-2)
	}

	response := parseStrToResponseInfo(readMsg(conn))
	fmt.Println("收到对方服务响应：【" + response.Message + "】")

	for {
		response = parseStrToResponseInfo(readMsg(conn))
		if response.Ok == true {
			break
		} else {
			fmt.Println("对方无法建立目录数据【" + response.Message + "】")
		}
	}
	fmt.Println("同步目录数据结束...")

	defer conn.Close()
}

//显示文件传输进度条
func showSyncFileBar() {
	defer wg.Done()

	nowTime := time.Now()
	var last = int32(0)
	var total = totalFileNum
	for {
		var bar = ""
		current := finishFileNum
		tNum := current - last
		last = current

		percent := (float32(current) / float32(total)) * float32(100)
		for i := 0; i < int(percent)/2; i++ {
			bar += "#"
		}
		fmt.Printf("\r总进度[%-50s] => %.2f%% => %d个文件/s", bar, percent, tNum)
		if current >= total {
			fmt.Println()
			fmt.Println("同步文件完毕...文件数【", total, "】")
			fmt.Println("耗时:%v", time.Since(nowTime))
			break
		}

		time.Sleep(1 * time.Second)
	}
}

//同步文件
func syncFile(fileList []CFileInfo) {
	finishFileNum = 0
	totalFileNum = int32(len(fileList))
	taskNum = make(chan CFileInfo, totalFileNum)

	//任务输送
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, file := range fileList {
			taskNum <- file
		}
	}()

	// show 进度条
	wg.Add(1)
	go func() {
		showSyncFileBar()
	}()

	//多任务启动
	wg.Add(goroutineNum)
	for i := 0; i < goroutineNum; i++ {
		go func() {
			defer wg.Done()

			//向服务器发起请求
			connectIP := (remoteIp + ":" + remotePort)
			conn, err1 := net.Dial("tcp", connectIP)
			if err1 != nil {
				fmt.Println("远程服务[", connectIP, "]连接失败")
				panic(err1)
				os.Exit(-1)
			}

			//发送传输请求
			writeMsg(conn, "f")

			for {
				tLen := len(taskNum)
				if tLen == 0 { //这意味着没有文件可以传输了
					//fmt.Println("======文件传输结束======")
					writeMsg(conn, "c")
					break
				}

				//获取传输任务
				task, _ := <-taskNum

				//发送开始传输请求【s】标志
				writeMsg(conn, "s")

				//发送文件信息 fileInfo
				writeMsg(conn, parseCFileInfoToJsonStr(task))

				//读取服务响应 检查信息
				response := parseStrToResponseInfo(readMsg(conn))
				if response.Status == "Exist" { //存在 不传
					//任务完成+1
					atomic.AddInt32(&finishFileNum, 1)
					fmt.Println(response.Message)
					continue
				} else if response.Status == "diffSize" {
					fmt.Println(response.Message)
				}

				//读取接收方 建立文件信息
				response = parseStrToResponseInfo(readMsg(conn))
				if response.Ok == false { //mean 没有成功
					fmt.Println(response.Message)
					continue
				}

				//fmt.Println("收到服务请求传输文件响应：" + rcvMsg + " 开始传输文件【" + task.Path + "】")

				if task.Size != 0 {
					//传输文件
					file, err0 := os.Open(task.Path)
					if err0 != nil {
						fmt.Println("文件【" + task.Path + "】打开失败，不传输")
						panic(err0)
						break
					}

					buf := make([]byte, 4096)
					current := int64(0)
					for {
						rLen, err1 := file.Read(buf)
						current += int64(rLen)

						_, err2 := conn.Write(buf[:rLen])

						if err1 != nil {
							if err1 == io.EOF {
								break
							} else {
								fmt.Println("读取本地文件数据失败:", err1)
							}
						}

						if err2 != nil {
							fmt.Println("向服务器发送文件数据错误:", err2)
							break
						}

						if current == task.Size {
							break
						}

					}

					//关闭文件
					file.Close()
				}

				//任务完成+1
				atomic.AddInt32(&finishFileNum, 1)

				//接收服务响应完成
				response = parseStrToResponseInfo(readMsg(conn))
				//fmt.Println("接收到Server传输文件" + task.Path + "完成响应：" + rcvMsg)
				//fmt.Println()

			}
			//关闭网络流
			conn.Close()

		}()
	}

}

//做多文件传输
func doSendMutliFile(path string) {
	nowTime := time.Now()
	fileList, err := getFileList(path)
	if err != nil {
		panic(err)
		return
	}

	var dirList []CFileInfo
	var fList []CFileInfo
	for _, info := range fileList {
		if info.IsDir == true {
			dirList = append(dirList, info)
		} else {
			fList = append(fList, info)
		}
	}

	syncDir(dirList)
	fmt.Println("同步目录完毕...目录数【", len(dirList), "】 耗时:%v", time.Since(nowTime))
	syncFile(fList)

	//fmt.Println("同步文件完毕...文件数", len(fList))

}

func main() {
	args := os.Args
	argLen := len(args)

	if argLen == 3 {
		filePath := args[1]
		targetIp := args[2] + ":8848"
		parseSingleFileInfo(filePath, targetIp)
	} else if argLen >= 4 {
		doWhat := args[1]
		if doWhat == "-r" {
			sendPath := args[2]
			remoteIp = args[3]

			if argLen == 5 {
				num, _ := strconv.Atoi(args[4])
				if num <= 0 || num > 2147483647 {
					fmt.Println("goroutine数目错误,可根据机器性能设置[1,2147483647]")
					return
				}
				goroutineNum = num
			}

			doSendMutliFile(sendPath)

		} else {
			fmt.Println("args format must be => ysend -r 文件夹 目标ip地址 [goroutine数]")
		}

	} else {
		fmt.Println("args format must be => ysend 文件 目标ip地址")
		fmt.Println("args format must be => ysend -r 文件夹 目标ip地址 [goroutine数]")
	}

	wg.Wait()
}
