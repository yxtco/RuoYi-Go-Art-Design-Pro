package initdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go-fin-server/internal/config"
	"go-fin-server/internal/db"

	"github.com/go-sql-driver/mysql"
)

// 初始化库文件路径，可用环境变量 RUOYI_INIT_SQL 覆盖
var initSQLPath = func() string {
	if p := os.Getenv("RUOYI_INIT_SQL"); p != "" {
		return p
	}
	return "data/ry_init_full.sql"
}()

// EnsureInitialized 启动时数据库自动初始化：
//  1. 确保 release.yaml 等配置对应的 master 库存在（不存在则创建）
//  2. 检测该库是否存在数据表（information_schema.tables 计数）
//  3. 若没有表（全新库）→ 自动导入 data/ry_init_full.sql（结构 + 种子数据），免去人工导入
//
// 需在 db.SetupDatabase 之前调用（因为可能要先建库）。
func EnsureInitialized() error {
	cfg := masterConfig()
	if cfg == nil {
		return errors.New("master 数据库配置不存在，无法初始化")
	}
	if cfg.Driver != "mysql" {
		return fmt.Errorf("数据库自动初始化暂仅支持 MySQL 驱动（当前: %s）", cfg.Driver)
	}

	// 1) 确保目标库存在
	if err := ensureDatabaseExists(cfg); err != nil {
		return fmt.Errorf("确保数据库存在失败: %w", err)
	}

	// 2) 检测表是否存在
	count, err := tableCount(cfg)
	if err != nil {
		return fmt.Errorf("检测数据库表数量失败: %w", err)
	}
	if count > 0 {
		fmt.Printf("[initdb] 数据库 %s 已有 %d 张表，跳过初始化\n", cfg.Database, count)
		return nil
	}

	// 3) 空库 → 自动导入初始化库
	return importInitSQL(cfg)
}

// masterConfig 返回 master 数据库配置
func masterConfig() *db.DatabaseConfig {
	for i := range config.GlobalConfig.DatabaseList {
		if config.GlobalConfig.DatabaseList[i].Name == "master" {
			return &config.GlobalConfig.DatabaseList[i]
		}
	}
	return nil
}

// buildDSN 构造 MySQL DSN；includeDB 决定是否附带库名；multi 开启多语句支持
func buildDSN(cfg *db.DatabaseConfig, includeDB, multi bool) string {
	dbName := ""
	if includeDB {
		dbName = cfg.Database
	}
	m := mysql.Config{
		User:            cfg.Username,
		Passwd:          cfg.Password,
		Net:             "tcp",
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:          dbName,
		ParseTime:       true,
		Loc:             time.Local,
		MultiStatements: multi,
		Params:          map[string]string{"charset": "utf8mb4"},
	}
	return m.FormatDSN()
}

// ensureDatabaseExists 创建目标库（若不存在）
func ensureDatabaseExists(cfg *db.DatabaseConfig) error {
	conn, err := sql.Open("mysql", buildDSN(cfg, false, false))
	if err != nil {
		return err
	}
	defer conn.Close()
	ddl := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci",
		cfg.Database,
	)
	_, err = conn.Exec(ddl)
	return err
}

// tableCount 查询目标库的表数量
func tableCount(cfg *db.DatabaseConfig) (int, error) {
	conn, err := sql.Open("mysql", buildDSN(cfg, true, false))
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	var count int
	err = conn.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?",
		cfg.Database,
	).Scan(&count)
	return count, err
}

// importInitSQL 读取并执行初始化库文件（多语句）
func importInitSQL(cfg *db.DatabaseConfig) error {
	content, err := os.ReadFile(initSQLPath)
	if err != nil {
		return fmt.Errorf("读取初始化库文件 %s 失败: %w（请先运行 dbexport 生成，或设置 RUOYI_INIT_SQL）", initSQLPath, err)
	}

	conn, err := sql.Open("mysql", buildDSN(cfg, true, true))
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Printf("[initdb] 数据库 %s 为空库，开始自动导入初始化库 %s ...\n", cfg.Database, initSQLPath)
	// 剥离文件内硬编码的 CREATE DATABASE / USE 语句，确保初始化只作用于当前连接的数据库
	if _, err := conn.Exec(stripSessionStatements(string(content))); err != nil {
		return fmt.Errorf("导入初始化库失败: %w", err)
	}
	fmt.Printf("[initdb] 初始化库导入完成\n")
	return nil
}

// stripSessionStatements 移除初始化库文件里的 CREATE DATABASE / USE 语句，
// 避免导入到与导出源库名不同的数据库时被劫持到错误的库。
func stripSessionStatements(content string) string {
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "CREATE DATABASE") || strings.HasPrefix(t, "USE `") {
			continue
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}
