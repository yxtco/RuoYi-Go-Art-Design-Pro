package middleware

import (
	"bytes"
	"go-fin-server/internal/config"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/addressutils"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CustomWriter 自定义的 ResponseWriter
type CustomWriter struct {
	gin.ResponseWriter
	Body []byte
}

func (w *CustomWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

func ReadRequestBody(body io.Reader) ([]byte, error) {
	// 读取 body 数据
	bodyData, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	// 重新放回 body 数据
	return bodyData, nil
}

var dbMutex sync.Mutex

// OperationLogMiddleware 记录用户操作日志
func OperationLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()
		var requestBody string
		contentType := c.Request.Header.Get("Content-Type")
		//  排除记录上传文件
		if strings.Contains(contentType, "multipart/form-data") {
			requestBody = ""
		} else {
			// 读取请求 body 数据
			requestBodyByte, _ := ReadRequestBody(c.Request.Body)
			requestBody = string(requestBodyByte)
			// 重新设置 body 数据
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyByte))
		}

		// 使用自定义的 ResponseWriter
		c.Writer = &CustomWriter{ResponseWriter: c.Writer}
		var loginUser = new(model.LoginUser)
		claims, exists := c.Get("loginUser")
		if exists {
			loginUser = claims.(*model.LoginUser)
		}
		// 处理请求
		c.Next()
		// 请求处理完成后获取响应体和状态码
		cw := c.Writer.(*CustomWriter)

		// 记录请求结束时间
		//endTime := time.Now()
		// 计算请求处理时间
		duration := time.Since(startTime).Milliseconds()
		// 获取请求和响应信息
		ip := c.ClientIP()
		requestMethod := c.Request.Method
		requestURL := c.Request.URL.String()
		//requestBody := c.Request.Form.Encode()
		//requestUa   := c.Request.UserAgent()
		responseStatus := c.Writer.Status()
		responseBody := string(cw.Body)
		if responseStatus == 404 || responseBody == "" {
			return
		}
		// 如果响应的body code 不为 200 则记录error 信息
		var errMsg string
		var status = 0 // 0 正常 1异常
		// 检查响应的 Content-Type
		responseContentType := c.Writer.Header().Get("Content-Type")
		// 不记录导出excel的响应
		if responseContentType == "application/octet-stream" {
			responseBody = ""
		} else {
			responseBodyMap := utils.StringToMap(responseBody)
			if v, ok := responseBodyMap["code"]; ok {
				if v.(float64) != 200 {
					if msg, ok2 := responseBodyMap["msg"]; ok2 {
						errMsg = msg.(string)
					}
					status = 1
				}
			}
		}

		//business_type 业务类型（0其它 1新增 2修改 3删除）
		var businessType int
		switch c.Request.Method {
		case "GET":
			businessType = 0
		case "POST":
			businessType = 1
		case "PUT":
			businessType = 2
		case "DELETE":
			businessType = 3
		default:
			businessType = 0
		}
		var operTitle = "操作日志"
		// 模块标题
		if val, exists := c.Get("operTitle"); exists {
			operTitle = val.(string)
		}
		operLocation := addressutils.GetRealAddressByIP(ip, config.GlobalConfig.IsAddressEnabled)
		//操作类别（0其它 1后台用户 2手机端用户）
		// 使用协程插入操作日志到数据库和本地文件
		go func() {
			dbMutex.Lock()
			defer dbMutex.Unlock()
			insertLogQuery := "INSERT INTO sys_oper_log (title,business_type,method, request_method,operator_type,oper_name,dept_name,oper_url,oper_ip,oper_location,oper_param,json_result,status,error_msg,oper_time,cost_time) VALUES (?, ?, ?, ?, ?, ?, ?,?,?,?,?, ?, ?, ?, ?,?)"
			if err := db.Exec(insertLogQuery, operTitle, businessType, requestURL, requestMethod, 1, loginUser.UserName, loginUser.DeptName, requestURL, ip, operLocation, requestBody, "", status, errMsg, utils.GetCurrentDateTime(), duration).Error; err != nil {
				pkg.Logger.Errorf("failed to insert log into database: %v", err)
			}
			// 写入操作日志到本地 operation.log 文件
			pkg.OperationLogger.Infow("操作日志",
				"title", operTitle,
				"method", requestURL,
				"requestMethod", requestMethod,
				"operator", loginUser.UserName,
				"deptName", loginUser.DeptName,
				"operIp", ip,
				"operLocation", operLocation,
				"requestBody", requestBody,
				"status", status,
				"errMsg", errMsg,
				"costTime", duration,
			)
		}()
	}
}
