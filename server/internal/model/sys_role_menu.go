package model

import "go-fin-server/internal/db"

type SysRoleMenu struct {
	RoleId uint64 `gorm:"column:role_id" json:"roleId"        description:"角色ID"`
	MenuId uint64 `gorm:"column:menu_id" json:"menuId"        description:"菜单ID"`
}

func (c SysRoleMenu) TableName() string {
	return "sys_role_menu"
}

func (c SysRoleMenu) BatchRoleMenu(list []SysRoleMenu) error {
	if len(list) > 0 {
		d := db.DBConnections["master"].Create(&list)
		return d.Error
	}
	return nil
}

func (c SysRoleMenu) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysRoleMenu{})
	return d.Error
}

func (c SysRoleMenu) CheckMenuExistRole(query string, args ...interface{}) (int64, error) {
	var count int64
	if err := db.DBConnections["master"].Model(&SysRoleMenu{}).Where(query, args...).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}
