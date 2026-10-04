package common

type Common struct {
	IdRequest
	LoginRequest
	IdUriRequest
	IdsUriRequest
	ChangeStatusRequest
}

type IdRequest struct {
	Id uint64 `json:"id" binding:"required" err_msg:"id 必填"`
}
type IdUriRequest struct {
	Id uint64 `uri:"id" binding:"required" err_msg:"id 必填"`
}
type IdsUriRequest struct {
	Ids string `uri:"ids" binding:"required" err_msg:"id 必填"`
}

type ChangeStatusRequest struct {
	Id     uint64 `json:"id" binding:"required" err_msg:"id 必填"`
	Status string `json:"status" binding:"required" err_msg:"状态 必填"`
}

type LoginRequest struct {
	UserName string `json:"username" binding:"required" err_msg:"请输入您的账号"`
	Password string `json:"password" binding:"required" err_msg:"请输入您的密码"`
	Code     string `json:"code" err_msg:"请输入验证码"`
	Uuid     string `json:"uuid"`
}

// PageReq 公共请求参数
type PageReq struct {
	DateRange []string `form:"dateRange"` //日期范围
	PageNum   int      `form:"pageNum"`   //当前页码
	PageSize  int      `form:"pageSize"`  //每页数
	OrderBy   string   //排序方式
}
