// Package tls 包含 TLS 安全策略配置
package tls

import "crypto/tls"

// DefaultConfig 返回安全的 TLS 配置
func DefaultConfig(cert tls.Certificate) *tls.Config {
	// 创建并配置新的TLS配置
	config := createBaseTLSConfig(cert)
	
	// 设置密码套件
	config.CipherSuites = getSecureCipherSuites()
	
	// 设置曲线偏好
	config.CurvePreferences = getSecureCurvePreferences()
	
	return config
}

// createBaseTLSConfig 创建基础TLS配置
func createBaseTLSConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}
}

// getSecureCipherSuites 获取安全的密码套件
func getSecureCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
}

// getSecureCurvePreferences 获取安全的曲线偏好
func getSecureCurvePreferences() []tls.CurveID {
	return []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	}
}
