package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alhkxsj/hserve/internal/certmanager"
	"github.com/Alhkxsj/hserve/internal/i18n"
	"github.com/Alhkxsj/hserve/internal/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hserve",
	Short: i18n.T(i18n.GetLanguage(), "hserve_desc"),
	Long:  i18n.T(i18n.GetLanguage(), "hserve_long_desc"),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果只执行根命令且没有参数，或者指定了版本标志
		if len(args) == 0 {
			if version {
				lang := i18n.GetLanguage()
				fmt.Printf("🌟 %s v1.2.3\n", i18n.T(lang, "https_server_title"))
				fmt.Println("👤 Author: 快手阿泠 (Alexa Haley)")
				fmt.Println("🏠 Project: https://github.com/Alhkxsj/hserve")
				fmt.Println(i18n.T(lang, "poem"))
				return
			}
			// 如果没有参数也没有指定版本，显示帮助
			cmd.Help()
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 在命令执行前处理语言设置
		if lang != "" {
			switch lang {
			case "en", "EN", "eng":
				i18n.SetLanguage(i18n.EN)
			case "zh", "ZH", "ch", "cn":
				i18n.SetLanguage(i18n.ZH)
			}
		}
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

var (
	port        int
	dir         string
	quiet       bool
	force       bool
	version     bool
	lang        string
	allowList   []string
	tlsCertFile string
	tlsKeyFile  string
	autoGen     bool
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: i18n.T(i18n.GetLanguage(), "serve_desc"),
	Long:  i18n.T(i18n.GetLanguage(), "serve_long_desc"),
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			lang := i18n.GetLanguage()
			fmt.Printf("🌟 %s v1.2.3\n", i18n.T(lang, "https_server_title"))
			fmt.Println("👤 Author: 快手阿泠 (Alexa Haley)")
			fmt.Println("🏠 Project: https://github.com/Alhkxsj/hserve")
			fmt.Println(i18n.T(lang, "poem"))
			return
		}

		// 如果指定了外挂证书，则跳过自动证书生成
		if tlsCertFile == "" || tlsKeyFile == "" {
			// 智能启动逻辑：如果证书不存在，自动调用gen-cert
			certPath, _ := certmanager.GetCertPaths()
			if !certmanager.CheckCertificateExists(certPath) {
				if autoGen {
					lang := i18n.GetLanguage()
					fmt.Println(i18n.T(lang, "cert_gen_auto"))
					if err := certmanager.Generate(false); err != nil {
						fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "cert_auto_gen_failed"), err)
						os.Exit(1)
					}
					// 安装到Termux信任库（如果在Termux环境中）
					if certmanager.IsInTermux() {
						caCertPath := certmanager.GetCACertPath()
						prefix := os.Getenv("PREFIX")
						termuxCertDir := prefix + "/etc/tls/certs/"
						if err := os.MkdirAll(termuxCertDir, 0755); err != nil {
							fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "termux_cert_dir_failed"), err)
						} else {
							caCertName := "hserve_ca.crt"
							termuxCaCertPath := filepath.Join(termuxCertDir, caCertName)
							if err := copyFile(caCertPath, termuxCaCertPath); err != nil {
								fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "install_ca_failed"), err)
							} else {
								fmt.Println(i18n.T(i18n.GetLanguage(), "ca_installed_auto"))
							}
						}
					}
				} else {
					lang := i18n.GetLanguage()
					fmt.Println(i18n.T(lang, "cert_not_found"))
					fmt.Println(i18n.T(lang, "run_gen_cert"))
					fmt.Println(i18n.T(lang, "auto_gen_tip"))
					os.Exit(1)
				}
			}
		}

		root, err := server.GetAbsPath(dir)
		if err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "get_path_failed"), err)
			os.Exit(1)
		}

		// 获取证书路径（除非使用外挂证书）
		var certPath, keyPathValue string
		if tlsCertFile == "" || tlsKeyFile == "" {
			certPath, keyPathValue = certmanager.GetCertPaths()
		} else {
			certPath = tlsCertFile
			keyPathValue = tlsKeyFile
		}

		opts := server.Options{
			Addr:        fmt.Sprintf(":%d", port),
			Root:        root,
			Quiet:       quiet,
			CertPath:    certPath,
			KeyPath:     keyPathValue,
			AllowList:   allowList,
			TlsCertFile: tlsCertFile,
			TlsKeyFile:  tlsKeyFile,
		}

		if err := server.Run(opts); err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "server_start_failed"), err)
			os.Exit(1)
		}
	},
}

func initServeCmd() {
	serveCmd.SetUsageFunc(func(*cobra.Command) error {
		lang := i18n.GetLanguage()
		fmt.Printf("🚀 %s\n", i18n.T(lang, "https_server_title"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "usage"))
		fmt.Printf("  %s [OPTIONS]\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "available_options"))
		fmt.Println("  -port int")
		fmt.Printf("      %s\n", i18n.T(lang, "port_desc"))
		fmt.Println("  -dir string")
		fmt.Printf("      %s\n", i18n.T(lang, "dir_desc"))
		fmt.Println("  -quiet")
		fmt.Printf("      %s\n", i18n.T(lang, "quiet_desc"))
		fmt.Println("  -auto-gen")
		fmt.Printf("      %s\n", i18n.T(lang, "auto_gen_desc"))
		fmt.Println("  -allow stringArray")
		fmt.Printf("      %s\n", i18n.T(lang, "allow_desc"))
		fmt.Println("  -tls-cert-file string")
		fmt.Printf("      %s\n", i18n.T(lang, "tls_cert_file_desc"))
		fmt.Println("  -tls-key-file string")
		fmt.Printf("      %s\n", i18n.T(lang, "tls_key_file_desc"))
		fmt.Println("  -lang string")
		fmt.Printf("      %s\n", i18n.T(lang, "lang_desc"))
		fmt.Println("  -version")
		fmt.Printf("      %s\n", i18n.T(lang, "version_desc"))
		fmt.Println("  -help")
		fmt.Printf("      %s\n", i18n.T(lang, "help_desc"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "tip_cert_first"))
		fmt.Println(i18n.T(lang, "poem"))
		return nil
	})
}

var genCertCmd = &cobra.Command{
	Use:   "gen-cert",
	Short: i18n.T(i18n.GetLanguage(), "gen_cert_desc"),
	Long:  i18n.T(i18n.GetLanguage(), "gen_cert_long_desc"),
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			lang := i18n.GetLanguage()
			fmt.Printf("🔐 %s v1.2.3\n", i18n.T(lang, "https_server_title"))
			fmt.Println("👤 Author: 快手阿泠 (Alexa Haley)")
			fmt.Println("🏠 Project: https://github.com/Alhkxsj/hserve")
			fmt.Println(i18n.T(lang, "poem"))
			return
		}

		lang := i18n.GetLanguage()
		fmt.Printf("🔐 %s - %s\n", i18n.T(lang, "https_server_title"), i18n.T(lang, "tip_external_cert"))
		fmt.Println(i18n.T(lang, "poem"))
		fmt.Println(i18n.T(lang, "cert_gen_progress"))

		if err := certmanager.Generate(force); err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "cert_gen_failed"), err)
			os.Exit(1)
		}

		fmt.Println("================================")
	},
}

func initGenCertCmd() {
	genCertCmd.SetUsageFunc(func(*cobra.Command) error {
		lang := i18n.GetLanguage()
		fmt.Printf("🔐 %s - %s\n", i18n.T(lang, "https_server_title"), i18n.T(lang, "tip_external_cert"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "usage"))
		fmt.Printf("  %s [OPTIONS]\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "available_options"))
		fmt.Println("  -force")
		fmt.Printf("      %s\n", i18n.T(lang, "force_desc"))
		fmt.Println("  -lang string")
		fmt.Printf("      %s\n", i18n.T(lang, "lang_desc"))
		fmt.Println("  -version")
		fmt.Printf("      %s\n", i18n.T(lang, "version_desc"))
		fmt.Println("  -help")
		fmt.Printf("      %s\n", i18n.T(lang, "help_desc"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "tip_external_cert"))
		fmt.Println(i18n.T(lang, "poem"))
		return nil
	})
}

var installCaCmd = &cobra.Command{
	Use:   "install-ca",
	Short: i18n.T(i18n.GetLanguage(), "install_ca_desc"),
	Long:  i18n.T(i18n.GetLanguage(), "install_ca_long_desc"),
	Run: func(cmd *cobra.Command, args []string) {
		// 检查是否在Termux环境中
		if !certmanager.IsInTermux() {
			fmt.Println(i18n.T(i18n.GetLanguage(), "termux_only"))
			return
		}

		// 获取CA证书路径
		caCertPath := certmanager.GetCACertPath()
		if !certmanager.CheckCertificateExists(caCertPath) {
			fmt.Println(i18n.T(i18n.GetLanguage(), "ca_not_found"))
			fmt.Println(i18n.T(i18n.GetLanguage(), "run_gen_cert"))
			os.Exit(1)
		}

		// 检查Termux证书目录
		prefix := os.Getenv("PREFIX")
		termuxCertDir := prefix + "/etc/tls/certs/"
		if err := os.MkdirAll(termuxCertDir, 0755); err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "termux_cert_dir_failed"), err)
			os.Exit(1)
		}

		// 复制CA证书到Termux证书目录
		caCertName := "hserve_ca.crt"
		termuxCaCertPath := filepath.Join(termuxCertDir, caCertName)

		if err := copyFile(caCertPath, termuxCaCertPath); err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "install_ca_failed"), err)
			os.Exit(1)
		}

		fmt.Println(i18n.T(i18n.GetLanguage(), "ca_installed_success"))
	},
}

func initInstallCaCmd() {
	installCaCmd.SetUsageFunc(func(*cobra.Command) error {
		lang := i18n.GetLanguage()
		fmt.Printf("🔐 %s\n", i18n.T(lang, "https_server_title"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "usage"))
		fmt.Printf("  %s [OPTIONS]\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "available_options"))
		fmt.Println("  -lang string")
		fmt.Printf("      %s\n", i18n.T(lang, "lang_desc"))
		fmt.Println("  -version")
		fmt.Printf("      %s\n", i18n.T(lang, "version_desc"))
		fmt.Println("  -help")
		fmt.Printf("      %s\n", i18n.T(lang, "help_desc"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "install_ca_desc"))
		fmt.Println(i18n.T(lang, "poem"))
		return nil
	})
}

var exportCaCmd = &cobra.Command{
	Use:   "export-ca",
	Short: i18n.T(i18n.GetLanguage(), "export_ca_desc"),
	Long:  i18n.T(i18n.GetLanguage(), "export_ca_long_desc"),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取CA证书路径
		caCertPath := certmanager.GetCACertPath()
		if !certmanager.CheckCertificateExists(caCertPath) {
			fmt.Println(i18n.T(i18n.GetLanguage(), "ca_not_found"))
			fmt.Println(i18n.T(i18n.GetLanguage(), "run_gen_cert"))
			os.Exit(1)
		}

		// 默认导出到用户存储目录
		storageDir := filepath.Join(os.Getenv("HOME"), "storage", "downloads")
		if _, err := os.Stat(storageDir); os.IsNotExist(err) {
			// 如果存储目录不存在，尝试创建
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "get_home_dir_failed"), err)
				os.Exit(1)
			}
			storageDir = filepath.Join(homeDir, "hserve-ca.crt")
		} else {
			storageDir = filepath.Join(storageDir, "hserve-ca.crt")
		}

		if err := copyFile(caCertPath, storageDir); err != nil {
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "export_ca_failed"), err)
			os.Exit(1)
		}

		fmt.Printf("%s: %s\n", i18n.T(i18n.GetLanguage(), "export_ca_success"), storageDir)
		fmt.Println()
		lang := i18n.GetLanguage()
		fmt.Printf("%s\n", i18n.T(lang, "android_install_steps"))
		fmt.Printf("%s\n", i18n.T(lang, "android_install_step1"))
		fmt.Printf("%s\n", i18n.T(lang, "android_install_step2"))
		fmt.Printf("%s\n", i18n.T(lang, "android_install_step3"))
		fmt.Printf("%s\n", i18n.T(lang, "android_install_step4"))
		fmt.Printf("%s\n", i18n.T(lang, "android_install_step5"))
		fmt.Println()
		fmt.Println(i18n.T(lang, "poem"))
	},
}

func initExportCaCmd() {
	exportCaCmd.SetUsageFunc(func(*cobra.Command) error {
		lang := i18n.GetLanguage()
		fmt.Printf("🔐 %s\n", i18n.T(lang, "https_server_title"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "usage"))
		fmt.Printf("  %s [OPTIONS]\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "available_options"))
		fmt.Println("  -lang string")
		fmt.Printf("      %s\n", i18n.T(lang, "lang_desc"))
		fmt.Println("  -version")
		fmt.Printf("      %s\n", i18n.T(lang, "version_desc"))
		fmt.Println("  -help")
		fmt.Printf("      %s\n", i18n.T(lang, "help_desc"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "export_ca_desc"))
		fmt.Println(i18n.T(lang, "poem"))
		return nil
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 设置目标文件权限
	return os.Chmod(dst, 0644)
}

// languageCmd 定义语言切换命令
var languageCmd = &cobra.Command{
	Use:   "language [en|zh]",
	Short: i18n.T(i18n.GetLanguage(), "language_desc_short"),
	Long:  i18n.T(i18n.GetLanguage(), "language_desc_long"),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		langArg := args[0]
		var newLang i18n.LangType
		var successMessage string

		switch langArg {
		case "en", "EN", "eng", "english":
			newLang = i18n.EN
			i18n.SetLanguage(i18n.EN)
			successMessage = i18n.T(i18n.EN, "language_switched_en")
		case "zh", "ZH", "ch", "cn", "chinese":
			newLang = i18n.ZH
			i18n.SetLanguage(i18n.ZH)
			successMessage = i18n.T(i18n.ZH, "language_switched_zh")
		default:
			fmt.Printf("%s: %s\n", i18n.T(i18n.GetLanguage(), "invalid_lang_error"), langArg)
			os.Exit(1)
		}

		// 尝试将语言设置保存到配置文件
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "hserve")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			// 如果无法创建配置目录，只在当前会话中设置语言
			fmt.Println(successMessage)
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "config_save_failed"), err)
			return
		}

		defaultLangFile := filepath.Join(configDir, "default_lang")
		if err := os.WriteFile(defaultLangFile, []byte(langArg), 0644); err != nil {
			// 如果无法写入配置文件，只在当前会话中设置语言
			fmt.Println(successMessage)
			fmt.Printf("%s: %v\n", i18n.T(i18n.GetLanguage(), "config_save_failed"), err)
			return
		}

		fmt.Println(successMessage)
		fmt.Printf("%s: %s\n", i18n.T(newLang, "config_saved"), defaultLangFile)
	},
}

func initLanguageCmd() {
	languageCmd.SetUsageFunc(func(*cobra.Command) error {
		lang := i18n.GetLanguage()
		fmt.Printf("🌐 %s\n", i18n.T(lang, "https_server_title"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "usage"))
		fmt.Printf("  %s language [en|zh]\n", filepath.Base(os.Args[0]))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "available_options"))
		fmt.Println("  en    English language")
		fmt.Println("  zh    Chinese language")
		fmt.Println("  -lang string")
		fmt.Printf("      %s\n", i18n.T(lang, "lang_desc"))
		fmt.Println("  -version")
		fmt.Printf("      %s\n", i18n.T(lang, "version_desc"))
		fmt.Println("  -help")
		fmt.Printf("      %s\n", i18n.T(lang, "help_desc"))
		fmt.Println()
		fmt.Printf("%s\n", i18n.T(lang, "language_desc_long"))
		fmt.Println(i18n.T(lang, "poem"))
		return nil
	})
}

func init() {
	// 检查是否有配置文件设置默认语言
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "hserve")
	defaultLangFile := filepath.Join(configDir, "default_lang")

	// 尝试读取默认语言设置
	defaultLang := "en" // 默认为英文
	if _, err := os.Stat(defaultLangFile); err == nil {
		// 配置文件存在，读取内容
		if content, err := os.ReadFile(defaultLangFile); err == nil {
			defaultLang = string(content)
			// 去除可能的空白字符和换行符
			defaultLang = strings.TrimSpace(defaultLang)
		}
	} else {
		// 如果配置文件不存在，尝试检测系统语言
		if i18n.GetSystemLanguage() == i18n.ZH {
			defaultLang = "zh"
		}
	}

	// 根据配置文件设置默认语言
	if defaultLang == "zh" {
		i18n.SetLanguage(i18n.ZH) // 设置为中文
	} else {
		i18n.SetLanguage(i18n.EN) // 默认为英文
	}

	// 检查命令行参数中的语言设置（这会覆盖配置文件设置）
	for i, arg := range os.Args {
		if arg == "--lang" || arg == "-l" {
			if i+1 < len(os.Args) {
				langArg := os.Args[i+1]
				switch langArg {
				case "en", "EN", "eng":
					i18n.SetLanguage(i18n.EN)
				case "zh", "ZH", "ch", "cn":
					i18n.SetLanguage(i18n.ZH)
				}
				break
			}
		}
	}

	// 添加版本标志到根命令
	rootCmd.PersistentFlags().BoolVar(&version, "version", false, i18n.T(i18n.GetLanguage(), "version_desc"))
	rootCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "", i18n.T(i18n.GetLanguage(), "lang_desc"))

	// serve 命令的标志
	serveCmd.Flags().IntVarP(&port, "port", "p", 8443, i18n.T(i18n.GetLanguage(), "port_desc"))
	serveCmd.Flags().StringVarP(&dir, "dir", "d", ".", i18n.T(i18n.GetLanguage(), "dir_desc"))
	serveCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, i18n.T(i18n.GetLanguage(), "quiet_desc"))
	serveCmd.Flags().StringSliceVar(&allowList, "allow", []string{}, i18n.T(i18n.GetLanguage(), "allow_desc"))
	serveCmd.Flags().StringVar(&tlsCertFile, "tls-cert-file", "", i18n.T(i18n.GetLanguage(), "tls_cert_file_desc"))
	serveCmd.Flags().StringVar(&tlsKeyFile, "tls-key-file", "", i18n.T(i18n.GetLanguage(), "tls_key_file_desc"))
	serveCmd.Flags().BoolVar(&autoGen, "auto-gen", false, i18n.T(i18n.GetLanguage(), "auto_gen_desc"))

	// gen-cert 命令的标志
	genCertCmd.Flags().BoolVarP(&force, "force", "f", false, i18n.T(i18n.GetLanguage(), "force_desc"))

	// 初始化命令的使用函数
	initServeCmd()
	initGenCertCmd()
	initInstallCaCmd()
	initExportCaCmd()
	initLanguageCmd()

	// 添加子命令到根命令
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(genCertCmd)
	rootCmd.AddCommand(installCaCmd)
	rootCmd.AddCommand(exportCaCmd)
	rootCmd.AddCommand(languageCmd)
}
