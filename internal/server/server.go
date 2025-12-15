package server

import (
	"fmt"
	"net/http"
)

type Options struct {
	Addr     string
	Root     string
	Quiet    bool
	CertPath string
	KeyPath  string
}

func Run(opt Options) error {
	tlsConfig, err := LoadTLSConfig(opt.CertPath, opt.KeyPath)
	if err != nil {
		return err
	}

	handler := NewHandler(opt.Root, opt.Quiet)

	srv := &http.Server{
		Addr:      opt.Addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	if !opt.Quiet {
		fmt.Printf("🚀 HTTPS 服务器已启动\n")
		fmt.Printf("📁 共享目录: %s\n", opt.Root)
		fmt.Printf("🌐 访问地址: https://localhost%s\n", opt.Addr)
		fmt.Printf("🔐 监听地址: %s\n", opt.Addr)
		fmt.Println("💡 提示: 在浏览器中打开访问地址即可浏览文件")
		fmt.Println("🛑 按 Ctrl+C 停止\n")
	}

	return srv.ListenAndServeTLS("", "")
}