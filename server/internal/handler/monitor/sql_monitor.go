package monitor

import (
	"go-fin-server/internal/db"
	"go-fin-server/internal/response"

	"github.com/gin-gonic/gin"
)

type SQLMonitorHandler struct{}

func NewSQLMonitorHandler() *SQLMonitorHandler {
	return &SQLMonitorHandler{}
}

// GetSQLStats 获取 SQL 监控统计数据
// GetSQLStats 获取 SQL 监控统计数据
//
//	@Summary	获取 SQL 监控统计数据
//	@Tags		SQL监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"SQL监控统计"
//	@Router		/monitor/druid/sqlStats [get]
//	@Security	BearerAuth
func (h *SQLMonitorHandler) GetSQLStats(c *gin.Context) {
	response.SetOperTitle(c, "获取SQL监控统计数据")
	stats := db.GetSQLStats()
	recent := db.GetRecentRecords()
	slow := db.GetSlowRecords()

	var avgDuration float64
	if stats.TotalCount > 0 {
		avgDuration = float64(stats.TotalDuration) / float64(stats.TotalCount)
	}

	response.Data(c, map[string]interface{}{
		"totalCount":     stats.TotalCount,
		"totalDuration":  stats.TotalDuration,
		"avgDuration":    roundTo2(avgDuration),
		"slowCount":      stats.SlowCount,
		"selectCount":    stats.SelectCount,
		"insertCount":    stats.InsertCount,
		"updateCount":    stats.UpdateCount,
		"deleteCount":    stats.DeleteCount,
		"selectDuration": stats.SelectDuration,
		"insertDuration": stats.InsertDuration,
		"updateDuration": stats.UpdateDuration,
		"deleteDuration": stats.DeleteDuration,
		"recent":         recent,
		"slow":           slow,
	})
}

// ClearSQLStats 清空 SQL 监控统计数据
// ClearSQLStats 清空 SQL 监控统计数据
//
//	@Summary	清空 SQL 监控统计数据
//	@Tags		SQL监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/druid/sqlStats [delete]
//	@Security	BearerAuth
func (h *SQLMonitorHandler) ClearSQLStats(c *gin.Context) {
	response.SetOperTitle(c, "清空SQL监控统计数据")
	db.ClearSQLStats()
	response.Data(c, nil)
}

func roundTo2(val float64) float64 {
	return float64(int(val*100+0.5)) / 100
}
