package redistool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Info 获取 Redis 服务器信息
func Info(client *redis.Client) (string, error) {
	ctx := context.Background()
	val, err := client.Info(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("Error getting Redis info: %v", err)
	}
	return val, nil
}

// CommandStats 获取 Redis 命令统计信息
func CommandStats(client *redis.Client) (string, error) {
	ctx := context.Background()
	val, err := client.Info(ctx, "commandstats").Result()
	if err != nil {
		return "", fmt.Errorf("Error getting Redis command stats: %v", err)
	}
	return val, nil
}

// DBSize 获取 Redis 数据库大小
func DBSize(client *redis.Client) (int64, error) {
	ctx := context.Background()
	val, err := client.DBSize(ctx).Result()
	if err != nil {
		return val, fmt.Errorf("Error getting Redis DB size: %v", err)
	}
	return val, nil
}

// Set 设置 key-value 到 Redis，传递 Redis 客户端
func Set(client *redis.Client, key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key '%s': %v", key, err)
	}
	return nil
}

// Get 获取 key 对应的值，传递 Redis 客户端
func Get(client *redis.Client, key string) (string, error) {
	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key '%s' not found", key)
		}
		return "", fmt.Errorf("failed to get key '%s': %v", key, err)
	}
	return val, nil
}

// Del 删除指定 key，传递 Redis 客户端
func Del(client *redis.Client, keys ...string) error {
	if len(keys) <= 0 {
		return nil
	}
	ctx := context.Background()
	err := client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %v", err)
	}
	return nil
}

// Exists 检查 key 是否存在，传递 Redis 客户端
func Exists(client *redis.Client, key string) (bool, error) {
	ctx := context.Background()
	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %v", err)
	}
	return exists > 0, nil
}

// HSet 设置哈希表 key 中的字段 field 的值为 value，传递 Redis 客户端
func HSet(client *redis.Client, key, field string, value interface{}) error {
	ctx := context.Background()
	err := client.HSet(ctx, key, field, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field: %v", err)
	}
	return nil
}

// HGet 获取哈希表 key 中字段 field 的值，传递 Redis 客户端
func HGet(client *redis.Client, key, field string) (string, error) {
	ctx := context.Background()
	val, err := client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("hash field '%s' not found", field)
		}
		return "", fmt.Errorf("failed to get hash field '%s': %v", field, err)
	}
	return val, nil
}

// HMSet 批量设置哈希表 key 中的字段，传递 Redis 客户端
func HMSet(client *redis.Client, key string, fields map[string]interface{}) error {
	ctx := context.Background()
	err := client.HMSet(ctx, key, fields).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash fields: %v", err)
	}
	return nil
}

// HGetAll 获取哈希表 key 中所有的字段和值，传递 Redis 客户端
func HGetAll(client *redis.Client, key string) (map[string]string, error) {
	ctx := context.Background()
	result, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields: %v", err)
	}
	return result, nil
}

// LPush 将一个或多个值插入到列表头部，传递 Redis 客户端
func LPush(client *redis.Client, key string, values ...interface{}) error {
	ctx := context.Background()
	err := client.LPush(ctx, key, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to push values to list: %v", err)
	}
	return nil
}

// LPop 移出并获取列表的第一个元素，传递 Redis 客户端
func LPop(client *redis.Client, key string) (string, error) {
	ctx := context.Background()
	val, err := client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list '%s' is empty", key)
		}
		return "", fmt.Errorf("failed to pop value from list '%s': %v", key, err)
	}
	return val, nil
}

// LRange 获取列表指定范围内的所有元素，传递 Redis 客户端
func LRange(client *redis.Client, key string, start, stop int64) ([]string, error) {
	ctx := context.Background()
	result, err := client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range of list '%s': %v", key, err)
	}
	return result, nil
}

// ZAdd 将一个或多个成员元素及其分数值加入到有序集 key 中，传递 Redis 客户端
func ZAdd(client *redis.Client, key string, members ...redis.Z) error {
	ctx := context.Background()
	err := client.ZAdd(ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to add members to sorted set: %v", err)
	}
	return nil
}

// ZRangeByScore 获取有序集 key 中指定分数范围的成员，传递 Redis 客户端
func ZRangeByScore(client *redis.Client, key string, opt *redis.ZRangeBy) ([]string, error) {
	ctx := context.Background()
	result, err := client.ZRangeByScore(ctx, key, opt).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range by score from sorted set '%s': %v", key, err)
	}
	return result, nil
}

// Keys 获取keys
func Keys(client *redis.Client, pattern string) ([]string, error) {
	ctx := context.Background()
	result, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys '%s': %v", pattern, err)
	}
	return result, nil
}

func Incr(client *redis.Client, key string) (int64, error) {
	ctx := context.Background()
	result, err := client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to incr key '%s': %v", key, err)
	}
	return result, nil
}

// HIncrBy 对哈希表 key 中的字段 field 增加增量
func HIncrBy(client *redis.Client, key, field string, incr int64) error {
	ctx := context.Background()
	err := client.HIncrBy(ctx, key, field, incr).Err()
	if err != nil {
		return fmt.Errorf("failed to hincrby key '%s' field '%s': %v", key, field, err)
	}
	return nil
}

// LTrim 修剪列表，只保留指定范围内的元素
func LTrim(client *redis.Client, key string, start, stop int64) error {
	ctx := context.Background()
	err := client.LTrim(ctx, key, start, stop).Err()
	if err != nil {
		return fmt.Errorf("failed to ltrim key '%s': %v", key, err)
	}
	return nil
}

func TTL(client *redis.Client, key string) (time.Duration, error) {
	ctx := context.Background()
	result, err := client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to ttl key '%s': %v", key, err)
	}
	return result, nil
}

func Expire(client *redis.Client, key string, expiration time.Duration) error {
	ctx := context.Background()
	err := client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetCacheObject(client *redis.Client, key string, obj interface{}) error {
	val, err := Get(client, key)
	if err != nil {
		return err
	}
	// 解析JSON字符串到结构体
	err = json.Unmarshal([]byte(val), &obj)
	if err != nil {
		return fmt.Errorf("Error parsing JSON:", err)
	}
	return nil
}

func SetCacheObject(client *redis.Client, key string, obj interface{}, expireTime time.Duration) error {
	value, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	err = Set(client, key, string(value), expireTime)
	if err != nil {
		return err
	}
	return nil
}
