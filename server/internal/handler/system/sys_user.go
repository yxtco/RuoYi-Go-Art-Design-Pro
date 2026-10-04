package system

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/config"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/fileuploadtool"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"io"
	"mime"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Services service.Services
	Api      api.Api
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// ListUser 查询用户列表
//
//	@Summary	查询用户列表
//	@Tags		用户管理
//	@Produce	json
//	@Param		userName	query	string	false	"用户名称"
//	@Param		phonenumber	query	string	false	"手机号码"
//	@Param		status		query	string	false	"状态"
//	@Param		deptId		query	string	false	"部门ID"
//	@Param		beginTime	query	string	false	"开始时间"
//	@Param		endTime		query	string	false	"结束时间"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"用户列表"
//	@Router		/system/user/list [get]
//	@Security	BearerAuth
func (s *UserHandler) ListUser(c *gin.Context) {
	response.SetOperTitle(c, "查询用户列表")
	req := s.Api.SysUser.GetUserListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, total, err := s.Services.SysUserService.SelectDataScopeUserList(conditions, sortConditions, req.PageNum, req.PageSize, loginUser.User)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}

// GetUser 查询用户详细信息
//
//	@Summary	查询用户详细信息
//	@Tags		用户管理
//	@Produce	json
//	@Param		id	path	int	true	"用户ID"
//	@Success	200	{object}	map[string]interface{}"用户详细信息"
//	@Router		/system/user/{id} [get]
//	@Security	BearerAuth
func (s *UserHandler) GetUser(c *gin.Context) {
	response.SetOperTitle(c, "查询用户详细信息")
	req := s.Api.SysUser.UserIdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	var user model.SysUser
	var data *model.SysUserAll
	if *req.Id != 0 {
		err := s.Services.SysUserService.CheckUserDataScope(*req.Id)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		data2, err := s.Services.SysUserService.SelectUserById(*req.Id)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		data = &data2
	}

	// 获取所有的角色
	roles, err := s.Services.SysRoleService.SelectRoleAllList(map[string]interface{}{
		"status = ?": "0",
	})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	roleData := make([]model.SysRole, 0)
	if !user.IsAdmin(*req.Id) {
		for _, role := range roles {
			if !role.IsAdmin(role.Id) {
				roleData = append(roleData, role)
			}
		}
	} else {
		roleData = roles
	}
	// 获取所有的岗位
	posts, err := s.Services.SysPostService.SelectPostAllList(map[string]interface{}{
		"status = ?": "0",
	})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	if data == nil {
		response.DataExpand(c, data, map[string]interface{}{
			"roles": roleData,
			"posts": posts,
		})
	} else {
		data.Password = ""
		response.DataExpand(c, data, map[string]interface{}{
			"postIds": data.PostIds,
			"roleIds": data.RoleIds,
			"roles":   roleData,
			"posts":   posts,
		})
	}

}

// DelUser 删除用户
//
//	@Summary	删除用户
//	@Tags		用户管理
//	@Produce	json
//	@Param		ids	path	string	true	"用户ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/{ids} [delete]
//	@Security	BearerAuth
func (s *UserHandler) DelUser(c *gin.Context) {
	response.SetOperTitle(c, "删除用户")
	req := s.Api.SysUser.UserIdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)

	ids := strings.Split(req.Ids, ",")
	if stringutils.StringInSlice(strconv.FormatUint(loginUser.UserId, 10), ids) {
		response.Error(c, "当前用户不能删除")
		return
	}
	err := s.Services.SysUserService.DeleteUserByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}

// AddUser 新增用户
//
//	@Summary	新增用户
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"用户信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user [post]
//	@Security	BearerAuth
func (s *UserHandler) AddUser(c *gin.Context) {
	response.SetOperTitle(c, "新增用户")
	var req model.SysUserAll
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	if !s.Services.SysUserService.CheckUserNameUnique(req.Id, req.UserName) {
		response.Error(c, "新增用户'"+req.UserName+"'失败，登录账号已存在")
		return
	} else if req.Phonenumber != "" && !s.Services.SysUserService.CheckPhoneUnique(req.Id, req.Phonenumber) {
		response.Error(c, "新增用户'"+req.UserName+"'失败，手机号码已存在")
		return
	} else if req.Email != "" && !s.Services.SysUserService.CheckEmailUnique(req.Id, req.Email) {
		response.Error(c, "新增用户'"+req.UserName+"'失败，邮箱账号已存在")
		return
	}
	req.CreateBy = loginUser.UserName
	password, err := utils.EncryptPassword(req.Password)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	req.Password = password
	err = s.Services.SysUserService.InsertUser(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}

// UpdateUser 修改用户
//
//	@Summary	修改用户
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"用户信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user [put]
//	@Security	BearerAuth
func (s *UserHandler) UpdateUser(c *gin.Context) {
	response.SetOperTitle(c, "修改用户")
	req := s.Api.SysUser.UpdateUserRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	err := s.Services.SysUserService.CheckUserAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysUserService.CheckUserDataScope(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	if !s.Services.SysUserService.CheckUserNameUnique(req.Id, req.UserName) {
		response.Error(c, "修改用户'"+req.UserName+"'失败，登录账号已存在")
		return
	} else if req.Phonenumber != "" && !s.Services.SysUserService.CheckPhoneUnique(req.Id, req.Phonenumber) {
		response.Error(c, "修改用户'"+req.UserName+"'失败，手机号码已存在")
		return
	} else if req.Email != "" && !s.Services.SysUserService.CheckEmailUnique(req.Id, req.Email) {
		response.Error(c, "修改用户'"+req.UserName+"'失败，邮箱账号已存在")
		return
	}
	//req.UpdateBy = loginUser.UserName
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysUserService.UpdateUser(req.Id, upd, req.RoleIds, req.PostIds)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}

// ResetUserPwd 重置用户密码
//
//	@Summary	重置用户密码
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含id和password"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/resetPwd [put]
//	@Security	BearerAuth
func (s *UserHandler) ResetUserPwd(c *gin.Context) {
	response.SetOperTitle(c, "重置用户密码")
	req := s.Api.SysUser.ResetPwdRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	err := s.Services.SysUserService.CheckUserAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysUserService.CheckUserDataScope(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	password, err := utils.EncryptPassword(req.Password)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	upd := utils.StructToMapWithGormColumn(req)
	upd["password"] = password
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysUserService.ResetPwd(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// ChangeUserStatus 修改用户状态
//
//	@Summary	修改用户状态
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含id和status"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/changeStatus [put]
//	@Security	BearerAuth
func (s *UserHandler) ChangeUserStatus(c *gin.Context) {
	response.SetOperTitle(c, "修改用户状态")
	req := s.Api.SysUser.ChangeUserStatusRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err := s.Services.SysUserService.CheckUserAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysUserService.CheckUserDataScope(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysUserService.UpdateUserStatus(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// GetUserProfile 查询用户个人信息
//
//	@Summary	查询用户个人信息
//	@Tags		用户管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"个人信息"
//	@Router		/system/user/profile [get]
//	@Security	BearerAuth
func (s *UserHandler) GetUserProfile(c *gin.Context) {
	response.SetOperTitle(c, "查询用户个人信息")
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	user, err := s.Services.SysUserService.SelectUserById(loginUser.UserId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	group, err := s.Services.SysUserService.SelectUserRoleGroup(user.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	postGroup, err := s.Services.SysUserService.SelectUserPostGroup(user.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.DataExpand(c, user, map[string]interface{}{
		"roleGroup": group,
		"postGroup": postGroup,
	})
}

// UpdateUserProfile 修改用户个人信息
//
//	@Summary	修改用户个人信息
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"个人信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/profile [put]
//	@Security	BearerAuth
func (s *UserHandler) UpdateUserProfile(c *gin.Context) {
	response.SetOperTitle(c, "修改用户个人信息")
	req := s.Api.SysUser.UpdateUserProfileRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	if req.Phonenumber != "" && !s.Services.SysUserService.CheckPhoneUnique(req.Id, req.Phonenumber) {
		response.Error(c, "修改用户'"+req.UserName+"'失败，手机号码已存在")
		return
	} else if req.Email != "" && !s.Services.SysUserService.CheckEmailUnique(req.Id, req.Email) {
		response.Error(c, "修改用户'"+req.UserName+"'失败，邮箱账号已存在")
		return
	}
	tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysUserService.UpdateUserProfile(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser.User.NickName = req.NickName
	loginUser.User.Email = req.Email
	loginUser.User.Phonenumber = req.Phonenumber
	loginUser.User.Sex = req.Sex
	// 更新缓存用户信息
	err = tokenService.SetLoginUser(loginUser)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// UpdateUserPwd 修改用户密码
//
//	@Summary	修改用户密码
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含oldPassword和newPassword"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/profile/updatePwd [put]
//	@Security	BearerAuth
func (s *UserHandler) UpdateUserPwd(c *gin.Context) {
	response.SetOperTitle(c, "修改用户密码")
	req := s.Api.SysUser.UpdateUserPwdRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	userName := loginUser.UserName
	password := loginUser.User.Password
	if !utils.CheckPassword(req.OldPassword, password) {
		response.Error(c, "修改密码失败，旧密码错误")
		return
	}
	if !utils.CheckPassword(req.NewPassword, password) {
		response.Error(c, "修改密码失败，旧密码错误")
		return
	}
	newPassword, err := utils.EncryptPassword(req.NewPassword)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysUserService.ResetUserPwd(userName, newPassword)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser.User.Password = newPassword
	tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
	// 更新缓存用户密码
	err = tokenService.SetLoginUser(loginUser)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// UploadAvatar 用户头像上传
//
//	@Summary	用户头像上传
//	@Tags		用户管理
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		avatarfile	formData	file	true	"头像文件"
//	@Success	200			{object}	map[string]interface{}"头像URL"
//	@Router		/system/user/profile/avatar [post]
//	@Security	BearerAuth
func (s *UserHandler) UploadAvatar(c *gin.Context) {
	response.SetOperTitle(c, "用户头像上传")
	file, err := c.FormFile("avatarfile")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "获取文件失败")
		return
	}
	contentType := file.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "获取文件类型失败")
		return
	}
	//  检查文件大小
	if fileuploadtool.CheckFileSize(file.Size) {
		response.Error(c, "文件过大")
		return
	}
	//  检查文件大小
	if fileuploadtool.CheckFileNameLength(len(file.Filename)) {
		response.Error(c, "文件名过长")
		return
	}

	// 获取上传目录（基础路径 + 日期子目录）
	uploadDir := fileuploadtool.GetUploadDir()
	if err := fileuploadtool.EnsureDir(uploadDir); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "创建上传目录失败")
		return
	}

	// 生成文件名
	filename := fileuploadtool.GenerateFilename(file.Filename + extensions[0])
	savePath := filepath.Join(uploadDir, filename)
	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, "无法保存文件")
		return
	}

	// 构建相对路径（用于前端存储）
	relativePath := filepath.Join(fileuploadtool.GetDateSubDir(), filename)
	avatar := fmt.Sprintf("http://%s:%s/assets/%s", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port, relativePath)
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err = s.Services.SysUserService.UpdateUserProfile(loginUser.UserId, map[string]interface{}{"avatar": avatar})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
	loginUser.User.Avatar = avatar
	// 更新缓存用户密码
	err = tokenService.SetLoginUser(loginUser)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "上传图片异常，请联系管理员")
		return
	}
	response.DataExpand(c, nil, map[string]interface{}{
		"imgUrl": avatar,
	})
}

// GetAuthRole 查询用户授权角色
//
//	@Summary	查询用户授权角色
//	@Tags		用户管理
//	@Produce	json
//	@Param		id	path	int	true	"用户ID"
//	@Success	200	{object}	map[string]interface{}"角色列表"
//	@Router		/system/user/authRole/{id} [get]
//	@Security	BearerAuth
func (s *UserHandler) GetAuthRole(c *gin.Context) {
	response.SetOperTitle(c, "查询用户授权角色")
	req := s.Api.SysUser.UserIdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	var user model.SysUser
	userObj, err := user.Get("id = ?", *req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	roles, err := s.Services.SysRoleService.SelectRolesByUserId(*req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	roleData := make([]model.SysRole, 0)
	if !user.IsAdmin(*req.Id) {
		for _, role := range roles {
			if !role.IsAdmin(role.Id) {
				roleData = append(roleData, role)
			}
		}
	} else {
		roleData = roles
	}
	userObj.Password = ""
	response.DataExpand(c, nil, map[string]interface{}{
		"roles": roleData,
		"user":  userObj,
	})

}
// UpdateAuthRole 保存授权角色
//
//	@Summary	保存授权角色
//	@Tags		用户管理
//	@Accept		json
//	@Produce	json
//	@Param		id		body	int		true	"用户ID"
//	@Param		roleIds	body	[]int	true	"角色ID列表"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/user/authRole [put]
//	@Security	BearerAuth
func (s *UserHandler) UpdateAuthRole(c *gin.Context) {
	response.SetOperTitle(c, "保存授权角色")
	req := s.Api.SysUser.UpdateAuthRoleRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err := s.Services.SysUserService.CheckUserDataScope(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysUserService.InsertUserRole(req.Id, req.RoleIds)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)

}
// DeptTreeSelect 查询部门下拉树结构
//
//	@Summary	查询部门下拉树结构
//	@Tags		用户管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"部门树"
//	@Router		/system/user/deptTree [get]
//	@Security	BearerAuth
func (s *UserHandler) DeptTreeSelect(c *gin.Context) {
	response.SetOperTitle(c, "查询部门下拉树结构")
	var dept model.SysDeptTreeNode

	data, err := dept.GetDeptTreeList()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	dataTree := dept.BuildSysDeptTree(data)
	response.Data(c, dataTree)
}

// ImportTemplate 下载导入模板
//
//	@Summary	下载导入模板
//	@Tags		用户管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"模板数据"
//	@Router		/system/user/importTemplate [post]
//	@Security	BearerAuth
func (s *UserHandler) ImportTemplate(c *gin.Context) {
	response.SetOperTitle(c, "下载导入模板")
	response.Data(c, nil)
}

// Export 导出用户数据
//
//	@Summary	导出用户数据
//	@Tags		用户管理
//	@Produce	application/octet-stream
//	@Param		userName	query	string	false	"用户名称"
//	@Param		phonenumber	query	string	false	"手机号码"
//	@Param		status		query	string	false	"状态"
//	@Success	200			{file}	binary"Excel文件"
//	@Router		/system/user/export [post]
//	@Security	BearerAuth
func (s *UserHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出用户数据")
	req := s.Api.SysUser.GetUserListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	userList, err := s.Services.SysUserService.SelectUserAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"用户编号", "用户名称", "用户昵称", "部门", "手机号码", "状态", "创建时间"}
	var data [][]string
	for _, user := range userList {
		data = append(data, []string{
			fmt.Sprint(user.Id),
			user.UserName,
			user.NickName,
			user.SysDept.DeptName,
			user.Phonenumber,
			user.Status,
			fmt.Sprint(user.CreatedAt),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "用户数据")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}

// ImportData 导入用户数据
//
//	@Summary	导入用户数据
//	@Tags		用户管理
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		file	formData	file	true	"Excel文件"
//	@Success	200		{object}	map[string]interface{}"导入结果"
//	@Router		/system/user/importData [post]
//	@Security	BearerAuth
func (s *UserHandler) ImportData(c *gin.Context) {
	response.SetOperTitle(c, "导入用户数据")
	req := s.Api.SysUser.ImportUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	defer file.Close()

	// 读取上传的文件到内存中
	fileContent, err := io.ReadAll(file)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	// 调用导入函数
	data, err := exceltool.ReadExcelFile(fileContent)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
