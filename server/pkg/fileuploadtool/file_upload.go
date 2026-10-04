package fileuploadtool

import (
	"fmt"
	"go-fin-server/internal/config"
	"go-fin-server/pkg/utils/stringutils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultMaxSize        = 50 * 1024 * 1024
	DefaultFileNameLength = 100
	DefaultBaseDir        = "./uploads" // 默认基础目录
)

// GetBasePath 获取文件上传基础路径（从配置读取）
func GetBasePath() string {
	if config.GlobalConfig.Upload.BasePath != "" {
		return config.GlobalConfig.Upload.BasePath
	}
	return DefaultBaseDir
}

// GetMaxSize 获取最大上传文件大小（字节）
func GetMaxSize() int64 {
	if config.GlobalConfig.Upload.MaxSize > 0 {
		return config.GlobalConfig.Upload.MaxSize * 1024 * 1024
	}
	return DefaultMaxSize
}

// GetDateSubDir 获取日期子目录路径（格式：YYYY/MM/DD）
func GetDateSubDir() string {
	now := time.Now()
	return filepath.Join(
		fmt.Sprintf("%d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
	)
}

// EnsureDir 确保目录存在
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// GetUploadDir 获取完整的上传目录（基础路径 + 日期子目录）
func GetUploadDir() string {
	return filepath.Join(GetBasePath(), GetDateSubDir())
}

// GenerateFilename 生成唯一文件名
func GenerateFilename(originalName string) string {
	extension := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, extension)
	now := time.Now().Format("20060102150405")
	uuidWithHyphen := uuid.New()
	uuidStr := strings.ReplaceAll(uuidWithHyphen.String(), "-", "")
	return fmt.Sprintf("%s_%s_%s", now, uuidStr, strings.ToLower(name)) + extension
}

// CheckFileSize 检查文件大小是否超限
func CheckFileSize(size int64) bool {
	return size > GetMaxSize()
}

// CheckFileNameLength 检查文件名长度是否超限
func CheckFileNameLength(nameLength int) bool {
	return nameLength > DefaultFileNameLength
}

func isAllowedExtension(extension string, allowedExtension []string) bool {
	return stringutils.StringInSlice(extension, allowedExtension)
}

// IsImageFile 判断是否为图片文件
func IsImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico"}
	return isAllowedExtension(strings.ToLower(ext), imageExts)
}

// IsDocumentFile 判断是否为文档文件
func IsDocumentFile(ext string) bool {
	docExts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".md", ".csv"}
	return isAllowedExtension(strings.ToLower(ext), docExts)
}
