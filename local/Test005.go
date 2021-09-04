package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Sex  int    `json:"sex"`
}

type CFileInfo struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
	Path  string `json:"path"`
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

func main() {

	p1 := Person{
		Id:   "11",
		Name: "Black",
		Age:  10,
		Sex:  1,
	}
	p2 := Person{
		Id:   "12",
		Name: "Green",
		Age:  15,
		Sex:  2,
	}
	ss := CFileInfo{Path: "sddf", IsDir: false, Size: 35, Name: "343"}

	fmt.Println("name=" + parseToJsonStr(ss))

	bb1, _ := json.Marshal(p1)
	bb2, _ := json.Marshal(p2)
	json1 := string(bb1)
	json2 := string(bb2)
	fmt.Println(json1)
	fmt.Println(json2)
	var person1, person2 Person
	json.Unmarshal([]byte(json1), &person1)
	json.Unmarshal([]byte(json2), &person2)
	fmt.Println(person1)
	fmt.Println(person2)
}
