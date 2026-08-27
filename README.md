#它已被废弃，我意识到这个项目只是随手做着玩，而且很复杂，没有实际应用场景，所以我正式宣布，不会再维护了。


# hserve

一个快速搭建本地HTTPS服务器的工具。

## 使用说明
安装新版本后，系统会自动为您生成证书，无需手动操作。
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
