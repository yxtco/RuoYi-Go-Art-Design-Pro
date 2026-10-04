package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-fin-server/api/common"
	"go-fin-server/api/monitor"
	"go-fin-server/api/system"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Api struct {
	Common        common.Common
	SysUser       system.SysUser
	SysRole       system.SysRole
	SysMenu       system.SysMenu
	SysDept       system.SysDept
	SysPost       system.SysPost
	SysNotice     system.SysNotice
	SysDictType   system.SysDictType
	SysDictData   system.SysDictData
	SysConfig     system.SysConfig
	SysOperLog    monitor.SysOperLog
	SysUserOnline monitor.SysUserOnline
	SysLoginInfo  monitor.SysLoginInfo
	SysJob        monitor.SysJob
	SysJobLog     monitor.SysJobLog
}

func ShouldBindError(c *gin.Context, req interface{}) error {
	// 使用 ShouldBind 绑定请求数据到结构体
	if err := c.ShouldBind(req); err != nil {
		// 检查是否为 JSON 解析错误 jsonSyntaxErr
		if _, ok := err.(*json.SyntaxError); ok {
			return errors.New("Invalid JSON syntax")
		}

		// 检查是否为参数类型错误
		if jsonTypeErr, ok := err.(*json.UnmarshalTypeError); ok {
			return errors.New(fmt.Sprintf("Invalid type for field %s, expected %s", jsonTypeErr.Field, jsonTypeErr.Type))
		}

		// 判断是否是 validator.ValidationErrors 类型
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				// 获取字段名
				fieldName := fe.Field()

				// 获取字段的自定义标签
				customTag := fe.Tag()

				// 获取结构体字段的自定义错误消息
				customMsg, _ := getCustomErrorMessage(reflect.TypeOf(req).Elem(), fieldName)

				// 根据字段的自定义标签或结构体字段的自定义错误消息生成自定义错误消息
				errorMsg := customMsg
				if errorMsg == "" {
					errorMsg = fmt.Sprintf("'%s' is %s", fieldName, customTag)
				}
				return errors.New(errorMsg)
			}
		} else {
			// 如果不是 validator.ValidationErrors 类型，直接返回错误
			return err
		}
	}
	return nil
}

func ShouldBindUriError(c *gin.Context, req interface{}) error {
	// 使用 ShouldBind 绑定请求数据到结构体
	if err := c.ShouldBindUri(req); err != nil {
		// 检查是否为 JSON 解析错误 jsonSyntaxErr
		if _, ok := err.(*json.SyntaxError); ok {
			return errors.New("Invalid JSON syntax")
		}

		// 检查是否为参数类型错误
		if jsonTypeErr, ok := err.(*json.UnmarshalTypeError); ok {
			return errors.New(fmt.Sprintf("Invalid type for field %s, expected %s", jsonTypeErr.Field, jsonTypeErr.Type))
		}

		// 判断是否是 validator.ValidationErrors 类型
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			for _, fe := range ve {
				// 获取字段名
				fieldName := fe.Field()

				// 获取字段的自定义标签
				customTag := fe.Tag()

				// 获取结构体字段的自定义错误消息
				customMsg, _ := getCustomErrorMessage(reflect.TypeOf(req).Elem(), fieldName)

				// 根据字段的自定义标签或结构体字段的自定义错误消息生成自定义错误消息
				errorMsg := customMsg
				if errorMsg == "" {
					errorMsg = fmt.Sprintf("'%s' is %s", fieldName, customTag)
				}
				return errors.New(errorMsg)
			}
		} else {
			// 如果不是 validator.ValidationErrors 类型，直接返回错误
			return err
		}
	}
	return nil
}

// getCustomErrorMessage 根据结构体和字段名获取自定义错误消息
func getCustomErrorMessage(structType reflect.Type, fieldName string) (string, error) {
	// 获取字段的 struct tag
	field, ok := structType.FieldByName(fieldName)
	if !ok {
		return "", fmt.Errorf("field not found: %s", fieldName)
	}
	// 获取自定义的 custom_msg 标签
	customMsg := field.Tag.Get("err_msg")
	return customMsg, nil
}
