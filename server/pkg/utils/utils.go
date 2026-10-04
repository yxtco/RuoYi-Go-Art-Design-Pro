package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// RemoveEmptyConditions 移除条件中值为空的项
func RemoveEmptyConditions(conditions map[string]interface{}) {
	for key, value := range conditions {
		if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
			delete(conditions, key)
		}
	}
}

// StructToMapWithGormColumn 将结构体转换为map[string]interface{}，处理gorm:"column:xxx"标签
func StructToMapWithGormColumn(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		fmt.Println("Expected struct")
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")
		if gormTag == "-" {
			continue
		}
		var columnName string

		// 解析gorm标签找到column的名称
		if strings.Contains(gormTag, "column:") {
			tagParts := strings.Split(gormTag, ";")
			for _, part := range tagParts {
				if strings.Contains(part, "column:") {
					columnName = strings.Split(part, ":")[1]
					break
				}
			}
		}

		// 如果没有找到gorm的column标签，使用驼峰转下划线
		if columnName == "" {
			columnName = CamelToSnake(field.Name)
		}

		value := v.Field(i)
		result[columnName] = value.Interface()
	}

	return result
}

// StructToMap 将结构体字段转换为 map，支持 JSON 标签
func StructToMap(obj interface{}, conditions map[string]interface{}) {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		value := objValue.Field(i).Interface()

		// 获取字段的 JSON 标签，如果没有则使用字段名
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		// 这里可以根据需求进行类型转换，例如将 int 转为字符串
		conditions[tag] = value
	}
}

// StringToMap 字符串json转map
func StringToMap(str string) map[string]interface{} {
	var resMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &resMap)
	if err != nil {
		fmt.Println("string转map失败", err)
	}
	return resMap
}

func PageParse(pageNum, pageSize int, maxPageSize ...int) (int, int) {
	maxSize := 5
	if len(maxPageSize) > 0 {
		maxSize = maxPageSize[0]
	}
	if pageNum == 0 {
		pageNum = 1
	}
	if pageSize == 0 || pageSize > maxSize {
		pageSize = 10
	}
	return (pageNum - 1) * pageSize, pageSize
}

// BuildConditions 根据结构体字段的自定义 Tag 构建查询条件
func BuildConditions(input interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Struct {
		return result // 如果不是结构体，直接返回空结果
	}

	for i := 0; i < val.NumField(); i++ {
		//fieldName := val.Type().Field(i).Name
		fieldValue := val.Field(i).Interface()

		// 如果 fieldValue 为空，跳过该条件
		if isEmptyValue(fieldValue) {
			continue
		}

		// 获取 Struct Tag
		tag := val.Type().Field(i).Tag.Get("query")
		if tag == "" {
			continue
		}

		// 解析 Struct Tag，获取列名、操作符和额外信息
		parts := strings.Split(tag, ":")
		if len(parts) >= 2 {
			columnName := parts[0]
			operator := parts[1]

			// 根据不同的操作符构建条件
			switch operator {
			case "=":
				result[columnName+" = ?"] = fieldValue
			case "!=":
				result[columnName+" != ?"] = fieldValue
			case "like":
				result[columnName+" LIKE ?"] = fmt.Sprintf("%%%v%%", fieldValue)
			case ">":
				result[columnName+" > ?"] = fieldValue
			case "<":
				result[columnName+" < ?"] = fieldValue
			case ">=":
				result[columnName+" >= ?"] = fieldValue
			case "<=":
				result[columnName+" <= ?"] = fieldValue
			case "in":
				result[columnName+" IN (?)"] = fieldValue
			case "not_in":
				result[columnName+" NOT IN (?)"] = fieldValue
			// 添加更多操作符的情况...
			default:
				// 如果没有匹配的操作符，默认使用等于操作符
				result[columnName] = fieldValue
			}
		}
	}

	return result
}

func BuildSortConditions(sort string) []string {
	sortConditions := strings.Split(sort, ",")
	orderList := make([]string, 0)
	for _, condition := range sortConditions {
		fields := strings.Split(condition, ":")
		if len(fields) != 2 {
			continue // 跳过格式不正确的条件
		}
		column, direction := fields[0], strings.ToLower(fields[1])
		if direction != "asc" && direction != "desc" {
			continue // 跳过排序方向不正确的条件
		}
		// 应用排序条件
		orderList = append(orderList, CamelToSnake(column)+" "+direction)
	}
	return orderList
}

// CamelToSnake 将驼峰命名字符串转换为下划线命名
func CamelToSnake(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// isEmptyValue 检查值是否为空
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	// 如果是指针类型，检查指针是否为 nil
	if reflect.ValueOf(value).Kind() == reflect.Ptr {
		return reflect.ValueOf(value).IsNil()
	}

	// 检查值是否为零值
	switch reflect.ValueOf(value).Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return reflect.ValueOf(value).Len() == 0
	default:
		zero := reflect.Zero(reflect.TypeOf(value)).Interface()
		return reflect.DeepEqual(value, zero)
	}
}

// BuildTree 将扁平的列表数据转换为树形结构
func BuildTree(nodes []map[string]interface{}, idKey, parentIDKey, childrenKey string) []map[string]interface{} {

	nodeMap := make(map[int64]map[string]interface{})

	// 构建节点映射
	for _, node := range nodes {
		id, ok := node[idKey].(int64)
		if !ok {
			// 处理错误
			continue
		}
		nodeMap[id] = node
	}

	// 将节点连接起来
	var tree []map[string]interface{}
	for _, node := range nodes {
		parentID, ok := node[parentIDKey].(int64)
		if !ok {
			// 处理错误
			continue
		}

		parent, exists := nodeMap[parentID]
		if !exists || parentID == 0 {
			// 如果没有父节点或者父节点为 0，则将自身作为根节点添加
			tree = append(tree, node)
		} else {
			// 将自身作为子节点添加到父节点的 Children 中
			children, ok := parent[childrenKey].([]map[string]interface{})
			if !ok {
				children = make([]map[string]interface{}, 0)
			}
			children = append(children, node)
			parent[childrenKey] = children
		}
	}
	// 对树进行排序，确保父节点在子节点之前
	sort.Slice(tree, func(i, j int) bool {
		return tree[i][parentIDKey].(int64) < tree[j][parentIDKey].(int64)
	})
	return tree
}

func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	mergedMap := make(map[string]interface{})

	// 遍历第一个map，将其键值对添加到合并后的map中
	for key, value := range map1 {
		mergedMap[key] = value
	}

	// 遍历第二个map，将其键值对添加到合并后的map中
	for key, value := range map2 {
		mergedMap[key] = value
	}

	return mergedMap
}

func SliceUint64ContainsElement(arr []uint64, element uint64) bool {
	for _, value := range arr {
		if value == element {
			return true
		}
	}
	return false
}
