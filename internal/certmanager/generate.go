package certmanager

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

	"github.com/Alhkxsj/hserve/internal/i18n"
)

// Generate 生成证书
func Generate(force bool) error {
	certPath, keyPath := GetCertPaths()
	caCertPath := GetCACertPath()

	if !force && CheckCertificateExists(certPath) && CheckCertificateExists(caCertPath) {
		fmt.Println(i18n.T(i18n.GetLanguage(), "cert_exists"))
		ShowInstructions(caCertPath)
		return nil
	}

	// 确保证书目录存在
	certDir := filepath.Dir(certPath)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf(i18n.T(i18n.GetLanguage(), "cert_dir_failed"), err)
	}

	// 确保CA证书目录存在
	caCertDir := filepath.Dir(caCertPath)
	if certDir != caCertDir {
		if err := os.MkdirAll(caCertDir, 0755); err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "ca_cert_dir_failed"), err)
		}
	}

	// 生成CA密钥
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %v", err)
	}

	// 创建CA证书
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
		return fmt.Errorf("failed to create CA certificate: %v", err)
	}

	// 生成CA证书文件
	if err := writePem(caCertPath, "CERTIFICATE", caCertDER, 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %v", err)
	}

	// 生成服务器密钥
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate server key: %v", err)
	}

	// 创建服务器证书模板
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(30, 0, 0), // 服务器证书有效期30年
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"localhost", "127.0.0.1", "0.0.0.0"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1"), net.ParseIP("0.0.0.0")},
	}

	// 签名服务器证书
	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create server certificate: %v", err)
	}

	// 写入服务器证书和私钥
	if err := writePem(certPath, "CERTIFICATE", serverCertDER, 0644); err != nil {
		return fmt.Errorf("failed to write server certificate: %v", err)
	}
	if err := writePem(keyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey), 0600); err != nil {
		return fmt.Errorf("failed to write server key: %v", err)
	}

	fmt.Println(i18n.T(i18n.GetLanguage(), "cert_gen_success"))
	fmt.Printf("💡 %s\n", i18n.T(i18n.GetLanguage(), "cert_gen_tip"))
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
