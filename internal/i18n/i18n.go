package i18n

import (
	"os"
)

// 语言类型
type LangType string

const (
	ZH LangType = "zh"
	EN LangType = "en"
)

// 全局语言变量
var currentLang LangType = EN // 默认英文

// 获取当前语言环境
func GetLanguage() LangType {
	return currentLang
}

// 设置语言
func SetLanguage(lang LangType) {
	currentLang = lang
}

// 获取系统语言环境
func GetSystemLanguage() LangType {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}

	// 默认英文
	if lang != "" && (lang[:2] == "zh" || lang[:2] == "zn") {
		return ZH
	}
	return EN
}

// 翻译函数
func T(lang LangType, key string) string {
	switch key {
	case "https_server_title":
		if lang == EN {
			return "HTTPS File Server - Making file sharing simple and secure"
		}
		return "HTTPS 文件服务器 - 让文件分享变得简单而安全"
	case "usage":
		if lang == EN {
			return "📖 Usage:"
		}
		return "📖 使用方法:"
	case "available_options":
		if lang == EN {
			return "✨ Available Options:"
		}
		return "✨ 可用选项:"
	case "port_desc":
		if lang == EN {
			return "Listening port (default 8443)"
		}
		return "监听端口（默认 8443）"
	case "dir_desc":
		if lang == EN {
			return "Shared directory (default current directory)"
		}
		return "共享目录（默认当前目录）"
	case "quiet_desc":
		if lang == EN {
			return "Quiet mode (no access logs)"
		}
		return "安静模式（不输出访问日志）"
	case "help_desc":
		if lang == EN {
			return "Show help information"
		}
		return "显示此帮助信息"
	case "version_desc":
		if lang == EN {
			return "Show version information"
		}
		return "显示版本信息"
	case "gen_cert_desc":
		if lang == EN {
			return "Generate HTTPS certificates"
		}
		return "生成HTTPS证书"
	case "force_desc":
		if lang == EN {
			return "Force re-generate certificates"
		}
		return "强制重新生成证书"
	case "install_ca_desc":
		if lang == EN {
			return "Install CA certificate to Termux trust store"
		}
		return "将CA证书部署到Termux信任库"
	case "export_ca_desc":
		if lang == EN {
			return "Export CA certificate for manual installation"
		}
		return "导出CA证书到指定目录"
	case "serve_desc":
		if lang == EN {
			return "Start HTTPS file server"
		}
		return "启动HTTPS文件服务器"
	case "auto_gen_desc":
		if lang == EN {
			return "Automatically generate certificates for first run"
		}
		return "自动为首次运行生成证书"
	case "allow_desc":
		if lang == EN {
			return "Allowed directory paths (can be specified multiple times)"
		}
		return "允许访问的目录路径（可多次指定）"
	case "tls_cert_file_desc":
		if lang == EN {
			return "External TLS certificate file path"
		}
		return "外部TLS证书文件路径"
	case "tls_key_file_desc":
		if lang == EN {
			return "External TLS private key file path"
		}
		return "外部TLS私钥文件路径"
	case "tip_cert_first":
		if lang == EN {
			return "💡 Tip: Run 'hserve gen-cert' first to generate certificates"
		}
		return "💡 小贴士: 首次使用前请运行 'hserve gen-cert' 生成证书哦~"
	case "tip_external_cert":
		if lang == EN {
			return "💡 Tip: The certificates are used for hserve tool's HTTPS connection"
		}
		return "💡 小贴士: 生成的证书用于 hserve 工具的 HTTPS 连接哦~"
	case "android_install_steps":
		if lang == EN {
			return "📱 Android Certificate Installation Steps:"
		}
		return "📱 安卓证书安装步骤:"
	case "android_install_step1":
		if lang == EN {
			return "1. Open Settings"
		}
		return "1. 打开 设置"
	case "android_install_step2":
		if lang == EN {
			return "2. Security → Encryption & credentials"
		}
		return "2. 安全 → 加密与凭据"
	case "android_install_step3":
		if lang == EN {
			return "3. Install certificates → CA certificates"
		}
		return "3. 安装证书 → CA证书"
	case "android_install_step4":
		if lang == EN {
			return "4. Select the hserve-ca.crt file"
		}
		return "4. 选择 hserve-ca.crt 文件"
	case "android_install_step5":
		if lang == EN {
			return "5. Name the certificate (e.g., hserve CA)"
		}
		return "5. 命名证书（例如：hserve CA）"
	case "launch_example":
		if lang == EN {
			return "🎮 Launch server example:"
		}
		return "🎮 启动服务器示例:"
	case "poem":
		if lang == EN {
			return "🌟 May code be like poetry, life be like a song ~"
		}
		return "🌟 愿代码如诗，生活如歌 ~"
	case "cert_exists":
		if lang == EN {
			return "✅ Certificates already exist, no need to regenerate"
		}
		return "✅ 证书已存在，无需重新生成"
	case "cert_gen_success":
		if lang == EN {
			return "✅ Certificate generation completed"
		}
		return "✅ 证书生成完成"
	case "cert_gen_tip":
		if lang == EN {
			return "💡 Tip: Please keep your certificate files safe"
		}
		return "💡 温馨提示: 请妥善保管您的证书文件"
	case "server_started":
		if lang == EN {
			return "🚀 HTTPS server started"
		}
		return "🚀 HTTPS 服务器已启动"
	case "shared_dir":
		if lang == EN {
			return "📁 Shared directory:"
		}
		return "📁 共享目录:"
	case "access_whitelist":
		if lang == EN {
			return "✅ Access whitelist:"
		}
		return "✅ 访问白名单:"
	case "access_address":
		if lang == EN {
			return "🌐 Access address:"
		}
		return "🌐 访问地址:"
	case "listen_address":
		if lang == EN {
			return "🔐 Listen address:"
		}
		return "🔐 监听地址:"
	case "tip_open_browser":
		if lang == EN {
			return "💡 Tip: Open the access address in your browser to browse files"
		}
		return "💡 提示: 在浏览器中打开访问地址即可浏览文件"
	case "tip_stop_server":
		if lang == EN {
			return "🛑 Press Ctrl+C to stop"
		}
		return "🛑 按 Ctrl+C 停止"
	case "ca_installed_success":
		if lang == EN {
			return "✅ CA certificate has been deployed to Termux trust store"
		}
		return "✅ CA证书已成功部署到Termux信任库"
	case "export_ca_success":
		if lang == EN {
			return "✅ CA certificate exported to:"
		}
		return "✅ CA证书已导出到:"
	case "cert_not_found":
		if lang == EN {
			return "⚠️  Server certificate not detected"
		}
		return "⚠️  未检测到服务器证书"
	case "run_gen_cert":
		if lang == EN {
			return "Please run: hserve gen-cert"
		}
		return "请先运行：hserve gen-cert"
	case "auto_gen_tip":
		if lang == EN {
			return "Or use --auto-gen flag to automatically generate certificates for you"
		}
		return "或者使用 --auto-gen 标志自动为您生成证书"
	case "cert_gen_auto":
		if lang == EN {
			return "⚠️  Server certificate not detected, automatically generating for you..."
		}
		return "⚠️  未检测到服务器证书，正在自动为您生成..."
	case "ca_installed_auto":
		if lang == EN {
			return "✅ CA certificate automatically installed to Termux trust store"
		}
		return "✅ CA证书已自动安装到Termux信任库"
	case "termux_only":
		if lang == EN {
			return "⚠️  This command is only available in Termux environment"
		}
		return "⚠️  此命令仅在Termux环境中可用"
	case "ca_not_found":
		if lang == EN {
			return "⚠️  CA certificate not detected"
		}
		return "⚠️  未检测到CA证书"
	case "path_not_allowed":
		if lang == EN {
			return "Directory %s is not in the access whitelist"
		}
		return "目录 %s 不在访问白名单中"
	case "forbidden_access":
		if lang == EN {
			return "403 Forbidden - Access path not in whitelist"
		}
		return "403 Forbidden - 访问路径不在白名单中"
	case "cert_dir_failed":
		if lang == EN {
			return "❌ Create certificate directory failed: %s"
		}
		return "❌ 创建证书目录失败: %s"
	case "ca_cert_dir_failed":
		if lang == EN {
			return "❌ Create CA certificate directory failed: %s"
		}
		return "❌ 创建CA证书目录失败: %s"
	case "cert_gen_failed":
		if lang == EN {
			return "❌ Certificate generation failed: %s"
		}
		return "❌ 证书生成失败: %s"
	case "server_start_failed":
		if lang == EN {
			return "❌ Start HTTPS server failed: %s"
		}
		return "❌ 启动 HTTPS 服务器失败: %s"
	case "get_path_failed":
		if lang == EN {
			return "❌ Get directory path failed: %s"
		}
		return "❌ 获取目录路径失败: %s"
	case "cert_auto_gen_failed":
		if lang == EN {
			return "❌ Certificate auto-generation failed: %s"
		}
		return "❌ 证书自动生成失败: %s"
	case "termux_cert_dir_failed":
		if lang == EN {
			return "⚠️  Create Termux certificate directory failed: %s"
		}
		return "⚠️  创建Termux证书目录失败: %s"
	case "install_ca_failed":
		if lang == EN {
			return "⚠️  Install CA certificate to Termux trust store failed: %s"
		}
		return "⚠️  安装CA证书到Termux信任库失败: %s"
	case "copy_file_failed":
		if lang == EN {
			return "❌ Copy file failed: %s"
		}
		return "❌ 复制文件失败: %s"
	case "export_ca_failed":
		if lang == EN {
			return "❌ Export CA certificate failed: %s"
		}
		return "❌ 导出CA证书失败: %s"
	case "cert_file_not_exists":
		if lang == EN {
			return "Certificate file does not exist: %s"
		}
		return "证书文件不存在: %s"
	case "key_file_not_exists":
		if lang == EN {
			return "Private key file does not exist: %s"
		}
		return "私钥文件不存在: %s"
	case "tls_config_failed":
		if lang == EN {
			return "Load TLS configuration failed: %s"
		}
		return "加载TLS配置失败: %s"
	case "user_error":
		if lang == EN {
			return "❌ Error:"
		}
		return "❌ 错误:"
	case "cert_exists_tip":
		if lang == EN {
			return "Please run hserve gen-cert to generate certificates first"
		}
		return "请先运行 hserve gen-cert 生成证书"
	case "hserve_desc":
		if lang == EN {
			return "A quick setup local HTTPS server tool"
		}
		return "一个快速搭建本地HTTPS服务器的工具"
	case "hserve_long_desc":
		if lang == EN {
			return "hserve is a zero-configuration HTTPS static file server designed specifically for the Termux environment."
		}
		return "hserve 是一个专为Termux环境设计的零配置HTTPS静态文件服务器。"
	case "serve_long_desc":
		if lang == EN {
			return "Start HTTPS file server to provide secure file sharing service"
		}
		return "启动HTTPS文件服务器，提供安全的文件共享服务"
	case "gen_cert_long_desc":
		if lang == EN {
			return "Generate self-signed CA and server certificates"
		}
		return "生成自签名CA证书和服务器证书"
	case "install_ca_long_desc":
		if lang == EN {
			return "Copy CA certificate to Termux's trust store to make it trusted by internal Termux tools"
		}
		return "将CA证书复制到Termux的证书目录，使其在Termux内部工具中受信任"
	case "export_ca_long_desc":
		if lang == EN {
			return "Copy CA certificate to specified directory for manual installation to Android system"
		}
		return "将CA证书复制到指定目录，便于手动安装到安卓系统"
	case "cert_gen_progress":
		if lang == EN {
			return "🌟 Generating secure certificates, please wait..."
		}
		return "🌟 正在为您生成安全证书，请稍候..."
	case "directory_not_exists":
		if lang == EN {
			return "Directory does not exist: %s"
		}
		return "目录不存在: %s"
	case "get_home_dir_failed":
		if lang == EN {
			return "❌ Failed to get user home directory: %s"
		}
		return "❌ 获取用户主目录失败: %s"
	case "lang_desc":
		if lang == EN {
			return "Language (en/zh)"
		}
		return "语言 (en/zh)"
	case "invalid_lang_error":
		if lang == EN {
			return "Invalid language. Use 'en' or 'zh'"
		}
		return "语言无效。请使用 'en' 或 'zh'"
	case "language_desc_short":
		if lang == EN {
			return "Switch language between English and Chinese"
		}
		return "在英文和中文之间切换语言"
	case "language_desc_long":
		if lang == EN {
			return "Change the language of the hserve tool interface between English and Chinese"
		}
		return "在英文和中文之间切换 hserve 工具界面语言"
	case "language_switched_en":
		if lang == EN {
			return "Language switched to English"
		}
		return "语言已切换为英文"
	case "language_switched_zh":
		if lang == EN {
			return "Language switched to Chinese"
		}
		return "语言已切换为中文"
	case "config_save_failed":
		if lang == EN {
			return "Failed to save configuration"
		}
		return "保存配置失败"
	case "config_saved":
		if lang == EN {
			return "Configuration saved to"
		}
		return "配置已保存到"
	default:
		return key // 返回键本身作为默认值
	}
}
