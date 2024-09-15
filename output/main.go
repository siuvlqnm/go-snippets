package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/schollz/progressbar/v3"
)

const (
	filename = "output.csv"
	endMark  = "END"
)

func main() {
	// 打开或创建文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()

	// 创建GBK编码转换器
	encoder := mahonia.NewEncoder("gbk")
	if encoder == nil {
		fmt.Println("无法创建GBK编码转换器")
		return
	}

	// 创建CSV写入器
	writer := csv.NewWriter(encoder.NewWriter(file))
	defer writer.Flush()

	// 获取用户输入
	fmt.Println("请输入文字（以“END”结束输入）:")
	input := getUserInput()

	// 解析输入
	records := parseInput(input)
	totalRecords := len(records)

	// 创建进度条
	bar := progressbar.Default(int64(totalRecords))

	// 写入CSV文件
	for _, record := range records {
		err := writer.Write(record)
		if err != nil {
			fmt.Println("写入文件失败:", err)
			return
		}
		bar.Add(1)
	}
}

func getUserInput() string {
	var inputLines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == endMark {
			break
		}
		inputLines = append(inputLines, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取输入失败:", err)
	}

	return strings.Join(inputLines, "\n")
}

func parseInput(input string) [][]string {
	lines := strings.Split(input, "\n")
	var records [][]string
	for _, line := range lines {
		// 按空格分隔字段
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			records = append(records, fields)
		}
	}
	return records
}
