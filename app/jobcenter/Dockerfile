# syntax=docker/dockerfile:1.7
#
# Jobcenter 服务 Dockerfile
#
# 功能说明：
#   Jobcenter 是异步任务处理中心，负责：
#   - 消费 Kafka 提现事件，执行 BTC 提现
#   - 定时同步 OKX 汇率和 K 线数据
#   - 定时刷新最新价格缓存
#
#   本 Dockerfile 用于构建 Jobcenter 的容器镜像。
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
# 服务依赖：
#   - MySQL：查询和更新提现记录状态
#   - Redis：缓存 txid 检查点、最新价格
#   - MongoDB：存储 K 线历史数据
#   - Kafka：消费提现事件
#   - Market RPC：查询币种信息
#   - Ucenter RPC：查询钱包地址
#   - Bitcoin Core：执行 BTC 提现交易
#
# 使用方式：
#   docker build -t mscoin-go/jobcenter:local \
#     --build-arg GO_IMAGE=golang:1.26.3 \
#     -f app/jobcenter/Dockerfile .
#
# 注意事项：
#   - 构建必须在仓库根目录执行，因为需要访问 pkg/ 和 idl/ 目录
#   - Jobcenter 不对外暴露端口，仅作为后台任务运行
#   - Jobcenter 需要等待 bitcoin-wallet-init 完成后才能启动

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
# 编译参数说明：
#   CGO_ENABLED=0：禁用 CGO，生成纯静态二进制文件
#   -trimpath：去除文件路径信息
#   -ldflags="-s -w"：去除符号表和调试信息
#   ./app/jobcenter：编译目标
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/server ./app/jobcenter

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
COPY --from=builder /workspace/app/jobcenter/etc /app/etc

# --- 步骤 7：设置启动命令 ---
ENTRYPOINT ["/app/server"]
CMD ["-f", "/app/etc/jobcenter.yaml"]