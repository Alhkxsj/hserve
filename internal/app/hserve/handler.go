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

// WriteHeader 实现 ResponseWriter 接口，记录状态码
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// NewHandler 创建一个新的 HTTP 处理器，提供文件服务功能
func NewHandler(root string, quiet bool, paths []string) http.Handler {
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, fs, root, quiet, paths)
	})
}

// handleRequest 处理 HTTP 请求的主要逻辑
func handleRequest(w http.ResponseWriter, r *http.Request, fs http.Handler, root string, quiet bool, paths []string) {
	start := time.Now()

	// 包装 ResponseWriter 以捕获状态码
	lrw := createLoggingResponseWriter(w)

	// 安全头部
	secureHeaders(lrw)

	// 检查请求安全性
	if !isRequestAllowed(r.URL.Path, root, paths, len(paths) > 0) {
		http.Error(lrw, "Forbidden", http.StatusForbidden)
		logRequest(r, lrw.statusCode, time.Since(start), quiet)
		return
	}

	fs.ServeHTTP(lrw, r)

	logRequest(r, lrw.statusCode, time.Since(start), quiet)
}

// createLoggingResponseWriter 创建日志响应写入器
func createLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// isRequestAllowed 检查请求是否被允许
func isRequestAllowed(path, root string, paths []string, hasSpecificPaths bool) bool {
	// 路径安全检查
	if !isPathSafe(path, root) {
		return false
	}

	// 如果指定了特定路径，检查请求的路径是否在允许的路径中
	if hasSpecificPaths && !isPathAllowed(path, paths, root) {
		return false
	}

	// 阻止访问隐藏文件和系统文件
	if isHiddenFileRequest(path) {
		return false
	}

	return true
}

// isHiddenFileRequest 检查请求是否为隐藏文件
func isHiddenFileRequest(path string) bool {
	cleanPath := filepath.Clean(path)
	segments := strings.Split(cleanPath, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ".") && segment != "." && segment != ".." {
			return true
		}
	}
	return false
}

// logRequest 记录 HTTP 请求信息
func logRequest(r *http.Request, statusCode int, duration time.Duration, quiet bool) {
	if !quiet {
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
	// 如果没有指定允许的路径，则允许所有路径
	if len(allowedPaths) == 0 {
		return true
	}

	// 检查是否请求根路径
	cleanRequestPath := filepath.Clean(requestPath)
	if cleanRequestPath == "/" || cleanRequestPath == "." {
		return true
	}

	// 检查请求路径是否在允许的路径中
	reqPath := strings.TrimPrefix(cleanRequestPath, "/")

	for _, allowedPath := range allowedPaths {
		// 检查请求路径是否是允许路径的子路径
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		relPath, err := filepath.Rel(root, absAllowedPath)
		if err != nil {
			continue
		}

		if reqPath == relPath || strings.HasPrefix(reqPath, relPath+"/") {
			return true
		}
	}

	return false
}

// isSubPathOf 检查请求路径是否是允许路径的子路径
func isSubPathOf(reqPath, allowedPath, root string) bool {
	absAllowedPath, err := filepath.Abs(allowedPath)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(root, absAllowedPath)
	if err != nil {
		return false
	}

	return reqPath == relPath || strings.HasPrefix(reqPath, relPath+"/")
}

// isPathSafe 确保请求路径不会访问到 root 目录之外的文件
func isPathSafe(requestPath, rootDir string) bool {
	// 清理请求路径
	cleanRequestPath := filepath.Clean(requestPath)

	// 获取完整的文件系统路径
	fullPath := filepath.Join(rootDir, cleanRequestPath)

	// 检查基本路径安全性
	relPath, err := filepath.Rel(rootDir, fullPath)
	if err != nil {
		return false
	}

	// 检查相对路径是否以 ".." 开头，这表示访问父目录
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	// 检查符号链接安全性
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
		// 检查是否支持gzip
		if !supportsGzip(r) {
			next.ServeHTTP(w, r)
			return
		}

		// 处理gzip响应
		handleGzipResponse(w, r, next)
	})
}

// supportsGzip 检查请求是否支持gzip
func supportsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// handleGzipResponse 处理gzip响应
func handleGzipResponse(w http.ResponseWriter, r *http.Request, next http.Handler) {
	gz := gzip.NewWriter(w)
	defer closeGzipWriter(gz)

	gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
	next.ServeHTTP(gzw, r)
}

// closeGzipWriter 关闭gzip写入器
func closeGzipWriter(gz *gzip.Writer) {
	// 忽略关闭错误，因为响应已经发送
	_ = gz.Close()
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
			// 处理身份验证
			if shouldSkipAuth(username, password) {
				next.ServeHTTP(w, r)
				return
			}

			if !isAuthenticated(r, username, password, realm, w) {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkipAuth 检查是否应该跳过身份验证
func shouldSkipAuth(username, password string) bool {
	return username == "" && password == ""
}

// isAuthenticated 检查请求是否已通过身份验证
func isAuthenticated(r *http.Request, username, password, realm string, w http.ResponseWriter) bool {
	user, pass, ok := r.BasicAuth()
	if !ok || !isCredentialsValid(user, pass, username, password) {
		sendUnauthorizedResponse(w, realm)
		return false
	}
	return true
}

// isCredentialsValid 检查凭据是否有效
func isCredentialsValid(user, pass, username, password string) bool {
	return subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1 &&
		   subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1
}

// sendUnauthorizedResponse 发送未授权响应
func sendUnauthorizedResponse(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// LimitRequestBodySize 中间件限制请求体大小
func LimitRequestBodySize(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查请求体大小
			if !isRequestBodySizeAcceptable(r, maxSize) {
				sendTooLargeResponse(w)
				return
			}

			// 包装请求体以限制读取大小
			wrapRequestBody(r, w, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// isRequestBodySizeAcceptable 检查请求体大小是否可接受
func isRequestBodySizeAcceptable(r *http.Request, maxSize int64) bool {
	return r.ContentLength <= maxSize
}

// sendTooLargeResponse 发送请求体过大响应
func sendTooLargeResponse(w http.ResponseWriter) {
	http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
}

// wrapRequestBody 包装请求体以限制读取大小
func wrapRequestBody(r *http.Request, w http.ResponseWriter, maxSize int64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
}
