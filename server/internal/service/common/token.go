package common

import (
	"context"
	"encoding/json"
	"fmt"
	"go-fin-server/internal/config"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"
	"go-fin-server/pkg/utils/addressutils"
	"go-fin-server/pkg/utils/httputils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mssola/useragent"
	"github.com/redis/go-redis/v9"
)

type TokenService struct {
	Request      *gin.Context
	Client       *redis.Client
	Secret       string
	ExpireMinute int64
}

func (c TokenService) New(request *gin.Context, client *redis.Client, secret string, expireMinute int64) *TokenService {
	return &TokenService{
		request,
		client,
		secret,
		expireMinute,
	}
}

// Claims 定义JWT的Payload
type Claims struct {
	LoginUserKey string `json:"loginUserKey"`
	jwt.RegisteredClaims
}

// GetLoginUser 获取用户身份信息
func (c TokenService) GetLoginUser() (*model.LoginUser, error) {
	var user model.LoginUser
	token := c.GetToken()
	if token == "" {
		return nil, fmt.Errorf("token 为空，无法获取用户信息")
	}
	claims, err := c.ParseToken(token)
	if err != nil {
		return nil, err
	}
	userKey := c.GetTokenKey(claims.LoginUserKey)
	err = redistool.GetCacheObject(c.Client, userKey, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SetLoginUser 设置用户身份信息
func (c TokenService) SetLoginUser(loginUser *model.LoginUser) error {
	if loginUser != nil && loginUser.Token != "" {
		now := time.Now()
		expireTime := now.Add(time.Duration(c.ExpireMinute) * time.Minute)
		err := c.RefreshToken(loginUser, now.Unix(), expireTime.Unix())
		if err != nil {
			return err
		}
	}
	return nil
}

// DelLoginUser 删除用户身份信息
func (c TokenService) DelLoginUser(token string) error {
	if token != "" {
		userKey := c.GetTokenKey(token)
		err := redistool.Del(c.Client, userKey)
		if err != nil {
			return err
		}
	}
	return nil
}

// DelLoginUserAndMapping 删除用户身份信息和用户 token 映射
// 用于退出登录时完整清理 Redis 中的 login_tokens 和 login_user_token
func (c TokenService) DelLoginUserAndMapping(token string, userId uint64) error {
	// 收集需要删除的 key
	keys := make([]string, 0, 2)
	if token != "" {
		keys = append(keys, c.GetTokenKey(token))
	}
	if userId != 0 {
		keys = append(keys, c.GetUserTokenKey(userId))
	}
	if len(keys) > 0 {
		err := redistool.Del(c.Client, keys...)
		if err != nil {
			pkg.Logger.Errorf("退出登录清理 Redis 失败: keys=%v, error=%v", keys, err)
			return err
		}
		pkg.Logger.Infof("退出登录清理 Redis 成功: keys=%v", keys)
	}
	return nil
}

// CreateToken 创建令牌
// 实现单点登录：新登录时自动踢掉之前的会话
func (c TokenService) CreateToken(loginUser *model.LoginUser) (string, error) {
	now := time.Now()
	expireTime := now.Add(time.Duration(c.ExpireMinute) * time.Minute)

	// 单点登录：踢掉之前的会话
	c.kickPreviousSession(loginUser.UserId)

	token := uuid.New().String()
	loginUser.Token = token
	c.SetUserAgent(loginUser)
	err := c.RefreshToken(loginUser, now.Unix(), expireTime.Unix())
	if err != nil {
		return "", err
	}

	// 不设置过期读取redis
	claims := &Claims{
		LoginUserKey: token,
		//StandardClaims: jwt.StandardClaims{
		//	ExpiresAt: expireTime.Unix(),
		//},
	}
	newToken, err := c.CreateTokenByClaims(claims)
	if err != nil {
		return "", err
	}
	return newToken, nil
}

// VerifyToken 滑动续期：会话有效期内按需刷新，但会话总时长不超过1天
func (c TokenService) VerifyToken(loginUser *model.LoginUser) error {
	now := time.Now().Unix()
	// 会话起点（登录时间），作为1天硬上限的锚点，不随续期改变
	start := loginUser.LoginTime
	if start <= 0 {
		start = now
	}
	deadline := start + 24*60*60 // 1天有效期上限
	if now >= deadline {
		return fmt.Errorf("会话已超过1天有效期，请重新登录")
	}
	// 剩余不足20分钟才续期，避免每次请求都写 Redis
	if loginUser.ExpireTime-now > 20*60 {
		return nil
	}
	// 滑动续期：续到 min(now+窗口, deadline)，接近上限时 TTL 逐步收缩
	newExpire := now + c.ExpireMinute*60
	if newExpire > deadline {
		newExpire = deadline
	}
	return c.RefreshToken(loginUser, now, newExpire)
}

// RefreshToken 刷新令牌有效期
func (c TokenService) RefreshToken(loginUser *model.LoginUser, now, expireTime int64) error {
	// 保留登录时间（会话起点），用于在线用户展示与1天有效期上限
	if loginUser.LoginTime == 0 {
		loginUser.LoginTime = now
	}
	loginUser.ExpireTime = expireTime
	// 根据uuid将loginUser缓存到 login_tokens，TTL 为真实剩余时长
	userKey := c.GetTokenKey(loginUser.Token)
	ttl := expireTime - now
	if ttl <= 0 {
		ttl = 1
	}
	err := redistool.SetCacheObject(c.Client, userKey, loginUser, time.Duration(ttl)*time.Second)
	if err != nil {
		return err
	}
	// 同步刷新 login_user_token 映射的 TTL，保持与 login_tokens 一致
	if loginUser.UserId != 0 {
		c.recordUserToken(loginUser.UserId, loginUser.Token)
	}
	return nil
}

// SetUserAgent 设置用户代理信息
func (c TokenService) SetUserAgent(loginUser *model.LoginUser) {
	userAgent := useragent.New(c.Request.GetHeader("User-Agent"))
	name, _ := userAgent.Browser()
	ip := httputils.GetClientIP(c.Request)
	loginUser.Ipaddr = ip
	loginUser.LoginLocation = addressutils.GetRealAddressByIP(ip, config.GlobalConfig.IsAddressEnabled)
	loginUser.Browser = name
	loginUser.Os = userAgent.OS()
}

// CreateTokenByClaims 从数据声明生成令牌
func (c TokenService) CreateTokenByClaims(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(c.Secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// ParseToken 从令牌中获取数据声明
func (c TokenService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// GetUsernameFromToken 从令牌中获取用户名
func (c TokenService) GetUsernameFromToken(token string) (string, error) {
	claims, err := c.ParseToken(token)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// GetToken 获取请求token
func (c TokenService) GetToken() string {
	return strings.Replace(c.Request.GetHeader("Authorization"), "Bearer ", "", -1)
}

// GetTokenKey 获取token key
func (c TokenService) GetTokenKey(uuid string) string {
	return constant.CACHE_LOGIN_TOKEN_KEY + uuid
}

// GetUserTokenKey 获取用户当前 token 映射的 key
func (c TokenService) GetUserTokenKey(userId uint64) string {
	return constant.CACHE_LOGIN_USER_TOKEN_KEY + fmt.Sprintf("%d", userId)
}

// kickPreviousSession 踢掉用户之前的登录会话
func (c TokenService) kickPreviousSession(userId uint64) {
	if userId == 0 {
		pkg.Logger.Warn("踢掉旧会话: userId 为 0，跳过")
		return
	}
	userTokenKey := c.GetUserTokenKey(userId)
	ctx := context.Background()
	// 获取用户之前活跃的 token UUID
	oldTokenUuid, err := c.Client.Get(ctx, userTokenKey).Result()
	if err != nil {
		pkg.Logger.Debugf("踢掉旧会话: userId=%d, 无旧会话映射(key=%s), err=%v", userId, userTokenKey, err)
		return // 没有旧会话，无需踢人
	}
	if oldTokenUuid == "" {
		pkg.Logger.Debugf("踢掉旧会话: userId=%d, 旧会话映射为空", userId)
		return
	}
	// 历史写入（recordUserToken 通过 SetCacheObject）的值为 JSON 编码字符串，如 `"uuid"` 带引号，
	// 需还原为原始 uuid，否则构造的 login_tokens:"uuid" 是错误 key，导致旧会话无法被踢出。
	var rawUuid string
	if jsonErr := json.Unmarshal([]byte(oldTokenUuid), &rawUuid); jsonErr == nil && rawUuid != "" {
		oldTokenUuid = rawUuid
	}
	pkg.Logger.Infof("踢掉旧会话: userId=%d, oldToken=%s, oldLoginKey=%s", userId, oldTokenUuid, c.GetTokenKey(oldTokenUuid))
	// 删除旧的登录信息和用户 token 映射
	oldLoginKey := c.GetTokenKey(oldTokenUuid)
	err = redistool.Del(c.Client, oldLoginKey, userTokenKey)
	if err != nil {
		pkg.Logger.Errorf("踢掉旧会话清理 Redis 失败: userId=%d, error=%v", userId, err)
	} else {
		pkg.Logger.Infof("踢掉旧会话成功: userId=%d, 已删除 keys=[%s, %s]", userId, oldLoginKey, userTokenKey)
	}
}

// recordUserToken 记录用户当前活跃的 token
func (c TokenService) recordUserToken(userId uint64, tokenUuid string) {
	if userId == 0 {
		return
	}
	userTokenKey := c.GetUserTokenKey(userId)
	// 存储用户当前活跃的 token UUID（纯字符串，不用 JSON 编码，便于踢人时直接读取），过期时间与登录 token 一致
	err := redistool.Set(c.Client, userTokenKey, tokenUuid, time.Duration(c.ExpireMinute)*time.Minute)
	if err != nil {
		pkg.Logger.Errorf("记录用户 token 映射失败: key=%s, token=%s, error=%v", userTokenKey, tokenUuid, err)
	} else {
		pkg.Logger.Infof("记录用户 token 映射成功: key=%s, token=%s", userTokenKey, tokenUuid)
	}
}
