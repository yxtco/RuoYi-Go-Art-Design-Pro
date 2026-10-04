package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils"
)

type SysLoginInfo struct {
	Id            int64           `gorm:"column:id"  json:"id"        description:"访问ID"`
	UserName      string          `gorm:"column:user_name" json:"userName"     description:"登录账号"`
	Ipaddr        string          `gorm:"column:ipaddr" json:"ipaddr"        description:"登录IP地址"`
	LoginLocation string          `gorm:"column:login_location" json:"loginLocation" description:"登录地点"`
	Browser       string          `gorm:"column:browser" json:"browser"       description:"浏览器类型"`
	Os            string          `gorm:"column:os" json:"os"            description:"操作系统"`
	Status        string          `gorm:"column:status;default:'0'" json:"status"        description:"登录状态（0成功 1失败）"`
	Msg           string          `gorm:"column:msg" json:"msg"           description:"提示消息"`
	LoginTime     types.LocalTime `gorm:"column:login_time" json:"loginTime"     description:"登录时间"`
	//Module        string          `json:"module"        description:"登录模块"`
}

func (c SysLoginInfo) TableName() string {
	return "sys_login_info"
}

func (c SysLoginInfo) SelectLoginInfoList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysLoginInfo, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysLoginInfo{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysLoginInfo, 0)
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

func (c SysLoginInfo) SelectLoginInfoAllList(conditions map[string]interface{}) ([]SysLoginInfo, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysLoginInfo{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysLoginInfo, 0)
	d := query.Order("login_time desc").Find(&data)
	return data, d.Error
}

func (c SysLoginInfo) Get(query string, args ...interface{}) (SysLoginInfo, error) {
	var data SysLoginInfo
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysLoginInfo) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(SysLoginInfo{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysLoginInfo) Create() error {
	return db.DBConnections["master"].Create(&c).Error
}

func (c SysLoginInfo) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysLoginInfo{})
	return d.Error
}

func (c SysLoginInfo) Clean() error {
	sql := `
		truncate table sys_login_info;
	`
	d := db.DBConnections["master"].Exec(sql)
	return d.Error
}
