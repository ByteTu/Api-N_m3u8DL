# Api-N_m3u8DL

[![GitHub release](https://img.shields.io/github/v/release/ByteTu/Api-N_m3u8DL?include_prereleases)](https://github.com/ByteTu/Api-N_m3u8DL/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/isouu/api_m3u8dl)](https://hub.docker.com/r/isouu/api_m3u8dl)
[![Go Version](https://img.shields.io/badge/Go-1.22.5-00ADD8?logo=go)](https://go.dev/)

基于 [N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE) 的 m3u8 视频下载服务，提供 API 接口和 Web 管理界面，支持任务管理、实时日志、暂停/恢复/终止等完整功能。

适合爬虫批量下载、自动化下载场景使用。

如果对你有帮助请帮忙 Star 一下 ⭐

## 功能特性

- **API 调用** — 通过 HTTP POST 接口提交下载任务，支持所有 N_m3u8DL-RE 命令行参数
- **Web 管理界面** — 可视化操作页面，实时查看下载进度与日志
- **任务管理** — 支持暂停、恢复、终止下载任务，任务历史记录持久化
- **批量验证** — 自动检测任务进程状态，修复残留的运行标记
- **日志系统** — 实时日志面板 + 日志文件查看器，支持时间筛选与增量获取
- **文件管理** — 查看已下载的视频目录与文件列表
- **Docker 部署** — 提供官方 Docker 镜像

## 版本记录

| 版本 | 日期 | 更新内容 |
| :--- | :--- | :--- |
| **v3.0** | 2026/04/18 | 升级 N_m3u8DL-RE v0.5.1-beta；重构下载调用逻辑；重构 Web 界面；新增暂停/恢复/终止任务；新增任务历史与批量验证；新增日志文件查看器与已下载文件管理 |
| **v2.0** | 2024/09/12 | 移除配置文件与 CLI 版本支持；升级 N_m3u8DL-RE v0.2.1-beta；支持跨平台/Docker/Web/API |
| **v1.1** | 2023/11/14 | 修复特殊字符 `&` 引起的问题 |
| **v1.0** | 2023/11/06 | 发布版本；升级 N_m3u8DL-RE v0.2.0-beta；仅支持 Windows |

## 快速开始

### 下载发行版

**GitHub Releases:** https://github.com/ByteTu/Api-N_m3u8DL/releases

> 需要将 `N_m3u8DL-RE` 可执行文件和 `ffmpeg` 放在同一目录下运行。

### Docker 部署

```shell
# 拉取镜像
docker pull isouu/api_m3u8dl:v3.0

# 启动容器
docker run --restart=always --name api_m3u8dl \
  -p 5050:5050 \
  -v /opt/api_m3u8dl/data:/goApp/data \
  -v /opt/api_m3u8dl/logs:/goApp/Logs \
  -d isouu/api_m3u8dl:v3.0
```

> 注意：需要在容器中放置 `N_m3u8DL-RE` (Linux版) 和 `ffmpeg` (Linux版) 可执行文件，可通过 volume 挂载或额外 COPY。

| 说明 | 容器路径 |
| :--- | :--- |
| 服务端口 | `5050` |
| N_m3u8DL-RE 运行日志 | `/goApp/Logs` |
| 下载文件保存路径 | `/goApp/data/videos` |
| 临时文件保存路径 | `/goApp/data/temp_dir` |

### 本地编译

```shell
# 需要 Go 1.22.5+
go build -o Api-N_m3u8DL .
```

## 界面预览

![首页](/Img/image01.png "Web 管理界面")
![API 文档](/Img/image02.png "API 文档弹窗")

## API 接口

服务启动后默认监听 `5050` 端口，所有接口返回 JSON 格式数据。

---

### 新建下载任务

```
POST /download
```

#### 请求参数

**必填参数：**

| 参数 | 类型 | 说明 |
| :--- | :--- | :--- |
| `must` | string | m3u8 视频 URL 地址 |
| `save-dir` | string | 保存视频目录（在 `/data/videos/` 下创建） |
| `save-name` | string | 保存视频名称 |

**基础选项：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `base-url` | string | | 设置 BaseURL，用于相对路径拼接 |
| `max-speed` | string | 不限速 | 限速，如 `15M`、`100K` |
| `thread-count` | int | CPU线程数 | 下载线程数 |
| `header` | string | | 自定义 HTTP 请求头 |

**合并选项：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `binary-merge` | bool | `false` | 二进制合并 |
| `use-ffmpeg-concat-demuxer` | bool | `false` | 使用 ffmpeg concat demuxer 合并 |
| `del-after-done` | bool | `true` | 完成后删除临时文件 |
| `no-date-info` | bool | `false` | 混流时不写入日期信息 |
| `write-meta-json` | bool | `false` | 输出 meta JSON 文件 |

**高级选项：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `skip-merge` | bool | `false` | 跳过合并（仅下载分片） |
| `skip-download` | bool | `false` | 跳过下载（仅探测信息） |
| `check-segments-count` | bool | `false` | 检测分片数量一致性 |
| `auto-select` | bool | `false` | 自动选择最佳轨道 |
| `no-ansi-color` | bool | `false` | 去除 ANSI 颜色 |
| `no-log` | bool | `false` | 关闭日志文件输出 |
| `sub-only` | bool | `false` | 仅选取字幕轨道 |
| `append-url-params` | bool | `false` | 将 URL 参数追加到分片 |
| `concurrent-download` | bool | `false` | 并发下载音视频字幕 |

**流选择：**

| 参数 | 类型 | 说明 |
| :--- | :--- | :--- |
| `select-video` | string | 正则选择视频流，如 `.*1080.*` |
| `select-audio` | string | 正则选择音频流，如 `.*128kbps.*` |
| `select-subtitle` | string | 正则选择字幕流，如 `.*zh.*` |

**字幕选项：**

| 参数 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `sub-format` | string | `SRT` | 字幕格式：`SRT` 或 `VTT` |
| `auto-subtitle-fix` | bool | `false` | 自动修正字幕 |

**网络选项：**

| 参数 | 类型 | 说明 |
| :--- | :--- | :--- |
| `custom-proxy` | string | 自定义代理，如 `http://127.0.0.1:8888` |
| `custom-range` | string | 下载分片范围，如 `bytes=0-1000000` |
| `task-start-at` | string | 任务开始时间，格式 `20241201123000` |
| `http-timeout` | int | HTTP 超时秒数，默认 100 |
| `download-retry` | int | 分片下载重试次数，默认 3 |

**其他选项：**

| 参数 | 类型 | 说明 |
| :--- | :--- | :--- |
| `ffmpeg-binary-path` | string | ffmpeg 程序路径 |
| `decryption-binary-path` | string | mp4decrypt 程序路径 |
| `ui-language` | string | UI 语言：`zh-CN` / `zh-TW` / `en-US` |
| `log-level` | string | 日志级别：`DEBUG` / `INFO` / `WARN` / `ERROR` / `OFF` |
| `disable-update` | bool | 禁用版本更新检测 |

#### 请求示例

```json
POST http://127.0.0.1:5050/download

{
  "must": "https://example.com/video.m3u8",
  "save-name": "视频文件名称",
  "save-dir": "download"
}
```

#### 返回示例

```json
{
  "code": 0,
  "msg": "任务已在后台启动，日志面板实时显示下载进度",
  "console": "./N_m3u8DL-RE.exe https://example.com/video.m3u8 --tmp-dir ./data/temp_dir/ --save-dir ./data/videos/download --save-name 视频文件名称 --log-file-path ./Logs/视频文件名称_20260418_120000.log",
  "taskId": "T000001",
  "pid": 12345
}
```

---

### 查询任务状态

```
GET /task-status?taskId=T000001
```

返回指定任务的运行状态、退出码和实时日志。

---

### 获取任务列表

```
GET /task-list
```

返回所有正在运行的任务列表。

---

### 获取任务日志（增量）

```
GET /task-logs?taskId=T000001&offset=0
```

支持增量获取，`offset` 为上次获取的总行数。返回 `total`、`offset`、`logs`、`running`。

---

### 暂停任务

```
POST /task-pause?taskId=T000001
```

---

### 恢复任务

```
POST /task-resume?taskId=T000001
```

---

### 终止任务

```
POST /task-stop?taskId=T000001
```

---

### 获取任务历史

```
GET /api/task-history
```

返回所有历史任务记录（含已完成、失败、已终止），按时间倒序。

---

### 批量验证任务状态

```
POST /api/task-verify
Content-Type: application/json

{
  "taskIds": ["T000001", "T000002"]
}
```

批量检查任务进程是否存活，自动修复残留的 `running` 状态。

---

### 清空已完成的历史记录

```
POST /api/task-history/clear-done
```

---

### 获取日志文件列表

```
GET /api/logs
```

返回 `Logs` 目录下的日志文件列表，按修改时间倒序。

---

### 获取日志文件内容

```
GET /api/logs/{filename}
```

返回指定日志文件的完整内容，自动处理 GBK/UTF-8 编码。

---

### 清空旧日志

```
POST /api/logs/clear-old
```

删除超过 1 天的日志文件。

---

### 获取已下载视频目录

```
GET /api/videos/dirs
```

返回 `data/videos` 下的所有子目录。

---

### 获取已下载视频文件

```
GET /api/videos/files?dir=download
```

返回指定子目录下的视频文件列表（`.mp4` `.mkv` `.avi` `.flv` `.mov` `.ts` `.webm` 等格式）。

---

### 获取版本号

```
GET /version
```

返回当前服务版本。

```json
{
  "version": "v3.0"
}
```

## 目录结构

```
.
├── main.go              # 主程序（路由 + 下载逻辑 + 任务管理）
├── process_windows.go   # Windows 进程管理（暂停/恢复/终止/检测）
├── process_linux.go     # Linux进程管理
├── go.mod / go.sum      # Go 依赖管理
├── Dockerfile           # Docker 构建文件
├── templates/
│   └── index.html       # Web 界面
├── static/
│   ├── css/             # 样式文件
│   └── js/              # JavaScript (jQuery)
├── Img/                 # 截图资源
├── data/
│   ├── videos/          # 下载文件保存目录
│   ├── temp_dir/        # 临时文件目录
│   └── task_history.json # 任务历史记录
└── Logs/                # N_m3u8DL-RE 运行日志
```

## 感谢

- **[N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE)** — 强大的 m3u8 下载器