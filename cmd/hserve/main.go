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

// fatal 打印错误信息并退出程序
func fatal(msg string, err error) {
	printErrorMessage(msg, err)
	exitProgram()
}

// printErrorMessage 打印错误消息
func printErrorMessage(msg string, err error) {
	fmt.Fprintln(os.Stderr, "❌ 错误:", msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "   详情:", err.Error())
	}
}

// exitProgram 退出程序
func exitProgram() {
	os.Exit(1)
}

// main 程序主入口点
func main() {
	if len(os.Args) < 2 {
		// 没有参数时，运行服务器使用默认设置
		runServerWithArgs([]string{"-port", "8443", "-dir", "."})
		return
	}

	subCommand := os.Args[1]
	args := os.Args[2:]

	// 处理命令
	handleCommand(subCommand, args)
}

// handleCommand 处理命令行命令
func handleCommand(subCommand string, args []string) {
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
		// 处理默认情况（端口号或传递给服务器）
		handleDefaultCommand(subCommand, args)
	}
}

// handleDefaultCommand 处理默认命令情况
func handleDefaultCommand(subCommand string, args []string) {
	// 检查是否是端口号简写（如 hserve 4444）
	if isPortNumber(subCommand) {
		// 这是一个端口号，使用默认目录启动
		runServerWithArgs(append([]string{"-port", subCommand}, args...))
	} else {
		// 如果不是已知的子命令，则将所有参数传递给服务器运行
		runServerWithArgs(os.Args[1:])
	}
}

// isPortNumber 检查字符串是否为有效的端口号
func isPortNumber(s string) bool {
	port, err := strconv.Atoi(s)
	return err == nil && port > 0 && port < 65536
}

// showHelp 显示帮助信息
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

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("🌟 hserve v1.2.5")
	fmt.Println("👤 作者: 快手阿泠好困想睡觉")
	fmt.Println("🏠 项目地址: https://github.com/Alhkxsj/hserve")
	fmt.Println("✨ 愿代码如诗，生活如歌 ~")
}

func runServerWithArgs(args []string) {
	opts, err := parseServerOptions(args)
	if err != nil {
		fatal("解析服务器参数失败", err)
		return
	}

	if err := server.Run(opts); err != nil {
		fatal("启动 HTTPS 服务器失败", err)
	}
}

// parseServerOptions 解析服务器参数
func parseServerOptions(args []string) (server.Options, error) {
	// 解析命令行参数
	flags, err := parseServerFlags(args)
	if err != nil {
		return server.Options{}, err
	}

	// 处理特殊标志（help, version）
	if result, handled := handleSpecialFlags(flags); handled {
		return result, fmt.Errorf("special flag handled")
	}

	// 获取证书路径
	certPath, keyPath := getCertificatePaths()

	// 验证证书存在
	if err := validateCertificates(certPath); err != nil {
		return server.Options{}, err
	}

	// 确定服务器配置
	config, err := buildServerConfig(flags, certPath, keyPath)
	if err != nil {
		return server.Options{}, err
	}

	return config, nil
}

// handleSpecialFlags 处理特殊标志（help, version）
func handleSpecialFlags(flags serverFlags) (server.Options, bool) {
	if flags.help {
		showServerHelp()
		return server.Options{}, true
	}

	if flags.version {
		showVersion()
		return server.Options{}, true
	}

	return server.Options{}, false
}

// validateCertificates 验证证书是否存在
func validateCertificates(certPath string) error {
	if !certgen.CheckCertificateExists(certPath) {
		fmt.Println("⚠️  未检测到服务器证书")
		fmt.Println("请先运行：hserve cert")
		os.Exit(1)
	}
	return nil
}

// buildServerConfig 构建服务器配置
func buildServerConfig(flags serverFlags, certPath, keyPath string) (server.Options, error) {
	// 确定根目录
	root, err := determineRootDir(flags.dir, flags.nonFlagArgs)
	if err != nil {
		return server.Options{}, err
	}

	// 解析超时时间
	readTimeoutDuration, _ := time.ParseDuration(flags.readTimeout)
	writeTimeoutDuration, _ := time.ParseDuration(flags.writeTimeout)
	idleTimeoutDuration, _ := time.ParseDuration(flags.idleTimeout)

	return server.Options{
		Addr:           fmt.Sprintf(":%d", flags.port),
		Root:           root,
		Quiet:          flags.quiet,
		CertPath:       certPath,
		KeyPath:        keyPath,
		Paths:          flags.nonFlagArgs, // 传递要分享的特定路径
		ReadTimeout:    readTimeoutDuration,
		WriteTimeout:   writeTimeoutDuration,
		IdleTimeout:    idleTimeoutDuration,
		MaxHeaderBytes: flags.maxHeaderBytes,
		MaxBodyBytes:   flags.maxBodyBytes,
		AuthUser:       flags.authUser,
		AuthPass:       flags.authPass,
		AuthRealm:      flags.authRealm,
	}, nil
}

// getCertificatePaths 获取证书路径
func getCertificatePaths() (string, string) {
	return certgen.GetCertPaths()
}

// timeoutValues 超时值结构
type timeoutValues struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

// sizeLimits 大小限制结构
type sizeLimits struct {
	maxHeaderBytes int
	maxBodyBytes   int64
}

// getTimeoutsAndLimits 获取超时时间和大小限制
func getTimeoutsAndLimits(flags serverFlags) (timeoutValues, sizeLimits) {
	// 解析超时时间
	readTimeoutDuration, _ := time.ParseDuration(flags.readTimeout)
	writeTimeoutDuration, _ := time.ParseDuration(flags.writeTimeout)
	idleTimeoutDuration, _ := time.ParseDuration(flags.idleTimeout)

	// 返回超时值和大小限制
	return timeoutValues{
			readTimeout:  readTimeoutDuration,
			writeTimeout: writeTimeoutDuration,
			idleTimeout:  idleTimeoutDuration,
		}, sizeLimits{
			maxHeaderBytes: flags.maxHeaderBytes,
			maxBodyBytes:   flags.maxBodyBytes,
		}
}

// serverFlags 定义服务器选项的结构
type serverFlags struct {
	port           int
	dir            string
	quiet          bool
	version        bool
	help           bool
	readTimeout    string
	writeTimeout   string
	idleTimeout    string
	maxHeaderBytes int
	maxBodyBytes   int64
	authUser       string
	authPass       string
	authRealm      string
	nonFlagArgs    []string
}

// parseServerFlags 解析服务器命令行参数
func parseServerFlags(args []string) (serverFlags, error) {
	// 创建新的 FlagSet 来解析参数，避免与全局 flag.CommandLine 冲突
	fs := flag.NewFlagSet("server", flag.ExitOnError)

	// 定义所有标志
	flags := defineFlags(fs)

	// 解析传入的参数
	if err := fs.Parse(args); err != nil {
		return serverFlags{}, err
	}

	// 返回解析后的标志值
	return serverFlags{
		port:           *flags.port,
		dir:            *flags.dir,
		quiet:          *flags.quiet,
		version:        *flags.version,
		help:           *flags.help,
		readTimeout:    *flags.readTimeout,
		writeTimeout:   *flags.writeTimeout,
		idleTimeout:    *flags.idleTimeout,
		maxHeaderBytes: *flags.maxHeaderBytes,
		maxBodyBytes:   *flags.maxBodyBytes,
		authUser:       *flags.authUser,
		authPass:       *flags.authPass,
		authRealm:      *flags.authRealm,
		nonFlagArgs:    fs.Args(),
	}, nil
}

// flagPointers 存储所有标志的指针
type flagPointers struct {
	port, maxHeaderBytes *int
	dir, readTimeout, writeTimeout, idleTimeout, authUser, authPass, authRealm *string
	quiet, version, help *bool
	maxBodyBytes *int64
}

// defineFlags 定义命令行标志
func defineFlags(fs *flag.FlagSet) flagPointers {
	return flagPointers{
		port:           fs.Int("port", 8443, "监听端口（默认 8443）"),
		dir:            fs.String("dir", "", "共享目录"),
		quiet:          fs.Bool("quiet", false, "安静模式（不输出访问日志）"),
		version:        fs.Bool("version", false, "显示版本信息"),
		help:           fs.Bool("help", false, "显示此帮助信息"),
		readTimeout:    fs.String("read-timeout", "30s", "请求读取超时时间（默认 30s）"),
		writeTimeout:   fs.String("write-timeout", "30s", "响应写入超时时间（默认 30s）"),
		idleTimeout:    fs.String("idle-timeout", "120s", "连接空闲超时时间（默认 120s）"),
		maxHeaderBytes: fs.Int("max-header-bytes", 1048576, "最大请求头大小（字节，默认 1MB）"),
		maxBodyBytes:   fs.Int64("max-body-bytes", 10<<20, "最大请求体大小（字节，默认 10MB）"),
		authUser:       fs.String("auth-user", "", "基本身份验证用户名"),
		authPass:       fs.String("auth-pass", "", "基本身份验证密码"),
		authRealm:      fs.String("auth-realm", "hserve-secure-area", "身份验证领域"),
	}
}

// determineRootDir 确定服务器根目录
func determineRootDir(dir string, nonFlagArgs []string) (string, error) {
	var root string

	if dir != "" {
		// 如果指定了 -dir 参数，则使用该目录
		var err error
		root, err = filepath.Abs(dir)
		if err != nil {
			return "", err
		}
	} else if len(nonFlagArgs) > 0 {
		// 如果有非标志参数，使用当前目录作为根目录
		// 但会限制只访问指定的文件/目录
		var err error
		root, err = filepath.Abs(".")
		if err != nil {
			return "", err
		}
	} else {
		// 没有指定目录或文件，使用当前目录
		var err error
		root, err = filepath.Abs(".")
		if err != nil {
			return "", err
		}
	}

	return root, nil
}

// showServerHelp 显示服务器帮助信息
func showServerHelp() {
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
	fmt.Println("      身份验证领域（默认 \"hserve-secure-area\"")
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
}

// runCertGen 执行证书生成命令
func runCertGen(args []string) {
	// 解析参数
	opts, err := parseCertGenOptions(args)
	if err != nil {
		fatal("解析证书生成参数失败", err)
		return
	}

	// 处理特殊选项
	if handled := handleCertGenSpecialOptions(opts); handled {
		return
	}

	// 执行证书生成
	executeCertGeneration(opts)
}

// certGenOptions 证书生成选项
type certGenOptions struct {
	force   bool
	version bool
	help    bool
	args    []string
}

// parseCertGenOptions 解析证书生成选项
func parseCertGenOptions(args []string) (certGenOptions, error) {
	// 创建新的 FlagSet 来解析参数，避免与全局 flag.CommandLine 冲突
	certFlags := flag.NewFlagSet("certgen", flag.ExitOnError)

	force := certFlags.Bool("force", false, "强制重新生成证书")
	version := certFlags.Bool("version", false, "显示版本信息")
	help := certFlags.Bool("help", false, "显示此帮助信息")

	// 解析传入的参数
	if err := certFlags.Parse(args); err != nil {
		return certGenOptions{}, err
	}

	return certGenOptions{
		force:   *force,
		version: *version,
		help:    *help,
		args:    certFlags.Args(),
	}, nil
}

// handleCertGenSpecialOptions 处理证书生成特殊选项
func handleCertGenSpecialOptions(opts certGenOptions) bool {
	if opts.help {
		showCertHelp()
		return true
	}

	if opts.version {
		showVersion()
		return true
	}

	return false
}

// executeCertGeneration 执行证书生成
func executeCertGeneration(opts certGenOptions) {
	fmt.Println("🔐 HTTPS 证书生成工具 - 为您的安全访问保驾护航")
	fmt.Println("🌟 正在为您生成安全证书，请稍候...")

	if err := certgen.Generate(opts.force); err != nil {
		fatal("证书生成失败", err)
	}

	fmt.Println("================================")
}

// showCertHelp 显示证书生成帮助信息
func showCertHelp() {
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
}
