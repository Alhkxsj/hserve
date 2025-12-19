package server

import (
	"fmt"
	"net"
	"os"
)

// checkPort 检测端口是否可用
func checkPort(addr string) error {
	ln, err := createListener(addr)
	if err != nil {
		return formatPortError(addr, err)
	}
	
	cleanupListener(ln)
	return nil
}

// createListener 创建TCP监听器
func createListener(addr string) (net.Listener, error) {
	return net.Listen("tcp", addr)
}

// formatPortError 格式化端口错误信息
func formatPortError(addr string, err error) error {
	return fmt.Errorf("端口 %s 无法监听，可能已被占用", addr)
}

// cleanupListener 清理监听器
func cleanupListener(ln net.Listener) {
	_ = ln.Close()
}

// PreflightCheck 运行前环境自检
func PreflightCheck(addr, certPath, keyPath string) error {
	// 检查证书和私钥文件
	if err := checkCertificateFiles(certPath, keyPath); err != nil {
		return err
	}

	// 检查端口可用性
	if err := checkPort(addr); err != nil {
		return err
	}

	return nil
}

// checkCertificateFiles 检查证书和私钥文件是否存在
func checkCertificateFiles(certPath, keyPath string) error {
	if _, err := os.Stat(certPath); err != nil {
		return fmt.Errorf("未找到证书文件：%s\n请先运行 hserve cert 生成证书", certPath)
	}

	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("未找到私钥文件：%s\n请先运行 hserve cert 生成证书", keyPath)
	}

	return nil
}
