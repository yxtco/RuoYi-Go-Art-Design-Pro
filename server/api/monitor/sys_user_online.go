package monitor

import "go-fin-server/api/common"

type SysUserOnline struct {
	GetOnlineListRequest
	ForceLogoutRequest
}

type GetOnlineListRequest struct {
	Ipaddr   string `form:"ipaddr"`
	UserName string `form:"userName"`
	common.PageReq
}

type ForceLogoutRequest struct {
	TokenId string `uri:"tokenId" binding:"required" err_msg:"tokenId 必填"`
}
