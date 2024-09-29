package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Response represents the JSON structure returned by the server
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// 常量定义
const (
	gender       = "2"
	personID     = "130"
	followPerson = "1"
	formURL      = "https://huayu.qitawangluo.cn/manage.php/sign/bill/add"
)

// multiWriter 用于将日志输出到多个地方
func multiWriter(logFile *os.File) io.Writer {
	return io.MultiWriter(logFile, os.Stdout)
}

// init 函数用于初始化日志和随机数种子
func init() {
	// 将日志输出到文件并在控制台打印
	f, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(multiWriter(f))
}

func main() {
	// 让用户输入会话Cookie
	// var sessionCookie string
	// var start, end uint16
	// fmt.Print("请输入会话Cookie: ")
	// fmt.Scan(&sessionCookie)
	// fmt.Print("开始: ")
	// fmt.Scan(&start)
	// fmt.Print("结束: ")
	// fmt.Scan(&end)

	sessionCookie := "PHPSESSID=7qpqi5f6ajdnvgfqr15hi7ilvi;"

	// 打开CSV文件
	file, err := os.Open("members.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 读取CSV内容
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	// 处理批量记录
	err = processBatch(records[1:], sessionCookie)
	if err != nil {
		fmt.Println("Error processing batch:", err)
		return
	}
}

// processBatch 处理CSV中的每条记录
func processBatch(records [][]string, sessionCookie string) error {
	var cardID string
	for _, record := range records {
		name := record[0]
		mobile := record[1]
		money := record[2]

		// 判断金额并选择相应的卡ID
		switch money {
		case "100":
			cardID = "27"
		case "198":
			cardID = "28"
		case "888":
			cardID = "33"
		default:
			log.Printf("卡项识别错误，会员名: %s，手机号: %s，金额为: %s\n", name, mobile, money)
			continue
		}

		// 创建表单数据
		formData := url.Values{
			"row[user_id]":             {"0"},
			"row[name]":                {name},
			"row[mobile]":              {mobile},
			"row[gender]":              {gender},
			"row[card_number]":         {generateOrderNumber()},
			"row[card_id2]":            {cardID},
			"row[contract_no]":         {generateOrderNumber()},
			"row[person_id]":           {personID},
			"row[is_follow_person_id]": {followPerson},
			"row[business_allot]":      {money},
			"row[eid]":                 {""},
			"row[credit_amount]":       {""},
			"row[give_day]":            {""},
			"row[give_number]":         {""},
			"row[give_remark]":         {""},
			"row[remark]":              {""},
		}

		// 提交表单
		if err := submitForm(formURL, formData, sessionCookie); err != nil {
			log.Printf("信息录入失败，会员名: %s，手机号: %s, 错误: %v\n", name, mobile, err)
		} else {
			fmt.Printf("信息录入成功，会员名: %s，手机号: %s\n", name, mobile)
		}

		// 避免频繁请求，等待一秒
		time.Sleep(1 * time.Second)
	}
	return nil
}

// submitForm 提交表单数据并检查结果
func submitForm(submitURL string, formData url.Values, sessionCookie string) error {
	// 创建HTTP POST请求
	req, err := http.NewRequest("POST", submitURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", sessionCookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("提交失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 判断返回的响应是否为JSON格式
	if isJSON(body) {
		var response Response
		if err := json.Unmarshal(body, &response); err != nil {
			return fmt.Errorf("解析JSON失败: %v", err)
		}
		if response.Code != 1 {
			return fmt.Errorf("提交卡失败: %s", response.Msg)
		}
	} else {
		if strings.Contains(string(body), "system-message success") {
			return nil // 提交成功
		} else if strings.Contains(string(body), "system-message error") {
			return fmt.Errorf("错误信息: %s", extractErrorMessage(string(body)))
		} else {
			return fmt.Errorf("未知的响应: %s", string(body))
		}
	}

	return nil
}

// extractErrorMessage 提取错误信息
func extractErrorMessage(html string) string {
	errorPattern := regexp.MustCompile(`(?s)<div class="system-message error">.*?<h1>(.*?)</h1>.*?</div>`)
	matches := errorPattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "unknown error"
	}
	return matches[1]
}

// isJSON 检查响应是否为JSON格式
func isJSON(data []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(data, &js) == nil
}

// generateOrderNumber 生成订单号
func generateOrderNumber() string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return time.Now().Format("20060102150405") + fmt.Sprintf("%04d", rng.Intn(10000))
}
