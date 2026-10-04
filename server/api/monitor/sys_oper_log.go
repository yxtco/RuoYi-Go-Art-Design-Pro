package monitor

import "go-fin-server/api/common"

type SysOperLog struct {
	GetOperLogListRequest
	OperLogIdsUriRequest
}

type GetOperLogListRequest struct {
	OperIp       string `form:"operIp" query:"oper_ip:like"`
	Title        string `form:"title" query:"title:like"`
	OperName     string `form:"operName" query:"oper_name:like"`
	BusinessType *int   `form:"businessType" query:"business_type:="`
	Status       *int   `form:"status" query:"status:="`
	BeginTime    string `form:"beginTime" query:"date(create_time):>="`
	EndTime      string `form:"endTime" query:"date(create_time):<="`
	Sort         string `form:"sort,default=oper_time:desc"`
	common.PageReq
}

type OperLogIdsUriRequest struct {
	Ids string `uri:"ids" binding:"required" err_msg:"id 必填"`
}
