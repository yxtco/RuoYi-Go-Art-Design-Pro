package migrate

import (
	"errors"
	"fmt"

	"go-fin-server/internal/db"
	"go-fin-server/migrations"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Run 对 master 数据库连接执行 golang-migrate 迁移（自包含，迁移文件内嵌进二进制）
//   - 以现有 master 连接打开 migrate 实例（mysql 驱动 + iofs 内嵌文件源）
//   - 执行 Migrate up 至最新版本
//   - ErrNoChange 视为已是最新，正常返回
//   - 迁移记录存于数据库 schema_migrations 表，已应用的迁移不会重复执行
func Run() error {
	master, ok := db.DBConnections["master"]
	if !ok {
		return errors.New("master 数据库连接不存在，无法执行迁移")
	}
	sqlDB, err := master.DB()
	if err != nil {
		return fmt.Errorf("获取 master 数据库实例失败: %w", err)
	}

	// 内嵌迁移文件作为文件源
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("加载内嵌迁移文件失败: %w", err)
	}

	// 复用现有连接（库存在性由配置阶段保证）
	driver, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("初始化迁移驱动失败: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "mysql", driver)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil // 已是最新版本，无需迁移
		}
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	return nil
}
