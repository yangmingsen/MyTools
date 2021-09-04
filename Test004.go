package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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

	for _, tp := range tmpPath {
		resPath += (tp + spa)
	}

	return resPath
}

func main() {
	var str = "a\\b\\c"
	separator := replaceSeparator(str)

	fmt.Println(separator)

}

func main3() {
	fileName := "network\\ydect\\dir\\ydect2.go"
	create, err := os.Create(fileName)
	if err != nil {
		fmt.Println("创建失败", err)
	} else {
		fmt.Println("suceess")
	}
	create.Close()
}

func main21() {
	list, err := getDirList("./network")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range list {
		fmt.Println(v)
	}

}

type FileInfo1 struct {
	name  string
	isDir bool
	size  int64
	path  string
}

// os.Mkdir("abc", os.ModePerm)              //创建目录
// os.MkdirAll("dir1/dir2/dir3", os.ModePerm)   //创建多级目录

//获取文件列表
func getDirList(dirpath string) ([]FileInfo1, error) {
	var fileList []FileInfo1
	dir_err := filepath.Walk(dirpath,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}

			var fileInfo FileInfo1 = FileInfo1{name: f.Name(), isDir: false, size: f.Size(), path: path}
			if f.IsDir() {
				fileInfo.isDir = true
			}
			fileList = append(fileList, fileInfo)
			return nil
		})
	return fileList, dir_err
}

func main1() {
	// 读取当前目录中的所有文件和子目录
	files, err := ioutil.ReadDir(`./network`)
	if err != nil {
		panic(err)
	}
	// 获取文件，并输出它们的名字
	for _, file := range files {
		isD := file.IsDir()
		if isD == true {
			fmt.Println()
		}
		println(file.Name())
	}
}
