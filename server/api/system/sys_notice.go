package system

import "go-fin-server/api/common"

type SysNotice struct {
	ListNoticeRequest
	UpdateNoticeRequest
}

type ListNoticeRequest struct {
	NoticeTitle string `form:"noticeTitle" query:"noticeTitle:like"`
	CreateBy    string `form:"createBy" query:"createBy:like"`
	NoticeType  string `form:"noticeType" query:"noticeType:="`
	common.PageReq
}

type UpdateNoticeRequest struct {
	Id            uint64 `gorm:"column:id;primary_key" json:"id"    description:"公告ID" binding:"required" err_msg:"公告ID 必填"`
	NoticeTitle   string `gorm:"column:notice_title" json:"noticeTitle"  description:"公告标题" binding:"required" err_msg:"公告标题 必填"`
	NoticeType    string `gorm:"column:notice_type" json:"noticeType"  description:"公告类型（1通知 2公告）" binding:"required" err_msg:"公告类型 必填"`
	NoticeContent string `gorm:"column:notice_content" json:"noticeContent"  description:"公告内容"`
	Status        string `gorm:"column:status" json:"status"    description:"公告状态（0正常 1关闭）"`
	Remark        string `gorm:"column:remark" json:"remark"    description:"备注"`
}
