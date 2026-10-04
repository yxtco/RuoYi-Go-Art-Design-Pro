package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go-fin-server/internal/response"
	"io"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func SignVerificationMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params map[string]string
		contentType := c.GetHeader("Content-Type")
		// 如果是 multipart/form-data 类型，不执行签名校验
		if strings.Contains(contentType, "multipart/form-data") {
			c.Next()
			return
		}
		// json 类型
		if strings.Contains(contentType, "application/json") {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				response.ErrorCode(c, 401, "unable to read request body")
				c.Abort()
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if err := json.Unmarshal(bodyBytes, &params); err != nil {
				response.ErrorCode(c, 401, "invalid JSON data")
				c.Abort()
				return
			}
		} else {
			if err := c.Request.ParseForm(); err != nil {
				response.ErrorCode(c, 401, "error parsing form data")
				c.Abort()
				return
			}
			for key, value := range c.Request.Form {
				_, _, err := c.Request.FormFile(key)
				if err != nil {
					params[key] = strings.Join(value, ",")
				}
			}
		}
		// 提取并移除sign参数
		clientSign, exists := params["sign"]
		if !exists {
			response.ErrorCode(c, 401, "sign is required")
			c.Abort()
			return
		}
		delete(params, "sign")
		serverSign := generateSign(params, secretKey)
		if serverSign != clientSign {
			response.ErrorCode(c, 401, "invalid sign")
			c.Abort()
			return
		}
		c.Next()
	}
}

func generateSign(params map[string]string, secretKey string) string {
	var keys []string
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signStr string
	for _, k := range keys {
		signStr += k + params[k]
	}
	signStr += secretKey

	h := sha256.New()
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}
