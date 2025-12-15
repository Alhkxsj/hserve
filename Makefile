.PHONY: all build clean install termux-deb

APP_NAME := https-server
VERSION  := 1.2.1

PREFIX ?= /data/data/com.termux/files/usr
BIN_DIR := build/bin
DIST_DIR := dist
PKG_DIR := build/pkg

all: build

build:
	@echo "🔧 构建程序..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/https-server ./cmd/https-server
	go build -o $(BIN_DIR)/https-certgen ./cmd/https-certgen
	@echo "✅ 构建完成"

install: build
	@echo "📦 安装到 Termux..."
	install -Dm755 $(BIN_DIR)/https-server $(PREFIX)/bin/https-server
	install -Dm755 $(BIN_DIR)/https-certgen $(PREFIX)/bin/https-certgen
	mkdir -p $(PREFIX)/etc/https-server
	@echo "✅ 安装完成"

deb: build
	@echo "📦 构建 Termux deb 包..."
	rm -rf $(PKG_DIR)
	mkdir -p $(PKG_DIR)/DEBIAN
	mkdir -p $(PKG_DIR)$(PREFIX)/bin
	mkdir -p $(PKG_DIR)$(PREFIX)/etc/https-server

	cp packaging/termux/control  $(PKG_DIR)/DEBIAN/
	cp packaging/termux/postinst $(PKG_DIR)/DEBIAN/
	cp packaging/termux/prerm    $(PKG_DIR)/DEBIAN/
	chmod 755 $(PKG_DIR)/DEBIAN
	chmod 755 $(PKG_DIR)/DEBIAN/*

	cp $(BIN_DIR)/https-server     $(PKG_DIR)$(PREFIX)/bin/
	cp $(BIN_DIR)/https-certgen    $(PKG_DIR)$(PREFIX)/bin/

	dpkg-deb --build $(PKG_DIR) $(DIST_DIR)/$(APP_NAME)_$(VERSION)_aarch64.deb
	@echo "✅ deb 构建完成"

clean:
	rm -rf build dist