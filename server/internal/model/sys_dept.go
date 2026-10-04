package model

import (
	"fmt"
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
	"sort"
	"strings"
)

type SysDept struct {
	Id        uint64 `gorm:"column:id;primary_key;" json:"id"        description:"部门id"`
	ParentId  uint64 `gorm:"column:parent_id" json:"parentId"  description:"父部门id" binding:"required" err_msg:"上级部门 必填"`
	Ancestors string `gorm:"column:ancestors" json:"ancestors" description:"祖级列表"`
	DeptName  string `gorm:"column:dept_name" json:"deptName"  description:"部门名称" binding:"required" err_msg:"部门名称 必填"`
	OrderNum  *int   `gorm:"column:order_num" json:"orderNum"  description:"显示顺序" binding:"required" err_msg:"显示排序 必填"`
	Leader    string `gorm:"column:leader" json:"leader"    description:"负责人"`
	Phone     string `gorm:"column:phone" json:"phone"     description:"联系电话"`
	Email     string `gorm:"column:email" json:"email"     description:"邮箱"`
	Status    string `gorm:"column:status;default:'0'" json:"status"    description:"部门状态（0正常 1停用）"`
	BaseModel
}

func (c SysDept) TableName() string {
	return "sys_dept"
}

func (c SysDept) Get(query string, args ...interface{}) (SysDept, error) {
	var data SysDept
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysDept) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(SysDept{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysDept) Create(data *SysDept) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysDept) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysDept{})
	return d.Error
}

func (c SysDept) HasChildByDeptId(query string, args ...interface{}) (int64, error) {
	var count int64
	if err := db.DBConnections["master"].Model(&SysDept{}).Where(query, args...).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}

func (c SysDept) CheckDeptExistUser(query string, args ...interface{}) (int64, error) {
	var count int64
	if err := db.DBConnections["master"].Model(&SysUser{}).Where(query, args...).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}

type SysDeptTreeNode struct {
	Id       uint64             `gorm:"column:id" json:"id"     description:"部门id"`
	DeptName string             `gorm:"column:dept_name" json:"label"  description:"部门名称"`
	ParentId uint64             `gorm:"column:parent_id" json:"-"  description:"父部门id"`
	OrderNum int                `gorm:"column:order_num" json:"-"  description:"显示顺序"`
	Children []*SysDeptTreeNode `gorm:"-" json:"children,omitempty"`
}

func (c SysDeptTreeNode) TableName() string {
	return "sys_dept"
}

func (c SysDept) SelectDeptList(conditions map[string]interface{}, sortConditions []string) ([]SysDept, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDept{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysDept, 0)
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Find(&data)

	return data, d.Error
}

func (c SysDept) SelectDataScopeDeptList(conditions map[string]interface{}, sortConditions []string, user SysUser) ([]SysDept, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDept{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	query = query.Scopes(DataScope(user, c.TableName()))

	data := make([]SysDept, 0)
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Find(&data)
	return data, d.Error
}

func (c SysDept) SelectNormalChildrenDeptById(deptId uint64) (int64, error) {
	sql := `
		        select count(*) from sys_dept where status = '0' and delete_time is NULL and find_in_set(?, ancestors)
	`
	var count int64
	d := db.DBConnections["master"].Raw(sql, deptId).Scan(&count)
	return count, d.Error
}

func (c SysDept) SelectChildrenDeptById(deptId uint64) ([]SysDept, error) {
	sql := `
    select * from sys_dept where find_in_set(?, ancestors)
	`
	data := make([]SysDept, 0)
	d := db.DBConnections["master"].Raw(sql, deptId).Scan(&data)
	return data, d.Error
}

func (c SysDept) UpdateDeptChildren(depts []SysDept) error {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString("update sys_dept set ancestors = ")
	sqlBuilder.WriteString("CASE dept_id\n")
	deptIds := make([]uint64, 0)
	for _, item := range depts {
		sqlBuilder.WriteString(fmt.Sprintf("    WHEN %d THEN '%s'\n", item.Id, item.Ancestors))
		deptIds = append(deptIds, item.Id)
	}
	sqlBuilder.WriteString("END")
	sqlBuilder.WriteString(" where dept_id in (?)")
	d := db.DBConnections["master"].Exec(sqlBuilder.String(), deptIds)
	return d.Error
}

func (c SysDeptTreeNode) GetDeptTreeList() ([]*SysDeptTreeNode, error) {
	data := make([]*SysDeptTreeNode, 0)
	d := db.DBConnections["master"].Select("id,dept_name,parent_id,order_num").Model(&SysDeptTreeNode{}).Find(&data)
	return data, d.Error
}

func (c SysDeptTreeNode) BuildSysDeptTree(items []*SysDeptTreeNode) []*SysDeptTreeNode {
	dataMap := make(map[uint64]*SysDeptTreeNode)
	// 构建菜单映射
	for _, item := range items {
		dataMap[item.Id] = item
	}
	// 将菜单项连接起来
	var tree []*SysDeptTreeNode
	for _, item := range dataMap {
		if item.ParentId == 0 {
			// 如果没有父菜单项，则将自身作为根节点添加
			tree = append(tree, item)
		} else {
			// 将自身作为子菜单项添加到父菜单项的 Children 中
			parent, exists := dataMap[item.ParentId]
			if exists {
				parent.Children = append(parent.Children, item)
			}
		}
	}
	// 对最外层排序
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].OrderNum < tree[j].OrderNum
	})
	// 对子排序
	for _, d := range tree {
		d.SortChildren()
	}
	return tree
}

func (c SysDeptTreeNode) SortChildren() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].OrderNum < c.Children[j].OrderNum
	})

	// Recursively sort children's children
	for _, child := range c.Children {
		child.SortChildren()
	}
}
