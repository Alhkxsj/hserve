package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Alhkxsj/https-server/pkg/certgen"
)

func fatal(msg string, err error) {
	fmt.Println("❌ 错误:", msg)
	if err != nil {
		fmt.Println("   详情:", err.Error())
	}
	os.Exit(1)
}

func main() {
	flag.Usage = func() {
		fmt.Println("🔐 HTTPS 证书生成工具 - 为您的安全访问保驾护航")
		fmt.Println()
		fmt.Println("📖 使用方法:")
		fmt.Printf("  %s [选项]\n", "https-certgen")
		fmt.Println()
		fmt.Println("✨ 可用选项:")
		fmt.Println("  -force")
		fmt.Println("      强制重新生成证书")
		fmt.Println("  -version")
		fmt.Println("      显示版本信息")
		fmt.Println("  -help")
		fmt.Println("      显示此帮助信息")
		fmt.Println()
		fmt.Println("💡 小贴士: 生成的证书用于 https-server 工具的 HTTPS 连接哦~")
		fmt.Println("🌟 愿代码如诗，生活如歌 ~")
	}
	
	force := flag.Bool("force", false, "强制重新生成证书")
	version := flag.Bool("version", false, "显示版本信息")
	help := flag.Bool("help", false, "显示此帮助信息")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	if *version {
		fmt.Println("🔐 HTTPS 证书生成工具 v1.2.1")
		fmt.Println("👤 作者: 快手阿泠 (Alexa Haley)")
		fmt.Println("🏠 项目地址: https://github.com/Alhkxsj/https-server")
		fmt.Println("✨ 愿代码如诗，生活如歌 ~")
		return
	}

	fmt.Println("🔐 HTTPS 证书生成工具 - 为您的安全访问保驾护航")
	fmt.Println("🌟 正在为您生成安全证书，请稍候...")

	if err := certgen.Generate(*force); err != nil {
		fatal("证书生成失败", err)
	}

	fmt.Println("================================")
}