package monitor

import (
	"fmt"
	"go-fin-server/internal/config"
	"go-fin-server/internal/response"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type ServerHandler struct {
}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

// GetServer 获取服务器信息
//
//	@Summary	获取服务器信息（CPU、内存、磁盘、运行时）
//	@Tags		服务监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"服务器信息"
//	@Router		/monitor/server [get]
//	@Security	BearerAuth
func (s *ServerHandler) GetServer(c *gin.Context) {
	response.SetOperTitle(c, "获取服务器信息")
	response.Data(c, map[string]interface{}{
		"cpu":      getCPUInfo()[0],
		"mem":      getMemoryInfo(),
		"sys":      getHostInfo(),
		"sysFiles": getDiskInfo(),
		"jvm":      getRuntimeInfo(),
	})
}

func getCPUInfo() []map[string]interface{} {
	cpuStats, _ := cpu.Times(false) // false 表示获取整体CPU使用情况
	cpuInfo, _ := cpu.Info()
	data := make([]map[string]interface{}, 0)
	for _, info := range cpuInfo {
		data = append(data, map[string]interface{}{
			"cpuNum": info.Cores,
		})
	}
	for index, stat := range cpuStats {
		totalDelta := stat.User + stat.System + stat.Idle
		data[index]["used"] = fmt.Sprintf("%.2f", (stat.User/totalDelta)*100)
		data[index]["sys"] = fmt.Sprintf("%.2f", (stat.System/totalDelta)*100)
		data[index]["free"] = fmt.Sprintf("%.2f", (stat.Idle/totalDelta)*100)
	}
	return data
}

func getMemoryInfo() map[string]interface{} {
	vmStat, _ := mem.VirtualMemory()
	return map[string]interface{}{
		"total": fmt.Sprintf("%.2f", float64(vmStat.Total)/float64(1024*1024*1024)),
		"used":  fmt.Sprintf("%.2f", float64(vmStat.Used)/float64(1024*1024*1024)),
		"free":  fmt.Sprintf("%.2f", float64(vmStat.Free)/float64(1024*1024*1024)),
		"usage": fmt.Sprintf("%.2f", vmStat.UsedPercent),
	}
}

func getHostInfo() map[string]interface{} {
	hostInfo, _ := host.Info()
	addrs, _ := net.InterfaceAddrs()
	var ip string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}
	userDir, _ := os.Getwd()
	return map[string]interface{}{
		"computerName": hostInfo.Hostname,
		"computerIp":   ip,
		"osName":       hostInfo.OS,
		"osArch":       runtime.GOARCH,
		"userDir":      userDir,
	}
}

func getDiskInfo() []map[string]interface{} {
	partitions, _ := disk.Partitions(true) // true 表示获取所有分区
	data := make([]map[string]interface{}, 0)
	for _, part := range partitions {
		diskStat, _ := disk.Usage(part.Mountpoint)
		data = append(data, map[string]interface{}{
			"dirName":     part.Mountpoint,
			"sysTypeName": part.Fstype,
			"typeName":    "",
			"total":       fmt.Sprintf("%.2f GB", float64(diskStat.Total)/float64(1024*1024*1024)),
			"free":        fmt.Sprintf("%.2f GB", float64(diskStat.Free)/float64(1024*1024*1024)),
			"used":        fmt.Sprintf("%.2f GB", float64(diskStat.Used)/float64(1024*1024*1024)),
			"usage":       fmt.Sprintf("%.2f", diskStat.UsedPercent),
		})
	}
	return data
}

func getRuntimeInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	totalMemory := m.Sys / 1024 / 1024
	usedMemory := m.Alloc / 1024 / 1024
	freeMemory := (m.Sys - m.Alloc) / 1024 / 1024
	memoryUsage := fmt.Sprintf("%.2f", float64(m.Alloc)/float64(m.Sys)*100)

	// 使用当前时间作为结束时间
	endTime := time.Now()
	// 计算运行时长
	duration := endTime.Sub(config.GlobalConfig.StartTime)

	// 将运行时长转换为天数、小时数和分钟数
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return map[string]interface{}{
		"total":     totalMemory,
		"free":      freeMemory,
		"version":   runtime.Version(),
		"home":      runtime.GOROOT(),
		"startTime": config.GlobalConfig.StartTime.Format("2006-01-02 15:04:05"),
		"used":      usedMemory,
		"runTime":   fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes),
		"inputArgs": os.Args,
		"usage":     memoryUsage,
		"name":      "Go",
	}
}
