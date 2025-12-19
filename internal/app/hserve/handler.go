package server

import (
	"compress/gzip"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewHandler(root string, quiet bool, paths []string) http.Handler {
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装 ResponseWriter 以捕获状态码
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 安全头部
		secureHeaders(lrw)

		// 路径安全检查
		if !isPathSafe(r.URL.Path, root) {
			http.Error(lrw, "Forbidden", http.StatusForbidden)
			logRequest(r, lrw.statusCode, time.Since(start), quiet)
			return
		}

		// 如果指定了特定路径，检查请求的路径是否在允许的路径中
		if len(paths) > 0 && !isPathAllowed(r.URL.Path, paths, root) {
			http.Error(lrw, "Forbidden", http.StatusForbidden)
			logRequest(r, lrw.statusCode, time.Since(start), quiet)
			return
		}

		// 阻止访问隐藏文件和系统文件
		cleanPath := filepath.Clean(r.URL.Path)
		segments := strings.Split(cleanPath, "/")
		for _, segment := range segments {
			if strings.HasPrefix(segment, ".") && segment != "." && segment != ".." {
				// 阻止访问以点开头的隐藏文件/目录
				http.Error(lrw, "Forbidden", http.StatusForbidden)
				logRequest(r, lrw.statusCode, time.Since(start), quiet)
				return
			}
		}

		fs.ServeHTTP(lrw, r)

		logRequest(r, lrw.statusCode, time.Since(start), quiet)
	})
}

func logRequest(r *http.Request, statusCode int, duration time.Duration, quiet bool) {
	if !quiet {
		// 记录请求信息
		fmt.Printf("[%s] %s %s %d %v\n",
			time.Now().Format("15:04:05"),
			r.Method,
			r.URL.Path,
			statusCode,
			duration.Round(time.Millisecond))
	}
}

// isPathAllowed 检查请求的路径是否在允许的路径列表中
func isPathAllowed(requestPath string, allowedPaths []string, root string) bool {
	if len(allowedPaths) == 0 {
		return true
	}

	cleanRequestPath := filepath.Clean(requestPath)

	// 如果请求的是根路径，允许访问
	if cleanRequestPath == "/" || cleanRequestPath == "." {
		return true
	}

	// 如果请求路径以允许的路径为前缀，允许访问
	for _, allowedPath := range allowedPaths {
		// 将相对路径转换为绝对路径
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		// 获取相对于根目录的路径
		relPath, err := filepath.Rel(root, absAllowedPath)
		if err != nil {
			continue
		}

		// 处理请求路径，确保格式一致
		reqPath := strings.TrimPrefix(cleanRequestPath, "/")

		// 检查请求路径是否是允许路径的子路径
		if reqPath == relPath || strings.HasPrefix(reqPath, relPath+"/") {
			return true
		}
	}

	return false
}

// isPathSafe 确保请求路径不会访问到 root 目录之外的文件
func isPathSafe(requestPath, rootDir string) bool {
	// 清理请求路径
	cleanRequestPath := filepath.Clean(requestPath)

	// 获取完整的文件系统路径
	fullPath := filepath.Join(rootDir, cleanRequestPath)

	// 检查是否在 rootDir 下
	relPath, err := filepath.Rel(rootDir, fullPath)
	if err != nil {
		return false
	}

	// 检查相对路径是否以 ".." 开头，这表示访问父目录
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	// 检查目标路径是否仍在 rootDir 内
	resolvedPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil && !os.IsNotExist(err) {
		// 如果是符号链接错误，检查原始路径
		resolvedPath = fullPath
	}

	// 检查解析后的路径是否仍在 rootDir 内
	relResolvedPath, err := filepath.Rel(rootDir, resolvedPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relResolvedPath, "..")
}

func secureHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; font-src 'self' data:; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self';")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// GzipMiddleware 中间件提供 Gzip 压缩功能
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer func() {
			// 忽略关闭错误，因为响应已经发送
			_ = gz.Close()
		}()

		gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter 用于处理 gzip 响应
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
	wrote  bool
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gzw.wrote {
		gzw.Header().Del("Content-Length") // 压缩后的长度未知
		gzw.Header().Set("Content-Encoding", "gzip")
		gzw.wrote = true
	}
	return gzw.Writer.Write(b)
}

func (gzw *gzipResponseWriter) WriteHeader(statusCode int) {
	if !gzw.wrote {
		gzw.Header().Del("Content-Length")
		gzw.Header().Set("Content-Encoding", "gzip")
		gzw.wrote = true
	}
	gzw.ResponseWriter.WriteHeader(statusCode)
}

// BasicAuthMiddleware 中间件提供基本身份验证
func BasicAuthMiddleware(username, password, realm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果用户名和密码都为空，则跳过身份验证
			if username == "" && password == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {

				w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LimitRequestBodySize 中间件限制请求体大小
func LimitRequestBodySize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// 包装请求体以限制读取大小
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}
