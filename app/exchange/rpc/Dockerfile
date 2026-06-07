# syntax=docker/dockerfile:1.7
#
# Exchange RPC 服务 Dockerfile
#
# 功能说明：
#   Exchange RPC 是交易订单的 gRPC 服务层，提供订单创建、订单查询、订单取消等 RPC 方法。
#   本 Dockerfile 用于构建 Exchange RPC 的容器镜像。
#
# 架构设计：
#   - 采用多阶段构建（Multi-stage Build），分离编译环境和运行环境
#   - 构建上下文为仓库根目录，以便访问共享模块和 Proto 生成的契约包
#   - 使用 distroless 镜像作为运行基础，最小化镜像体积和安全风险
#
# 构建流程：
#   1. 第一阶段（builder）：使用 Go 官方镜像编译服务
#   2. 第二阶段（runtime）：复制编译产物到最小化运行镜像
#
# 使用方式：
#   docker build -t mscoin-go/exchange-rpc:local \
#     --build-arg GO_IMAGE=golang:1.26.3 \
#     -f app/exchange/rpc/Dockerfile .
#
# 注意事项：
#   - 构建必须在仓库根目录执行，因为需要访问 pkg/ 和 idl/ 目录
#   - CGO_ENABLED=0 确保生成静态链接的二进制文件

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
# ./app/exchange/rpc：编译目标
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/exchange/rpc

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
COPY --from=builder /workspace/app/exchange/rpc/etc /app/etc

# --- 步骤 7：设置启动命令 ---
ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/exchange.yaml"]