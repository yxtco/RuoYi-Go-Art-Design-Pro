package router

import (
	"go-fin-server/pkg"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultWebDistDir 未配置 server.web_dist 时的默认前端产物目录（前端 build 输出到 server/web-dist）
const defaultWebDistDir = "./web-dist"

// ServeFrontend 让后端直接托管前端编译产物（SPA，无需 Nginx）。
// 支持：
//   - 静态资源（assets-web 等）直接返回
//   - 构建期生成的 .gz / .br 压缩文件：按 Accept-Encoding 返回压缩版本，加快加载
//   - 前端 history 路由回退到 index.html
//
// 注意：若配置的产物目录不存在（例如本地仅跑接口不托管页面），则跳过托管，
// 不影响后端 API 正常使用。
func ServeFrontend(r *gin.Engine, distDir string) {
	if distDir == "" {
		distDir = defaultWebDistDir
	}

	indexPath := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		pkg.Logger.Warnf("前端产物目录不存在，跳过静态托管: %s", distDir)
		return
	}

	// 前端编译静态资源（assets-web 等，与上传目录 /assets 区分开）
	staticDir := filepath.Join(distDir, "assets-web")
	if _, err := os.Stat(staticDir); err == nil {
		// 自定义静态服务：优先返回 .br / .gz 压缩文件
		r.GET("/assets-web/*filepath", func(c *gin.Context) {
			// filepath 形如 "/app.js" 或 "/js/x.js"；先检查穿越再规范化
			rel := c.Param("filepath")
			if strings.Contains(rel, "..") {
				c.Status(http.StatusForbidden)
				return
			}
			rel = filepath.Clean(filepath.FromSlash(strings.TrimPrefix(rel, "/")))
			// 构建产物文件名含 content hash，可安全设置长缓存
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			serveStaticCompressed(c, staticDir, rel)
		})
	}

	// favicon（每次启动可能变化，不做强缓存）
	if _, err := os.Stat(filepath.Join(distDir, "favicon.ico")); err == nil {
		r.GET("/favicon.ico", func(c *gin.Context) {
			c.Header("Cache-Control", "public, max-age=86400")
			c.File(filepath.Join(distDir, "favicon.ico"))
		})
	}

	// 首页（SPA 入口，不缓存以保证每次获取最新版本）
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.File(indexPath)
	})

	// history 路由回退：未匹配的 GET 请求返回 index.html（浏览器路由刷新/直达）
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Header("Cache-Control", "no-cache")
			c.File(indexPath)
			return
		}
		c.Status(http.StatusNotFound)
	})

	pkg.Logger.Infof("前端静态托管已启用: distDir=%s", distDir)
}

// serveStaticCompressed 返回静态资源，优先提供构建期生成的 .br / .gz 压缩版本。
// baseDir 为静态资源根目录（如 dist/assets-web），relPath 为其下的相对路径。
func serveStaticCompressed(c *gin.Context, baseDir, relPath string) {
	// 安全：禁止路径穿越（双重防御）
	if strings.Contains(relPath, "..") || filepath.IsAbs(relPath) {
		c.Status(http.StatusForbidden)
		return
	}

	originalPath := filepath.Join(baseDir, relPath)
	original, err := os.Open(originalPath)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer original.Close()

	st, err := original.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 按客户端 Accept-Encoding 协商压缩格式
	enc := negotiateEncoding(c.GetHeader("Accept-Encoding"))
	if enc != "" {
		ext := ".gz"
		if enc == "br" {
			ext = ".br"
		}
		if vPath := originalPath + ext; fileExists(vPath) {
			vFile, openErr := os.Open(vPath)
			if openErr == nil {
				defer vFile.Close()
				c.Header("Content-Encoding", enc)
				c.Header("Vary", "Accept-Encoding")
				http.ServeContent(c.Writer, c.Request, filepath.Base(originalPath), st.ModTime(), vFile)
				return
			}
		}
	}

	// 无压缩版本时返回原文件
	http.ServeContent(c.Writer, c.Request, filepath.Base(originalPath), st.ModTime(), original)
}

// negotiateEncoding 根据 Accept-Encoding 决定优先使用的压缩格式（br > gzip）。
func negotiateEncoding(acceptEncoding string) string {
	ae := strings.ToLower(acceptEncoding)
	switch {
	case strings.Contains(ae, "br"):
		return "br"
	case strings.Contains(ae, "gzip"):
		return "gzip"
	default:
		return ""
	}
}

// fileExists 判断文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
