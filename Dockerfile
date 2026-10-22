# ============ 构建阶段：交叉编译 Go 服务 ============
# CGO_ENABLED=0 产出纯静态二进制，因此运行阶段不必与构建阶段同发行版，
# amd64 / arm64 可以共用同一个运行阶段，只需在编译时切换 GOARCH。
FROM golang:1.22-bookworm AS builder

ARG TARGETARCH

WORKDIR /src

# 先只拷依赖清单，利用 Docker 层缓存：源码改动不必重新下载模块
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并按 buildx 注入的目标架构交叉编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
      go build -trimpath -ldflags="-s -w" -o /out/Api-N_m3u8DL .

# ============ 运行阶段 ============
# N_m3u8DL-RE 官方 Linux 二进制按 glibc 动态链接，所以运行阶段用 glibc 系基础镜像
# （alpine 是 musl，直接跑会报 "no such file or directory"）。
FROM debian:bookworm-slim AS runner

ARG TARGETARCH
# N_m3u8DL-RE 的版本 tag，升级只需改这里（或由 workflow 的 build-args 传入）
ARG N_M3U8DL_VERSION=v0.6.0-beta

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    TZ=Asia/Shanghai

# ffmpeg：N_m3u8DL-RE 合并音视频轨（mp4/mkv 混流）需要
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
         ca-certificates \
         curl \
         ffmpeg \
         tzdata \
    && rm -rf /var/lib/apt/lists/*

# 下载与当前架构匹配的 N_m3u8DL-RE 二进制。
# release 资产命名：N_m3u8DL-RE_<版本>_linux-<x64|arm64>_<构建日期>.tar.gz，
# 解压后是单个可执行文件 N_m3u8DL-RE；构建日期后缀会变，所以从 GitHub API
# 的 assets 列表里挑出匹配当前架构的那个，取日期最新的一个。
# 只认 "browser_download_url" 字段：release 说明正文的 Markdown 里也有
# 形似的下载链接，不能直接全文 grep。
RUN set -eux; \
    case "${TARGETARCH}" in \
      amd64) RE_ARCH="x64" ;; \
      arm64) RE_ARCH="arm64" ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    URL="$(curl -fsSL --retry 3 --retry-delay 2 \
        "https://api.github.com/repos/nilaoda/N_m3u8DL-RE/releases/tags/${N_M3U8DL_VERSION}" \
      | grep -oE '"browser_download_url": "https://[^"]*_linux-'${RE_ARCH}'_[0-9]+\.tar\.gz"' \
      | grep -oE 'https://[^"]*\.tar\.gz' | tail -n 1)"; \
    if [ -z "$URL" ]; then \
      echo "N_m3u8DL-RE ${N_M3U8DL_VERSION} 没有 linux-${RE_ARCH} 的 release 资产" >&2; \
      exit 1; \
    fi; \
    echo "下载 N_m3u8DL-RE: $URL"; \
    curl -fsSL --retry 3 --retry-delay 2 -o /tmp/re.tar.gz "$URL"; \
    mkdir -p /tmp/re; \
    tar -xzf /tmp/re.tar.gz -C /tmp/re; \
    rm -f /tmp/re.tar.gz

WORKDIR /goApp

# 主程序 + 前端资源。N_m3u8DL-RE 放在工作目录下：
# main.go 里按 ./N_m3u8DL-RE 相对路径调用
RUN install -m 0755 /tmp/re/N_m3u8DL-RE /goApp/N_m3u8DL-RE \
    && ln -sf /goApp/N_m3u8DL-RE /usr/local/bin/N_m3u8DL-RE \
    && rm -rf /tmp/re

COPY --from=builder /out/Api-N_m3u8DL ./Api-N_m3u8DL
COPY templates ./templates
COPY static ./static

# 数据/日志目录（建议挂载卷）：
#   /goApp/data/videos   下载文件
#   /goApp/data/temp_dir 临时分片
#   /goApp/Logs          N_m3u8DL-RE 运行日志
EXPOSE 5050

CMD ["./Api-N_m3u8DL"]
