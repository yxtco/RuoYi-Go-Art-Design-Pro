package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils"
)

type SysOperLog struct {
	Id            uint64          `gorm:"column:id;primary_key;" json:"id"        description:"日志主键"`
	Title         string          `gorm:"column:title" json:"title"         description:"模块标题"`
	BusinessType  int             `gorm:"column:business_type" json:"businessType"  description:"业务类型（0其它 1新增 2修改 3删除）"`
	Method        string          `gorm:"column:method" json:"method"        description:"方法名称"`
	RequestMethod string          `gorm:"column:request_method" json:"requestMethod" description:"请求方式"`
	OperatorType  int             `gorm:"column:operator_type" json:"operatorType"  description:"操作类别（0其它 1后台用户 2手机端用户）"`
	OperName      string          `gorm:"column:oper_name" json:"operName"      description:"操作人员"`
	DeptName      string          `gorm:"column:dept_name" json:"deptName"      description:"部门名称"`
	OperUrl       string          `gorm:"column:oper_url" json:"operUrl"       description:"请求URL"`
	OperIp        string          `gorm:"column:oper_ip" json:"operIp"        description:"主机地址"`
	OperLocation  string          `gorm:"column:oper_location" json:"operLocation"  description:"操作地点"`
	OperParam     string          `gorm:"column:oper_param" json:"operParam"     description:"请求参数"`
	JsonResult    string          `gorm:"column:json_result" json:"jsonResult"     description:"返回参数"`
	Status        int             `gorm:"column:status" json:"status"     description:"操作状态（0正常 1异常）"`
	ErrorMsg      string          `gorm:"column:error_msg" json:"errorMsg"      description:"错误消息"`
	OperTime      types.LocalTime `gorm:"column:oper_time" json:"operTime"      description:"操作时间"`
	CostTime      uint64          `gorm:"column:cost_time" json:"costTime"      description:"消耗时间"`
}

func (c SysOperLog) TableName() string {
	return "sys_oper_log"
}

func (c SysOperLog) SelectOperLogList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysOperLog, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysOperLog{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysOperLog, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysOperLog) SelectOperLogAllList(conditions map[string]interface{}) ([]SysOperLog, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysOperLog{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysOperLog, 0)
	d := query.Order("oper_time desc").Find(&data)
	return data, d.Error
}

func (c SysOperLog) Clean() error {
	sql := `
		truncate table sys_oper_log;
	`
	d := db.DBConnections["master"].Exec(sql)
	return d.Error
}

func (c SysOperLog) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysOperLog{})
	return d.Error
}
