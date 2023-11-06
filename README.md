# Api-N_m3u8DL

**一个简单易用的m3u8下载器**

**实现通过Api调用N_m3u8DL-CLI和N_m3u8DL-RE进行下载**

**适合爬虫批量下载或实现自动化下载使用**

# 下载使用

**发行版:** https://github.com/ByteTu/Api-N_m3u8DL/releases

**自己编译:**
``` 
git clone https://github.com/ByteTu/Api-N_m3u8DL.git
cd Api-N_m3u8DL
go build -o Api-N_m3u8DL.exe
``` 

# 使用教程

## 注意

- **请先设置好 api_config.toml 文件内cli_path和re_path路径信息**

## 必传参数

| 参数   | 内容                                   |
|------|--------------------------------------|
| app  | cli:调用N_m3u8DL-CLI; re:调用N_m3u8DL-RE |
| must | 视频地址<链接或文件内容>                        |

## 请求示例

``` 

POST http://127.0.0.1:5000/download

# 调用N_m3u8DL-CLI
{
  "app": "cli",
  "must": "https://qq.com/qq.m3u8",
  "--saveName": "这是名称",
  "--workDir": "这是保存路径"
}
# 调用N_m3u8DL-RE
{
  "app": "re",
  "must": "https://qq.com/qq.m3u8",
  "--save-name": "这是名称",
  "--save-dir": "这是保存路径"
}
```

## --请求参数说明

**非必传请求参数请根据官方命令行选项进行添加** </br>
**[https://github.com/nilaoda/N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE)** </br>
**[https://github.com/nilaoda/N_m3u8DL-CLI](https://github.com/nilaoda/N_m3u8DL-CLI)**

## N_m3u8DL-CLI 命令行选项

```
OPTIONS:
  --workDir                  设定程序工作目录
  --saveName                 设定存储文件名(不包括后缀)
  --baseUrl                  设定Baseurl
  --headers                  设定请求头，格式 key:value 使用|分割不同的key&value
  --maxThreads               (Default: 32) 设定程序的最大线程数
  --minThreads               (Default: 16) 设定程序的最小线程数
  --retryCount               (Default: 15) 设定程序的重试次数
  --timeOut                  (Default: 10) 设定程序网络请求的超时时间(单位为秒)
  --muxSetJson               使用外部json文件定义混流选项
  --useKeyFile               使用外部16字节文件定义AES-128解密KEY
  --useKeyBase64             使用Base64字符串定义AES-128解密KEY
  --useKeyIV                 使用HEX字符串定义AES-128解密IV
  --downloadRange            仅下载视频的一部分分片或长度
  --liveRecDur               直播录制时，达到此长度自动退出软件(HH:MM:SS)
  --stopSpeed                当速度低于此值时，重试(单位为KB/s)
  --maxSpeed                 设置下载速度上限(单位为KB/s)
  --proxyAddress             设置HTTP/SOCKS5代理, 如 http://127.0.0.1:8080
  --enableDelAfterDone       开启下载后删除临时文件夹的功能
  --enableMuxFastStart       开启混流mp4的FastStart特性
  --enableBinaryMerge        开启二进制合并分片
  --enableParseOnly          开启仅解析模式(程序只进行到meta.json)
  --enableAudioOnly          合并时仅封装音频轨道
  --disableDateInfo          关闭混流中的日期写入
  --disableIntegrityCheck    不检测分片数量是否完整
  --noMerge                  禁用自动合并
  --noProxy                  不自动使用系统代理
  --registerUrlProtocol      注册m3u8dl链接协议
  --unregisterUrlProtocol    取消注册m3u8dl链接协议
  --enableChaCha20           enableChaCha20
  --chaCha20KeyBase64        ChaCha20KeyBase64
  --chaCha20NonceBase64      ChaCha20NonceBase64
  --help                     Display this help screen.
  --version                  Display version information.
```

## N_m3u8DL-RE 命令行选项

```
Options:
  --tmp-dir <tmp-dir>                      设置临时文件存储目录
  --save-dir <save-dir>                    设置输出目录
  --save-name <save-name>                  设置保存文件名
  --base-url <base-url>                    设置BaseURL
  --thread-count <number>                  设置下载线程数 [default: 16]
  --download-retry-count <number>          每个分片下载异常时的重试次数 [default: 3]
  --auto-select                            自动选择所有类型的最佳轨道 [default: False]
  --skip-merge                             跳过合并分片 [default: False]
  --skip-download                          跳过下载 [default: False]
  --check-segments-count                   检测实际下载的分片数量和预期数量是否匹配 [default: True]
  --binary-merge                           二进制合并 [default: False]
  --del-after-done                         完成后删除临时文件 [default: True]
  --no-date-info                           混流时不写入日期信息 [default: False]
  --no-log                                 关闭日志文件输出 [default: False]
  --write-meta-json                        解析后的信息是否输出json文件 [default: True]
  --append-url-params                      将输入Url的Params添加至分片, 对某些网站很有用, 例如 kakao.com [default: False]
  -mt, --concurrent-download               并发下载已选择的音频、视频和字幕 [default: False]
  -H, --header <header>                    为HTTP请求设置特定的请求头, 例如:
                                           -H "Cookie: mycookie" -H "User-Agent: iOS"
  --sub-only                               只选取字幕轨道 [default: False]
  --sub-format <SRT|VTT>                   字幕输出类型 [default: SRT]
  --auto-subtitle-fix                      自动修正字幕 [default: True]
  --ffmpeg-binary-path <PATH>              ffmpeg可执行程序全路径, 例如 C:\Tools\ffmpeg.exe
  --log-level <DEBUG|ERROR|INFO|OFF|WARN>  设置日志级别 [default: INFO]
  --ui-language <en-US|zh-CN|zh-TW>        设置UI语言
  --urlprocessor-args <urlprocessor-args>  此字符串将直接传递给URL Processor
  --key <key>                              设置解密密钥, 程序调用mp4decrpyt/shaka-packager进行解密. 格式:
                                           --key KID1:KEY1 --key KID2:KEY2
  --key-text-file <key-text-file>          设置密钥文件,程序将从文件中按KID搜寻KEY以解密.(不建议使用特大文件)
  --decryption-binary-path <PATH>          MP4解密所用工具的全路径, 例如 C:\Tools\mp4decrypt.exe
  --use-shaka-packager                     解密时使用shaka-packager替代mp4decrypt [default: False]
  --mp4-real-time-decryption               实时解密MP4分片 [default: False]
  -M, --mux-after-done <OPTIONS>           所有工作完成时尝试混流分离的音视频. 输入 "--morehelp mux-after-done" 以查看详细信息
  --custom-hls-method <METHOD>             指定HLS加密方式 (AES_128|AES_128_ECB|CENC|CHACHA20|NONE|SAMPLE_AES|SAMPLE_AES_CTR|UNKNOWN)
  --custom-hls-key <FILE|HEX|BASE64>       指定HLS解密KEY. 可以是文件, HEX或Base64
  --custom-hls-iv <FILE|HEX|BASE64>        指定HLS解密IV. 可以是文件, HEX或Base64
  --use-system-proxy                       使用系统默认代理 [default: True]
  --custom-proxy <URL>                     设置请求代理, 如 http://127.0.0.1:8888
  --custom-range <RANGE>                   仅下载部分分片. 输入 "--morehelp custom-range" 以查看详细信息
  --task-start-at <yyyyMMddHHmmss>         在此时间之前不会开始执行任务
  --live-perform-as-vod                    以点播方式下载直播流 [default: False]
  --live-real-time-merge                   录制直播时实时合并 [default: False]
  --live-keep-segments                     录制直播并开启实时合并时依然保留分片 [default: True]
  --live-pipe-mux                          录制直播并开启实时合并时通过管道+ffmpeg实时混流到TS文件 [default: False]
  --live-fix-vtt-by-audio                  通过读取音频文件的起始时间修正VTT字幕 [default: False]
  --live-record-limit <HH:mm:ss>           录制直播时的录制时长限制
  --live-wait-time <SEC>                   手动设置直播列表刷新间隔
  --mux-import <OPTIONS>                   混流时引入外部媒体文件. 输入 "--morehelp mux-import" 以查看详细信息
  -sv, --select-video <OPTIONS>            通过正则表达式选择符合要求的视频流. 输入 "--morehelp select-video" 以查看详细信息
  -sa, --select-audio <OPTIONS>            通过正则表达式选择符合要求的音频流. 输入 "--morehelp select-audio" 以查看详细信息
  -ss, --select-subtitle <OPTIONS>         通过正则表达式选择符合要求的字幕流. 输入 "--morehelp select-subtitle" 以查看详细信息
  -dv, --drop-video <OPTIONS>              通过正则表达式去除符合要求的视频流.
  -da, --drop-audio <OPTIONS>              通过正则表达式去除符合要求的音频流.
  -ds, --drop-subtitle <OPTIONS>           通过正则表达式去除符合要求的字幕流.
  --morehelp <OPTION>                      查看某个选项的详细帮助信息
  --version                                Show version information
  -?, -h, --help                           Show help and usage information
```
