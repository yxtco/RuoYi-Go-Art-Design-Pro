package taskregistry

import (
	"encoding/json"
	"fmt"

	"go-fin-server/internal/service"
	"go-fin-server/pkg"
)

// TaskFunc 任务执行函数类型
type TaskFunc func(args string) error

// taskFunctions 全局任务函数注册表
var taskFunctions = map[string]TaskFunc{
	// 示例函数
	"ryNoParams": func(args string) error {
		pkg.Logger.Info("执行 ryNoParams 任务")
		return nil
	},
	"ryParams": func(args string) error {
		pkg.Logger.Infof("执行 ryParams 任务，参数：%s", args)
		return nil
	},
	"test1": func(args string) error {
		pkg.Logger.Info("success1")
		return nil
	},
	"test2": func(args string) error {
		var params struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Errorf("failed to unmarshal args: %v", err)
		}
		pkg.Logger.Infof("test2: %+v", params)
		return nil
	},
	"test3": func(args string) error {
		var params struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return fmt.Errorf("failed to unmarshal args: %v", err)
		}
		pkg.Logger.Infof("test3: %+v", params)
		return nil
	},
	"cleanOperLog": func(args string) error {
		pkg.Logger.Info("执行清除操作日志任务")
		err := service.Services{}.SysOperLogService.CleanOperLog()
		if err != nil {
			return fmt.Errorf("清除操作日志失败: %v", err)
		}
		pkg.Logger.Info("操作日志清除完成")
		return nil
	},
}

// Register 注册一个任务函数
func Register(name string, fn TaskFunc) {
	taskFunctions[name] = fn
}

// Execute 执行指定的任务函数
func Execute(target string, args string) error {
	fn, exists := taskFunctions[target]
	if !exists {
		return fmt.Errorf("任务函数 %s 不存在", target)
	}
	return fn(args)
}

// Exists 检查任务函数是否存在
func Exists(target string) bool {
	_, exists := taskFunctions[target]
	return exists
}
