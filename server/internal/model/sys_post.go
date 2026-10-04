package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysPost struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"id"    description:"岗位ID"`
	PostCode string `gorm:"column:post_code" json:"postCode"  description:"岗位编码"  binding:"required" err_msg:"岗位编码 必填"`
	PostName string `gorm:"column:post_name" json:"postName"  description:"岗位名称" binding:"required" err_msg:"岗位名称 必填"`
	PostSort *int   `gorm:"column:post_sort" json:"postSort"  description:"显示顺序" binding:"required" err_msg:"岗位顺序 必填"`
	Status   string `gorm:"column:status" json:"status"    description:"状态（0正常 1停用）"`
	Remark   string `gorm:"column:remark" json:"remark"    description:"备注"`
	BaseModel
}

func (c SysPost) TableName() string {
	return "sys_post"
}

func (c SysPost) Get(query string, args ...interface{}) (SysPost, error) {
	var data SysPost
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysPost) Create(data *SysPost) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysPost) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysPost{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysPost) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysPost{})
	return d.Error
}

func (c SysPost) GetAll(query string, args ...interface{}) ([]SysPost, error) {
	data := make([]SysPost, 0)
	d := db.DBConnections["master"].Model(&SysPost{}).Where(query, args...).Find(&data)
	return data, d.Error
}

func (c SysPost) SelectPostList(conditions map[string]interface{}, pageNum, pageSize int) ([]SysPost, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysPost{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysPost, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	d := query.Order("post_sort").Offset(offset).Limit(limit).Find(&data)

	return data, count, d.Error
}

func (c SysPost) SelectPostAllList(conditions map[string]interface{}) ([]SysPost, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysPost{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysPost, 0)
	d := query.Order("post_sort").Find(&data)

	return data, d.Error
}

func (c SysPost) SelectPostsByUserName(userName string) ([]SysPost, error) {
	data := make([]SysPost, 0)
	sql := ` select p.id, p.post_name, p.post_code
		from sys_post p
			 left join sys_user_post up on up.post_id = p.id
			 left join sys_user u on u.id = up.user_id
		where u.user_name = ? `
	d := db.DBConnections["master"].Raw(sql, userName).Scan(&data)
	return data, d.Error
}
