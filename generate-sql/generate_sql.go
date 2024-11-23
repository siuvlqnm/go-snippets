package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Record struct {
	CustomerName string
	ProductName  string
	Amount       float64
	EmployeeName string
	Date         string
}

var productIDMap = map[string]string{
	"租教室":   "118597476499460096",
	"私教课":   "118597423726727168",
	"120次卡": "118597349890199552",
	"120次":  "118597349890199552",
	"90次卡":  "118597324833427456",
	"90次":   "118597324833427456",
	"60次卡":  "118597300909117440",
	"60次":   "118597300909117440", // 别名
	"30次卡":  "118597270961786880",
	"30次":   "118597270961786880", // 别名
	"年卡":    "118597151357014016",
	"半年卡":   "118597130192556032",
	"季卡":    "118597109074235392",
	"月卡":    "118597087125442560",
	"298月卡": "118597019819446272",
}

var employeeIDMap = map[string]string{
	"莎莎": "118596246427537408",
	"雪梨": "118596190848815104",
	"崔崔": "118596161929089024",
	"前后": "118596107981950976",
}

func main() {
	file, err := os.Create("output.sql")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("BEGIN TRANSACTION;\n\n")
	writer.WriteString(`INSERT INTO sales_records (
    order_no,
    user_id,
    store_id,
    actual_amount,
    submit_ts,
    customer_name,
    product_id,
    created_at
) VALUES
`)

	// 读取并解析输入数据
	isFirst := true
	i := 0 // Add counter
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--") || len(line) < 3 {
			continue
		}

		// 去掉开头的 "--"
		line = strings.TrimSpace(line[2:])

		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			continue
		}

		customerName := fields[0]
		productName := fields[1]
		amount := fields[2]
		employeeName := fields[3]
		date := fields[4]

		// 生成订单号，使用递增的计数器确保唯一性
		orderNo := fmt.Sprintf("SO%d%03d%d",
			time.Now().UnixMilli(),
			rand.Intn(1000), i)
		i++ // Increment counter

		// 转换日期为时间戳
		dateStr := fmt.Sprintf("2024-%s-%s",
			strings.Split(date, ".")[1],
			strings.Split(date, ".")[2])
		t, _ := time.Parse("2006-01-02", dateStr)
		timestamp := t.Unix() * 1000

		if !isFirst {
			writer.WriteString(",\n")
		}
		isFirst = false

		writer.WriteString(fmt.Sprintf(`    ('%s', '%s', '118595381885014016', %s, %d, '%s', '%s', datetime('now', '+8 hours'))`,
			orderNo,
			employeeIDMap[employeeName],
			amount,
			timestamp,
			customerName,
			productIDMap[productName]))
	}

	writer.WriteString(";\n\nCOMMIT;")
	writer.Flush()
}

const input = `
-- 李昭铭	半年卡	3380	崔崔	2024.10.02
`
