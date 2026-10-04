package service

import (
	"go-fin-server/internal/service/common"
	"go-fin-server/internal/service/monitor"
	"go-fin-server/internal/service/system"
)

type Services struct {
	TokenService         common.TokenService
	PermissionService    common.PermissionService
	SequenceService      common.SequenceService
	SysUserService       system.SysUserService
	SysRoleService       system.SysRoleService
	SysDeptService       system.SysDeptService
	SysPostService       system.SysPostService
	SysNoticeService     system.SysNoticeService
	SysDictTypeService   system.SysDictTypeService
	SysDictDataService   system.SysDictDataService
	SysConfigService     system.SysConfigService
	SysMenuService       system.SysMenuService
	SysLoginService      system.SysLoginService
	SysOperLogService    monitor.SysOperLogService
	SysUserOnlineService monitor.SysUserOnlineService
	SysLoginInfoService  monitor.SysLoginInfoService
	SysJobService        monitor.SysJobService
	SysJobLogService     monitor.SysJobLogService
}
