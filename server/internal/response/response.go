package response

import (
	"bytes"
	"go-fin-server/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
)

// Data 返回成功的 JSON 响应
func Data(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	})
}

// DataMsg 返回成功的 JSON 响应
func DataMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  msg,
		"data": data,
	})
}

func DataExpand(c *gin.Context, data interface{}, other map[string]interface{}) {
	c.JSON(200, utils.MergeMaps(gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	}, other))
}

func PageData(c *gin.Context, data interface{}, total int64) {
	c.JSON(200, gin.H{
		"total": total,
		"code":  200,
		"msg":   "success",
		"rows":  data,
	})
}

func Error(c *gin.Context, message string) {
	c.JSON(200, gin.H{
		"code": 400,
		"msg":  message,
		"data": nil,
	})
}

// Error 返回错误的 JSON 响应
func ErrorCode(c *gin.Context, code int, message string) {
	c.JSON(200, gin.H{
		"code": code,
		"msg":  message,
		"data": nil,
	})
}

// ExportExcel 导出 Excel 数据并写入响应体
func ExportExcel(c *gin.Context, filename string, file *xlsx.File) {
	// 创建内存缓冲区
	buffer := new(bytes.Buffer)

	// 将 xlsx.File 写入到缓冲区
	if err := file.Write(buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	// 使用缓冲区的内容创建 io.Reader
	reader := bytes.NewReader(buffer.Bytes())

	// 直接使用 http.ServeContent 将 Excel 数据写入响应体
	http.ServeContent(c.Writer, c.Request, filename, time.Now(), reader)
}

// SetOperTitle 设置当前操作标题（供操作日志中间件记录）
func SetOperTitle(c *gin.Context, title string) {
	c.Set("operTitle", title)
}
