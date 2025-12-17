package certmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

// GetCertPaths 返回证书和私钥路径
func GetCertPaths() (string, string) {
	var certPath, keyPath string
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			certPath = filepath.Join(prefix, "etc", "hserve", "certs", "server.crt")
			keyPath = filepath.Join(prefix, "etc", "hserve", "certs", "server.key")
		} else {
			certPath = filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "server.crt")
			keyPath = filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "server.key")
		}
	} else {
		certPath = filepath.Join("/etc/hserve/certs", "server.crt")
		keyPath = filepath.Join("/etc/hserve/certs", "server.key")
	}
	return certPath, keyPath
}

// GetCACertPath 返回 CA 证书路径
func GetCACertPath() string {
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			return filepath.Join(prefix, "etc", "hserve", "certs", "ca.crt")
		} else {
			return filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "ca.crt")
		}
	} else {
		return filepath.Join("/etc/hserve/certs", "ca.crt")
	}
}

// ShowInstructions 显示安装证书说明
func ShowInstructions(caCertPath string) {
	lang := i18n.GetLanguage()
	fmt.Println()
	fmt.Printf("%s\n", i18n.T(lang, "android_install_steps"))
	fmt.Printf("1. %s: %s\n", i18n.T(lang, "android_install_step1"), caCertPath)
	fmt.Printf("2. %s\n", i18n.T(lang, "android_install_step2"))
	fmt.Printf("3. %s\n", i18n.T(lang, "android_install_step3"))
	fmt.Printf("4. %s\n", i18n.T(lang, "android_install_step4"))
	fmt.Printf("5. %s\n", i18n.T(lang, "android_install_step5"))
	fmt.Println()
	fmt.Printf("%s\n", i18n.T(lang, "launch_example"))
	fmt.Println("  cd /path/to/website")
	fmt.Println("  hserve")
	fmt.Println()
	fmt.Println(i18n.T(lang, "poem"))
}
