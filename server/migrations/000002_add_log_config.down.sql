-- 回滚：删除日志设置配置项
DELETE FROM sys_config WHERE config_key IN ('sys.log.level', 'sys.log.format');
