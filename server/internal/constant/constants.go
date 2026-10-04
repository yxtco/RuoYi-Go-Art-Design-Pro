package constant

const (
	// UTF8 字符集
	UTF8 = "UTF-8"

	// GBK 字符集
	GBK = "GBK"

	// www 主域
	WWW = "www."

	// http 请求
	HTTP = "http://"

	// https 请求
	HTTPS = "https://"

	// 通用成功标识
	SUCCESS = "0"

	// 通用失败标识
	FAIL = "1"

	// 登录成功
	LOGIN_SUCCESS = "Success"

	// 注销
	LOGOUT = "Logout"

	// 注册
	REGISTER = "Register"

	// 登录失败
	LOGIN_FAIL = "Error"

	// 所有权限标识
	ALL_PERMISSION = "*:*:*"

	// 管理员角色权限标识
	SUPER_ADMIN = "admin"

	// 角色权限分隔符
	ROLE_DELIMITER = ","

	// 权限标识分隔符
	PERMISSION_DELIMITER = ","

	// 验证码有效期（分钟）
	CAPTCHA_EXPIRATION = 2

	// 令牌
	TOKEN = "token"

	// 令牌前缀
	TOKEN_PREFIX = "Bearer "

	// 登录用户键
	LOGIN_USER_KEY = "login_user_key"

	// 用户ID
	JWT_USERID = "userid"

	// 用户名称
	//JWT_USERNAME = jwt.StandardClaims.

	// 用户头像
	JWT_AVATAR = "avatar"

	// 创建时间
	JWT_CREATED = "created"

	// 用户权限
	JWT_AUTHORITIES = "authorities"

	// 资源映射路径前缀
	RESOURCE_PREFIX = "/profile"

	// RMI 远程方法调用
	LOOKUP_RMI = "rmi:"

	// LDAP 远程方法调用
	LOOKUP_LDAP = "ldap:"

	// LDAPS 远程方法调用
	LOOKUP_LDAPS = "ldaps:"

	// 自动识别 JSON 对象白名单配置
	JSON_WHITELIST_STR1 = "org.springframework"
	JSON_WHITELIST_STR2 = "com.ruoyi"

	// 定时任务白名单配置
	JOB_WHITELIST_STR = "com.ruoyi"

	// 定时任务违规的字符
	JOB_ERROR_STR1 = "java.net.URL"
	JOB_ERROR_STR2 = "javax.naming.InitialContext"
	JOB_ERROR_STR3 = "org.yaml.snakeyaml"
	JOB_ERROR_STR4 = "org.springframework"
	JOB_ERROR_STR5 = "org.apache"
	JOB_ERROR_STR6 = "com.ruoyi.common.utils.file"
	JOB_ERROR_STR7 = "com.ruoyi.common.config"
)

const (
	// 登录用户 redis key
	CACHE_LOGIN_TOKEN_KEY = "login_tokens:"
	// 登录用户当前 token 映射 redis key（用于单点登录踢人）
	// 格式: login_user_token:<userId> -> tokenUuid
	CACHE_LOGIN_USER_TOKEN_KEY = "login_user_token:"
	// 验证码 redis key
	CACHE_CAPTCHA_CODE_KEY = "captcha_codes:"
	// 参数管理 cache key
	CACHE_SYS_CONFIG_KEY = "sys_config:"
	// 字典管理 cache key
	CACHE_SYS_DICT_KEY = "sys_dict:"
	// 防重提交 redis key
	CACHE_REPEAT_SUBMIT_KEY = "repeat_submit:"
	// 限流 redis key
	CACHE_RATE_LIMIT_KEY = "rate_limit:"
	// 登录账户密码错误次数 redis key
	CACHE_PWD_ERR_CNT_KEY = "pwd_err_cnt:"
	// 序号生成 redis key 前缀
	CACHE_SEQUENCE_KEY = "sequence:"
	// 在线用户 redis hash key
	// 格式: online_users -> { userId: userInfoJson }
	CACHE_ONLINE_USERS_KEY = "online_users"
)
