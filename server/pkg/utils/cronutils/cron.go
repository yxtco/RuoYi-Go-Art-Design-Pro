package cronutils

import (
	"go-fin-server/pkg"
	"time"

	"github.com/robfig/cron/v3"
)

// IsValid 校验是否正确的cron表达式
func IsValid(expr string) bool {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expr)
	if err != nil {
		pkg.Logger.Error(err)
		return false
	}
	return true
}

// GetNextExecution 获取下一次执行时间
func GetNextExecution(expr string) (string, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		pkg.Logger.Error(err)
		return "", err
	}
	now := time.Now()
	nextRun := schedule.Next(now)
	return nextRun.Format("2006-01-02 15:04:05"), nil
}
