package common

import (
	"go-fin-server/internal/db"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/internal/service/common"
	"go-fin-server/pkg"

	"github.com/gin-gonic/gin"
)

type SequenceHandler struct {
	Services service.Services
}

func NewSequenceHandler() *SequenceHandler {
	return &SequenceHandler{}
}

// NextSequence 生成下一个序号
//
//	@Summary	生成下一个序号
//	@Tags		序号生成
//	@Produce	json
//	@Param		businessType	query	string	true	"业务类型（如 ORDER, INVOICE）"
//	@Param		prefix			query	string	false	"序号前缀（如 ORD, INV）"
//	@Param		datePattern		query	string	false	"日期格式（如 20060102, 200601, 空串不拼接日期）"
//	@Param		seqLength		query	int		false	"序号位数（默认3，如 3 → 001）"
//	@Param		resetCycle		query	string	false	"重置周期: daily/monthly/yearly/never（默认 daily）"
//	@Success	200				{object}	map[string]interface{}	"sequence:生成的完整序号"
//	@Router		/common/sequence/next [get]
//	@Security	BearerAuth
func (h *SequenceHandler) NextSequence(c *gin.Context) {
	response.SetOperTitle(c, "生成下一个序号")
	var req common.SequenceRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	seq, err := h.Services.SequenceService.NextSequence(db.RedisConnections["master"], req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	response.Data(c, map[string]interface{}{
		"sequence": seq,
	})
}

// GetCurrentSequence 获取当前序号值（不自增）
//
//	@Summary	获取当前序号值
//	@Tags		序号生成
//	@Produce	json
//	@Param		businessType	query	string	true	"业务类型"
//	@Param		prefix			query	string	false	"序号前缀"
//	@Param		datePattern		query	string	false	"日期格式"
//	@Param		seqLength		query	int		false	"序号位数"
//	@Param		resetCycle		query	string	false	"重置周期"
//	@Success	200				{object}	map[string]interface{}	"currentSequence:当前序号值"
//	@Router		/common/sequence/current [get]
//	@Security	BearerAuth
func (h *SequenceHandler) GetCurrentSequence(c *gin.Context) {
	response.SetOperTitle(c, "获取当前序号值")
	var req common.SequenceRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	seq, err := h.Services.SequenceService.GetCurrentSequence(db.RedisConnections["master"], req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	response.Data(c, map[string]interface{}{
		"currentSequence": seq,
	})
}

// ResetSequence 重置指定业务类型的序号
//
//	@Summary	重置序号
//	@Tags		序号生成
//	@Produce	json
//	@Param		businessType	query	string	true	"业务类型"
//	@Param		resetCycle		query	string	false	"重置周期"
//	@Success	200				{object}	map[string]interface{}	"操作结果"
//	@Router		/common/sequence/reset [delete]
//	@Security	BearerAuth
func (h *SequenceHandler) ResetSequence(c *gin.Context) {
	response.SetOperTitle(c, "重置序号")
	var req common.SequenceRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.Services.SequenceService.ResetSequence(db.RedisConnections["master"], req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	response.Data(c, nil)
}

// BatchSequences 批量生成序号
//
//	@Summary	批量生成序号
//	@Tags		序号生成
//	@Accept		json
//	@Produce	json
//	@Param		body	body		object		true	"批量生成参数 {businessType, prefix, datePattern, seqLength, resetCycle, count}"
//	@Success	200		{object}	map[string]interface{}	"sequences:序号数组"
//	@Router		/common/sequence/batch [post]
//	@Security	BearerAuth
func (h *SequenceHandler) BatchSequences(c *gin.Context) {
	response.SetOperTitle(c, "批量生成序号")
	var req struct {
		common.SequenceRequest
		Count int `json:"count" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.Count <= 0 || req.Count > 100 {
		response.Error(c, "批量数量必须在 1-100 之间")
		return
	}

	sequences := make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		seq, err := h.Services.SequenceService.NextSequence(db.RedisConnections["master"], req.SequenceRequest)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		sequences = append(sequences, seq)
	}

	response.Data(c, map[string]interface{}{
		"sequences": sequences,
	})
}
