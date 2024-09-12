FROM ubuntu:22.04

# 设置镜像维护者信息
MAINTAINER https://github.com/ByteTu

# 更新包列表
RUN apt-get update && apt-get install -y libicu-dev

# 工作目录
WORKDIR /goApp

# 将二进制文件复制到工作目录
COPY . .

# 容器暴露的端口
EXPOSE 5050

# 启动容器时运行的命令
CMD ["./Api-N_m3u8DL"]