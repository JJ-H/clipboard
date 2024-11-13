.PHONY: dev build clean

# 默认目标
all: build

# 开发模式
dev:
	wails dev

# 构建项目
build:
	wails build

# 清理构建文件
clean:
	rm -rf build/
	cd frontend && rm -rf dist/ node_modules/

# 安装依赖
install:
	cd frontend && npm install

# 更新依赖
update:
	cd frontend && npm update
	go get -u ./...

# 运行测试
test:
	go test ./...

# 生成 Windows 版本
build-windows:
	wails build -platform windows/amd64

# 生成 macOS 版本
build-mac:
	wails build -platform darwin/universal

# 生成 Linux 版本
build-linux:
	wails build -platform linux/amd64

# 帮助信息
help:
	@echo "可用的命令:"
	@echo "  make dev          - 运行开发模式"
	@echo "  make build        - 构建项目"
	@echo "  make clean        - 清理构建文件"
	@echo "  make install      - 安装依赖"
	@echo "  make update       - 更新依赖"
	@echo "  make test         - 运行测试"
	@echo "  make build-windows - 构建 Windows 版本"
	@echo "  make build-mac    - 构建 macOS 版本"
	@echo "  make build-linux  - 构建 Linux 版本" 