package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

type Options struct {
	Addr        string
	Root        string
	Quiet       bool
	CertPath    string
	KeyPath     string
	AllowList   []string
	TlsCertFile string
	TlsKeyFile  string
}

// GetAbsPath 获取绝对路径
func GetAbsPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// 确保路径存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf(i18n.T(i18n.GetLanguage(), "directory_not_exists"), absPath)
	}
	return absPath, nil
}

// CheckAccess 检查访问权限
func CheckAccess(root string, allowList []string) error {
	if !isPathAllowed(root, allowList) {
		return fmt.Errorf(i18n.T(i18n.GetLanguage(), "path_not_allowed"), root)
	}
	return nil
}

func Run(opt Options) error {
	// 检查访问权限
	if err := CheckAccess(opt.Root, opt.AllowList); err != nil {
		return err
	}

	handler := NewHandler(opt.Root, opt.Quiet, opt.AllowList)

	srv := &http.Server{
		Addr:    opt.Addr,
		Handler: handler,
	}

	if !opt.Quiet {
		lang := i18n.GetLanguage()
		fmt.Printf("🚀 %s\n", i18n.T(lang, "server_started"))
		fmt.Printf("📁 %s: %s\n", i18n.T(lang, "shared_dir"), opt.Root)
		if len(opt.AllowList) > 0 {
			fmt.Printf("✅ %s: %v\n", i18n.T(lang, "access_whitelist"), opt.AllowList)
		}
		fmt.Printf("🌐 %s: https://localhost%s\n", i18n.T(lang, "access_address"), opt.Addr)
		fmt.Printf("🔐 %s: %s\n", i18n.T(lang, "listen_address"), opt.Addr)
		fmt.Printf("💡 %s\n", i18n.T(lang, "tip_open_browser"))
		fmt.Printf("%s\n", i18n.T(lang, "tip_stop_server"))
		fmt.Println()
	}

	// 如果提供了外挂证书，则使用外挂证书，否则使用内置证书
	if opt.TlsCertFile != "" && opt.TlsKeyFile != "" {
		// 验证外挂证书文件是否存在
		if _, err := os.Stat(opt.TlsCertFile); err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "cert_file_not_exists"), opt.TlsCertFile)
		}
		if _, err := os.Stat(opt.TlsKeyFile); err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "key_file_not_exists"), opt.TlsKeyFile)
		}
		return srv.ListenAndServeTLS(opt.TlsCertFile, opt.TlsKeyFile)
	} else {
		// 使用内置证书
		tlsConfig, err := LoadTLSConfig(opt.CertPath, opt.KeyPath)
		if err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "tls_config_failed"), err)
		}
		srv.TLSConfig = tlsConfig
		return srv.ListenAndServeTLS("", "")
	}
}
