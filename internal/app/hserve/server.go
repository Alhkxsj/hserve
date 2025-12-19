package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Options struct {
	Addr           string
	Root           string
	Quiet          bool
	CertPath       string
	KeyPath        string
	Paths          []string      // 指定要分享的特定路径列表
	ReadTimeout    time.Duration // 读取超时
	WriteTimeout   time.Duration // 写入超时
	IdleTimeout    time.Duration // 空闲超时
	MaxHeaderBytes int           // 最大请求头大小
	MaxBodyBytes   int64         // 最大请求体大小
	AuthUser       string        // 基本身份验证用户名
	AuthPass       string        // 基本身份验证密码
	AuthRealm      string        // 基本身份验证领域
}

func Run(opt Options) error {
	if err := PreflightCheck(opt.Addr, opt.CertPath, opt.KeyPath); err != nil {
		return err
	}

	tlsConfig, err := LoadTLSConfig(opt.CertPath, opt.KeyPath)
	if err != nil {
		return err
	}

	handler := NewHandler(opt.Root, opt.Quiet, opt.Paths)

	// 设置默认值
	readTimeout := opt.ReadTimeout
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second
	}
	writeTimeout := opt.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 30 * time.Second
	}
	idleTimeout := opt.IdleTimeout
	if idleTimeout <= 0 {
		idleTimeout = 120 * time.Second
	}
	maxHeaderBytes := opt.MaxHeaderBytes
	if maxHeaderBytes <= 0 {
		maxHeaderBytes = 1 << 20 // 1 MB
	}
	maxBodyBytes := opt.MaxBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = 10 << 20 // 10 MB
	}

	// 应用中间件：限制请求体大小，然后是基本身份验证，最后是 Gzip 压缩
	handler = LimitRequestBodySize(maxBodyBytes)(handler)

	// 如果配置了身份验证，则应用身份验证中间件
	if opt.AuthUser != "" || opt.AuthPass != "" {
		authRealm := opt.AuthRealm
		if authRealm == "" {
			authRealm = "hserve-secure-area"
		}
		handler = BasicAuthMiddleware(opt.AuthUser, opt.AuthPass, authRealm)(handler)
	}

	handler = GzipMiddleware(handler)

	srv := &http.Server{
		Addr:           opt.Addr,
		Handler:        handler,
		TLSConfig:      tlsConfig,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// 设置优雅关闭
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("\n⏳ 正在优雅关闭服务器...")

		// 创建5秒的超时上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 关闭服务器，这会停止接受新连接
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf("❌ 服务器关闭出错: %v\n", err)
			// 如果优雅关闭失败，强制关闭
			if closeErr := srv.Close(); closeErr != nil {
				fmt.Printf("❌ 服务器强制关闭出错: %v\n", closeErr)
			}
		} else {
			fmt.Println("✅ 服务器已优雅关闭")
		}
		close(idleConnsClosed)
	}()

	if !opt.Quiet {
		fmt.Printf("🚀 hserve 已启动\n")
		fmt.Printf("📁 共享目录: %s\n", opt.Root)
		if len(opt.Paths) > 0 {
			fmt.Printf("🎯 分享路径: %v\n", opt.Paths)
		}
		fmt.Printf("🌐 访问地址: https://localhost%s\n", opt.Addr)
		fmt.Printf("🔐 监听地址: %s\n", opt.Addr)
		fmt.Printf("⏱️  超时设置: 读取=%v, 写入=%v, 空闲=%v\n", readTimeout, writeTimeout, idleTimeout)
		fmt.Printf("📊 大小限制: 最大请求体=%v, 最大请求头=%v\n", maxBodyBytes, maxHeaderBytes)
		if opt.AuthUser != "" {
			fmt.Printf("🔐 身份验证: 已启用 (用户: %s)\n", opt.AuthUser)
		}
		fmt.Println("💡 提示: 在浏览器中打开访问地址即可浏览文件")
		fmt.Print("🛑 按 Ctrl+C 停止\n\n")
	}

	// 启动服务器
	if err := srv.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
		return err
	}

	// 等待优雅关闭完成
	<-idleConnsClosed
	return nil
}
