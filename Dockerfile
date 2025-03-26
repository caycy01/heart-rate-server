ARG DOCKER_URL=docker.xihan.website
ARG GO_VERSION=1.24
FROM --platform=$BUILDPLATFORM $DOCKER_URL/golang:${GO_VERSION} AS build
WORKDIR /src

ENV HTTP_PROXY=http://192.168.43.113:20808
ENV HTTPS_PROXY=http://192.168.43.113:20808
ENV ALL_PROXY=socks5://192.168.43.113:20808

# 设置 Go 的自定义镜像源
ENV GOPROXY=https://mirrors.xihan.website/repository/go/
ENV GOSUMDB=sum.golang.org

# 复制 Go 源代码到容器中
COPY . .

# 下载依赖作为单独的步骤，以利用 Docker 的缓存
# 使用缓存挂载 /go/pkg/mod/ 以加快后续构建
# 使用绑定挂载 go.sum 和 go.mod 以避免将它们复制到容器中
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# 这是构建目标的架构，由构建器传递
# 将它放在这里允许在不同架构之间缓存前面的步骤
ARG TARGETARCH

# 构建应用程序
# 使用缓存挂载 /go/pkg/mod/ 以加快后续构建
# 使用绑定挂载当前目录以避免将源代码复制到容器中
RUN --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server .

################################################################################
# 创建一个新的阶段来运行应用程序，它包含应用程序的最小运行时依赖
# 这通常使用与构建阶段不同的基础镜像，其中必要文件从构建阶段复制
FROM $DOCKER_URL/alpine:latest AS final

# 安装运行应用程序所需的任何运行时依赖
# 使用缓存挂载 /var/cache/apk/ 以加快后续构建
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.cernet.edu.cn/alpine#g' /etc/apk/repositories && \
    apk --update add \
        ca-certificates \
        musl-locales \
        tzdata \
        && \
        update-ca-certificates

# 创建应用程序将运行的非特权用户
# 请参阅 https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

# 从 "build" 阶段复制可执行文件
COPY --from=build /bin/server /bin/

# 暴露应用程序监听的端口
EXPOSE 8080

# 容器启动时运行的命令
ENTRYPOINT [ "/bin/server" ]
