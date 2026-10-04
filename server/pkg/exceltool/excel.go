package exceltool

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

// CreateExcelFile 创建一个新的 Excel 文件并填充数据
func CreateExcelFile(headers []string, data [][]string, sheetName string) (*xlsx.File, error) {
	if sheetName == "" {
		sheetName = ""
	}
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)
	if err != nil {
		return nil, err
	}

	// 添加表头
	headerRow := sheet.AddRow()
	for _, header := range headers {
		cell := headerRow.AddCell()
		cell.Value = header
	}

	// 添加数据
	for _, rowData := range data {
		row := sheet.AddRow()
		for _, cellData := range rowData {
			cell := row.AddCell()
			cell.Value = cellData
		}
	}

	return file, nil
}

// ReadExcelFile 读取 Excel 文件并返回处理后的数据
func ReadExcelFile(fileContent []byte) ([][]string, error) {
	// 读取 Excel 文件
	excelFile, err := xlsx.OpenBinary(fileContent)
	if err != nil {
		return nil, fmt.Errorf("无法打开 Excel 文件: %v", err)
	}

	// 存储处理后的数据
	var result [][]string

	// 处理 Excel 数据，这里简单输出每个单元格的值
	for _, sheet := range excelFile.Sheets {
		for rowIndex, row := range sheet.Rows {
			// 跳过表头数据（第一行）
			if rowIndex == 0 {
				continue
			}

			var rowData []string
			for _, cell := range row.Cells {
				cellValue := cell.String()
				rowData = append(rowData, cellValue)
				//fmt.Printf("%s\t", cellValue)
			}
			//fmt.Println()
			result = append(result, rowData)
		}
	}

	return result, nil
}
