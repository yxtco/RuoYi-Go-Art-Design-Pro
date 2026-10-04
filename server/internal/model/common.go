package model

import (
	"go-fin-server/pkg/types"
	"sort"
	"strconv"

	"gorm.io/gorm"
)

type BaseModel struct {
	CreateBy  string          `gorm:"column:create_by" json:"createBy"     description:"创建者"`
	UpdateBy  string          `gorm:"column:update_by" json:"updateBy"     description:"更新者"`
	CreatedAt types.LocalTime `gorm:"column:create_time" json:"createTime"     description:"创建时间"`
	UpdatedAt types.LocalTime `gorm:"column:update_time" json:"updateTime"     description:"更新时间"`
	DeletedAt gorm.DeletedAt  `gorm:"column:delete_time" json:"-"      description:"删除时间"`
}

type BaseNoDelModel struct {
	CreateBy  string          `gorm:"column:create_by" json:"createBy"     description:"创建者"`
	UpdateBy  string          `gorm:"column:update_by" json:"updateBy"     description:"更新者"`
	CreatedAt types.LocalTime `gorm:"column:create_time" json:"createTime"     description:"创建时间"`
	UpdatedAt types.LocalTime `gorm:"column:update_time" json:"updateTime"     description:"更新时间"`
}

type LoginUser struct {
	UserId        uint64   `json:"userId"`        // 用户ID
	UserName      string   `json:"userName"`      // 用户名称
	DeptId        uint64   `json:"deptId"`        // 部门ID
	DeptName      string   `json:"deptName"`      // 部门名称
	Token         string   `json:"token"`         // 用户唯一标识
	LoginTime     int64    `json:"loginTime"`     // 登录时间
	ExpireTime    int64    `json:"expireTime"`    // 过期时间
	Ipaddr        string   `json:"ipaddr"`        // 登录IP地址
	LoginLocation string   `json:"loginLocation"` // 登录地点
	Browser       string   `json:"browser"`       // 浏览器类型
	Os            string   `json:"os"`            // 操作系统
	Permissions   []string `json:"permissions"`   // 权限列表 set
	User          SysUser  `json:"user"`          // 用户信息
}

// DataScope 数据权限作用域
func DataScope(user SysUser, tableName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if user.IsAdmin(user.Id) {
			return db
		}
		// 获取当前用户的所有角色data_scope 取最小值
		roles, err := SysUserRole{}.GetRolesByUserId(user.Id)
		if err != nil {
			return nil
		}
		dataScopes := make([]int, 0)
		roleIds := make([]uint64, 0)
		for _, item := range roles {
			num, _ := strconv.Atoi(item.SysRole.DataScope)
			dataScopes = append(dataScopes, num)
			roleIds = append(roleIds, item.SysRole.Id)
		}
		if len(dataScopes) > 0 {
			sort.Ints(dataScopes)
			switch tableName {
			case "sys_user":
				switch dataScopes[0] {
				case 1: // 全部数据权限
					return db
				case 2: // 自定数据权限
					return db.Or("dept_id IN ( SELECT dept_id FROM sys_role_dept WHERE role_id in (?) )", roleIds)
				case 3: // 部门数据权限
					return db.Or("dept_id = ?", user.DeptId)
				case 4: // 部门及以下数据权限
					return db.Or("dept_id IN ( SELECT id FROM sys_dept WHERE id = ? or find_in_set( ? , ancestors ) )", user.DeptId, user.DeptId)
				case 5: // 仅本人数据权限
					return db.Or("id = ?", user.Id)
				default:
					return db
				}
			case "sys_dept":
				switch dataScopes[0] {
				case 1: // 全部数据权限
					return db
				case 2: // 自定数据权限
					return db.Or("id IN ( SELECT dept_id FROM sys_role_dept WHERE role_id in (?) )", roleIds)
				case 3: // 部门数据权限
					return db.Or("id = ?", user.DeptId)
				case 4: // 部门及以下数据权限
					return db.Or("id IN ( SELECT id FROM sys_dept WHERE id = ? or find_in_set( ? , ancestors ) )", user.DeptId, user.DeptId)
				case 5: // 仅本人数据权限
					return db.Or("id = ?", user.DeptId)
				default:
					return db
				}
			case "sys_role":
				switch dataScopes[0] {
				case 1: // 全部数据权限
					return db
				case 2: // 自定数据权限
					return db.Or("id in (SELECT role_id FROM sys_user_role where user_id in (SELECT id from sys_user where dept_id in (SELECT dept_id FROM sys_role_dept WHERE role_id in (?)) ))", roleIds)
				case 3: // 部门数据权限
					return db.Or("id in (SELECT role_id FROM sys_user_role where user_id in (SELECT id from sys_user where dept_id = ?))", user.DeptId)
				case 4: // 部门及以下数据权限
					return db.Or("id in (SELECT role_id FROM sys_user_role where user_id in (SELECT id from sys_user where dept_id in (SELECT id FROM sys_dept WHERE id = ? or find_in_set( ? , ancestors ) ) ))", user.DeptId, user.DeptId)
				case 5: // 仅本人数据权限
					return db.Or("id in (SELECT role_id FROM sys_user_role where user_id = ?)", user.Id)
				default:
					return db
				}
			}
		}
		return db
	}
}
