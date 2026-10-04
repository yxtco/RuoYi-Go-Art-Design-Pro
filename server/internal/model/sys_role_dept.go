package model

import "go-fin-server/internal/db"

type SysRoleDept struct {
	RoleId uint64 `gorm:"column:role_id" json:"roleId" description:"角色ID"`
	DeptId uint64 `gorm:"column:dept_id" json:"deptId" description:"部门ID"`
}

func (c SysRoleDept) TableName() string {
	return "sys_role_dept"
}

func (c SysRoleDept) SelectRoleDeptAllList(conditions map[string]interface{}) ([]SysRoleDept, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysRoleDept{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysRoleDept, 0)
	d := query.Find(&data)
	return data, d.Error
}

func (c SysRoleDept) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysRoleDept{})
	return d.Error
}

func (c SysRoleDept) BatchRoleDept(list []SysRoleDept) error {
	if len(list) > 0 {
		d := db.DBConnections["master"].Create(&list)
		return d.Error
	}
	return nil
}
