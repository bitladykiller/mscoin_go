# syntax=docker/dockerfile:1.7
#
# Ucenter RPC 服务 Dockerfile
#
# 功能说明：
#   Ucenter RPC 是用户中心的 gRPC 服务层，提供用户注册、登录、钱包管理、提现申请等 RPC 方法。
#   本 Dockerfile 用于构建 Ucenter RPC 的容器镜像。
#
# 架构设计：
#   - 采用多阶段构建（Multi-stage Build），分离编译环境和运行环境
#   - 构建上下文为仓库根目录，以便访问共享模块和 Proto 生成的契约包
#   - 使用 distroless 镜像作为运行基础，最小化镜像体积和安全风险
#
# 服务依赖：
#   - MySQL：存储会员、钱包、交易记录、提现记录
#   - Redis：缓存验证码、提现验证码
#   - Kafka：发布提现事件
#   - Market RPC：查询币种信息
#   - Bitcoin Core：分配 BTC 地址
#
# 使用方式：
#   docker build -t mscoin-go/ucenter-rpc:local \
#     --build-arg GO_IMAGE=golang:1.26.3 \
#     -f app/ucenter/rpc/Dockerfile .
#
# 注意事项：
#   - 构建必须在仓库根目录执行，因为需要访问 pkg/ 和 idl/ 目录
#   - Ucenter RPC 需要等待 bitcoin-wallet-init 完成后才能启动

# ============================================================================
# 第一阶段：构建阶段（Builder Stage）
# ============================================================================

# 定义 Go 编译镜像的基础版本
ARG GO_IMAGE=golang:1.26.3

# 使用 Go 官方镜像作为构建环境
FROM ${GO_IMAGE} AS builder

# 定义目标操作系统和架构
ARG TARGETOS=linux
ARG TARGETARCH

# 设置工作目录
WORKDIR /workspace

# --- 步骤 1：复制依赖定义文件 ---
COPY go.mod go.sum go.work go.work.sum ./

# --- 步骤 2：下载 Go 模块依赖 ---
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# --- 步骤 3：复制源代码 ---
COPY app ./app
COPY idl ./idl
COPY pkg ./pkg

# --- 步骤 4：编译服务 ---
# CGO_ENABLED=0：禁用 CGO，生成纯静态二进制
# -trimpath：去除文件路径信息
# -ldflags="-s -w"：去除符号表和调试信息
# ./app/ucenter/rpc：编译目标
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/ucenter/rpc

# ============================================================================
# 第二阶段：运行阶段（Runtime Stage）
# ============================================================================

# 使用 Google 的 distroless 镜像作为运行基础
FROM gcr.io/distroless/static-debian12:nonroot

# 设置工作目录
WORKDIR /app

# --- 步骤 5：复制编译产物 ---
COPY --from=builder /out/server /app/server

# --- 步骤 6：复制配置文件 ---
COPY --from=builder /workspace/app/ucenter/rpc/etc /app/etc

# --- 步骤 7：设置启动命令 ---
ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/ucenter.yaml"]