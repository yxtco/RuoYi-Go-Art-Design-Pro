package monitor

import "go-fin-server/api/common"

type SysLoginInfo struct {
	GetLoginInfoListRequest
	UserNameUriRequest
}

type GetLoginInfoListRequest struct {
	Ipaddr    string `form:"ipaddr" query:"ipaddr:like"`
	UserName  string `form:"userName" query:"user_name:like"`
	Status    *int   `form:"status" query:"status:="`
	BeginTime string `form:"beginTime" query:"date(create_time):>="`
	EndTime   string `form:"endTime" query:"date(create_time):<="`
	Sort      string `form:"sort,default=login_time:desc"`
	common.PageReq
}

type UserNameUriRequest struct {
	UserName string `uri:"userName" binding:"required" err_msg:"userName 必填"`
}
