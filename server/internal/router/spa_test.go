package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-fin-server/pkg"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	// 测试环境不初始化真实日志，用 no-op logger 避免 nil 引用
	pkg.Logger = zap.NewNop().Sugar()
}

func setupTestDist(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>index</html>"), 0o644)
	assets := filepath.Join(dir, "assets-web")
	os.MkdirAll(assets, 0o755)
	raw := []byte("console.log('hello world hello world hello world')")
	os.WriteFile(filepath.Join(assets, "app.js"), raw, 0o644)
	os.WriteFile(filepath.Join(assets, "app.js.gz"), []byte("GZIP-BYTES"), 0o644)
	os.WriteFile(filepath.Join(assets, "app.js.br"), []byte("BROTLI-BYTES"), 0o644)
	return dir
}

func newFrontendEngine(dir string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ServeFrontend(r, dir)
	return r
}

func TestServeFrontend_CompressionNegotiation(t *testing.T) {
	dir := setupTestDist(t)
	r := newFrontendEngine(dir)

	cases := []struct {
		name    string
		accept  string
		wantCE  string // 期望的 Content-Encoding，空表示不压缩
		wantBody string
	}{
		{"brotli优先", "gzip, deflate, br", "br", "BROTLI-BYTES"},
		{"gzip", "gzip, deflate", "gzip", "GZIP-BYTES"},
		{"不压缩", "", "", "console.log('hello world hello world hello world')"},
		{"仅br", "br", "br", "BROTLI-BYTES"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/assets-web/app.js", nil)
			if c.accept != "" {
				req.Header.Set("Accept-Encoding", c.accept)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("状态码 = %d, 期望 200", w.Code)
			}
			ce := w.Header().Get("Content-Encoding")
			if ce != c.wantCE {
				t.Errorf("Content-Encoding = %q, 期望 %q", ce, c.wantCE)
			}
			if w.Body.String() != c.wantBody {
				t.Errorf("响应体 = %q, 期望 %q", w.Body.String(), c.wantBody)
			}
			// 压缩响应应带 Vary: Accept-Encoding
			if c.wantCE != "" && !strings.Contains(w.Header().Get("Vary"), "Accept-Encoding") {
				t.Errorf("压缩响应缺少 Vary: Accept-Encoding")
			}
			// assets-web 下的构建产物应带长缓存头
			cc := w.Header().Get("Cache-Control")
			if !strings.Contains(cc, "max-age=31536000") {
				t.Errorf("Cache-Control = %q, 期望包含 max-age=31536000", cc)
			}
		})
	}
}

func TestServeFrontend_CacheHeaders(t *testing.T) {
	dir := setupTestDist(t)
	r := newFrontendEngine(dir)

	// 首页应带 no-cache
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("首页 Cache-Control = %q, 期望 no-cache", cc)
	}

	// SPA 回退应带 no-cache
	req = httptest.NewRequest(http.MethodGet, "/some/route", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if cc := w.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("SPA 回退 Cache-Control = %q, 期望 no-cache", cc)
	}

	// assets-web 资源应带长缓存
	req = httptest.NewRequest(http.MethodGet, "/assets-web/app.js", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("静态资源 Cache-Control = %q, 期望包含 immutable", cc)
	}
}

func TestServeFrontend_SPAAndTraversal(t *testing.T) {
	dir := setupTestDist(t)
	r := newFrontendEngine(dir)

	// SPA 回退：未知路径返回 index.html
	req := httptest.NewRequest(http.MethodGet, "/user-center/detail", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "index") {
		t.Errorf("history 回退失败: code=%d body=%q", w.Code, w.Body.String())
	}

	// 路径穿越应被拒绝
	req = httptest.NewRequest(http.MethodGet, "/assets-web/../index.html", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Errorf("路径穿越未被拦截: code=%d", w.Code)
	}
}
