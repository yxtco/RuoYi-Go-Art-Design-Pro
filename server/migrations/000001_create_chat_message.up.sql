-- 聊天消息持久化表（WebSocket 聊天消息落库）
-- 使用 IF NOT EXISTS：兼容已有数据库（避免重复建表报错）
CREATE TABLE IF NOT EXISTS chat_message (
    id            bigint       NOT NULL AUTO_INCREMENT COMMENT '主键',
    from_user_id  bigint       NOT NULL DEFAULT 0 COMMENT '发送者用户ID',
    from_name     varchar(50)  DEFAULT NULL COMMENT '发送者名称',
    from_avatar   varchar(255) DEFAULT NULL COMMENT '发送者头像',
    to_user_id    bigint       NOT NULL DEFAULT 0 COMMENT '私聊目标用户ID（群聊为0）',
    group_name    varchar(100) DEFAULT NULL COMMENT '群聊名称（私聊为空）',
    content       text         COMMENT '消息内容',
    msg_type      varchar(20)  DEFAULT 'text' COMMENT '消息类型: text/image/file/emoji',
    file_url      varchar(500) DEFAULT NULL COMMENT '文件/图片URL',
    file_name     varchar(255) DEFAULT NULL COMMENT '文件名',
    timestamp     bigint       NOT NULL DEFAULT 0 COMMENT '时间戳(毫秒)',
    PRIMARY KEY (id),
    KEY idx_from_to (from_user_id, to_user_id),
    KEY idx_group (group_name),
    KEY idx_ts (timestamp)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '聊天消息持久化记录';
