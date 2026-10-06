package middleware

import (
	"bytes"
	"go-fin-server/internal/config"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/addressutils"
	"go-fin-server/pkg/utils/httputils"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// maxRequestBodySize 请求体最大读取大小（1MB），防止恶意大请求体占用内存
const maxRequestBodySize = 1 << 20

// CustomWriter 自定义的 ResponseWriter，用于捕获响应体
type CustomWriter struct {
	gin.ResponseWriter
	Body []byte
}

func (w *CustomWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

// OperationLogMiddleware 记录用户操作日志
// - quiet（安静）：不记录任何操作日志
// - standard（标准）：仅记录 POST/PUT/DELETE（写操作）
// - detailed（详细）：记录所有请求（含 GET）
func OperationLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestMethod := c.Request.Method

		// 根据日志级别判断是否需要记录当前请求
		if !pkg.ShouldLogOperation(requestMethod) {
			c.Next()
			return
		}

		// 读取请求 body（排除文件上传）
		requestBody := readRequestBody(c)

		// 使用自定义 ResponseWriter 捕获响应体
		cw := &CustomWriter{ResponseWriter: c.Writer}
		c.Writer = cw

		// 获取登录用户信息
		loginUser := getLoginUser(c)

		// 处理请求
		c.Next()

		// 获取响应信息
		responseStatus := cw.Status()
		responseBody := string(cw.Body)

		// 跳过无效响应（404 或空响应体）
		if responseStatus == 404 || responseBody == "" {
			return
		}

		// 解析响应状态和错误信息
		status, errMsg := parseResponse(c.Writer.Header().Get("Content-Type"), responseBody)

		// 计算业务类型
		businessType := parseBusinessType(requestMethod)

		// 获取模块标题
		operTitle := "操作日志"
		if val, exists := c.Get("operTitle"); exists {
			operTitle = val.(string)
		}

		// 获取操作信息
		ip := httputils.GetClientIP(c)
		requestURL := c.Request.URL.String()
		duration := time.Since(startTime).Milliseconds()
		operLocation := addressutils.GetRealAddressByIP(ip, config.GlobalConfig.IsAddressEnabled)

		// 异步写入操作日志（数据库 + 本地文件）
		go writeOperationLog(db, operTitle, businessType, requestURL, requestMethod,
			loginUser, ip, operLocation, requestBody, status, errMsg, duration)
	}
}

// readRequestBody 安全读取请求体，排除文件上传类型
func readRequestBody(c *gin.Context) string {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		return ""
	}

	bodyData, err := io.ReadAll(io.LimitReader(c.Request.Body, maxRequestBodySize))
	if err != nil {
		return ""
	}
	// 将 body 数据放回，供后续 handler 使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyData))
	return string(bodyData)
}

// getLoginUser 从 Gin Context 获取登录用户信息
func getLoginUser(c *gin.Context) *model.LoginUser {
	if claims, exists := c.Get("loginUser"); exists {
		return claims.(*model.LoginUser)
	}
	return &model.LoginUser{}
}

// parseResponse 解析响应内容，返回操作状态和错误信息
// status: 0=正常, 1=异常
func parseResponse(contentType, responseBody string) (int, string) {
	// 不记录导出 Excel 的响应体
	if contentType == "application/octet-stream" {
		return 0, ""
	}

	responseMap := utils.StringToMap(responseBody)
	if code, ok := responseMap["code"]; ok {
		if codeVal, isFloat := code.(float64); isFloat && codeVal != 200 {
			if msg, hasMsg := responseMap["msg"]; hasMsg {
				if msgStr, isStr := msg.(string); isStr {
					return 1, msgStr
				}
			}
			return 1, ""
		}
	}
	return 0, ""
}

// parseBusinessType 根据 HTTP 方法解析业务类型
// 0=其它 1=新增 2=修改 3=删除
func parseBusinessType(method string) int {
	switch method {
	case "POST":
		return 1
	case "PUT":
		return 2
	case "DELETE":
		return 3
	default:
		return 0
	}
}

// writeOperationLog 异步写入操作日志到数据库和本地文件
func writeOperationLog(db *gorm.DB, title string, businessType int, url, method string,
	loginUser *model.LoginUser, ip, location, requestBody string, status int, errMsg string, duration int64) {

	// 二次校验日志模式，防止异步执行时模式已变更为 quiet
	if !pkg.ShouldLogOperation(method) {
		return
	}

	// 写入数据库（GORM 内部线程安全，无需外部 mutex）
	insertSQL := `INSERT INTO sys_oper_log 
		(title, business_type, method, request_method, operator_type, oper_name, dept_name, 
		 oper_url, oper_ip, oper_location, oper_param, json_result, status, error_msg, oper_time, cost_time) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if err := db.Exec(insertSQL,
		title, businessType, url, method, 1,
		loginUser.UserName, loginUser.DeptName,
		url, ip, location,
		requestBody, "", status, errMsg,
		utils.GetCurrentDateTime(), duration,
	).Error; err != nil {
		pkg.Logger.Errorf("写入操作日志到数据库失败: %v", err)
	}

	// 写入本地 operation.log 文件
	pkg.OperationLogger.Infow("操作日志",
		"title", title,
		"method", url,
		"requestMethod", method,
		"operator", loginUser.UserName,
		"deptName", loginUser.DeptName,
		"operIp", ip,
		"operLocation", location,
		"requestBody", requestBody,
		"status", status,
		"errMsg", errMsg,
		"costTime", duration,
	)
}
