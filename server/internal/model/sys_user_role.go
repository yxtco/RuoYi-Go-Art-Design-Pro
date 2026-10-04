package model

import (
	"go-fin-server/internal/db"
)

type SysUserRole struct {
	UserId  uint64  `gorm:"column:user_id" json:"userId"        description:"用户ID"`
	RoleId  uint64  `gorm:"column:role_id" json:"roleId"        description:"角色ID"`
	SysRole SysRole `gorm:"foreignKey:id;references:role_id"`
}

func (c SysUserRole) TableName() string {
	return "sys_user_role"
}

// GetRolesByUserId 根据用户id获取角色
func (c SysUserRole) GetRolesByUserId(userId uint64) ([]SysUserRole, error) {
	data := make([]SysUserRole, 0)
	d := db.DBConnections["master"].Preload("SysRole").Where("user_id = ?", userId).Find(&data)
	return data, d.Error
}

func (c SysUserRole) DeleteUserRole(userIds []uint64) error {
	d := db.DBConnections["master"].Where("user_id  in (?)", userIds).Delete(&SysUserRole{})
	return d.Error
}

func (c SysUserRole) DeleteUserRoleByUserId(userId uint64) error {
	d := db.DBConnections["master"].Where("user_id = ?", userId).Delete(&SysUserRole{})
	return d.Error
}

func (c SysUserRole) BatchUserRole(list []SysUserRole) error {
	if len(list) > 0 {
		d := db.DBConnections["master"].Create(&list)
		return d.Error
	}
	return nil
}

func (c SysUserRole) CountUserRoleByRoleId(roleId uint64) (int64, error) {
	var count int64
	d := db.DBConnections["master"].Model(&SysUserRole{}).Where("role_id = ?", roleId).Count(&count)
	return count, d.Error
}

func (c SysUserRole) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysUserRole{})
	return d.Error
}
