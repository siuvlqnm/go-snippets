package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	sessionCookie := "PHPSESSID=njnd8l8rv4gif9864p9k65fh8l;"

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

	err = processBatch(records[1:], sessionCookie)

	if err != nil {
		fmt.Println("Error processing batch:", err)
		return
	}
}

func processBatch(records [][]string, sessionCookie string) error {
	formURL := "https://huayu.qitawangluo.cn/manage.php/sign/bill/add"
	var cardID string

	for _, record := range records {
		name := record[0]
		mobile := record[1]
		money := record[2]

		// 其他参数可根据需要添加
		if money == "100" {
			cardID = "27"
		} else if money == "198" {
			cardID = "28"
		} else if money == "298" {
			cardID = "24"
		} else {
			return fmt.Errorf("卡项识别错误，会员名: %s，手机号: %s，错误信息: %v", name, mobile, "表格内金额为: "+money)
		}

		// 创建表单数据
		formData := url.Values{
			"row[user_id]":             {"0"},
			"row[name]":                {name},
			"row[mobile]":              {mobile},
			"row[gender]":              {"2"},
			"row[card_number]":         {generateOrderNumber()},
			"row[card_id2]":            {cardID},
			"row[eid]":                 {""},
			"row[credit_amount]":       {""},
			"row[give_day]":            {""},
			"row[give_number]":         {""},
			"row[give_remark]":         {""},
			"row[contract_no]":         {generateOrderNumber()},
			"row[person_id]":           {"121"},
			"row[is_follow_person_id]": {"1"},
			"row[help_person_id]":      {""},
			"row[business_allot]":      {money},
			"row[remark]":              {""},
		}

		// 提交表单并检查结果
		err := submitForm(formURL, formData, sessionCookie)
		if err != nil {
			return fmt.Errorf("信息录入失败，失败信息: %v，会员名: %s，手机号: %s", err, name, mobile)
		} else {
			fmt.Printf("信息录入成功，会员名: %s，手机号: %s\n", name, mobile)
		}

		// 等待一秒钟
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

	// 执行请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("form submission failed with status code: %d, body: %s", resp.StatusCode, body)
	}

	// 调试输出响应体
	// fmt.Printf("Response Status: %s\n", resp.Status)
	// fmt.Println("Response Headers:")
	// for key, values := range resp.Header {
	// 	for _, value := range values {
	// 		fmt.Printf("%s: %s\n", key, value)
	// 	}
	// }
	// fmt.Printf("Response Body: %s\n", body)

	if isJSON(body) {
		// 解析JSON响应体
		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			return fmt.Errorf("failed to parse response JSON: %s", err)
		}
		if response.Code == 0 {
			return fmt.Errorf("send card failed: %s", response.Msg)
		}
	} else {
		if strings.Contains(string(body), "system-message success") {

		} else if strings.Contains(string(body), "system-message error") {
			// 提取错误信息
			errorMessage := extractErrorMessage(string(body))
			return fmt.Errorf("%s", errorMessage)
		} else {
			return fmt.Errorf("unexpected response: %s", body)
		}
	}

	return nil
}

// extractErrorMessage 提取错误信息
func extractErrorMessage(html string) string {
	// errorPattern := regexp.MustCompile(`<div class="system-message error">.*?<h1>(.*?)</h1>.*?`)
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
	return time.Now().Format("20060102150405") + fmt.Sprintf("%04d", rand.Intn(10000))
}
