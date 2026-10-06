package system

import (
	"go-fin-server/internal/config"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils/addressutils"
	"go-fin-server/pkg/utils/httputils"
	"go-fin-server/pkg/utils/stringutils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mssola/useragent"
)

type SysLoginService struct {
	Request *gin.Context
}

func (c SysLoginService) New(request *gin.Context) *SysLoginService {
	return &SysLoginService{
		request,
	}
}

// InsertLoginInfo 新增登录日志
func (c SysLoginService) InsertLoginInfo(userName, status, msg string) {
	userAgent := useragent.New(c.Request.GetHeader("User-Agent"))
	ip := httputils.GetClientIP(c.Request)
	os := userAgent.OS()
	browser, _ := userAgent.Browser()
	log := model.SysLoginInfo{
		UserName:      userName,
		Ipaddr:        ip,
		LoginLocation: addressutils.GetRealAddressByIP(ip, config.GlobalConfig.IsAddressEnabled),
		Browser:       browser,
		Os:            os,
		Msg:           msg,
		LoginTime:     types.LocalTime{Time: time.Now()},
	}
	// 日志状态
	if stringutils.StringInSlice(status, []string{constant.LOGIN_SUCCESS, constant.LOGOUT, constant.REGISTER}) {
		log.Status = constant.SUCCESS
	} else if constant.LOGIN_FAIL == status {
		log.Status = constant.FAIL
	}
	err := log.Create()
	if err != nil {
		pkg.Logger.Error(err)
	}
}
