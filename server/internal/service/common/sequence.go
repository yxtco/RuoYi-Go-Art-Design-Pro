package common

import (
	"fmt"
	"go-fin-server/internal/constant"
	"go-fin-server/pkg/redistool"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// SequenceService 序号生成服务
type SequenceService struct{}

// SequenceRequest 序号生成请求参数
type SequenceRequest struct {
	BusinessType string `form:"businessType" binding:"required"` // 业务类型（如 ORDER, INVOICE）
	Prefix       string `form:"prefix"`                          // 序号前缀（如 "ORD", "INV"）
	DatePattern  string `form:"datePattern"`                     // 日期格式（如 "20060102", "200601", ""）
	SeqLength    int    `form:"seqLength"`                       // 序号位数（如 3 → "001"）
	ResetCycle   string `form:"resetCycle"`                      // 重置周期: daily/monthly/yearly/never
}

// NextSequence 生成下一个序号
// 返回完整的序号字符串（包含前缀+日期+自增序号）
func (s SequenceService) NextSequence(client *redis.Client, req SequenceRequest) (string, error) {
	// 设置默认值
	if req.SeqLength <= 0 {
		req.SeqLength = 3
	}
	if req.ResetCycle == "" {
		req.ResetCycle = "daily"
	}

	// 构建 Redis key
	key := s.buildKey(req)

	// 原子自增
	seq, err := redistool.Incr(client, key)
	if err != nil {
		return "", fmt.Errorf("生成序号失败: %v", err)
	}

	// 首次创建时设置过期时间（确保 key 自动清理）
	if seq == 1 {
		expiration := s.getExpiration(req.ResetCycle)
		if expiration > 0 {
			_ = redistool.Expire(client, key, expiration)
		}
	}

	// 格式化序号（零填充）
	seqStr := fmt.Sprintf("%0*d", req.SeqLength, seq)

	// 组装完整序号
	var sb strings.Builder
	if req.Prefix != "" {
		sb.WriteString(req.Prefix)
	}
	if req.DatePattern != "" {
		sb.WriteString(time.Now().Format(req.DatePattern))
	}
	sb.WriteString(seqStr)

	return sb.String(), nil
}

// GetCurrentSequence 获取当前序号值（不自增）
func (s SequenceService) GetCurrentSequence(client *redis.Client, req SequenceRequest) (int64, error) {
	if req.SeqLength <= 0 {
		req.SeqLength = 3
	}
	if req.ResetCycle == "" {
		req.ResetCycle = "daily"
	}

	key := s.buildKey(req)

	val, err := redistool.Get(client, key)
	if err != nil {
		return 0, nil // key 不存在返回 0
	}

	var seq int64
	_, err = fmt.Sscanf(val, "%d", &seq)
	if err != nil {
		return 0, fmt.Errorf("解析序号失败: %v", err)
	}

	return seq, nil
}

// ResetSequence 重置指定业务类型的序号
func (s SequenceService) ResetSequence(client *redis.Client, req SequenceRequest) error {
	if req.ResetCycle == "" {
		req.ResetCycle = "daily"
	}

	key := s.buildKey(req)
	return redistool.Del(client, key)
}

// buildKey 构建 Redis key
// 格式: sequence:{businessType}:{dateSuffix}
func (s SequenceService) buildKey(req SequenceRequest) string {
	var sb strings.Builder
	sb.WriteString(constant.CACHE_SEQUENCE_KEY)
	sb.WriteString(req.BusinessType)

	// 根据重置周期添加日期后缀
	switch req.ResetCycle {
	case "daily":
		sb.WriteString(":")
		sb.WriteString(time.Now().Format("20060102"))
	case "monthly":
		sb.WriteString(":")
		sb.WriteString(time.Now().Format("200601"))
	case "yearly":
		sb.WriteString(":")
		sb.WriteString(time.Now().Format("2006"))
	case "never":
		// 不添加日期后缀
	default:
		// 默认按天重置
		sb.WriteString(":")
		sb.WriteString(time.Now().Format("20060102"))
	}

	return sb.String()
}

// getExpiration 根据重置周期获取 key 过期时间
func (s SequenceService) getExpiration(resetCycle string) time.Duration {
	now := time.Now()
	switch resetCycle {
	case "daily":
		// 到明天凌晨
		tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		return tomorrow.Sub(now) + time.Hour // 多加1小时缓冲
	case "monthly":
		// 到下个月1号
		nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
		return nextMonth.Sub(now) + time.Hour
	case "yearly":
		// 到明年1月1号
		nextYear := time.Date(now.Year()+1, 1, 1, 0, 0, 0, 0, now.Location())
		return nextYear.Sub(now) + time.Hour
	case "never":
		return 0 // 不过期
	default:
		return 25 * time.Hour
	}
}
