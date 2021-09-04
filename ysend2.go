package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type CFileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
}

var remoteIp string
var remotePort string
var goroutineNum int
var totalFileNum int32
var finishFileNum int32
var taskNum chan CFileInfo
var wg sync.WaitGroup

const msgCloseFlag = "\r"

func init() {
	remotePort = "9949"
	goroutineNum = 1
	totalFileNum = 0
	finishFileNum = 0
}

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

func parseToJsonStr(info CFileInfo) string {
	bytes, _ := json.Marshal(info)
	return string(bytes)
}

func parseByteToCFileInfo(bytes []byte) CFileInfo {
	var result CFileInfo

	json.Unmarshal(bytes, &result)
	return result
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
	conn, err1 := net.Dial("tcp", remoteIp+":"+remotePort)
	if err1 != nil {
		fmt.Println("远程服务连接失败")
		panic(err1)
		os.Exit(-1)
	}

	writeMsg(conn, "d")

	ok := writeMsg(conn, sendMsgStr)
	if ok == false {
		panic(err1)
		os.Exit(-2)
	}

	rcvMsg := readMsg(conn)
	fmt.Println("收到服务服务响应：" + rcvMsg)

	defer conn.Close()

}

func showSyncFileBar() {
	defer wg.Done()

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
			break
		}

		time.Sleep(1 * time.Second)
	}
}

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
			conn, err1 := net.Dial("tcp", remoteIp+":"+remotePort)
			if err1 != nil {
				fmt.Println("远程服务连接失败")
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

				//发送开始传输请求
				writeMsg(conn, "s")

				//发送文件信息
				writeMsg(conn, parseToJsonStr(task))

				//读取服务响应
				rcvMsg := readMsg(conn)
				if rcvMsg == "no" { //存在 不传
					continue
				}

				rcvMsg = readMsg(conn)
				if rcvMsg == "continue" {
					fmt.Println("远程Server创建文件失败."+"文件【", task.Path, "】传输失败")
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

						if err1 != nil {
							if err1 == io.EOF {
								break
							} else {
								fmt.Println("读取本地文件数据失败:", err1)
							}
						}

						_, err2 := conn.Write(buf[:rLen])
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
				rcvMsg = readMsg(conn)
				//fmt.Println("接收到Server传输文件" + task.Path + "完成响应：" + rcvMsg)
				//fmt.Println()

			}
			//关闭网络流
			conn.Close()

		}()
	}

}

func main() {
	args := os.Args
	argLen := len(args)
	fmt.Println(args)
	if argLen != 3 {
		fmt.Println("argError, maybe => yrecv2 remoteIP ./resouces")
	}
	remoteIp = args[1]
	path := args[2]
	//remotePort = "9949"

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
	fmt.Println("同步目录完毕...目录数", len(dirList))
	syncFile(fList)

	wg.Wait()
	fmt.Println("同步文件完毕...目录数", len(fList))

}
