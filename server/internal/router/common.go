package router

import (
	"go-fin-server/internal/handler/common"

	"github.com/gin-gonic/gin"
)

// SetupCommonRoutes 设置公共路由（文件上传/下载、序号生成等）
func SetupCommonRoutes(r *gin.Engine) {
	commonHandler := common.NewCommonHandler()
	sequenceHandler := common.NewSequenceHandler()

	// 公共上传下载路由组
	commonGroup := r.Group("/common")
	{
		commonGroup.POST("upload", commonHandler.Upload)                      // 通用文件上传
		commonGroup.POST("upload/wangeditor", commonHandler.UploadWangEditor) // WangEditor图片上传
		commonGroup.GET("download", commonHandler.Download)                   // 通用文件下载
		commonGroup.GET("download/resource", commonHandler.DownloadResource)  // 资源文件下载

		// 序号生成路由
		sequenceGroup := commonGroup.Group("/sequence")
		{
			sequenceGroup.GET("next", sequenceHandler.NextSequence)            // 生成下一个序号
			sequenceGroup.GET("current", sequenceHandler.GetCurrentSequence)   // 获取当前序号
			sequenceGroup.DELETE("reset", sequenceHandler.ResetSequence)       // 重置序号
			sequenceGroup.POST("batch", sequenceHandler.BatchSequences)        // 批量生成序号
		}
	}
}
