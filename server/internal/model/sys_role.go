package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysRole struct {
	Id                uint64   `gorm:"column:id;primary_key;" json:"id"        description:""`
	RoleName          string   `gorm:"column:role_name" json:"roleName"        description:"角色名称" binding:"required" err_msg:"角色名称 必填"`
	RoleKey           string   `gorm:"column:role_key" json:"roleKey"        description:"角色权限字符串" binding:"required" err_msg:"权限字符 必填"`
	RoleSort          *uint    `gorm:"column:role_sort" json:"roleSort"        description:"显示顺序" binding:"required" err_msg:"角色顺序 必填"`
	DataScope         string   `gorm:"column:data_scope;default:'1'" json:"dataScope" description:"数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限）"`
	MenuCheckStrictly bool     `gorm:"column:menu_check_strictly;default:1" json:"menuCheckStrictly"        description:"菜单树选择项是否关联显示"`
	DeptCheckStrictly bool     `gorm:"column:dept_check_strictly;default:1" json:"deptCheckStrictly"        description:"部门树选择项是否关联显示"`
	Status            string   `gorm:"column:status" json:"status"    description:"角色状态（0正常 1停用）"`
	Remark            string   `gorm:"column:remark" json:"remark"        description:"备注"`
	Flag              bool     `gorm:"-" json:"flag"        description:"是否已拥有的角色"`
	MenuIds           []uint64 `gorm:"-" json:"menuIds"        description:"角色菜单"`
	BaseModel
}

func (c SysRole) TableName() string {
	return "sys_role"
}

type SysRoleAll struct {
	SysRole
	MenuIds []uint64 `gorm:"-" json:"menuIds"`
	DeptIds []uint64 `gorm:"-" json:"deptIds"`
}

func (c SysRoleAll) TableName() string {
	return "sys_role"
}

func (c SysRole) SelectRoleList(conditions map[string]interface{}, pageNum, pageSize int) ([]SysRole, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysRole{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	var count int64
	data := make([]SysRole, 0)
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	d := query.Order("role_sort").Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysRole) SelectDataScopeRoleList(conditions map[string]interface{}, pageNum, pageSize int, user SysUser) ([]SysRole, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysRole{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	query = query.Scopes(DataScope(user, c.TableName()))
	var count int64
	data := make([]SysRole, 0)
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	d := query.Order("role_sort").Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysRole) SelectRoleAllList(conditions map[string]interface{}) ([]SysRole, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysRole{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysRole, 0)
	d := query.Order("role_sort").Find(&data)
	return data, d.Error
}

func (c SysRole) Get(query string, args ...interface{}) (SysRole, error) {
	var data SysRole
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysRole) GetAllList(query string, args ...interface{}) ([]SysRole, error) {
	data := make([]SysRole, 0)
	d := db.DBConnections["master"].Model(&SysRole{}).Where(query, args...).Find(&data)
	return data, d.Error
}

func (c SysRole) Create(role *SysRole) error {
	return db.DBConnections["master"].Create(&role).Error
}

func (c SysRole) IsAdmin(roleId uint64) bool {
	return 1 == roleId
}

func (c SysRole) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysRole{})
	return d.Error
}

func (c SysRole) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysRole{}).Where(query, args...).Updates(upd)
	return d.Error
}

const selectRoleVo = `
 select distinct r.id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly,
            r.status, r.delete_time, r.create_time, r.remark 
        from sys_role r
	        left join sys_user_role ur on ur.role_id = r.id
	        left join sys_user u on u.id = ur.user_id
	        left join sys_dept d on u.dept_id = d.id
`

func (c SysRole) SelectRolePermissionByUserId(userId uint64) ([]SysRole, error) {
	data := make([]SysRole, 0)
	sql := selectRoleVo + ` where r.delete_time is null and  ur.user_id = ?`
	d := db.DBConnections["master"].Raw(sql, userId).Scan(&data)
	return data, d.Error
}

func (c SysRole) SelectRolesByUserName(userName string) ([]SysRole, error) {
	data := make([]SysRole, 0)
	sql := selectRoleVo + ` where r.delete_time is null and  u.user_name = ?`
	d := db.DBConnections["master"].Raw(sql, userName).Scan(&data)
	return data, d.Error
}

func (c SysRole) SelectDeptListByRoleId(roleId uint64, isDeptCheckStrictly bool) ([]uint64, error) {
	args := make([]interface{}, 0)
	args = append(args, roleId)
	sql := `
		select d.id
        from sys_dept d
            left join sys_role_dept rd on d.id = rd.dept_id
        where rd.role_id = ?
	`
	if isDeptCheckStrictly {
		sql += ` and d.id not in (select d.parent_id from sys_dept d inner join sys_role_dept rd on d.id = rd.dept_id and rd.role_id = ?)`
		args = append(args, roleId)
	}
	sql += ` order by d.parent_id, d.order_num`

	data := make([]uint64, 0)
	d := db.DBConnections["master"].Raw(sql, args...).Scan(&data)
	return data, d.Error
}
