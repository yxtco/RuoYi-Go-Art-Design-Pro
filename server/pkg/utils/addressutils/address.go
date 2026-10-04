package addressutils

import (
	"encoding/json"
	"fmt"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils/httputils"
	"net"
)

const ipUrl = "http://www.ip-api.com/json/"
const UNKNOWN = "XX XX"

func GetRealAddressByIP(ip string, isAddressEnabled bool) string {
	if internalIp(ip) {
		return "内网IP"
	}
	if isAddressEnabled {
		httpClient := httputils.NewHttpClient()
		getResponse, err := httpClient.Get(ipUrl+ip, map[string]string{
			"lang": "zh-CN",
		})
		if err != nil {
			pkg.Logger.Error("获取地理位置异常 " + ip + " " + err.Error())
			return UNKNOWN
		}
		var responseMap map[string]interface{}
		err = json.Unmarshal(getResponse, &responseMap)
		if err != nil {
			pkg.Logger.Error("解析 JSON 响应失败: " + err.Error())
			return UNKNOWN
		}
		if responseMap["status"] != "success" {
			pkg.Logger.Error("IP查询失败: " + ip)
			return UNKNOWN
		}
		return fmt.Sprintf("%s %s", responseMap["regionName"], responseMap["city"])
	}
	return UNKNOWN
}

func internalIp(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true
	}
	return parsedIP.IsLoopback() || parsedIP.IsPrivate()
}
