package system

import (
	"go-fin-server/api/common"
)

type SysUser struct {
	GetUserListRequest
	UserIdsUriRequest
	UserIdUriRequest
	ChangeUserStatusRequest
	UpdateAuthRoleRequest
	ImportUserRequest
	ResetPwdRequest
	UpdateUserPwdRequest
	UpdateUserRequest
	UpdateUserProfileRequest
}

type UpdateUserRequest struct {
	Id          uint64   `gorm:"column:id;primary_key" json:"id"            description:"" binding:"required" err_msg:"用户Id 必填"`
	UserName    string   `gorm:"-" json:"userName"      description:"用户名" binding:"required" err_msg:"用户名 必填"`
	DeptId      uint64   `gorm:"column:dept_id" json:"deptId"        description:"部门id"`
	NickName    string   `gorm:"column:nick_name" json:"nickName"  description:"用户昵称" binding:"required" err_msg:"用户昵称 必填"`
	Email       string   `gorm:"column:email" json:"email"     description:"用户登录邮箱"`
	Phonenumber string   `gorm:"column:phonenumber" json:"phonenumber"        description:"中国手机不带国家代码，国际手机号格式为：国家代码-手机号"`
	Sex         string   `gorm:"column:sex" json:"sex"           description:"性别;0:保密,1:男,2:女"`
	Status      string   `gorm:"column:status" json:"status"    description:"用户状态;0:禁用,1:正常,2:未验证"`
	Remark      string   `gorm:"column:remark" json:"remark"        description:"备注"`
	RoleIds     []uint64 `gorm:"-" json:"roleIds"`
	PostIds     []int64  `gorm:"-" json:"postIds"`
}

type UpdateUserProfileRequest struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"id"            description:"" binding:"required" err_msg:"用户Id 必填"`
	UserName string `gorm:"-" json:"userName"      description:"用户名" binding:"required" err_msg:"用户名 必填"`
	//DeptId      uint64   `gorm:"column:dept_id" json:"deptId"        description:"部门id"`
	NickName    string `gorm:"column:nick_name" json:"nickName"  description:"用户昵称" binding:"required" err_msg:"用户昵称 必填"`
	Email       string `gorm:"column:email" json:"email"     description:"用户登录邮箱" binding:"required" err_msg:"邮箱 必填"`
	Phonenumber string `gorm:"column:phonenumber" json:"phonenumber"        description:"中国手机不带国家代码，国际手机号格式为：国家代码-手机号" binding:"required" err_msg:"手机号 必填"`
	Sex         string `gorm:"column:sex" json:"sex"           description:"性别;0:保密,1:男,2:女"`
}

type GetUserListRequest struct {
	DeptId      string `form:"deptId"  query:"dept_id:="`
	UserName    string `form:"userName" query:"user_name:like"`
	PhoneNumber string `form:"phonenumber" query:"phonenumber:like"`
	Status      string `form:"status" query:"status:="`
	BeginTime   string `form:"beginTime" query:"date(create_time):>="`
	EndTime     string `form:"endTime" query:"date(create_time):<="`
	Sort        string `form:"sort,default=create_time:desc"`
	common.PageReq
}

type UserIdsUriRequest struct {
	Ids string `uri:"ids" binding:"required" err_msg:"id 必填"`
}

type UserIdUriRequest struct {
	Id *uint64 `uri:"id" binding:"required" err_msg:"id 必填"`
}

type ChangeUserStatusRequest struct {
	Id     uint64 `gorm:"column:id" json:"id" binding:"required" err_msg:"id 必填"`
	Status string `gorm:"column:status" json:"status" binding:"required" err_msg:"状态 必填"`
}

type UpdateAuthRoleRequest struct {
	Id      uint64   `form:"id" binding:"required" err_msg:"userId 必填"`
	RoleIds []uint64 `form:"roleIds"`
}

type ImportUserRequest struct {
	UpdateSupport bool `form:"updateSupport"`
}

type ResetPwdRequest struct {
	Id       uint64 `gorm:"column:id" json:"id" binding:"required" err_msg:"id 必填"`
	Password string `gorm:"column:password" json:"password" binding:"required" err_msg:"密码 必填"`
}

type UpdateUserPwdRequest struct {
	OldPassword string `json:"oldPassword" binding:"required" err_msg:"旧密码 必填"`
	NewPassword string `json:"newPassword" binding:"required" err_msg:"新密码 必填"`
}
