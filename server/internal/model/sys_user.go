package model

import (
	"fmt"
	"go-fin-server/internal/db"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils"
)

type SysUser struct {
	Id          uint64          `gorm:"column:id;primary_key" json:"id"            description:""`
	DeptId      uint64          `gorm:"column:dept_id" json:"deptId"        description:"部门id"`
	UserName    string          `gorm:"column:user_name" json:"userName"      description:"用户名" binding:"required" err_msg:"用户名称 必填"`
	NickName    string          `gorm:"column:nick_name" json:"nickName"  description:"用户昵称" binding:"required" err_msg:"用户昵称 必填"`
	UserType    string          `gorm:"column:user_type;default:'00'" json:"userType"  description:"用户类型（00系统用户）"`
	Email       string          `gorm:"column:email" json:"email"     description:"用户登录邮箱"`
	Phonenumber string          `gorm:"column:phonenumber" json:"phonenumber"        description:"中国手机不带国家代码，国际手机号格式为：国家代码-手机号"`
	Sex         string          `gorm:"column:sex;default:'0'" json:"sex"           description:"性别;0:保密,1:男,2:女"`
	Avatar      string          `gorm:"column:avatar" json:"avatar"        description:"用户头像"`
	Password    string          `gorm:"column:password" json:"password"  description:"登录密码;cmf_password加密" binding:"required" err_msg:"用户密码 必填"`
	Status      string          `gorm:"column:status;default:'0'" json:"status"    description:"用户状态;0:禁用,1:正常,2:未验证"`
	LoginIp     string          `gorm:"column:login_ip" json:"loginIp"   description:"最后登录ip"`
	LoginDate   types.LocalTime `gorm:"column:login_date" json:"loginDate"   description:"最后登录时间"`
	Remark      string          `gorm:"column:remark" json:"remark"        description:"备注"`
	BaseModel
	SysDept SysDept `gorm:"foreignKey:id;references:dept_id" json:"dept"`
}

type SysUserAll struct {
	SysUser
	SysUserRoles []SysUserRole `gorm:"foreignKey:user_id;references:id" json:"-"`
	SysUserPosts []SysUserPost `gorm:"foreignKey:user_id;references:id" json:"-"`
	RoleIds      []uint64      `gorm:"-" json:"roleIds"`
	Roles        []SysRole     `gorm:"-" json:"roles"`
	PostIds      []int64       `gorm:"-" json:"postIds"`
	Posts        []SysPost     `gorm:"-" json:"posts"`
}

func (c SysUserAll) TableName() string {
	return "sys_user"
}

func (c SysUser) TableName() string {
	return "sys_user"
}

func (c SysUser) List(conditions map[string]interface{}) ([]SysUser, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysUser{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysUser, 0)
	d := query.Preload("SysDept").Order("create_time desc").Find(&data)

	return data, d.Error
}

func (c SysUser) SelectUserList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysUser, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysUser{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysUser, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Preload("SysDept").Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysUser) SelectDataScopeUserList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int, user SysUser) ([]SysUser, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysUser{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	query = query.Scopes(DataScope(user, c.TableName()))
	data := make([]SysUser, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Preload("SysDept").Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysUser) SelectUserAllList(conditions map[string]interface{}) ([]SysUser, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysUser{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysUser, 0)
	d := query.Preload("SysDept").Order("create_time desc").Find(&data)
	return data, d.Error
}

func (c SysUser) Get(query string, args ...interface{}) (SysUser, error) {
	var data SysUser
	d := db.DBConnections["master"].Preload("SysDept").Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysUser) GetAll(query string, args ...interface{}) (SysUserAll, error) {
	var data SysUserAll
	d := db.DBConnections["master"].Preload("SysDept").Preload("SysUserRoles").Preload("SysUserPosts").Where(query, args...).First(&data)
	if d.Error != nil {
		return data, d.Error
	}
	// 查询角色
	roleIds := make([]uint64, 0)
	for _, role := range data.SysUserRoles {
		roleIds = append(roleIds, role.RoleId)
	}
	roles := make([]SysRole, 0)
	d = db.DBConnections["master"].Model(&SysRole{}).Where("id in (?)", roleIds).Find(&roles)
	if d.Error != nil {
		return data, d.Error
	}
	data.Roles = roles
	// 查询岗位
	postIds := make([]int64, 0)
	for _, post := range data.SysUserPosts {
		postIds = append(postIds, post.PostId)
	}
	posts := make([]SysPost, 0)
	d = db.DBConnections["master"].Model(&SysPost{}).Where("id in (?)", postIds).Find(&posts)
	data.RoleIds = roleIds
	data.Roles = roles
	data.PostIds = postIds
	data.Posts = posts
	return data, d.Error
}

func (c SysUser) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysUser{})
	return d.Error
}

func (c SysUser) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(SysUser{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysUser) Create(user *SysUser) error {
	return db.DBConnections["master"].Create(&user).Error
}

func (c SysUser) IsAdmin(userId uint64) bool {
	return 1 == userId
}
func (c SysUser) SelectAllocatedList(roleId uint64, user SysUser, pageNum, pageSize int) ([]SysUser, int64, error) {
	args := make([]interface{}, 0)
	sql := `
			    select distinct u.id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time
	    from sys_user u
			 left join sys_dept d on u.dept_id = d.id
			 left join sys_user_role ur on u.id = ur.user_id
			 left join sys_role r on r.id = ur.role_id
	    where u.delete_time is NULL and r.id = ?
	`
	args = append(args, roleId)
	if user.UserName != "" {
		sql += ` AND u.user_name like concat('%', ?, '%')`
		args = append(args, user.UserName)
	}
	if user.Phonenumber != "" {
		sql += ` AND u.phonenumber like concat('%', ?, '%')`
		args = append(args, user.Phonenumber)
	}

	countSql := fmt.Sprintf(`select count(1) from (%s) as t1 `, sql)
	offset, limit := utils.PageParse(pageNum, pageSize)

	sql += fmt.Sprintf(` limit %d offset %d`, limit, offset)
	data := make([]SysUser, 0)
	var count int64
	if err := db.DBConnections["master"].Raw(countSql, args...).Scan(&count).Error; err != nil {
		return data, count, err
	}
	d := db.DBConnections["master"].Raw(sql, args...).Scan(&data)
	return data, count, d.Error
}

func (c SysUser) SelectUnallocatedList(roleId uint64, user SysUser, pageNum, pageSize int) ([]SysUser, int64, error) {
	args := make([]interface{}, 0)
	sql := `
		select distinct u.id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time
	    from sys_user u
			 left join sys_dept d on u.dept_id = d.id
			 left join sys_user_role ur on u.id = ur.user_id
			 left join sys_role r on r.id = ur.role_id
	    where u.delete_time is NULL  and (r.id != ? or r.id IS NULL)
	    and u.id not in (select u.id from sys_user u inner join sys_user_role ur on u.id = ur.user_id and ur.role_id = ?)
	`
	args = append(args, roleId)
	args = append(args, roleId)
	if user.UserName != "" {
		sql += ` AND u.user_name like concat('%', ?, '%')`
		args = append(args, user.UserName)
	}
	if user.Phonenumber != "" {
		sql += ` AND u.phonenumber like concat('%', ?, '%')`
		args = append(args, user.Phonenumber)
	}
	countSql := fmt.Sprintf(`select count(1) from (%s) as t1 `, sql)
	offset, limit := utils.PageParse(pageNum, pageSize)

	sql += fmt.Sprintf(` limit %d offset %d`, limit, offset)
	data := make([]SysUser, 0)
	var count int64
	if err := db.DBConnections["master"].Raw(countSql, args...).Scan(&count).Error; err != nil {
		return data, count, err
	}
	d := db.DBConnections["master"].Raw(sql, args...).Scan(&data)
	return data, count, d.Error
}
