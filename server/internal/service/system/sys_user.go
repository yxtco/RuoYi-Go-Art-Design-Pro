package system

import (
	"errors"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"strings"

	"gorm.io/gorm"
)

type SysUserService struct {
}

// SelectUserList 根据条件分页查询用户列表
func (c SysUserService) SelectUserList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	data, count, err := model.SysUser{}.SelectUserList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectDataScopeUserList 根据条件分页查询用户列表
func (c SysUserService) SelectDataScopeUserList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int, user model.SysUser) ([]model.SysUser, int64, error) {
	data, count, err := model.SysUser{}.SelectDataScopeUserList(conditions, sortConditions, pageNum, pageSize, user)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectUserAllList 根据条件查询用户列表
func (c SysUserService) SelectUserAllList(conditions map[string]interface{}) ([]model.SysUser, error) {
	data, err := model.SysUser{}.SelectUserAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectUserById 通过用户ID查询用户
func (c SysUserService) SelectUserById(userId uint64) (model.SysUserAll, error) {
	data, err := model.SysUser{}.GetAll("id = ?", userId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// DeleteUserByIds 批量删除用户信息
func (c SysUserService) DeleteUserByIds(userIds []uint64) error {
	for _, id := range userIds {
		err := c.CheckUserAllowed(id)
		if err != nil {
			return err
		}
		err = c.CheckUserDataScope(id)
		if err != nil {
			return err
		}
	}
	// 删除用户与角色关联
	err := model.SysUserRole{}.DeleteUserRole(userIds)
	if err != nil {
		return err
	}
	// 删除用户与岗位关联
	err = model.SysUserPost{}.DeleteUserPost(userIds)
	if err != nil {
		return err
	}

	err = model.SysUser{}.Delete("id in (?)", userIds)
	if err != nil {
		return err
	}
	return nil
}

// ResetPwd 重置密码
func (c SysUserService) ResetPwd(userId uint64, upd map[string]interface{}) error {
	err := model.SysUser{}.UpdateMap(upd, "id = ?", userId)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUserStatus 修改用户状态
func (c SysUserService) UpdateUserStatus(userId uint64, upd map[string]interface{}) error {
	err := model.SysUser{}.UpdateMap(upd, "id = ?", userId)
	if err != nil {
		return err
	}
	return nil
}

// CheckUserAllowed 校验用户是否允许操作
func (c SysUserService) CheckUserAllowed(id uint64) error {
	var user model.SysUser
	if user.IsAdmin(id) {
		return errors.New("不允许操作超级管理员用户")
	}
	return nil
}

// CheckUserDataScope 校验用户是否有数据权限
func (c SysUserService) CheckUserDataScope(id uint64) error {
	var user model.SysUser
	if !user.IsAdmin(id) {
		userList, err := user.SelectUserAllList(map[string]interface{}{
			"id = ?": id})
		if err != nil {
			return err
		}
		if len(userList) <= 0 {
			return errors.New("没有权限访问用户数据！")
		}
	}
	return nil
}

// UpdateUser 修改用户
func (c SysUserService) UpdateUser(userId uint64, upd map[string]interface{}, roleIds []uint64, PostIds []int64) error {
	// 删除用户与角色关联
	var userRole model.SysUserRole
	err := userRole.DeleteUserRoleByUserId(userId)
	if err != nil {
		return err
	}
	// 新增用户与角色管理
	err = c.InsertUserRole(userId, roleIds)
	if err != nil {
		return err
	}
	// 删除用户与岗位关联
	var userPost model.SysUserPost
	err = userPost.DeleteUserPostByUserId(userId)
	if err != nil {
		return err
	}
	// 新增用户与岗位管理
	err = c.InsertUserPost(userId, PostIds)
	if err != nil {
		return err
	}
	err = model.SysUser{}.UpdateMap(upd, "id = ?", userId)
	if err != nil {
		return err
	}
	return nil
}

// InsertUser 新增保存用户信息
func (c SysUserService) InsertUser(user *model.SysUserAll) error {
	// 新增用户信息
	err := user.SysUser.Create(&user.SysUser)
	if err != nil {
		return err
	}
	// 新增用户与角色管理
	err = c.InsertUserRole(user.SysUser.Id, user.RoleIds)
	if err != nil {
		return err
	}
	// 新增用户与岗位管理
	err = c.InsertUserPost(user.SysUser.Id, user.PostIds)
	if err != nil {
		return err
	}
	return nil
}

// CheckUserNameUnique 校验用户名称是否唯一
func (c SysUserService) CheckUserNameUnique(userId uint64, userName string) bool {
	var user model.SysUser
	userObj, err := user.Get("user_name = ?", userName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if userObj.Id != userId {
		return false
	}
	return true
}

// CheckPhoneUnique 校验手机号码是否唯一
func (c SysUserService) CheckPhoneUnique(userId uint64, phone string) bool {
	var user model.SysUser
	userObj, err := user.Get("phonenumber = ?", phone)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if userObj.Id != userId {
		return false
	}
	return true
}

// CheckEmailUnique 校验email是否唯一
func (c SysUserService) CheckEmailUnique(userId uint64, email string) bool {
	var user model.SysUser
	userObj, err := user.Get("email = ?", email)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if userObj.Id != userId {
		return false
	}
	return true
}

// InsertUserRole 新增用户角色信息
func (c SysUserService) InsertUserRole(userId uint64, roleIds []uint64) error {
	var userRole model.SysUserRole
	roleList := make([]model.SysUserRole, 0)
	for _, item := range roleIds {
		roleList = append(roleList, model.SysUserRole{
			RoleId: item,
			UserId: userId,
		})
	}
	err := userRole.BatchUserRole(roleList)
	if err != nil {
		return err
	}
	return nil
}

// InsertUserPost 新增用户角色信息
func (c SysUserService) InsertUserPost(userId uint64, postIds []int64) error {
	var userPost model.SysUserPost
	postList := make([]model.SysUserPost, 0)
	for _, item := range postIds {
		postList = append(postList, model.SysUserPost{
			PostId: item,
			UserId: userId,
		})
	}
	err := userPost.BatchUserPost(postList)
	if err != nil {
		return err
	}
	return nil
}

func (c SysUserService) insertUserAuth(userId uint64, roleIds []uint64) error {
	// 删除用户与角色关联
	err := model.SysUserRole{}.DeleteUserRoleByUserId(userId)
	if err != nil {
		return err
	}
	err = c.InsertUserRole(userId, roleIds)
	if err != nil {
		return err
	}
	return nil
}

// SelectAllocatedList 查询已分配用户角色列表
func (c SysUserService) SelectAllocatedList(roleId uint64, user model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	data, total, err := model.SysUser{}.SelectAllocatedList(roleId, user, pageNum, pageSize)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectUnallocatedList 查询未分配用户角色列表
func (c SysUserService) SelectUnallocatedList(roleId uint64, user model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	data, total, err := model.SysUser{}.SelectUnallocatedList(roleId, user, pageNum, pageSize)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectUserRoleGroup 查询用户所属角色组
func (c SysUserService) SelectUserRoleGroup(userName string) (string, error) {
	data, err := model.SysRole{}.SelectRolesByUserName(userName)
	if err != nil {
		return "", err
	}
	if len(data) <= 0 {
		return "", nil
	}
	roleNameList := make([]string, 0)
	for _, item := range data {
		roleNameList = append(roleNameList, item.RoleName)
	}
	return strings.Join(roleNameList, ","), nil
}

// SelectUserPostGroup 查询用户所属岗位组
func (c SysUserService) SelectUserPostGroup(userName string) (string, error) {
	data, err := model.SysPost{}.SelectPostsByUserName(userName)
	if err != nil {
		return "", err
	}
	if len(data) <= 0 {
		return "", nil
	}
	postNameList := make([]string, 0)
	for _, item := range data {
		postNameList = append(postNameList, item.PostName)
	}
	return strings.Join(postNameList, ","), nil
}

// UpdateUserProfile 修改用户基本信息
func (c SysUserService) UpdateUserProfile(userId uint64, upd map[string]interface{}) error {
	err := model.SysUser{}.UpdateMap(upd, "id = ?", userId)
	if err != nil {
		return err
	}
	return nil
}

// ResetUserPwd 重置用户密码
func (c SysUserService) ResetUserPwd(userName, password string) error {
	err := model.SysUser{}.UpdateMap(map[string]interface{}{
		"password": password,
	}, "user_name = ?", userName)
	if err != nil {
		return err
	}
	return nil
}
