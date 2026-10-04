package handler

import (
	"fmt"
	"go-fin-server/api"
	_ "go-fin-server/api/common" // swagger 文档引用
	"go-fin-server/internal/config"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/captcha"
	"go-fin-server/pkg/redistool"
	"go-fin-server/pkg/rsatool"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	Services service.Services
	Api      api.Api
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

// Login 用户登录
//
//	@Summary	用户登录
//	@Tags		登录
//	@Accept		json
//	@Produce	json
//	@Param		body	body		common.LoginRequest	true	"登录信息"
//	@Success	200		{object}	map[string]string	"token"
//	@Router		/login [post]
func (s *LoginHandler) Login(c *gin.Context) {
	response.SetOperTitle(c, "用户登录")
	req := s.Api.Common.LoginRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginService := s.Services.SysLoginService.New(c)
	// 读取验证码开关配置
	captchaEnabled, _ := s.Services.SysConfigService.SelectConfigByKey("sys.account.captchaEnabled")
	if captchaEnabled == "true" {
		// 验证码校验
		exists, err := redistool.Exists(db.RedisConnections["master"], constant.CACHE_CAPTCHA_CODE_KEY+req.Uuid)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		if !exists {
			loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "验证码已过期")
			response.Error(c, "验证码已过期")
			return
		}

		if !captcha.VerifyCaptcha(constant.CACHE_CAPTCHA_CODE_KEY+req.Uuid, req.Code, db.RedisConnections["master"]) {
			loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "验证码不正确")
			response.Error(c, "验证码不正确")
			return
		}
	}
	// IP黑名单校验
	// 账号校验-密码rsa解密
	password, err := rsatool.Decrypt(req.Password, config.GlobalConfig.Rsa.PrivateKey)
	if err != nil {
		loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "密码解析错误")
		pkg.Logger.Error(err)
		response.Error(c, "密码解析错误")
		return
	}
	// 账号校验-用户是否存在
	var user model.SysUser
	userObj, err := user.Get("user_name", req.UserName)
	if err != nil {
		loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "用户不存在")
		pkg.Logger.Error(err)
		response.Error(c, "用户不存在")
		return
	}
	// 验证用户状态
	if userObj.Status != "0" {
		loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "用户状态异常")
		response.Error(c, "用户状态异常")
		return
	}
	// 校验密码
	if !utils.CheckPassword(password, userObj.Password) {
		loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "密码错误")
		count, err := redistool.Incr(db.RedisConnections["master"], constant.CACHE_PWD_ERR_CNT_KEY+req.UserName)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		if count >= 5 {
			err := redistool.Expire(db.RedisConnections["master"], constant.CACHE_PWD_ERR_CNT_KEY+req.UserName, time.Duration(10)*time.Minute)
			if err != nil {
				pkg.Logger.Error(err)
				response.Error(c, err.Error())
				return
			}
		} else {
			response.Error(c, fmt.Sprintf("密码错误剩余次数%d", 5-count))
			return
		}

	}
	// 查询key是否有剩余过期时间
	ttl, err := redistool.TTL(db.RedisConnections["master"], constant.CACHE_PWD_ERR_CNT_KEY+req.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if ttl.Seconds() > 0 {
		response.Error(c, "密码错误次数过多，限制登录。请10分钟后刷新页面重新登录")
		return
	}

	loginUser := model.LoginUser{
		UserId:   userObj.Id,
		UserName: userObj.UserName,
		DeptId:   userObj.DeptId,
		DeptName: userObj.SysDept.DeptName,
		User:     userObj,
	}
	permissions, err := s.Services.PermissionService.GetMenuPermission(userObj)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "获取permissions错误")
		return
	}
	loginUser.Permissions = permissions
	tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
	token, err := tokenService.CreateToken(&loginUser)
	if err != nil {
		loginService.InsertLoginInfo(req.UserName, constant.LOGIN_FAIL, "生成token错误")
		pkg.Logger.Error(err)
		response.Error(c, "生成token错误")
		return
	}

	loginService.InsertLoginInfo(req.UserName, constant.LOGIN_SUCCESS, "user.login.success")
	err = s.Services.SysUserService.UpdateUserProfile(userObj.Id, map[string]interface{}{
		"login_ip":   c.ClientIP(),
		"login_date": types.LocalTime{Time: time.Now()},
	})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, map[string]string{
		"token": token,
	})
}

// GetRouters 获取路由菜单
//
//	@Summary	获取路由菜单
//	@Tags		登录
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"菜单树列表"
//	@Router		/getRouters [get]
//	@Security	BearerAuth
func (s *LoginHandler) GetRouters(c *gin.Context) {
	response.SetOperTitle(c, "获取路由菜单")
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	var menu model.SysMenu
	var user model.SysUser
	dataTree := make([]*model.SysMenuTreeNode, 0)
	// 找出目录和菜单
	if user.IsAdmin(loginUser.UserId) {
		data, err := menu.GetMenuAll()
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, "服务错误")
			return
		}
		dataTree = menu.BuildMenuTree(data)

	} else {
		data, err := menu.GetMenuByUserId(loginUser.UserId)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, "服务错误")
			return
		}
		dataTree = menu.BuildMenuTree(data)
	}

	response.Data(c, dataTree)
}

// Register 用户注册
//
//	@Summary	用户注册
//	@Tags		登录
//	@Router		/register [post]
func (s *LoginHandler) Register(c *gin.Context) {
	response.SetOperTitle(c, "用户注册")

}
// GetInfo 获取用户详细信息
//
//	@Summary	获取用户详细信息
//	@Tags		登录
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"用户信息、角色、权限"
//	@Router		/getInfo [get]
//	@Security	BearerAuth
func (s *LoginHandler) GetInfo(c *gin.Context) {
	response.SetOperTitle(c, "获取用户详细信息")
	loginUser := c.MustGet("loginUser").(*model.LoginUser)

	var user model.SysUser
	userObj, err := user.Get("id = ?", loginUser.UserId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "用户不存在")
		return
	}

	// 角色集合
	roles := make([]string, 0)
	var userRole model.SysUserRole
	userRoles, err := userRole.GetRolesByUserId(userObj.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "用户不存在")
		return
	}
	for _, item := range userRoles {
		roles = append(roles, item.SysRole.RoleKey)
	}
	// 权限集合
	var menu model.SysMenu
	permissions := make([]string, 0)
	if user.IsAdmin(userObj.Id) {
		permissions = append(permissions, "*:*:*")
	} else {
		if len(userRoles) > 0 {
			userRoleIds := make([]uint64, 0)
			for _, item := range userRoles {
				userRoleIds = append(userRoleIds, item.RoleId)
			}
			rolePerms, err := menu.GetMenuPermsByRoleIds(userRoleIds)
			if err != nil {
				pkg.Logger.Error(err)
				response.Error(c, "获取用户权限错误")
				return
			}
			permissions = append(permissions, rolePerms...)
		} else {
			userPerms, err := menu.GetMenuPermsByUserId(userObj.Id)
			if err != nil {
				pkg.Logger.Error(err)
				response.Error(c, "获取用户权限错误")
				return
			}
			permissions = append(permissions, userPerms...)
		}
	}

	response.Data(c, map[string]interface{}{
		"user":        userObj,
		"roles":       roles,
		"permissions": permissions,
	})

}
// LoginOut 退出登录
//
//	@Summary	退出登录
//	@Tags		登录
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"退出成功"
//	@Router		/logout [post]
//	@Security	Bearer Auth
func (s *LoginHandler) LoginOut(c *gin.Context) {
	response.SetOperTitle(c, "退出登录")
	claims, exists := c.Get("loginUser")
	if !exists || claims == nil {
		response.DataMsg(c, nil, "退出成功")
		return
	}
	loginUser, ok := claims.(*model.LoginUser)
	if !ok || loginUser == nil {
		response.DataMsg(c, nil, "退出成功")
		return
	}
	tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
	// 删除 token 和用户映射，如果 token 不存在（已被踢出）也视为成功
	err := tokenService.DelLoginUserAndMapping(loginUser.Token, loginUser.UserId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginService := s.Services.SysLoginService.New(c)
	loginService.InsertLoginInfo(loginUser.UserName, constant.LOGOUT, "退出成功")
	response.DataMsg(c, nil, "退出成功")
}
// GetCodeImg 获取验证码
//
//	@Summary	获取验证码图片
//	@Tags		登录
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"验证码图片base64和uuid"
//	@Router		/captchaImage [get]
func (s *LoginHandler) GetCodeImg(c *gin.Context) {
	response.SetOperTitle(c, "获取验证码")
	// 读取验证码开关配置
	captchaEnabled, _ := s.Services.SysConfigService.SelectConfigByKey("sys.account.captchaEnabled")
	if captchaEnabled != "true" {
		response.Data(c, map[string]interface{}{
			"captchaEnabled": false,
		})
		return
	}
	uuid, img, err := captcha.GenerateCaptcha(db.RedisConnections["master"])
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "生成验证码错误")
		return
	}

	response.Data(c, map[string]interface{}{
		"img":            strings.Split(img, ",")[1],
		"uuid":           uuid,
		"captchaEnabled": true,
	})
}
