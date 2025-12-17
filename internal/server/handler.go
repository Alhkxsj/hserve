package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

func NewHandler(root string, quiet bool, allowList []string) http.Handler {
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查路径是否在白名单中
		requestPath := filepath.Join(root, r.URL.Path)
		if !isPathAllowed(requestPath, allowList) {
			http.Error(w, i18n.T(i18n.GetLanguage(), "forbidden_access"), http.StatusForbidden)
			if !quiet {
				fmt.Printf("[%s] %s %s - FORBIDDEN (%s)\n",
					time.Now().Format("15:04:05"),
					r.Method,
					r.URL.Path,
					i18n.T(i18n.GetLanguage(), "forbidden_access"))
			}
			return
		}

		// 安全头部
		secureHeaders(w)

		// 日志记录
		if !quiet {
			fmt.Printf("[%s] %s %s\n",
				time.Now().Format("15:04:05"),
				r.Method,
				r.URL.Path)
		}

		// 修复路径遍历问题
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		upath = filepath.Join(root, filepath.Clean(upath))
		r.URL.Path = upath

		fs.ServeHTTP(w, r)
	})
}

// isPathAllowed 检查路径是否在白名单中
func isPathAllowed(requestPath string, allowList []string) bool {
	if len(allowList) == 0 {
		return true // 没有白名单则允许所有路径
	}

	// 将请求路径转换为绝对路径进行比较
	absRequestPath, err := filepath.Abs(requestPath)
	if err != nil {
		return false
	}

	for _, allowedPath := range allowList {
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		// 检查请求路径是否在允许的路径下
		rel, err := filepath.Rel(absAllowedPath, absRequestPath)
		if err != nil {
			continue
		}

		// 如果相对路径不以".."开头，则说明请求路径在允许路径下
		if !strings.HasPrefix(rel, "..") && !strings.Contains(rel, "/../") && rel != ".." {
			return true
		}
	}

	return false
}

func secureHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}
