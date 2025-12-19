package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Alhkxsj/hserve/internal/app/hserve"
	"github.com/Alhkxsj/hserve/pkg/certgen"
)

func fatal(msg string, err error) {
	fmt.Fprintln(os.Stderr, "❌ 错误:", msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "   详情:", err.Error())
	}
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		// 没有参数时，运行服务器使用默认设置
		runServerWithArgs([]string{"-port", "8443", "-dir", "."})
		return
	}

	subCommand := os.Args[1]
	args := os.Args[2:]

	switch strings.ToLower(subCommand) {
	case "serve":
		runServerWithArgs(args)
	case "cert":
		runCertGen(args)
	case "version", "-version", "--version":
		showVersion()
	case "help", "-help", "--help", "-h":
		showHelp()
	default:
		// 检查是否是端口号简写（如 hserve 4444）
		if port, err := strconv.Atoi(subCommand); err == nil && port > 0 && port < 65536 {
			// 这是一个端口号，使用默认目录启动
			runServerWithArgs(append([]string{"-port", subCommand}, args...))
		} else {
			// 如果不是已知的子命令，则将所有参数传递给服务器运行
			runServerWithArgs(os.Args[1:])
		}
	}
}

func showHelp() {
	fmt.Println("🚀 HTTPS 文件服务器 - 让文件分享变得简单而安全")
	fmt.Println()
	fmt.Println("📖 使用方法:")
	fmt.Printf("  hserve [选项] [路径...]")
	fmt.Println()
	fmt.Println("✨ 可用选项:")
	fmt.Println("  -port int")
	fmt.Println("      监听端口（默认 8443）")
	fmt.Println("  -dir string")
	fmt.Println("      共享目录")
	fmt.Println("  -quiet")
	fmt.Println("      安静模式（不输出访问日志）")
	fmt.Println()
	fmt.Println("💡 使用示例:")
	fmt.Println("  hserve                    # 在当前目录启动服务器")
	fmt.Println("  hserve 4444               # 在 4444 端口启动服务器")
	fmt.Println("  hserve -port 9999         # 在 9999 端口启动服务器")
	fmt.Println("  hserve /path/to/dir       # 分享指定目录")
	fmt.Println("  hserve /path/to/file.txt  # 分享单个文件")
	fmt.Println("  hserve /file1 /file2      # 分享多个文件")
	fmt.Println("  hserve /dir1 /dir2        # 分享多个目录")
	fmt.Println("  hserve -port 9999 /path/to/files")
	fmt.Println()
	fmt.Println("🌟 愿代码如诗，生活如歌 ~")
}

func showVersion() {
	fmt.Println("🌟 hserve v1.2.5")
	fmt.Println("👤 作者: 快手阿泠 (Alexa Haley)")
	fmt.Println("🏠 项目地址: https://github.com/Alhkxsj/hserve")
	fmt.Println("✨ 愿代码如诗，生活如歌 ~")
}

func runServerWithArgs(args []string) {
	// 创建新的 FlagSet 来解析参数，避免与全局 flag.CommandLine 冲突
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)

	port := serverFlags.Int("port", 8443, "监听端口（默认 8443）")
	dir := serverFlags.String("dir", "", "共享目录")
	quiet := serverFlags.Bool("quiet", false, "安静模式（不输出访问日志）")
	version := serverFlags.Bool("version", false, "显示版本信息")
	help := serverFlags.Bool("help", false, "显示此帮助信息")

	// 网络相关的高级选项
	readTimeout := serverFlags.String("read-timeout", "30s", "请求读取超时时间（默认 30s）")
	writeTimeout := serverFlags.String("write-timeout", "30s", "响应写入超时时间（默认 30s）")
	idleTimeout := serverFlags.String("idle-timeout", "120s", "连接空闲超时时间（默认 120s）")
	maxHeaderBytes := serverFlags.Int("max-header-bytes", 1048576, "最大请求头大小（字节，默认 1MB）")
	maxBodyBytes := serverFlags.Int64("max-body-bytes", 10<<20, "最大请求体大小（字节，默认 10MB）")

	// 身份验证选项
	authUser := serverFlags.String("auth-user", "", "基本身份验证用户名")
	authPass := serverFlags.String("auth-pass", "", "基本身份验证密码")
	authRealm := serverFlags.String("auth-realm", "hserve-secure-area", "身份验证领域")

	// 解析传入的参数
	if err := serverFlags.Parse(args); err != nil {
		fatal("解析服务器参数失败", err)
	}

	if *help {
		fmt.Println("📖 hserve - 启动 HTTPS 文件服务器")
		fmt.Println()
		fmt.Println("✨ 可用选项:")
		fmt.Println("  -port int")
		fmt.Println("      监听端口（默认 8443）")
		fmt.Println("  -dir string")
		fmt.Println("      共享目录")
		fmt.Println("  -quiet")
		fmt.Println("      安静模式（不输出访问日志）")
		fmt.Println("  -read-timeout string")
		fmt.Println("      请求读取超时时间（默认 30s）")
		fmt.Println("  -write-timeout string")
		fmt.Println("      响应写入超时时间（默认 30s）")
		fmt.Println("  -idle-timeout string")
		fmt.Println("      连接空闲超时时间（默认 120s）")
		fmt.Println("  -max-header-bytes int")
		fmt.Println("      最大请求头大小（字节，默认 1MB）")
		fmt.Println("  -max-body-bytes int64")
		fmt.Println("      最大请求体大小（字节，默认 10MB）")
		fmt.Println("  -auth-user string")
		fmt.Println("      基本身份验证用户名")
		fmt.Println("  -auth-pass string")
		fmt.Println("      基本身份验证密码")
		fmt.Println("  -auth-realm string")
		fmt.Println("      身份验证领域（默认 \"hserve-secure-area\"）")
		fmt.Println("  -version")
		fmt.Println("      显示版本信息")
		fmt.Println("  -help")
		fmt.Println("      显示此帮助信息")
		fmt.Println()
		fmt.Println("💡 使用示例:")
		fmt.Println("  hserve                    # 在当前目录启动服务器")
		fmt.Println("  hserve 4444               # 在 4444 端口启动服务器")
		fmt.Println("  hserve /path/to/dir       # 分享指定目录")
		fmt.Println("  hserve /path/to/file.txt  # 分享单个文件")
		fmt.Println("  hserve /file1 /file2      # 分享多个文件")
		fmt.Println("  hserve -port 9999 -read-timeout 60s -max-body-bytes 20971520 -dir /path/to/files")
		fmt.Println("  hserve -auth-user admin -auth-pass 123456 /path/to/secure/dir")
		return
	}

	if *version {
		showVersion()
		return
	}

	var root string

	// 检查是否有非标志参数（即文件/目录路径）
	nonFlagArgs := serverFlags.Args()

	if *dir != "" {
		// 如果指定了 -dir 参数，则使用该目录
		var err error
		root, err = filepath.Abs(*dir)
		if err != nil {
			fatal("获取目录路径失败", err)
		}
	} else if len(nonFlagArgs) > 0 {
		// 如果有非标志参数，使用当前目录作为根目录
		// 但会限制只访问指定的文件/目录
		var err error
		root, err = filepath.Abs(".")
		if err != nil {
			fatal("获取当前目录路径失败", err)
		}
	} else {
		// 没有指定目录或文件，使用当前目录
		var err error
		root, err = filepath.Abs(".")
		if err != nil {
			fatal("获取当前目录路径失败", err)
		}
	}

	// 解析超时时间
	readTimeoutDuration, err := time.ParseDuration(*readTimeout)
	if err != nil {
		fatal("无效的读取超时时间", err)
	}

	writeTimeoutDuration, err := time.ParseDuration(*writeTimeout)
	if err != nil {
		fatal("无效的写入超时时间", err)
	}

	idleTimeoutDuration, err := time.ParseDuration(*idleTimeout)
	if err != nil {
		fatal("无效的空闲超时时间", err)
	}

	certPath, keyPath := certgen.GetCertPaths()
	if !certgen.CheckCertificateExists(certPath) {
		fmt.Println("⚠️  未检测到服务器证书")
		fmt.Println("请先运行：hserve cert")
		os.Exit(1)
	}

	opts := server.Options{
		Addr:           fmt.Sprintf(":%d", *port),
		Root:           root,
		Quiet:          *quiet,
		CertPath:       certPath,
		KeyPath:        keyPath,
		Paths:          nonFlagArgs, // 传递要分享的特定路径
		ReadTimeout:    readTimeoutDuration,
		WriteTimeout:   writeTimeoutDuration,
		IdleTimeout:    idleTimeoutDuration,
		MaxHeaderBytes: *maxHeaderBytes,
		MaxBodyBytes:   *maxBodyBytes,
		AuthUser:       *authUser,
		AuthPass:       *authPass,
		AuthRealm:      *authRealm,
	}

	if err := server.Run(opts); err != nil {
		fatal("启动 HTTPS 服务器失败", err)
	}
}

func runCertGen(args []string) {
	// 创建新的 FlagSet 来解析参数，避免与全局 flag.CommandLine 冲突
	certFlags := flag.NewFlagSet("certgen", flag.ExitOnError)

	force := certFlags.Bool("force", false, "强制重新生成证书")
	version := certFlags.Bool("version", false, "显示版本信息")
	help := certFlags.Bool("help", false, "显示此帮助信息")

	// 解析传入的参数
	if err := certFlags.Parse(args); err != nil {
		fatal("解析证书生成参数失败", err)
	}

	if *help {
		fmt.Println("🔐 hserve cert - 生成 HTTPS 证书")
		fmt.Println()
		fmt.Println("✨ 可用选项:")
		fmt.Println("  -force")
		fmt.Println("      强制重新生成证书")
		fmt.Println("  -version")
		fmt.Println("      显示版本信息")
		fmt.Println("  -help")
		fmt.Println("      显示此帮助信息")
		fmt.Println()
		fmt.Println("💡 使用示例:")
		fmt.Println("  hserve cert")
		fmt.Println("  hserve cert -force")
		return
	}

	if *version {
		showVersion()
		return
	}

	fmt.Println("🔐 HTTPS 证书生成工具 - 为您的安全访问保驾护航")
	fmt.Println("🌟 正在为您生成安全证书，请稍候...")

	if err := certgen.Generate(*force); err != nil {
		fatal("证书生成失败", err)
	}

	fmt.Println("================================")
}
