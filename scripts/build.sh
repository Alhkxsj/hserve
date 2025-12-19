#!/bin/bash

# hserve build script
set -e

echo "🚀 Building hserve..."

# 获取项目版本
VERSION=$(grep 'VERSION :=' Makefile | cut -d' ' -f3)
if [ -z "$VERSION" ]; then
    VERSION="1.2.6-dev"
fi

# 创建构建目录
mkdir -p build/bin

# 构建主程序
echo "🔧 Building hserve..."
go build -ldflags="-X main.Version=$VERSION" -o build/bin/hserve ./cmd/hserve

echo "✅ Build completed successfully!"
echo "✨ Binary location: build/bin/hserve"