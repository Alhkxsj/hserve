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

	// 检查证书是否存在
	if shouldSkipGeneration(force, certPath, caCertPath) {
		fmt.Println("✅ 证书已存在，无需重新生成")
		ShowInstructions(caCertPath)
		return nil
	}

	// 生成证书
	return generateAndSaveCertificates(certPath, keyPath, caCertPath)
}

// shouldSkipGeneration 检查是否应该跳过证书生成
func shouldSkipGeneration(force bool, certPath, caCertPath string) bool {
	return !force && CheckCertificateExists(certPath) && CheckCertificateExists(caCertPath)
}

// generateAndSaveCertificates 生成并保存证书
func generateAndSaveCertificates(certPath, keyPath, caCertPath string) error {
	if err := ensureCertDirectory(filepath.Dir(certPath)); err != nil {
		return err
	}

	certData, err := createCertificateData()
	if err != nil {
		return err
	}

	if err := saveCertificates(certData, certPath, keyPath, caCertPath); err != nil {
		return err
	}

	fmt.Println("✅ 证书生成完成")
	fmt.Println("💡 温馨提示: 请妥善保管您的证书文件")
	ShowInstructions(caCertPath)
	return nil
}

// ensureCertDirectory 确保证书目录存在
func ensureCertDirectory(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// certificateData 包含证书生成所需的所有数据
type certificateData struct {
	caKey       *rsa.PrivateKey
	serverKey   *rsa.PrivateKey
	caCertDER   []byte
	serverCertDER []byte
}

// createCertificateData 创建证书数据
func createCertificateData() (certificateData, error) {
	// 生成证书对
	caKey, serverKey, err := generateCertificateKeys()
	if err != nil {
		return certificateData{}, err
	}

	// 生成 CA 证书
	caCertDER, err := generateCACertificate(caKey)
	if err != nil {
		return certificateData{}, err
	}

	// 生成服务器证书
	serverCertDER, err := generateServerCertificate(caKey, serverKey)
	if err != nil {
		return certificateData{}, err
	}

	return certificateData{
		caKey:         caKey,
		serverKey:     serverKey,
		caCertDER:     caCertDER,
		serverCertDER: serverCertDER,
	}, nil
}

// saveCertificates 保存所有证书文件
func saveCertificates(data certificateData, certPath, keyPath, caCertPath string) error {
	// 保存 CA 证书
	if err := writePem(caCertPath, "CERTIFICATE", data.caCertDER, 0644); err != nil {
		return err
	}

	// 保存服务器证书和私钥
	if err := writePem(certPath, "CERTIFICATE", data.serverCertDER, 0644); err != nil {
		return err
	}
	if err := writePem(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(data.serverKey), 0600); err != nil {
		return err
	}

	return nil
}

// generateCertificateKeys 生成证书密钥对
func generateCertificateKeys() (*rsa.PrivateKey, *rsa.PrivateKey, error) {
	// 生成CA私钥
	caKey, err := generateCAKey()
	if err != nil {
		return nil, nil, err
	}

	// 生成服务器私钥
	serverKey, err := generateServerKey()
	if err != nil {
		return nil, nil, err
	}

	return caKey, serverKey, nil
}

// generateCAKey 生成CA密钥
func generateCAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// generateServerKey 生成服务器密钥
func generateServerKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// generateCACertificate 生成CA证书
func generateCACertificate(caKey *rsa.PrivateKey) ([]byte, error) {
	// 创建CA证书模板
	caTemplate := createCACertificateTemplate()

	// 生成CA证书
	caCertDER, err := createCertificate(&caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	return caCertDER, nil
}

// createCACertificateTemplate 创建CA证书模板
func createCACertificateTemplate() x509.Certificate {
	return x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "Local HTTPS CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // CA证书有效期10年
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
}

// createCertificate 创建证书
func createCertificate(template, parent *x509.Certificate, pub interface{}, priv interface{}) ([]byte, error) {
	return x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
}

// generateServerCertificate 生成服务器证书
func generateServerCertificate(caKey *rsa.PrivateKey, serverKey *rsa.PrivateKey) ([]byte, error) {
	// 创建服务器证书模板
	serverTemplate := createServerCertificateTemplate()

	// 从CA证书模板获取
	caTemplate := createCACertificateTemplate()

	// 生成服务器证书
	serverCertDER, err := createCertificate(&serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	return serverCertDER, nil
}

// createServerCertificateTemplate 创建服务器证书模板
func createServerCertificateTemplate() x509.Certificate {
	return x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(30, 0, 0),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"localhost", "127.0.0.1"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
}

// writePem 写入 PEM 文件
func writePem(path, typ string, data []byte, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: typ, Bytes: data}); err != nil {
		return err
	}

	return f.Close()
}

// GetCertPaths 返回证书和私钥路径
func GetCertPaths() (string, string) {
	var certPath, keyPath string
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			certPath = prefix + "/etc/hserve/cert.pem"
			keyPath = prefix + "/etc/hserve/key.pem"
		} else {
			certPath = "/data/data/com.termux/files/usr/etc/hserve/cert.pem"
			keyPath = "/data/data/com.termux/files/usr/etc/hserve/key.pem"
		}
	} else {
		// 在非 Termux 环境中，使用用户主目录
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		certPath = filepath.Join(homeDir, ".hserve", "cert.pem")
		keyPath = filepath.Join(homeDir, ".hserve", "key.pem")
	}
	return certPath, keyPath
}

// GetCACertPath 返回 CA 证书路径
func GetCACertPath() string {
	if IsInTermux() {
		// 在 Termux 环境中，使用 Termux 的 home 目录
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		return filepath.Join(home, "hserve-ca.crt")
	} else {
		// 在非 Termux 环境中，使用用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			home = "/tmp"
		}
		return filepath.Join(home, "hserve-ca.crt")
	}
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
	fmt.Println("5. 选择证书文件，命名为 'hserve'")
	fmt.Println()
	fmt.Println("💡 温馨提示: 使用 deb 包安装会自动为您生成证书")
	fmt.Println("🎮 启动服务器示例:")
	fmt.Println("  cd /path/to/website")
	fmt.Println("  hserve")
	fmt.Println()
	fmt.Println("🌟 愿代码如诗，生活如歌 ~")
}

// IsInTermux 检测是否在 Termux 环境中
func IsInTermux() bool {
	return os.Getenv("PREFIX") != "" && os.Getenv("TERMUX_VERSION") != ""
}
