package server

import (
	"crypto/tls"
	"fmt"

	tlspolicy "github.com/Alhkxsj/hserve/internal/tls"
)

// LoadTLSConfig 加载并返回 TLS 配置
func LoadTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	// 加载证书
	cert, err := loadCertificate(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	
	// 创建TLS配置
	tlsConfig := createTLSConfig(cert)
	return tlsConfig, nil
}

// loadCertificate 加载TLS证书
func loadCertificate(certPath, keyPath string) (tls.Certificate, error) {
	// 使用系统函数加载证书和密钥
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		// 包装错误以便更好地调试
		return cert, fmt.Errorf("加载 TLS 证书失败: %w", err)
	}
	return cert, nil
}

// createTLSConfig 创建TLS配置
func createTLSConfig(cert tls.Certificate) *tls.Config {
	// 使用默认的安全策略创建TLS配置
	return tlspolicy.DefaultConfig(cert)
}
