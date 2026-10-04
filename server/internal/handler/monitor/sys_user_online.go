package monitor

import (
	"go-fin-server/api"
	"go-fin-server/internal/config"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"
	"sort"

	"github.com/gin-gonic/gin"
)

type UserOnlineHandler struct {
	Services service.Services
	Api      api.Api
}

func NewUserOnlineHandler() *UserOnlineHandler {
	return &UserOnlineHandler{}
}

// ListOnline 查询在线用户列表
//
//	@Summary	查询在线用户列表
//	@Tags		在线用户
//	@Produce	json
//	@Param		ipaddr		query	string	false	"IP地址"
//	@Param		userName	query	string	false	"用户名称"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"在线用户列表"
//	@Router		/monitor/online/list [get]
//	@Security	BearerAuth
func (s *UserOnlineHandler) ListOnline(c *gin.Context) {
	response.SetOperTitle(c, "查询在线用户列表")
	req := s.Api.SysUserOnline.GetOnlineListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	userOnlineList := make([]model.SysUserOnline, 0)
	keys, err := redistool.Keys(db.RedisConnections["master"], "login_tokens:*")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	for _, key := range keys {
		var loginUser model.LoginUser
		err := redistool.GetCacheObject(db.RedisConnections["master"], key, &loginUser)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		if req.Ipaddr != "" && req.UserName != "" {
			userOnline := s.Services.SysUserOnlineService.SelectOnlineByInfo(req.Ipaddr, req.UserName, loginUser)
			if userOnline != nil {
				userOnlineList = append(userOnlineList, *userOnline)
			}
		} else if req.Ipaddr != "" {
			userOnline := s.Services.SysUserOnlineService.SelectOnlineByIpaddr(req.Ipaddr, loginUser)
			if userOnline != nil {
				userOnlineList = append(userOnlineList, *userOnline)
			}
		} else if req.UserName != "" {
			userOnline := s.Services.SysUserOnlineService.SelectOnlineByUserName(req.UserName, loginUser)
			if userOnline != nil {
				userOnlineList = append(userOnlineList, *userOnline)
			}
		} else {
			userOnline := s.Services.SysUserOnlineService.LoginUserToUserOnline(loginUser)
			if userOnline != nil {
				userOnlineList = append(userOnlineList, *userOnline)
			}
		}
	}
	// 通过登录时间字段排序
	sort.Slice(userOnlineList, func(i, j int) bool {
		return userOnlineList[i].LoginTime > userOnlineList[j].LoginTime
	})
	// 计算分页的起始索引和结束索引
	//startIndex := (req.PageNum - 1) * req.PageSize
	//endIndex := req.PageNum * req.PageSize
	//pagedList := userOnlineList[startIndex:endIndex]
	response.PageData(c, userOnlineList, int64(len(userOnlineList)))

}

// ForceLogout 强退用户
//
//	@Summary	强退用户
//	@Tags		在线用户
//	@Produce	json
//	@Param		tokenId	path	string	true	"令牌ID"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/online/{tokenId} [delete]
//	@Security	BearerAuth
func (s *UserOnlineHandler) ForceLogout(c *gin.Context) {
	response.SetOperTitle(c, "强退用户")
	req := s.Api.SysUserOnline.ForceLogoutRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 先获取旧会话的用户信息，用于清理 login_user_token 映射
	var loginUser model.LoginUser
	loginTokenKey := constant.CACHE_LOGIN_TOKEN_KEY + req.TokenId
	err := redistool.GetCacheObject(db.RedisConnections["master"], loginTokenKey, &loginUser)
	if err != nil {
		pkg.Logger.Warnf("强退用户获取登录信息失败（可能已过期）: tokenId=%s, error=%v", req.TokenId, err)
	}
	// 删除 login_tokens
	err = redistool.Del(db.RedisConnections["master"], loginTokenKey)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 同时清理 login_user_token 映射，确保下次登录时能正确踢出
	if loginUser.UserId != 0 {
		tokenService := service.Services{}.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
		userTokenKey := tokenService.GetUserTokenKey(loginUser.UserId)
		_ = redistool.Del(db.RedisConnections["master"], userTokenKey)
		pkg.Logger.Infof("强退用户清理映射成功: userId=%d, tokenId=%s", loginUser.UserId, req.TokenId)
	}
	response.Data(c, nil)
}
