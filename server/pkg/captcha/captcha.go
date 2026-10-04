package captcha

import (
	"context"
	"fmt"
	"go-fin-server/internal/constant"
	"image/color"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/mojocn/base64Captcha"
)

// RedisStore 是 base64Captcha.Store 接口的实现，用于将验证码信息存储到 Redis
type RedisStore struct {
	client     *redis.Client
	expiration time.Duration // 添加超时设置
}

// Set 设置验证码信息到 Redis 中
func (rs *RedisStore) Set(id string, value string) error {
	ctx := context.Background()
	err := rs.client.Set(ctx, constant.CACHE_CAPTCHA_CODE_KEY+id, value, rs.expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get 获取验证码信息从 Redis 中
func (rs *RedisStore) Get(id string, clear bool) string {
	ctx := context.Background()
	val, err := rs.client.Get(ctx, id).Result()
	if err != nil {
		fmt.Printf("failed to get captcha value from Redis: %v\n", err)
		return ""
	}

	if clear {
		// 如果 clear 为 true，获取后从 Redis 中删除该验证码信息
		err := rs.client.Del(ctx, id).Err()
		if err != nil {
			fmt.Printf("failed to delete captcha value from Redis: %v\n", err)
		}
	}

	return val
}

func (rs *RedisStore) Verify(id, answer string, clear bool) bool {
	return rs.Get(id, clear) == answer
}

//var store = base64Captcha.DefaultMemStore

//var store *RedisStore

func GenerateCaptcha(client *redis.Client) (captchaID, imageURL string, err error) {
	var driver base64Captcha.Driver
	// 创建带有超时设置的 RedisStore 实例
	// 设置过期时间为 5 分钟
	store := &RedisStore{
		client:     client,
		expiration: 3 * time.Minute,
	}

	driverString := &base64Captcha.DriverString{
		Height:     40,
		Width:      100,
		NoiseCount: 0,
		Length:     4,
		Source:     "1234567890qwertyuioplkjhgfdsazxcvbnm",
		BgColor:    &color.RGBA{R: 0, G: 0, B: 0, A: 0},
		Fonts:      []string{"wqy-microhei.ttc"},
	}
	driver = driverString.ConvertFonts()
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := c.Generate()
	return id, b64s, err
}

func VerifyCaptcha(id, VerifyValue string, client *redis.Client) bool {
	store := &RedisStore{
		client:     client,
		expiration: 3 * time.Minute,
	}
	return store.Verify(id, VerifyValue, true)
}
