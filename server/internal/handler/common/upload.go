package common

import (
	"fmt"
	"go-fin-server/internal/config"
	"go-fin-server/internal/response"
	"go-fin-server/pkg"
	"go-fin-server/pkg/fileuploadtool"
	"mime"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type CommonHandler struct{}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

// Upload 通用文件上传（图片/文档）
//
//	@Summary	通用文件上传
//	@Tags		公共上传
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		file	formData	file	true	"上传文件"
//	@Success	200		{object}	map[string]interface{}	"fileName:文件路径,newFileName:文件名,url:完整URL,originalFilename:原始文件名"
//	@Router		/common/upload [post]
//	@Security	BearerAuth
func (h *CommonHandler) Upload(c *gin.Context) {
	response.SetOperTitle(c, "通用文件上传")
	file, err := c.FormFile("file")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "获取文件失败")
		return
	}

	// 检查文件大小
	if fileuploadtool.CheckFileSize(file.Size) {
		response.Error(c, "文件过大")
		return
	}

	// 检查文件名长度
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

	// 生成唯一文件名
	filename := fileuploadtool.GenerateFilename(file.Filename)
	fullPath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "保存文件失败")
		return
	}

	// 构建相对路径（用于前端存储）
	relativePath := filepath.Join(fileuploadtool.GetDateSubDir(), filename)

	// 构建完整访问URL
	url := fmt.Sprintf("http://%s:%s/assets/%s", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port, relativePath)

	response.DataExpand(c, nil, map[string]interface{}{
		"fileName":       relativePath,
		"newFileName":    filename,
		"url":            url,
		"originalFilename": file.Filename,
	})
}

// UploadWangEditor WangEditor 富文本编辑器图片上传
//
//	@Summary	WangEditor图片上传
//	@Tags		公共上传
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		file	formData	file	true	"图片文件"
//	@Success	200		{object}	map[string]interface{}	"url:图片地址,alt:图片名称,href:链接"
//	@Router		/common/upload/wangeditor [post]
//	@Security	BearerAuth
func (h *CommonHandler) UploadWangEditor(c *gin.Context) {
	response.SetOperTitle(c, "富文本编辑器图片上传")
	file, err := c.FormFile("file")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "获取文件失败")
		return
	}

	// 检查文件大小
	if fileuploadtool.CheckFileSize(file.Size) {
		response.Error(c, "文件过大")
		return
	}

	// 检查是否为图片
	ext := filepath.Ext(file.Filename)
	if !fileuploadtool.IsImageFile(ext) {
		response.Error(c, "仅支持上传图片格式文件")
		return
	}

	// 获取上传目录
	uploadDir := fileuploadtool.GetUploadDir()
	if err := fileuploadtool.EnsureDir(uploadDir); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "创建上传目录失败")
		return
	}

	// 生成唯一文件名
	filename := fileuploadtool.GenerateFilename(file.Filename)
	fullPath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "保存文件失败")
		return
	}

	// 构建完整访问URL
	relativePath := filepath.Join(fileuploadtool.GetDateSubDir(), filename)
	url := fmt.Sprintf("http://%s:%s/assets/%s", config.GlobalConfig.Server.Host, config.GlobalConfig.Server.Port, relativePath)

	// WangEditor 期望的响应格式
	response.Data(c, map[string]interface{}{
		"url":  url,
		"alt":  file.Filename,
		"href": "",
	})
}

// Download 通用文件下载
//
//	@Summary	通用文件下载
//	@Tags		公共上传
//	@Produce	application/octet-stream
//	@Param		fileName	query	string	true	"文件名称"
//	@Param		delete		query	bool	false	"下载后是否删除"
//	@Success	200			{file}	binary	"文件内容"
//	@Router		/common/download [get]
//	@Security	BearerAuth
func (h *CommonHandler) Download(c *gin.Context) {
	response.SetOperTitle(c, "通用文件下载")
	fileName := c.Query("fileName")
	if fileName == "" {
		response.Error(c, "文件名不能为空")
		return
	}

	deleteAfter := c.Query("delete") == "true"

	// 构建完整文件路径
	fullPath := filepath.Join(fileuploadtool.GetBasePath(), fileName)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		response.Error(c, "文件不存在")
		return
	}

	// 获取文件信息
	fileInfo, _ := os.Stat(fullPath)
	
	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Download-Filename", fileInfo.Name())

	// 发送文件
	c.File(fullPath)

	// 下载后删除
	if deleteAfter {
		os.Remove(fullPath)
	}
}

// DownloadResource 下载资源文件
//
//	@Summary	下载资源文件
//	@Tags		公共上传
//	@Produce	application/octet-stream
//	@Param		resource	query	string	true	"资源路径"
//	@Success	200			{file}	binary	"文件内容"
//	@Router		/common/download/resource [get]
//	@Security	BearerAuth
func (h *CommonHandler) DownloadResource(c *gin.Context) {
	response.SetOperTitle(c, "下载资源文件")
	resource := c.Query("resource")
	if resource == "" {
		response.Error(c, "资源路径不能为空")
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(fileuploadtool.GetBasePath(), resource)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		response.Error(c, "资源文件不存在")
		return
	}

	// 获取文件信息
	fileInfo, _ := os.Stat(fullPath)

	// 根据文件扩展名获取 Content-Type
	ext := filepath.Ext(fileInfo.Name())
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	c.Header("Content-Type", contentType)
	c.Header("Download-Filename", fileInfo.Name())

	// 发送文件
	c.File(fullPath)
}
