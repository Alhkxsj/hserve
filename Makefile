.PHONY: all build clean install termux-deb

APP_NAME := hserve
VERSION  := 1.2.4

PREFIX ?= /data/data/com.termux/files/usr
BIN_DIR := build/bin
DIST_DIR := dist
PKG_DIR := build/pkg

all: build

build:
	@echo "🔧 构建程序..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/hserve ./cmd/hserve
	@echo "✅ 构建完成"

install: build
	@echo "📦 安装到 Termux..."
	install -Dm755 $(BIN_DIR)/hserve $(PREFIX)/bin/hserve
	mkdir -p $(PREFIX)/etc/hserve
	@echo "✅ 安装完成"

deb: build
	@echo "📦 构建 Termux deb 包..."
	rm -rf $(PKG_DIR)
	mkdir -p $(PKG_DIR)/DEBIAN
	mkdir -p $(PKG_DIR)$(PREFIX)/bin
	mkdir -p $(PKG_DIR)$(PREFIX)/etc/hserve
	mkdir -p $(DIST_DIR)

	cp packaging/termux/control  $(PKG_DIR)/DEBIAN/
	cp packaging/termux/postinst $(PKG_DIR)/DEBIAN/
	cp packaging/termux/prerm    $(PKG_DIR)/DEBIAN/
	chmod 755 $(PKG_DIR)/DEBIAN
	chmod 755 $(PKG_DIR)/DEBIAN/*

	cp $(BIN_DIR)/hserve     $(PKG_DIR)$(PREFIX)/bin/

	dpkg-deb --build $(PKG_DIR) $(DIST_DIR)/$(APP_NAME)_$(VERSION)_aarch64.deb
	@echo "✅ deb 构建完成"

clean:
	rm -rf build dist