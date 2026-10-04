package db

import (
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SQLRecord SQL 执行记录
type SQLRecord struct {
	SQL      string `json:"sql"`
	Duration int64  `json:"duration"` // 毫秒
	Type     string `json:"type"`     // SELECT/INSERT/UPDATE/DELETE
	Time     string `json:"time"`
	Rows     int64  `json:"rows"`
}

// SQLStats SQL 执行统计（内存存储）
type SQLStats struct {
	TotalCount     int64 `json:"totalCount"`
	TotalDuration  int64 `json:"totalDuration"`
	SlowCount      int64 `json:"slowCount"`
	SelectCount    int64 `json:"selectCount"`
	InsertCount    int64 `json:"insertCount"`
	UpdateCount    int64 `json:"updateCount"`
	DeleteCount    int64 `json:"deleteCount"`
	SelectDuration int64 `json:"selectDuration"`
	InsertDuration int64 `json:"insertDuration"`
	UpdateDuration int64 `json:"updateDuration"`
	DeleteDuration int64 `json:"deleteDuration"`
}

// SQLMonitorStore 内存存储 SQL 监控数据
var SQLMonitorStore = struct {
	mu     sync.RWMutex
	stats  SQLStats
	recent []SQLRecord
	slow   []SQLRecord
}{
	recent: make([]SQLRecord, 0, 50),
	slow:   make([]SQLRecord, 0, 20),
}

// GetSQLStats 获取统计数据（线程安全）
func GetSQLStats() SQLStats {
	SQLMonitorStore.mu.RLock()
	defer SQLMonitorStore.mu.RUnlock()
	return SQLMonitorStore.stats
}

// GetRecentRecords 获取最近执行记录（线程安全）
func GetRecentRecords() []SQLRecord {
	SQLMonitorStore.mu.RLock()
	defer SQLMonitorStore.mu.RUnlock()
	result := make([]SQLRecord, len(SQLMonitorStore.recent))
	copy(result, SQLMonitorStore.recent)
	return result
}

// GetSlowRecords 获取慢查询记录（线程安全）
func GetSlowRecords() []SQLRecord {
	SQLMonitorStore.mu.RLock()
	defer SQLMonitorStore.mu.RUnlock()
	result := make([]SQLRecord, len(SQLMonitorStore.slow))
	copy(result, SQLMonitorStore.slow)
	return result
}

// ClearSQLStats 清空所有监控数据（线程安全）
func ClearSQLStats() {
	SQLMonitorStore.mu.Lock()
	defer SQLMonitorStore.mu.Unlock()
	SQLMonitorStore.stats = SQLStats{}
	SQLMonitorStore.recent = make([]SQLRecord, 0, 50)
	SQLMonitorStore.slow = make([]SQLRecord, 0, 20)
}

// SQLMonitorPlugin GORM 插件，用于记录 SQL 执行统计到内存
type SQLMonitorPlugin struct{}

func (p *SQLMonitorPlugin) Name() string {
	return "sql_monitor"
}

// Initialize 注册回调钩子
func (p *SQLMonitorPlugin) Initialize(db *gorm.DB) error {
	// 在所有 SQL 执行前记录开始时间
	beforeCallback := func(db *gorm.DB) {
		db.InstanceSet("sql_start_time", time.Now())
	}

	// 在所有 SQL 执行后记录统计
	afterCallback := func(db *gorm.DB) {
		if db.Statement.SQL.String() == "" {
			return
		}

		sqlStr := db.Statement.SQL.String()
		sqlType := detectSQLType(sqlStr)
		rows := db.Statement.RowsAffected

		// 计算执行时间
		var duration int64
		if startTimeVal, ok := db.InstanceGet("sql_start_time"); ok {
			if startTime, ok := startTimeVal.(time.Time); ok {
				duration = time.Since(startTime).Milliseconds()
			}
		}

		record := SQLRecord{
			SQL:      truncateSQL(sqlStr, 500),
			Duration: duration,
			Type:     sqlType,
			Time:     time.Now().Format("2006-01-02 15:04:05"),
			Rows:     rows,
		}

		// 写入内存存储
		SQLMonitorStore.mu.Lock()
		defer SQLMonitorStore.mu.Unlock()

		// 更新统计
		SQLMonitorStore.stats.TotalCount++
		SQLMonitorStore.stats.TotalDuration += duration
		switch sqlType {
		case "SELECT":
			SQLMonitorStore.stats.SelectCount++
			SQLMonitorStore.stats.SelectDuration += duration
		case "INSERT":
			SQLMonitorStore.stats.InsertCount++
			SQLMonitorStore.stats.InsertDuration += duration
		case "UPDATE":
			SQLMonitorStore.stats.UpdateCount++
			SQLMonitorStore.stats.UpdateDuration += duration
		case "DELETE":
			SQLMonitorStore.stats.DeleteCount++
			SQLMonitorStore.stats.DeleteDuration += duration
		}

		// 追加最近执行记录（最多 50 条）
		SQLMonitorStore.recent = append([]SQLRecord{record}, SQLMonitorStore.recent...)
		if len(SQLMonitorStore.recent) > 50 {
			SQLMonitorStore.recent = SQLMonitorStore.recent[:50]
		}

		// 慢查询记录（超过 100ms）
		if duration >= 100 {
			SQLMonitorStore.stats.SlowCount++
			SQLMonitorStore.slow = append([]SQLRecord{record}, SQLMonitorStore.slow...)
			if len(SQLMonitorStore.slow) > 20 {
				SQLMonitorStore.slow = SQLMonitorStore.slow[:20]
			}
		}
	}

	// 分别注册到不同的回调链以确保覆盖所有操作
	_ = db.Callback().Create().Before("gorm:create").Register("sql_monitor:before_create", beforeCallback)
	_ = db.Callback().Create().After("gorm:create").Register("sql_monitor:after_create", afterCallback)
	_ = db.Callback().Update().Before("gorm:update").Register("sql_monitor:before_update", beforeCallback)
	_ = db.Callback().Update().After("gorm:update").Register("sql_monitor:after_update", afterCallback)
	_ = db.Callback().Delete().Before("gorm:delete").Register("sql_monitor:before_delete", beforeCallback)
	_ = db.Callback().Delete().After("gorm:delete").Register("sql_monitor:after_delete", afterCallback)
	_ = db.Callback().Query().Before("gorm:query").Register("sql_monitor:before_query", beforeCallback)
	_ = db.Callback().Query().After("gorm:query").Register("sql_monitor:after_query", afterCallback)
	_ = db.Callback().Row().Before("gorm:row").Register("sql_monitor:before_row", beforeCallback)
	_ = db.Callback().Row().After("gorm:row").Register("sql_monitor:after_row", afterCallback)
	_ = db.Callback().Raw().Before("gorm:raw").Register("sql_monitor:before_raw", beforeCallback)
	_ = db.Callback().Raw().After("gorm:raw").Register("sql_monitor:after_raw", afterCallback)

	return nil
}

// detectSQLType 检测 SQL 语句类型
func detectSQLType(sql string) string {
	sql = strings.TrimSpace(sql)
	upper := strings.ToUpper(sql)
	switch {
	case strings.HasPrefix(upper, "SELECT"):
		return "SELECT"
	case strings.HasPrefix(upper, "INSERT"):
		return "INSERT"
	case strings.HasPrefix(upper, "UPDATE"):
		return "UPDATE"
	case strings.HasPrefix(upper, "DELETE"):
		return "DELETE"
	default:
		return "OTHER"
	}
}

// truncateSQL 截断过长的 SQL 语句
func truncateSQL(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "..."
}
