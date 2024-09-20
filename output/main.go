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
		fmt.Printf("无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	// 获取用户输入
	fmt.Println("请输入文字（以“END”结束输入）:")
	input := getUserInput()

	// 解析输入
	records := parseInput(input)
	totalRecords := len(records)

	if totalRecords == 0 {
		fmt.Println("没有有效的数据进行写入")
		return
	}

	// 创建GBK编码转换器
	encoder := mahonia.NewEncoder("gbk")
	if encoder == nil {
		fmt.Println("无法创建GBK编码转换器")
		return
	}

	// 创建CSV写入器
	writer := csv.NewWriter(encoder.NewWriter(file))
	defer writer.Flush()

	// 创建进度条
	bar := progressbar.NewOptions(totalRecords, progressbar.OptionSetPredictTime(false))

	// 批量写入CSV文件
	batchWriteRecords(writer, records, bar)
}

// 获取用户输入
func getUserInput() string {
	var inputLines []string
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			line := scanner.Text()
			if line == endMark {
				break
			}
			inputLines = append(inputLines, line)
		} else {
			if err := scanner.Err(); err != nil {
				fmt.Printf("读取输入失败: %v\n", err)
			}
			break
		}
	}
	return strings.Join(inputLines, "\n")
}

// 解析用户输入
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

// 批量写入记录并显示进度
func batchWriteRecords(writer *csv.Writer, records [][]string, bar *progressbar.ProgressBar) {
	batchSize := 100 // 设定批量写入的大小
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		if err := writer.WriteAll(records[i:end]); err != nil {
			fmt.Printf("写入文件失败: %v\n", err)
			return
		}
		bar.Add(end - i) // 更新进度条
	}
}
