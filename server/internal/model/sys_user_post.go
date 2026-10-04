package model

import "go-fin-server/internal/db"

type SysUserPost struct {
	UserId uint64 `gorm:"column:user_id" json:"userId" description:"用户ID"`
	PostId int64  `gorm:"column:post_id" json:"postId" description:"岗位ID"`
}

func (c SysUserPost) TableName() string {
	return "sys_user_post"
}

func (c SysUserPost) DeleteUserPost(userIds []uint64) error {
	d := db.DBConnections["master"].Where("user_id  in (?)", userIds).Delete(&SysUserPost{})
	return d.Error
}

func (c SysUserPost) DeleteUserPostByUserId(userId uint64) error {
	d := db.DBConnections["master"].Where("user_id = ?", userId).Delete(&SysUserPost{})
	return d.Error
}

func (c SysUserPost) BatchUserPost(list []SysUserPost) error {
	if len(list) > 0 {
		d := db.DBConnections["master"].Create(&list)
		return d.Error
	}
	return nil
}

func (c SysUserPost) CountUserPostById(postId uint64) (int64, error) {
	sql := `
		        select count(1) from sys_user_post where post_id= ?  
	`
	var count int64
	d := db.DBConnections["master"].Raw(sql, postId).Scan(&count)
	return count, d.Error
}
