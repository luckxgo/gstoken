# GSToken Golang权限认证框架 Makefile

# 项目信息
PROJECT_NAME := gstoken
VERSION := v1.0.0
GO_VERSION := 1.23

# 颜色定义
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "$(BLUE)GSToken Golang权限认证框架$(RESET)"
	@echo "$(BLUE)版本: $(VERSION)$(RESET)"
	@echo ""
	@echo "$(YELLOW)可用命令:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 环境检查
.PHONY: check-env
check-env: ## 检查开发环境
	@echo "$(BLUE)检查开发环境...$(RESET)"
	@go version
	@echo "$(GREEN)✓ Go环境正常$(RESET)"

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "$(BLUE)下载项目依赖...$(RESET)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ 依赖下载完成$(RESET)"

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "$(BLUE)更新项目依赖...$(RESET)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✓ 依赖更新完成$(RESET)"

# 代码检查
.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓ 代码格式化完成$(RESET)"

.PHONY: vet
vet: ## 代码静态检查
	@echo "$(BLUE)执行代码静态检查...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✓ 静态检查通过$(RESET)"

.PHONY: lint
lint: ## 代码规范检查 (需要安装golangci-lint)
	@echo "$(BLUE)执行代码规范检查...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✓ 代码规范检查完成$(RESET)"; \
	else \
		echo "$(YELLOW)⚠ golangci-lint未安装，跳过检查$(RESET)"; \
		echo "$(YELLOW)安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
	fi

# 测试相关
.PHONY: test
test: ## 运行所有测试
	@echo "$(BLUE)运行所有测试...$(RESET)"
	go test ./test -v
	@echo "$(GREEN)✓ 测试完成$(RESET)"

.PHONY: test-short
test-short: ## 运行快速测试
	@echo "$(BLUE)运行快速测试...$(RESET)"
	go test ./test -v -short
	@echo "$(GREEN)✓ 快速测试完成$(RESET)"

.PHONY: test-cover
test-cover: ## 运行测试并生成覆盖率报告
	@echo "$(BLUE)运行测试并生成覆盖率报告...$(RESET)"
	go test ./test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ 覆盖率报告生成完成: coverage.html$(RESET)"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	@echo "$(BLUE)运行竞态检测测试...$(RESET)"
	go test ./test -v -race
	@echo "$(GREEN)✓ 竞态检测测试完成$(RESET)"

.PHONY: bench
bench: ## 运行性能测试
	@echo "$(BLUE)运行性能测试...$(RESET)"
	go test ./test -v -bench=. -benchmem
	@echo "$(GREEN)✓ 性能测试完成$(RESET)"

.PHONY: bench-cpu
bench-cpu: ## 运行CPU性能分析
	@echo "$(BLUE)运行CPU性能分析...$(RESET)"
	go test ./test -bench=. -cpuprofile=cpu.prof
	@echo "$(GREEN)✓ CPU性能分析完成: cpu.prof$(RESET)"

.PHONY: bench-mem
bench-mem: ## 运行内存性能分析
	@echo "$(BLUE)运行内存性能分析...$(RESET)"
	go test ./test -bench=. -memprofile=mem.prof
	@echo "$(GREEN)✓ 内存性能分析完成: mem.prof$(RESET)"

# 构建相关
.PHONY: build
build: ## 构建项目
	@echo "$(BLUE)构建项目...$(RESET)"
	go build -v ./...
	@echo "$(GREEN)✓ 构建完成$(RESET)"

.PHONY: build-example
build-example: ## 构建示例程序
	@echo "$(BLUE)构建示例程序...$(RESET)"
	@if [ -d "examples" ]; then \
		go build -v ./examples/...; \
		echo "$(GREEN)✓ 示例程序构建完成$(RESET)"; \
	else \
		echo "$(YELLOW)⚠ examples目录不存在$(RESET)"; \
	fi

# 清理
.PHONY: clean
clean: ## 清理生成的文件
	@echo "$(BLUE)清理生成的文件...$(RESET)"
	go clean
	rm -f coverage.out coverage.html
	rm -f cpu.prof mem.prof
	rm -f *.test
	@echo "$(GREEN)✓ 清理完成$(RESET)"

# 开发工具
.PHONY: mod-init
mod-init: ## 初始化Go模块 (仅首次使用)
	@echo "$(BLUE)初始化Go模块...$(RESET)"
	go mod init $(PROJECT_NAME)
	@echo "$(GREEN)✓ Go模块初始化完成$(RESET)"

.PHONY: mod-verify
mod-verify: ## 验证依赖
	@echo "$(BLUE)验证项目依赖...$(RESET)"
	go mod verify
	@echo "$(GREEN)✓ 依赖验证通过$(RESET)"

.PHONY: mod-graph
mod-graph: ## 显示依赖图
	@echo "$(BLUE)显示依赖关系图...$(RESET)"
	go mod graph

# 文档生成
.PHONY: doc
doc: ## 生成并查看文档
	@echo "$(BLUE)生成项目文档...$(RESET)"
	godoc -http=:6060 &
	@echo "$(GREEN)✓ 文档服务启动: http://localhost:6060$(RESET)"
	@echo "$(YELLOW)按 Ctrl+C 停止文档服务$(RESET)"

# 完整检查流程
.PHONY: check
check: fmt vet test ## 执行完整的代码检查流程
	@echo "$(GREEN)✓ 所有检查通过$(RESET)"

.PHONY: ci
ci: deps fmt vet test-race test-cover ## CI/CD流水线检查
	@echo "$(GREEN)✓ CI检查完成$(RESET)"

# 快速开发
.PHONY: dev
dev: deps fmt test ## 开发环境快速检查
	@echo "$(GREEN)✓ 开发环境检查完成$(RESET)"

# 发布准备
.PHONY: release-check
release-check: clean deps fmt vet lint test-race test-cover bench ## 发布前完整检查
	@echo "$(GREEN)✓ 发布检查完成$(RESET)"

# 安装开发工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "$(BLUE)安装开发工具...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "$(GREEN)✓ 开发工具安装完成$(RESET)"

# 项目信息
.PHONY: info
info: ## 显示项目信息
	@echo "$(BLUE)项目信息:$(RESET)"
	@echo "  名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  Go版本要求: $(GO_VERSION)+"
	@echo "  当前Go版本: $(shell go version)"
	@echo "  项目路径: $(shell pwd)"
	@echo "  模块路径: $(shell go list -m)"