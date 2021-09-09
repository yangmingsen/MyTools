package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

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

type Bar struct {
	percent int64  //百分比
	cur     int64  //当前进度位置
	total   int64  //总进度
	rate    string //进度条
	graph   string //显示符号
	last    int64
	avgShow int
}

func (bar *Bar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) NewOptionWithGraph(start, total int64, graph string) {
	bar.graph = graph
	bar.NewOption(start, total)
}

//init
func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.last = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "#"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 1 {
		bar.rate += bar.graph //初始化进度条位置
	}

	var sizeChoose int
	//显示条选择
	if total < SizeB {
		sizeChoose = B
	} else if total < SizeKB {
		sizeChoose = KB
	} else if total < SizeMB {
		sizeChoose = MB
	} else if total < SizeGB {
		sizeChoose = GB
	}
	bar.avgShow = sizeChoose
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()

	//获取本次长度
	tmpLen := cur - bar.last

	//保存上次结果
	bar.last = cur

	if bar.percent != last { //&& bar.percent%2 == 0 {
		tmpRate := ""
		for i := 0; i < (int)(bar.percent)/2; i += 1 { //除以2的原因是输出条是 %-50s
			tmpRate += bar.graph
		}
		bar.rate = tmpRate //bar.graph//每次加载多少格子
	}
	//fmt.Printf("\r[%-50s]%3d%%  %5d/%d", bar.rate, bar.percent, bar.cur, bar.total)
	fmt.Printf("\r[%-50s]", bar.rate)

	a := float64(cur)
	b := float64(bar.total)

	switch bar.avgShow {
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

}

func (bar *Bar) Finish() {
	fmt.Println()
	fmt.Println("结束")
}

func doBar(cur *int64, total int64) {
	defer wg.Done()
	var bar Bar
	bar.NewOption(0, total)

	for {
		time.Sleep(1 * time.Second)
		bar.Play(int64(*cur))

		if *cur >= total {
			break
		}
	}

	bar.Finish()
}

var wg sync.WaitGroup

func main() {

	totalInt := 1000
	bT := time.Now() // 开始时间

	eT := time.Since(bT) // 从开始到当前所消耗的时间

	fmt.Println("Run time: ", eT)
	cur := int64(0)
	total := int64(totalInt)

	wg.Add(1) //一个 go 等待
	go func() {
		doBar(&cur, total)
	}()

	for i := 0; i <= totalInt; i += rand.Intn(50) {
		cur = int64(i)
		time.Sleep(600 * time.Millisecond)
	}

	wg.Wait()
}
