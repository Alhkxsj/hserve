# hserve Project Code

## Makefile
```
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=gofmt
BINARY_NAME=hserve
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the project
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/hserve

# Install the binary to system
install: build
	cp $(BINARY_NAME) $(HOME)/go/bin/ || cp $(BINARY_NAME) /usr/local/bin/ || echo "Please copy $(BINARY_NAME) to a directory in your PATH"

# Run tests
test: 
	$(GOTEST) -v ./...

# Run go vet
vet:
	$(GOVET) ./...

# Format code
fmt:
	$(GOFMT) -s -w ./

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run go mod tidy
tidy:
	$(GOMOD) tidy

# Build for multiple architectures
multiarch:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-amd64 -v ./cmd/hserve
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-arm64 -v ./cmd/hserve
	GOOS=linux GOARCH=arm $(GOBUILD) -o dist/$(BINARY_NAME)-linux-arm -v ./cmd/hserve
	GOOS=android GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-android-arm64 -v ./cmd/hserve
	GOOS=android GOARCH=arm $(GOBUILD) -o dist/$(BINARY_NAME)-android-arm -v ./cmd/hserve
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-amd64 -v ./cmd/hserve
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-arm64 -v ./cmd/hserve
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-windows-amd64.exe -v ./cmd/hserve

# Build deb package
deb:
	@echo "Building deb package..."
	@mkdir -p dist
	./scripts/build-deb.sh

# Install deb package
install-deb: deb
	sudo dpkg -i dist/*.deb

# Run all checks
check: vet test

# Generate certificates (for testing)
gen-cert:
	./$(BINARY_NAME) gen-cert

# Run server (for testing)
serve:
	./$(BINARY_NAME) serve

.PHONY: build install test vet fmt clean multiarch deb install-deb check gen-cert serve tidy
```

## cmd/hserve/main.go
```
package main

import (
	"fmt"
	"os"

	"github.com/Alhkxsj/hserve/cmd"
	"github.com/Alhkxsj/hserve/internal/i18n"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%s: %s\n", i18n.T(i18n.GetLanguage(), "user_error"), err.Error())
		os.Exit(1)
	}
}
```

## cmd/root.go
```
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
```

## internal/server/server.go
```
package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

type Options struct {
	Addr        string
	Root        string
	Quiet       bool
	CertPath    string
	KeyPath     string
	AllowList   []string
	TlsCertFile string
	TlsKeyFile  string
}

// GetAbsPath 获取绝对路径
func GetAbsPath(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// 确保路径存在
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf(i18n.T(i18n.GetLanguage(), "directory_not_exists"), absPath)
	}
	return absPath, nil
}

// CheckAccess 检查访问权限
func CheckAccess(root string, allowList []string) error {
	if !isPathAllowed(root, allowList) {
		return fmt.Errorf(i18n.T(i18n.GetLanguage(), "path_not_allowed"), root)
	}
	return nil
}

func Run(opt Options) error {
	// 检查访问权限
	if err := CheckAccess(opt.Root, opt.AllowList); err != nil {
		return err
	}

	handler := NewHandler(opt.Root, opt.Quiet, opt.AllowList)

	srv := &http.Server{
		Addr:    opt.Addr,
		Handler: handler,
	}

	if !opt.Quiet {
		lang := i18n.GetLanguage()
		fmt.Printf("🚀 %s\n", i18n.T(lang, "server_started"))
		fmt.Printf("📁 %s: %s\n", i18n.T(lang, "shared_dir"), opt.Root)
		if len(opt.AllowList) > 0 {
			fmt.Printf("✅ %s: %v\n", i18n.T(lang, "access_whitelist"), opt.AllowList)
		}
		fmt.Printf("🌐 %s: https://localhost%s\n", i18n.T(lang, "access_address"), opt.Addr)
		fmt.Printf("🔐 %s: %s\n", i18n.T(lang, "listen_address"), opt.Addr)
		fmt.Printf("💡 %s\n", i18n.T(lang, "tip_open_browser"))
		fmt.Printf("%s\n", i18n.T(lang, "tip_stop_server"))
		fmt.Println()
	}

	// 如果提供了外挂证书，则使用外挂证书，否则使用内置证书
	if opt.TlsCertFile != "" && opt.TlsKeyFile != "" {
		// 验证外挂证书文件是否存在
		if _, err := os.Stat(opt.TlsCertFile); err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "cert_file_not_exists"), opt.TlsCertFile)
		}
		if _, err := os.Stat(opt.TlsKeyFile); err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "key_file_not_exists"), opt.TlsKeyFile)
		}
		return srv.ListenAndServeTLS(opt.TlsCertFile, opt.TlsKeyFile)
	} else {
		// 使用内置证书
		tlsConfig, err := LoadTLSConfig(opt.CertPath, opt.KeyPath)
		if err != nil {
			return fmt.Errorf(i18n.T(i18n.GetLanguage(), "tls_config_failed"), err)
		}
		srv.TLSConfig = tlsConfig
		return srv.ListenAndServeTLS("", "")
	}
}
```

## internal/server/handler.go
```
package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

func NewHandler(root string, quiet bool, allowList []string) http.Handler {
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查路径是否在白名单中
		requestPath := filepath.Join(root, r.URL.Path)
		if !isPathAllowed(requestPath, allowList) {
			http.Error(w, i18n.T(i18n.GetLanguage(), "forbidden_access"), http.StatusForbidden)
			if !quiet {
				fmt.Printf("[%s] %s %s - FORBIDDEN (%s)\n",
					time.Now().Format("15:04:05"),
					r.Method,
					r.URL.Path,
					i18n.T(i18n.GetLanguage(), "forbidden_access"))
			}
			return
		}

		// 安全头部
		secureHeaders(w)

		// 日志记录
		if !quiet {
			fmt.Printf("[%s] %s %s\n",
				time.Now().Format("15:04:05"),
				r.Method,
				r.URL.Path)
		}

		// 修复路径遍历问题
		upath := r.URL.Path
		if !strings.HasPrefix(upath, "/") {
			upath = "/" + upath
			r.URL.Path = upath
		}
		upath = filepath.Join(root, filepath.Clean(upath))
		r.URL.Path = upath

		fs.ServeHTTP(w, r)
	})
}

// isPathAllowed 检查路径是否在白名单中
func isPathAllowed(requestPath string, allowList []string) bool {
	if len(allowList) == 0 {
		return true // 没有白名单则允许所有路径
	}

	// 将请求路径转换为绝对路径进行比较
	absRequestPath, err := filepath.Abs(requestPath)
	if err != nil {
		return false
	}

	for _, allowedPath := range allowList {
		absAllowedPath, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		// 检查请求路径是否在允许的路径下
		rel, err := filepath.Rel(absAllowedPath, absRequestPath)
		if err != nil {
			continue
		}

		// 如果相对路径不以".."开头，则说明请求路径在允许路径下
		if !strings.HasPrefix(rel, "..") && !strings.Contains(rel, "/../") && rel != ".." {
			return true
		}
	}

	return false
}

func secureHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
}
```

## internal/server/path.go
```
package server

import (
	"net/url"
	"path"
	"strings"
)

// cleanPath 防止路径穿越，但允许目录访问
func cleanPath(p string) string {
	decoded, _ := url.PathUnescape(p)
	clean := path.Clean("/" + decoded)
	if strings.Contains(clean, "..") {
		return "/"
	}
	return clean
}
```

## internal/server/tls.go
```
package server

import (
	"crypto/tls"
	"fmt"

	"github.com/Alhkxsj/hserve/internal/i18n"
	tlspolicy "github.com/Alhkxsj/hserve/internal/tls"
)

func LoadTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf(i18n.T(i18n.GetLanguage(), "tls_config_failed"), err)
	}
	return tlspolicy.DefaultConfig(cert), nil
}
```

## internal/server/env.go
```
package server

import (
	"fmt"
	"net"
	"os"
)

type RuntimeEnv struct {
	CertPath string
	KeyPath  string
}

// 检测端口是否可用
func checkPort(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("端口 %s 无法监听，可能已被占用", addr)
	}
	_ = ln.Close()
	return nil
}

// 运行前环境自检
func PreflightCheck(addr, certPath, keyPath string) error {
	if _, err := os.Stat(certPath); err != nil {
		return fmt.Errorf("未找到证书文件：%s\n请先运行 hserve gen-cert 生成证书", certPath)
	}

	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("未找到私钥文件：%s\n请先运行 hserve gen-cert 生成证书", keyPath)
	}

	if err := checkPort(addr); err != nil {
		return err
	}

	return nil
}
```

## internal/tls/policy.go
```
package tls

import "crypto/tls"

// DefaultConfig 返回安全的 TLS 配置
func DefaultConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,

		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}
}
```

## internal/certmanager/generate.go
```
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
```

## internal/certmanager/check.go
```
package certmanager

import "os"

// CheckCertificateExists 检查证书是否存在
func CheckCertificateExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsInTermux 检测是否在 Termux 环境中
func IsInTermux() bool {
	return os.Getenv("PREFIX") != "" && os.Getenv("TERMUX_VERSION") != ""
}
```

## internal/certmanager/install.go
```
package certmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Alhkxsj/hserve/internal/i18n"
)

// GetCertPaths 返回证书和私钥路径
func GetCertPaths() (string, string) {
	var certPath, keyPath string
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			certPath = filepath.Join(prefix, "etc", "hserve", "certs", "server.crt")
			keyPath = filepath.Join(prefix, "etc", "hserve", "certs", "server.key")
		} else {
			certPath = filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "server.crt")
			keyPath = filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "server.key")
		}
	} else {
		certPath = filepath.Join("/etc/hserve/certs", "server.crt")
		keyPath = filepath.Join("/etc/hserve/certs", "server.key")
	}
	return certPath, keyPath
}

// GetCACertPath 返回 CA 证书路径
func GetCACertPath() string {
	if IsInTermux() {
		prefix := os.Getenv("PREFIX")
		if prefix != "" {
			return filepath.Join(prefix, "etc", "hserve", "certs", "ca.crt")
		} else {
			return filepath.Join("/data/data/com.termux/files/usr/etc/hserve/certs", "ca.crt")
		}
	} else {
		return filepath.Join("/etc/hserve/certs", "ca.crt")
	}
}

// ShowInstructions 显示安装证书说明
func ShowInstructions(caCertPath string) {
	lang := i18n.GetLanguage()
	fmt.Println()
	fmt.Printf("%s\n", i18n.T(lang, "android_install_steps"))
	fmt.Printf("1. %s: %s\n", i18n.T(lang, "android_install_step1"), caCertPath)
	fmt.Printf("2. %s\n", i18n.T(lang, "android_install_step2"))
	fmt.Printf("3. %s\n", i18n.T(lang, "android_install_step3"))
	fmt.Printf("4. %s\n", i18n.T(lang, "android_install_step4"))
	fmt.Printf("5. %s\n", i18n.T(lang, "android_install_step5"))
	fmt.Println()
	fmt.Printf("%s\n", i18n.T(lang, "launch_example"))
	fmt.Println("  cd /path/to/website")
	fmt.Println("  hserve")
	fmt.Println()
	fmt.Println(i18n.T(lang, "poem"))
}
```

## internal/certmanager/certmanager_test.go
```
package certmanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInTermux(t *testing.T) {
	// 保存原始环境变量
	originalPrefix := os.Getenv("PREFIX")
	originalTermuxVersion := os.Getenv("TERMUX_VERSION")

	// 清理环境变量
	os.Unsetenv("PREFIX")
	os.Unsetenv("TERMUX_VERSION")

	// 测试非Termux环境
	if IsInTermux() {
		t.Error("Expected IsInTermux() to return false when not in Termux")
	}

	// 设置Termux环境
	os.Setenv("PREFIX", "/data/data/com.termux/files/usr")
	os.Setenv("TERMUX_VERSION", "0.118.0")

	if !IsInTermux() {
		t.Error("Expected IsInTermux() to return true when in Termux")
	}

	// 恢复原始环境变量
	os.Setenv("PREFIX", originalPrefix)
	os.Setenv("TERMUX_VERSION", originalTermuxVersion)
}

func TestGetCertPaths(t *testing.T) {
	// 这个测试会验证证书路径生成逻辑
	certPath, keyPath := GetCertPaths()

	// 检查路径是否包含正确的文件名
	if filepath.Ext(certPath) != ".crt" && filepath.Ext(certPath) != ".pem" {
		t.Errorf("Certificate path does not have expected extension: %s", certPath)
	}
	if filepath.Ext(keyPath) != ".key" && filepath.Ext(keyPath) != ".pem" {
		t.Errorf("Key path does not have expected extension: %s", keyPath)
	}
}

func TestCheckCertificateExists(t *testing.T) {
	// 测试不存在的文件
	if CheckCertificateExists("/nonexistent/path/to/cert") {
		t.Error("Expected non-existent file to not exist")
	}

	// 测试当前目录下不存在的文件
	if CheckCertificateExists("nonexistent.cert") {
		t.Error("Expected non-existent file to not exist")
	}
}
```

## internal/i18n/i18n.go
```
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
```

## internal/i18n/i18n_test.go
```
package i18n

import (
	"os"
	"testing"
)

func TestGetLanguage(t *testing.T) {
	// 保存原始环境变量
	originalLang := os.Getenv("LANG")
	originalLcAll := os.Getenv("LC_ALL")

	// 清理环境变量
	os.Unsetenv("LANG")
	os.Unsetenv("LC_ALL")

	// 默认情况下应返回英文
	defaultLang := GetLanguage()
	if defaultLang != EN {
		t.Errorf("Expected default language to be EN, got %s", defaultLang)
	}

	// 设置语言并测试
	SetLanguage(ZH)
	if GetLanguage() != ZH {
		t.Errorf("Expected language to be ZH after SetLanguage(ZH)")
	}

	SetLanguage(EN)
	if GetLanguage() != EN {
		t.Errorf("Expected language to be EN after SetLanguage(EN)")
	}

	// 测试系统语言获取
	os.Setenv("LANG", "zh_CN.UTF-8")
	if GetSystemLanguage() != ZH {
		t.Errorf("Expected system language to be ZH when LANG=zh_CN.UTF-8")
	}

	os.Setenv("LANG", "en_US.UTF-8")
	if GetSystemLanguage() != EN {
		t.Errorf("Expected system language to be EN when LANG=en_US.UTF-8")
	}

	// 恢复原始环境变量
	os.Setenv("LANG", originalLang)
	os.Setenv("LC_ALL", originalLcAll)

	// 确保恢复默认设置
	SetLanguage(EN) // 重置为默认英文
}

func TestT(t *testing.T) {
	// 测试中文翻译
	zhTranslation := T(ZH, "https_server_title")
	if zhTranslation != "HTTPS 文件服务器 - 让文件分享变得简单而安全" {
		t.Errorf("Expected Chinese translation, got: %s", zhTranslation)
	}

	// 测试英文翻译
	enTranslation := T(EN, "https_server_title")
	if enTranslation != "HTTPS File Server - Making file sharing simple and secure" {
		t.Errorf("Expected English translation, got: %s", enTranslation)
	}

	// 测试未定义的键
	undefinedKey := T(ZH, "undefined_key")
	if undefinedKey != "undefined_key" {
		t.Errorf("Expected undefined key to return itself, got: %s", undefinedKey)
	}
}
```

## docs/usage.md
```
1. Project Introduction

hserve is a simple and easy-to-use HTTPS file server:

Auto-generate CA and server certificates

Suitable for local development / LAN file sharing

Specifically adapted for Termux (Android) environment

No external CA dependency, no internet connection required



---

2. Installation

Termux

make termux-install

After installation, you will get:

hserve - HTTPS file server (with gen-cert subcommand for certificate generation)



---

3. Generate Certificate (Required)

Before first use, you must generate certificates:

hcertgen

Generated content:

CA root certificate (for installation to Android system)

Server certificate + private key (for server use)



---

4. Install CA Certificate to Android

See documentation: android-ca-install.md

[WARNING] Without installing CA, browsers will show "Not Secure" warning.


---

5. Start Server

hserve

Common parameters:

-port   Listening port (default 8443)
-dir    Shared directory (default current directory)
-quiet  Quiet mode (no access logs)

Example:

hserve -dir=/sdcard -port=9443
```

## docs/android-ca-install.md
```
1. Certificate File Location

After running hcertgen, a CA certificate file will be generated
The certificate file is placed in the home directory by default
~/hserve-ca.crt


---

2. Copy Certificate to Phone Storage

You can copy it to:

/storage/emulated/0/Download/
You don't have to copy it here specifically, just make sure you can find it when selecting the certificate file

---

3. Android Installation Steps

1. Open Settings


2. Security → Encryption & Credentials


3. Install Certificate → CA Certificate


4. Select hserve-ca.crt


5. If you can't find it, search in the settings top search box (certificate) and find the relevant certificate installation search results, then install the certificate.


---

6. Notes

Android will warn "Certificate can monitor traffic" - this is normal

Certificate is only for your locally generated HTTPS service

No upload, no internet, no sharing
```

## docs/security-model.md
```
1. Design Goals

The security model of this project follows these principles:

Local-first

Full user control

No third-party CA dependency

No complex PKI system



---

2. Trust Model

User
 └─ Install local CA (user actively trusts)
      └─ hserve (only valid locally)

CA is only generated on this device

Private key never leaves the device



---

3. Why Use Self-Signed CA

Reasons:

Let's Encrypt is not suitable for local / IP usage

Android local development certificate needs are clear

Self-signed CA = User actively trusts


This is a developer tool, not a public network service.


---

4. TLS Policy

TLS minimum version: TLS 1.2

Disable insecure protocols

Certificate validity is longer to reduce repeated operations


Specific parameters defined in:

internal/tls/policy.go


---

5. Things Not Done (Intentionally)

[X] Automatically install system certificates

[X] Bypass Android security prompts

[X] Background resident service


Users must clearly know what they are doing.


---

4. Applicable Scenarios Summary

Local HTTPS development testing

Android ↔ PC file sharing

LAN device access


Not suitable for:

Public network deployment

Commercial HTTPS services



---

5. Conclusion

This is a tool designed for clear-minded people.

No magic

No hidden behavior

All certificates and trusts are in your hands


If you understand HTTPS, you will like it.
```

## docs/usage_zh.md
```
1. 项目简介

hserve 是一个简单易用的 HTTPS 文件服务器：

自动生成 CA 与服务器证书

适合本地开发 / 局域网文件共享

特别适配 Termux（Android）环境

不依赖外部 CA，不联网



---

2. 安装

Termux

make termux-install

安装完成后您将获得：

hserve   HTTPS 文件服务器（包含 gen-cert 子命令用于证书生成）



---

3. 生成证书（必须）

首次使用前必须生成证书：

hcertgen

生成内容：

CA 根证书（用于安装到 Android 系统）

服务器证书 + 私钥（服务器使用）



---

4. 安装 CA 证书到 Android

见文档：android-ca-install.md

[警告] 不安装 CA，浏览器会提示"不安全连接"。


---

5. 启动服务器

hserve

常用参数：

-port   监听端口（默认 8443）
-dir    共享目录（默认当前目录）
-quiet  安静模式（不输出访问日志）

示例：

hserve -dir=/sdcard -port=9443
```

## docs/android-ca-install_zh.md
```
1. 证书文件位置

运行 hcertgen 后，会生成一个 CA 证书文件
证书文件默认放在home目录
~/hserve-ca.crt


---

2. 复制证书到手机存储

可以复制到：

/storage/emulated/0/Download/
也不一定非得复制到这里，只要在选择证书文件的时候，你能找得到在哪就行

---

3. Android 安装步骤

1. 打开 设置


2. 安全 → 加密与凭据


3. 安装证书 → CA 证书


4. 选择 hserve-ca.crt


5. 如果找不到，那在设置最上方的搜索框里搜索（证书）然后找到安装证书相关搜索结果，安装证书。


---

6. 注意事项

Android 会警告"证书可监控流量"——这是正常的

证书仅用于你本地生成的 HTTPS 服务

不会上传、不联网、不共享
```

## docs/security-model_zh.md
```
1. 设计目标

本项目安全模型遵循以下原则：

本地优先

用户完全可控

不依赖第三方 CA

不引入复杂 PKI 体系



---

2. 信任模型

用户
 └─ 安装本地 CA（用户主动信任）
      └─ hserve（仅本地有效）

CA 只在本机生成

私钥不离开设备



---

3. 为什么使用自签 CA

原因：

Let's Encrypt 不适合本地 / IP

Android 本地开发证书需求明确

自签 CA = 用户主动信任


这是 开发者工具，不是公网服务。


---

4. TLS 策略

TLS 最低版本：TLS 1.2

禁用不安全协议

证书有效期较长，减少重复操作


具体参数定义见：

internal/tls/policy.go


---

5. 不做的事情（刻意）

[X] 自动安装系统证书

[X] 绕过 Android 安全提示

[X] 后台常驻服务


用户必须明确知道自己在做什么。


---

四、适用场景总结

本地 HTTPS 开发测试

Android ↔ PC 文件共享

局域网设备访问


不适合：

公网部署

商业 HTTPS 服务



---

五、结束语

这是一个 为清楚的人准备的工具。

没有魔法

没有隐蔽行为

所有证书、信任都在你手里


如果你理解 HTTPS，那么你会喜欢它。
```