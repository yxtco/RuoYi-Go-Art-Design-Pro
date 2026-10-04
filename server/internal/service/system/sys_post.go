package system

import (
	"errors"
	"fmt"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"

	"gorm.io/gorm"
)

type SysPostService struct {
}

// SelectPostList 查询部门管理数据
func (c SysPostService) SelectPostList(conditions map[string]interface{}, pageNum, pageSize int) ([]model.SysPost, int64, error) {
	data, count, err := model.SysPost{}.SelectPostList(conditions, pageNum, pageSize)
	if err != nil {
		return nil, count, err
	}
	return data, count, nil
}

// SelectPostAllList 查询部门管理数据
func (c SysPostService) SelectPostAllList(conditions map[string]interface{}) ([]model.SysPost, error) {
	data, err := model.SysPost{}.SelectPostAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectPostById 根据岗位编号获取详细信息
func (c SysPostService) SelectPostById(postId uint64) (model.SysPost, error) {
	data, err := model.SysPost{}.Get("id = ?", postId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckPostNameUnique 校验岗位名称是否唯一
func (c SysPostService) CheckPostNameUnique(postId uint64, postName string) bool {
	var post model.SysPost
	postObj, err := post.Get("post_name = ?", postName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if postObj.Id != postId {
		return false
	}
	return true
}

// CheckPostCodeUnique 校验岗位编码是否唯一
func (c SysPostService) CheckPostCodeUnique(postId uint64, postCode string) bool {
	var post model.SysPost
	postObj, err := post.Get("post_code = ?", postCode)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if postObj.Id != postId {
		return false
	}
	return true
}

// InsertPost 新增保存岗位信息
func (c SysPostService) InsertPost(data *model.SysPost) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePost 修改保存岗位信息
func (c SysPostService) UpdatePost(postId uint64, upd map[string]interface{}) error {
	err := model.SysPost{}.UpdateMap(upd, "id = ?", postId)
	if err != nil {
		return err
	}
	return nil
}

// DeletePostByIds 批量删除岗位信息
func (c SysPostService) DeletePostByIds(postIds []uint64) error {
	for _, postId := range postIds {
		post, err := model.SysPost{}.Get("id = ?", postId)
		if err != nil {
			return err
		}
		count, err := c.CountUserPostById(postId)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New(fmt.Sprintf("%s已分配,不能删除", post.PostName))
		}
	}
	err := model.SysPost{}.Delete("id in (?)", postIds)
	if err != nil {
		return err
	}
	return nil
}

// CountUserPostById 通过岗位ID查询岗位使用数量
func (c SysPostService) CountUserPostById(postId uint64) (int64, error) {
	count, err := model.SysUserPost{}.CountUserPostById(postId)
	if err != nil {
		return count, err
	}
	return count, nil
}
