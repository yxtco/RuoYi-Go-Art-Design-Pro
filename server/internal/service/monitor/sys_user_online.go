package monitor

import (
	"go-fin-server/internal/model"
	"strings"
)

type SysUserOnlineService struct {
}

// SelectOnlineByInfo 通过登录地址/用户名称查询信息
func (c SysUserOnlineService) SelectOnlineByInfo(ipaddr, userName string, loginUser model.LoginUser) *model.SysUserOnline {
	if strings.Contains(loginUser.Ipaddr, ipaddr) && strings.Contains(loginUser.UserName, userName) {
		return c.LoginUserToUserOnline(loginUser)
	}
	return nil
}

// SelectOnlineByIpaddr 通过登录地址查询信息
func (c SysUserOnlineService) SelectOnlineByIpaddr(ipaddr string, loginUser model.LoginUser) *model.SysUserOnline {

	if strings.Contains(loginUser.Ipaddr, ipaddr) {
		return c.LoginUserToUserOnline(loginUser)
	}

	return nil
}

// SelectOnlineByUserName 通过用户名称查询信息
func (c SysUserOnlineService) SelectOnlineByUserName(userName string, loginUser model.LoginUser) *model.SysUserOnline {
	if strings.Contains(loginUser.UserName, userName) {
		return c.LoginUserToUserOnline(loginUser)
	}
	return nil
}

// LoginUserToUserOnline 设置在线用户信息
func (c SysUserOnlineService) LoginUserToUserOnline(loginUser model.LoginUser) *model.SysUserOnline {
	sysUserOnline := &model.SysUserOnline{
		TokenId:       loginUser.Token,
		UserName:      loginUser.UserName,
		Ipaddr:        loginUser.Ipaddr,
		LoginLocation: loginUser.LoginLocation,
		Browser:       loginUser.Browser,
		Os:            loginUser.Os,
		LoginTime:     loginUser.LoginTime,
		DeptName:      loginUser.DeptName,
	}
	return sysUserOnline
}
