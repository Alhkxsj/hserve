package certgen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Generate 生成证书
func Generate(force bool) error {
	certPath, keyPath := GetCertPaths()
	caCertPath := GetCACertPath()

	if !force && CheckCertificateExists(certPath) && CheckCertificateExists(caCertPath) {
		fmt.Println("✅ 证书已存在，无需重新生成")
		ShowInstructions(caCertPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}

	// 生成服务器证书
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 创建CA证书
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "Local HTTPS CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // CA证书有效期10年
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// 生成CA证书文件
	if err := writePem(caCertPath, "CERTIFICATE", caCertDER, 0644); err != nil {
		return err
	}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(30, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	if err := writePem(certPath, "CERTIFICATE", serverCertDER, 0644); err != nil {
		return err
	}
	if err := writePem(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey), 0600); err != nil {
		return err
	}

	fmt.Println("✅ 证书生成完成")
	fmt.Println("💡 温馨提示: 请妥善保管您的证书文件")
	ShowInstructions(caCertPath)
	return nil
}

// writePem 写入 PEM 文件
func writePem(path, typ string, data []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: typ, Bytes: data})
}

// GetCertPaths 返回证书和私钥路径
func GetCertPaths() (string, string) {
	var certPath, keyPath string
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			certPath = prefix + "/etc/https-server/cert.pem"
			keyPath = prefix + "/etc/https-server/key.pem"
		} else {
			certPath = "/data/data/com.termux/files/usr/etc/https-server/cert.pem"
			keyPath = "/data/data/com.termux/files/usr/etc/https-server/key.pem"
		}
	} else {
		certPath = "/etc/https-server/cert.pem"
		keyPath = "/etc/https-server/key.pem"
	}
	return certPath, keyPath
}

// GetCACertPath 返回 CA 证书路径
func GetCACertPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, "https-ca.crt")
}

// CheckCertificateExists 检查证书是否存在
func CheckCertificateExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ShowInstructions 显示安装证书说明
func ShowInstructions(caCertPath string) {
	fmt.Println()
	fmt.Println("📱 安卓证书安装步骤:")
	fmt.Println("1. 找到 CA 证书文件:", caCertPath)
	fmt.Println("2. 复制到手机存储")
	fmt.Println("3. 设置 → 安全 → 加密与凭据")
	fmt.Println("4. 安装证书 → CA证书")
	fmt.Println("5. 选择证书文件，命名为 'Local HTTPS'")
	fmt.Println()
	fmt.Println("🎮 启动服务器示例:")
	fmt.Println("  cd /path/to/website")
	fmt.Println("  https-server")
	fmt.Println()
	fmt.Println("🌟 愿代码如诗，生活如歌 ~")
}

// IsInTermux 检测是否在 Termux 环境中
func IsInTermux() bool {
	return os.Getenv("PREFIX") != "" && os.Getenv("TERMUX_VERSION") != ""
}