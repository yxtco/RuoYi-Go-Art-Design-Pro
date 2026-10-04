package system

import (
	"go-fin-server/api"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type NoticeHandler struct {
	Services service.Services
	Api      api.Api
}

func NewNoticeHandler() *NoticeHandler {
	return &NoticeHandler{}
}

// ListNotice 查询公告列表
//
//	@Summary	查询公告列表
//	@Tags		公告管理
//	@Produce	json
//	@Param		noticeTitle	query	string	false	"公告标题"
//	@Param		createBy		query	string	false	"创建者"
//	@Param		noticeType	query	string	false	"公告类型"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"公告列表"
//	@Router		/system/notice/list [get]
//	@Security	BearerAuth
func (s *NoticeHandler) ListNotice(c *gin.Context) {
	response.SetOperTitle(c, "查询公告列表")
	req := s.Api.SysNotice.ListNoticeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	data, total, err := s.Services.SysNoticeService.SelectNoticeList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}
// GetNotice 查询公告详细
//
//	@Summary	查询公告详细
//	@Tags		公告管理
//	@Produce	json
//	@Param		id	path	int	true	"公告ID"
//	@Success	200	{object}	map[string]interface{}"公告信息"
//	@Router		/system/notice/{id} [get]
//	@Security	BearerAuth
func (s *NoticeHandler) GetNotice(c *gin.Context) {
	response.SetOperTitle(c, "查询公告详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysNoticeService.SelectNoticeById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// AddNotice 新增公告
//
//	@Summary	新增公告
//	@Tags		公告管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"公告信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/notice [post]
//	@Security	BearerAuth
func (s *NoticeHandler) AddNotice(c *gin.Context) {
	response.SetOperTitle(c, "新增公告")
	var req model.SysNotice
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysNoticeService.InsertNotice(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)
}
// UpdateNotice 修改公告
//
//	@Summary	修改公告
//	@Tags		公告管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"公告信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/notice [put]
//	@Security	BearerAuth
func (s *NoticeHandler) UpdateNotice(c *gin.Context) {
	response.SetOperTitle(c, "修改公告")
	req := s.Api.SysNotice.UpdateNoticeRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysNoticeService.UpdateNotice(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// DelNotice 删除公告
//
//	@Summary	删除公告
//	@Tags		公告管理
//	@Produce	json
//	@Param		ids	path	string	true	"公告ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/notice/{ids} [delete]
//	@Security	BearerAuth
func (s *NoticeHandler) DelNotice(c *gin.Context) {
	response.SetOperTitle(c, "删除公告")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysNoticeService.DeleteNoticeByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}

// ListNoticeReadUsers 查询公告已读用户列表
//
//	@Summary	查询公告已读用户列表
//	@Tags		公告管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"已读用户列表"
//	@Router		/system/notice/readUsers/list [get]
//	@Security	BearerAuth
func (s *NoticeHandler) ListNoticeReadUsers(c *gin.Context) {
	response.SetOperTitle(c, "查询公告已读用户列表")
	// TODO: 待实现公告已读用户表后完善此接口
	response.PageData(c, make([]interface{}, 0), 0)
}

// MarkNoticeRead 标记公告已读
//
//	@Summary	标记公告已读
//	@Tags		公告管理
//	@Produce	json
//	@Param		noticeId	query	int	true	"公告ID"
//	@Success	200			{object}	map[string]interface{}"操作结果"
//	@Router		/system/notice/markRead [post]
//	@Security	BearerAuth
func (s *NoticeHandler) MarkNoticeRead(c *gin.Context) {
	response.SetOperTitle(c, "标记公告已读")
	noticeIdStr := c.Query("noticeId")
	if noticeIdStr == "" {
		response.Error(c, "noticeId 不能为空")
		return
	}
	noticeId, err := strconv.ParseUint(noticeIdStr, 10, 64)
	if err != nil {
		response.Error(c, "noticeId 格式错误")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err = s.Services.SysNoticeService.MarkNoticeRead(noticeId, loginUser.User.Id, loginUser.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// MarkNoticeReadAll 批量标记公告已读
//
//	@Summary	批量标记公告已读
//	@Tags		公告管理
//	@Produce	json
//	@Param		ids	query	string	true	"公告ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/notice/markReadAll [post]
//	@Security	BearerAuth
func (s *NoticeHandler) MarkNoticeReadAll(c *gin.Context) {
	response.SetOperTitle(c, "批量标记公告已读")
	idsStr := c.Query("ids")
	if idsStr == "" {
		response.Error(c, "ids 不能为空")
		return
	}
	ids := strings.Split(idsStr, ",")
	noticeIds := stringutils.StringSliceToUint64Slice(ids)
	if len(noticeIds) == 0 {
		response.Error(c, "ids 格式错误")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err := s.Services.SysNoticeService.MarkNoticeReadBatch(noticeIds, loginUser.User.Id, loginUser.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// GetUnreadCount 获取当前用户未读公告数量
//
//	@Summary	获取当前用户未读公告数量
//	@Tags		公告管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"未读数量"
//	@Router		/system/notice/unreadCount [get]
//	@Security	BearerAuth
func (s *NoticeHandler) GetUnreadCount(c *gin.Context) {
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	count, err := s.Services.SysNoticeService.CountUnreadNotice(loginUser.User.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, gin.H{"unreadCount": count})
}
