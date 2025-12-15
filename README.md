# HTTPS Server

一个快速搭建本地HTTPS服务器的工具。

## 使用说明
注意注意。每一次。安装新版本之后，都要重新生成证书，并从系统里删除以前的旧证书。重新安装新证书。
详细使用说明请参见 [使用说明文档](./docs/usage.md)。

## 安全模型

关于安全模型的信息请参见 [安全模型文档](./docs/security-model.md)。

## Android CA安装

在Android设备上安装CA证书的详细步骤请参见 [Android CA安装文档](./docs/android-ca-install.md)。

## 构建与安装

### 构建

```bash
make build
```

### 构建deb包

```bash
make deb
```

### 安装

```bash
make install
```

安装deb包：

```bash
dpkg -i dist/*.deb
```

## 许可证

本项目采用 [LICENSE](./LICENSE) 许可证。