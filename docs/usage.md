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

安装完成后会得到一个命令：

hserve         HTTPS 文件服务器（包含证书生成功能）



---

3. 生成证书（必须）

首次使用前必须生成证书：

hserve cert

或

hserve certgen

生成内容：

CA 根证书（用于安装到 Android 系统）

服务器证书 + 私钥（服务器使用）



---

4. 安装 CA 证书到 Android

见文档：android-ca-install.md

⚠️ 不安装 CA，浏览器会提示“不安全连接”。


---

5. 启动服务器

hserve serve

或直接运行 hserve（默认启动服务器）

常用参数：

-port   监听端口（默认 8443）
-dir    共享目录（默认当前目录）
-quiet  安静模式（不输出访问日志）

示例：

hserve serve -dir=/sdcard -port=9443

或

hserve -dir=/sdcard -port=9443


---

6. 命令帮助

查看所有可用命令：

hserve help

查看特定命令帮助：

hserve cert -help

hserve serve -help

hserve version


---