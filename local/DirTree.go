package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var rootNameLen = -1

func print3(len int) {
	for i := 0; i < len; i++ {
		print(" ")
	}
}

func errFunc(info string, err error) bool {
	if err != nil {
		log.Println("info:", info, " error:", err)
		return false
	}
	return true
}
func getDirFileList(path string) []os.FileInfo {
	curPath, err := os.OpenFile(path, os.O_RDONLY, os.ModeDir)
	if errFunc("opend dir error in openDir..", err) == false {
		os.Exit(2)
	}

	dir1, err2 := curPath.Readdir(-1)
	if !errFunc("read dir error in openDir..", err2) {
		os.Exit(2)
	}
	curPath.Close()
	return dir1
}

var arr [100]int

func drawTree2(path string, idx int) {
	fileList := getDirFileList(path)

	if len(fileList) > 0 {
		for i := 0; i < idx; i++ {
			print3(arr[i])
			print("|")
			//if i == 0 || i+1 == idx {
			//	print3(arr[i])
			//	print("|")
			//} else {
			//	print3(arr[i])
			//	print(" ")
			//}

		}
		println("")
	} else {
		return
	}

	i1 := 0
	for _, name := range fileList {
		for i := 0; i < idx-1; i++ {
			print3(arr[i])
			print("|")
			//if i==0 {
			//	print3(arr[i])
			//	print("|")
			//} else {
			//	print3(arr[i])
			//	print(" ")
			//}

		}
		print3(arr[idx-1])
		print("+--- ")
		println(name.Name()) //打印目录名

		if name.IsDir() {
			arr[idx] = 4 + (len(name.Name()) / 2) //空格长度
			drawTree2(path+"/"+name.Name(), idx+1)
		}

		//打印文件名换行后的竖线
		if i1+1 < len(fileList) {
			for i := 0; i < idx-1; i++ {

				if i == 0 {
					print3(arr[i])
					print("|")
				} else {
					print3(arr[i])
					print(" ")
				}

			}
			print3(arr[idx-1])
			println("|")
		}

		i1++

	}
}

func printSpace(n int) {
	for i := 0; i < n; i++ {
		print(" ")
	}
}

var spaceNum [100]int
var ifPrintVerticalline [100]bool

func drawTree3(path string, idx int) {
	fileList := getDirFileList(path)
	fileListLen := len(fileList) //获取文件/文件夹列表长度

	if fileListLen > 0 {
		for i := 0; i < idx; i++ {
			printSpace(spaceNum[i])
			if ifPrintVerticalline[i] {
				print("|")
			}
		}
		println("")
	} else {
		return
	}

	for ii, name := range fileList {

		for i := 0; i < idx; i++ {
			printSpace(spaceNum[i])
			if ifPrintVerticalline[i] && i+1 < idx {
				print("|")
			}
		}
		print("+--- ", name.Name(), "\n")

		if ii+1 == fileListLen {
			ifPrintVerticalline[idx] = false
		}

		if name.IsDir() {
			spaceNum[idx] = 4 + (len(name.Name()) / 2) //空格长度
			if ii+1 == fileListLen {
				ifPrintVerticalline[idx] = false
			} else {
				ifPrintVerticalline[idx] = true
			}
			drawTree3(path+"/"+name.Name(), idx+1)
		}

	}
}

func main() {
	args := os.Args
	var path string

	//如果只有3个选项 那么查找路径默认为当前路径
	if len(args) == 1 {
		workDir, err := os.Getwd() //获取当前工作路径
		if !errFunc("open wordir error in main", err) {
			return
		}
		path = workDir
	} else if len(args) == 2 {
		path = args[1]
	}

	splits := strings.Split(path, "\\")
	rootName := splits[len(splits)-1]

	rootNameLen = len(rootName)
	fmt.Println(rootName)

	arr[0] = len(rootName) / 2

	spaceNum[0] = len(rootName) / 2
	ifPrintVerticalline[0] = true

	drawTree3(path, 1)

}
