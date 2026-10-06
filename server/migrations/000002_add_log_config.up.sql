-- 新增日志设置配置项（日志级别、日志格式），用于网站设置页面动态调整日志输出
-- 使用 INSERT IGNORE：兼容已有数据库（避免重复插入报错）
INSERT IGNORE INTO sys_config (config_name, config_key, config_value, config_type, create_by, create_time, remark)
VALUES
    ('日志级别', 'sys.log.level', 'standard', 'N', 'admin', NOW(), '网站设置-日志级别：quiet(安静)/standard(标准)/detailed(详细)'),
    ('日志格式', 'sys.log.format', 'json', 'N', 'admin', NOW(), '网站设置-日志格式：json/console');
