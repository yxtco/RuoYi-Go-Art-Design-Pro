package httputils

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

// HttpClient 定义了一个 HTTP 客户端结构体，包含一个 http.Client 实例。
type HttpClient struct {
	client *http.Client
}

// NewHttpClient 创建并返回一个新的 HttpClient 实例。
func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Get 发起一个 GET 请求，并返回响应体的字节和错误。
// params 是一个 map，表示要传递的查询参数。
func (c *HttpClient) Get(baseURL string, params map[string]string) ([]byte, error) {
	// 解析基础 URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// 如果有参数，添加到 URL 的查询字符串中
	if params != nil {
		query := parsedURL.Query()
		for key, value := range params {
			query.Set(key, value)
		}
		parsedURL.RawQuery = query.Encode()
	}

	// 创建请求
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// 发起请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 尝试自动检测内容编码
	e, name, _ := charset.DetermineEncoding(body, resp.Header.Get("Content-Type"))
	// 如果检测到GBK编码，则进行转换
	if name == "gb18030" || name == "gbk" {
		transformer := e.NewDecoder()
		reader := transform.NewReader(bytes.NewReader(body), transformer)
		convertedBody, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		// 处理转换后的内容
		return convertedBody, nil
	}
	return body, nil
}

// Post 发起一个 POST 请求，并返回响应体的字节和错误。
// `contentType` 通常是 "application/json" 或 "application/x-www-form-urlencoded"
func (c *HttpClient) Post(url string, contentType string, body []byte) ([]byte, error) {
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	// 发起请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
