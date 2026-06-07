# syntax=docker/dockerfile:1.7
#
# Market API 服务 Dockerfile
#
# 功能说明：
#   Market API 是市场数据的 HTTP 接入层，提供币种信息、交易对行情、K线数据、汇率等查询接口。
#   本 Dockerfile 用于构建 Market API 的容器镜像。
#
# 架构设计：
#   - 采用多阶段构建（Multi-stage Build），分离编译环境和运行环境
#   - 构建上下文为仓库根目录，以便访问共享模块和跨服务包
#   - 使用 distroless 镜像作为运行基础，最小化镜像体积和安全风险
#
# 构建流程：
#   1. 第一阶段（builder）：使用 Go 官方镜像编译服务
#   2. 第二阶段（runtime）：复制编译产物到最小化运行镜像
#
# 使用方式：
#   docker build -t mscoin-go/market-api:local \
#     --build-arg GO_IMAGE=golang:1.26.3 \
#     -f app/market/api/Dockerfile .
#
# 注意事项：
#   - 构建必须在仓库根目录执行，因为需要访问 pkg/ 和 idl/ 目录
#   - CGO_ENABLED=0 确保生成静态链接的二进制文件
#   - -ldflags="-s -w" 去除调试信息，减小二进制体积

# ============================================================================
# 第一阶段：构建阶段（Builder Stage）
# ============================================================================

# 定义 Go 编译镜像的基础版本
# 可通过 --build-arg 覆盖，例如：--build-arg GO_IMAGE=golang:1.24
ARG GO_IMAGE=golang:1.26.3

# 使用 Go 官方镜像作为构建环境
# 该镜像包含完整的 Go 工具链（编译器、链接器等）
FROM ${GO_IMAGE} AS builder

# 定义目标操作系统和架构
# TARGETOS: 目标操作系统（默认 linux）
# TARGETARCH: 目标架构（默认 amd64，可指定 arm64 等）
# 这些参数用于跨平台编译（如 Mac 编译 Linux/amd64 镜像）
ARG TARGETOS=linux
ARG TARGETARCH

# 设置工作目录
# 所有后续操作都在 /workspace 目录下进行
WORKDIR /workspace

# --- 步骤 1：复制依赖定义文件 ---
# 仅复制 go.mod、go.sum、go.work、go.work.sum
# 这样可以利用 Docker 缓存机制：依赖不变时跳过重新下载
# go.work 和 go.work.sum 用于多模块工作区管理
COPY go.mod go.sum go.work go.work.sum ./

# --- 步骤 2：下载 Go 模块依赖 ---
# 使用 BuildKit 的缓存挂载，加速后续构建
# --mount=type=cache,target=/go/pkg/mod：缓存 Go 模块到指定目录
# 这样即使重新构建，已下载的模块也会被保留
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# --- 步骤 3：复制源代码 ---
# 复制所有必要的源代码目录
# app/: 业务应用代码（market、exchange、ucenter、jobcenter）
# idl/: 接口定义语言文件（.api 和 .proto）
# pkg/: 公共基础设施包（auth、btcx、redisx、mysqlx 等）
COPY app ./app
COPY idl ./idl
COPY pkg ./pkg

# --- 步骤 4：编译服务 ---
# 使用 BuildKit 缓存加速编译
# --mount=type=cache,target=/go/pkg/mod：模块缓存
# --mount=type=cache,target=/root/.cache/go-build：编译缓存
#
# 编译参数说明：
#   CGO_ENABLED=0：禁用 CGO，生成纯静态二进制文件
#   GOOS=${TARGETOS}：目标操作系统
#   GOARCH=${TARGETARCH:-amd64}：目标架构，默认 amd64
#   -trimpath：去除文件路径信息，增强可移植性
#   -ldflags="-s -w"：
#     -s：去除符号表（symbol table）
#     -w：去除 DWARF 调试信息
#     两者结合可减小二进制体积约 30%
#   -o /out/server：输出路径
#   ./app/market/api：编译目标（main.go 所在目录）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/market/api

# ============================================================================
# 第二阶段：运行阶段（Runtime Stage）
# ============================================================================

# 使用 Google 的 distroless 镜像作为运行基础
# distroless/static-debian12:nonroot 特点：
#   - 极小体积（约 2MB），仅包含必要的运行时组件
#   - 无 shell、无包管理器，减少安全攻击面
#   - nonroot 标签表示以非 root 用户运行，增强安全
#   - 基于 Debian 12，提供最新的安全补丁
FROM gcr.io/distroless/static-debian12:nonroot

# 设置工作目录
# 服务将在 /app 目录下运行
WORKDIR /app

# --- 步骤 5：复制编译产物 ---
# 从 builder 阶段复制编译好的二进制文件
# --from=builder：指定从哪个阶段复制
COPY --from=builder /out/server /app/server

# --- 步骤 6：复制配置文件 ---
# 复制服务的配置文件目录
# etc/ 目录包含 market-api.yaml 配置文件
COPY --from=builder /workspace/app/market/api/etc /app/etc

# --- 步骤 7：设置启动命令 ---
# ENTRYPOINT：容器启动时执行的主命令
# 这里直接执行编译好的服务二进制文件
ENTRYPOINT ["/app/server"]

# CMD：ENTRYPOINT 的默认参数
# 这里传入配置文件路径参数 -f /app/etc/market-api.yaml
# 可在运行时覆盖，例如：docker run ... -f /path/to/custom.yaml
CMD ["-f", "/app/etc/market-api.yaml"]