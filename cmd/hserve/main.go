package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		// 如果不是已知的子命令，则将所有参数传递给服务器运行
		// 这样用户可以直接使用 'hserve -port 9999' 这样的命令
		runServerWithArgs(os.Args[1:])
	}
}

func showHelp() {
	fmt.Println("🚀 HTTPS 文件服务器 - 让文件分享变得简单而安全")
	fmt.Println()
	fmt.Println("📖 使用方法:")
	fmt.Printf("  hserve [命令] [选项]\n")
	fmt.Println()
	fmt.Println("✨ 可用命令:")
	fmt.Println("  serve            启动 HTTPS 文件服务器（默认）")
	fmt.Println("  cert             生成证书")
	fmt.Println("  version          显示版本信息")
	fmt.Println("  help             显示此帮助信息")
	fmt.Println()
	fmt.Println("💡 小贴士: 首次使用前请先生成证书 'hserve cert'")
	fmt.Println("🌟 愿代码如诗，生活如歌 ~")
}

func showVersion() {
	fmt.Println("🌟 hserve v1.2.4")
	fmt.Println("👤 作者: 快手阿泠 (Alexa Haley)")
	fmt.Println("🏠 项目地址: https://github.com/Alhkxsj/hserve")
	fmt.Println("✨ 愿代码如诗，生活如歌 ~")
}

func runServer() {
	// 为默认运行模式设置默认参数
	defaultArgs := []string{"-port", "8443", "-dir", "."}
	runServerWithArgs(defaultArgs)
}

func runServerWithArgs(args []string) {
	// 创建新的 FlagSet 来解析参数，避免与全局 flag.CommandLine 冲突
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)

	port := serverFlags.Int("port", 8443, "监听端口（默认 8443）")
	dir := serverFlags.String("dir", ".", "共享目录（默认当前目录）")
	quiet := serverFlags.Bool("quiet", false, "安静模式（不输出访问日志）")
	version := serverFlags.Bool("version", false, "显示版本信息")
	help := serverFlags.Bool("help", false, "显示此帮助信息")

	// 解析传入的参数
	if err := serverFlags.Parse(args); err != nil {
		fatal("解析服务器参数失败", err)
	}

	if *help {
		fmt.Println("📖 hserve serve - 启动 HTTPS 文件服务器")
		fmt.Println()
		fmt.Println("✨ 可用选项:")
		fmt.Println("  -port int")
		fmt.Println("      监听端口（默认 8443）")
		fmt.Println("  -dir string")
		fmt.Println("      共享目录（默认当前目录）")
		fmt.Println("  -quiet")
		fmt.Println("      安静模式（不输出访问日志）")
		fmt.Println("  -version")
		fmt.Println("      显示版本信息")
		fmt.Println("  -help")
		fmt.Println("      显示此帮助信息")
		fmt.Println()
		fmt.Println("💡 使用示例:")
		fmt.Println("  hserve serve -dir=/path/to/files -port=9443")
		return
	}

	if *version {
		showVersion()
		return
	}

	root, err := filepath.Abs(*dir)
	if err != nil {
		fatal("获取目录路径失败", err)
	}

	certPath, keyPath := certgen.GetCertPaths()
	if !certgen.CheckCertificateExists(certPath) {
		fmt.Println("⚠️  未检测到服务器证书")
		fmt.Println("请先运行：hserve cert")
		os.Exit(1)
	}

	opts := server.Options{
		Addr:     fmt.Sprintf(":%d", *port),
		Root:     root,
		Quiet:    *quiet,
		CertPath: certPath,
		KeyPath:  keyPath,
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
