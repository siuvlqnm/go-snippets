package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	var version float32
	version = 0.5
	fmt.Printf("版本：%0.2f\n", version)
	fmt.Print("请输入文件名（包含文件后缀，如：data.xlsx）: ")
	var response string
	fmt.Scanln(&response)

	// 设置日志输出到文件
	logFile, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("无法创建日志文件:", err)
		waitForEnter()
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// 使用匿名函数包装主要逻辑，以便捕获和处理panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("程序发生严重错误: %v", r)
				fmt.Println("程序发生严重错误，详情请查看error.log文件")
			}
		}()

		// 主要逻辑
		f, err := excelize.OpenFile(response)
		if err != nil {
			log.Printf("无法打开文件: %v", err)
			fmt.Println("无法打开文件，请确保文件存在且未被其他程序占用")
			return
		}
		defer f.Close()

		// 获取第一个工作表
		sheetList := f.GetSheetList()
		if len(sheetList) == 0 {
			log.Fatalf("文件中没有工作表")
		}
		sheetName := sheetList[0]

		// 获取所有行
		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Fatalf("无法获取行: %v", err)
		}

		// 插入新列
		if err := insertColumns(f, sheetName); err != nil {
			log.Fatalf("无法插入新列: %v", err)
		}

		// 缓存正则表达式
		offlineCommissionRe := regexp.MustCompile(`提(\d+)`)

		for i, row := range rows {
			if i == 0 {
				setHeader(f, sheetName)
				continue
			}

			if len(row) > 3 {
				processRow(f, sheetName, i+1, row, offlineCommissionRe)
			}
		}

		// 删除不必要的列
		removeColumns(f, sheetName, []string{"Q", "U", "U"})

		// 保存修改后的文件
		if err := f.SaveAs("output.xlsx"); err != nil {
			log.Fatalf("无法保存文件: %v", err)
		}

		fmt.Println("处理完成，结果已保存到 output.xlsx")
	}()

	waitForEnter()
}

func waitForEnter() {
	fmt.Println("按回车键退出...")
	fmt.Scanln() // 等待用户按下回车
}

func insertColumns(f *excelize.File, sheetName string) error {
	if err := f.InsertCols(sheetName, "E", 4); err != nil {
		return err
	}
	if err := f.InsertCols(sheetName, "L", 5); err != nil {
		return err
	}
	return nil
}

func setHeader(f *excelize.File, sheetName string) {
	headers := []struct {
		col, title string
	}{
		{"E1", "省"},
		{"F1", "市"},
		{"G1", "县/区"},
		{"H1", "详细地址"},
		{"L1", "合计"},
		{"M1", "美团券"},
		{"N1", "好评"},
		{"O1", "是否贴画"},
		{"P1", "线下交提成"},
	}

	for _, header := range headers {
		f.SetCellValue(sheetName, header.col, header.title)
	}
}

func processRow(f *excelize.File, sheetName string, rowIndex int, row []string, offlineCommissionRe *regexp.Regexp) {
	// 处理地址
	province, city, district, detail := parseAddress(row[3])
	setCellValues(f, sheetName, rowIndex, map[string]interface{}{
		"E": province,
		"F": city,
		"G": district,
		"H": detail,
	})

	// 处理预约时间
	if len(row) > 14 {
		f.SetCellValue(sheetName, fmt.Sprintf("X%d", rowIndex), extractDate(row[14]))
	}

	// 处理跟单备注
	if len(row) > 15 {
		f.SetCellValue(sheetName, fmt.Sprintf("Y%d", rowIndex), strings.TrimSpace(strings.Replace(row[15], "后台导入", "", -1)))
	}

	// 处理回访内容
	if len(row) > 16 {
		processFeedback(f, sheetName, rowIndex, row[16], offlineCommissionRe)
	}

	// 处理支付状态
	if len(row) > 10 {
		processPaymentStatus(f, sheetName, rowIndex, row[10])
	}

	// 计算合计
	calculateTotal(f, sheetName, rowIndex)

	// 修改订单状态和派单师傅
	if len(row) > 11 {
		processOrderStatus(f, sheetName, rowIndex, row[11], city)
	}
}

func processFeedback(f *excelize.File, sheetName string, rowIndex int, feedback string, offlineCommissionRe *regexp.Regexp) {
	if strings.Contains(feedback, "验券") && (strings.Contains(feedback, "20") || strings.Contains(feedback, "25") || strings.Contains(feedback, "30")) {
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIndex), 50)
	}

	// 处理线下交提成
	offlineCommission := extractOfflineCommission(feedback, offlineCommissionRe)
	if offlineCommission > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIndex), offlineCommission)
	}

	// 处理好评
	if strings.Contains(feedback, "好评") {
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIndex), 1)
	}

	// 是否贴画
	if strings.Contains(feedback, "贴画") {
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIndex), 1)
	}
}

func processPaymentStatus(f *excelize.File, sheetName string, rowIndex int, paymentStatus string) {
	if paymentStatus == "无需支付" {
		clearCells(f, sheetName, rowIndex, []string{"R", "S", "T"})
	} else if paymentStatus == "未支付" {
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Color: "FF0000"},
		})
		f.SetCellStyle(sheetName, fmt.Sprintf("R%d", rowIndex), fmt.Sprintf("T%d", rowIndex), style)
	}
}

func calculateTotal(f *excelize.File, sheetName string, rowIndex int) {
	repairFee, _ := f.GetCellValue(sheetName, fmt.Sprintf("R%d", rowIndex))
	materialFee, _ := f.GetCellValue(sheetName, fmt.Sprintf("S%d", rowIndex))
	meituanCoupon, _ := f.GetCellValue(sheetName, fmt.Sprintf("M%d", rowIndex))

	total := sumFees(repairFee, materialFee, meituanCoupon)
	f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIndex), total)

	switch {
	case total >= 70:
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), "成功订单")
	case total > 0 && total < 70:
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), "只收上门费")
	default:
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), "待服务")
	}
}

func processOrderStatus(f *excelize.File, sheetName string, rowIndex int, masterDispatcher, city string) {
	switch masterDispatcher {
	case "测试1":
		f.SetCellValue(sheetName, fmt.Sprintf("U%d", rowIndex), city+"-未派出")
	case "邓姐":
		f.SetCellValue(sheetName, fmt.Sprintf("U%d", rowIndex), city+"-邓姐外派")
	}
}

func removeColumns(f *excelize.File, sheetName string, cols []string) {
	for _, col := range cols {
		if err := f.RemoveCol(sheetName, col); err != nil {
			log.Printf("无法删除%s列: %v", col, err)
		}
	}
}

func setCellValues(f *excelize.File, sheetName string, rowIndex int, values map[string]interface{}) {
	for col, value := range values {
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, rowIndex), value)
	}
}

func clearCells(f *excelize.File, sheetName string, rowIndex int, cols []string) {
	for _, col := range cols {
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, rowIndex), "")
	}
}

func parseAddress(address string) (province, city, district, detail string) {
	parts := strings.SplitN(address, "市", 2)
	if len(parts) == 2 {
		// 处理 "省" 或 "自治区"
		cityParts := strings.SplitN(parts[0], "省", 2)
		if len(cityParts) != 2 {
			cityParts = strings.SplitN(parts[0], "自治区", 2)
			if len(cityParts) == 2 {
				province = cityParts[0] + "自治区"
				city = cityParts[1] + "市"
			} else {
				city = parts[0] + "市"
			}
		} else {
			province = cityParts[0] + "省"
			city = cityParts[1] + "市"
		}

		// 处理 "区" 或 "县"
		districtParts := strings.SplitN(parts[1], "区", 2)
		if len(districtParts) == 2 {
			district = districtParts[0] + "区"
			detail = districtParts[1]
		} else {
			districtParts = strings.SplitN(parts[1], "县", 2)
			if len(districtParts) == 2 {
				district = districtParts[0] + "县"
				detail = districtParts[1]
			} else {
				detail = parts[1]
			}
		}
	} else {
		detail = address
	}
	return
}

func extractDate(dateTime string) string {
	parts := strings.Split(dateTime, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return dateTime
}

func extractOfflineCommission(feedback string, re *regexp.Regexp) float64 {
	match := re.FindStringSubmatch(feedback)
	if len(match) > 1 {
		commission, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			return commission
		}
	}
	return 0
}

func sumFees(fees ...string) float64 {
	var total float64
	for _, fee := range fees {
		if value, err := strconv.ParseFloat(fee, 64); err == nil {
			total += value
		}
	}
	return total
}
