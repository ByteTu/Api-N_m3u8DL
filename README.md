# Api-N_m3u8DL

## 关于

**`Api-N_m3u8DL`支持Api调用一个简单的m3u8下载器**

**`Api-N_m3u8DL`是实现通过调用`N_m3u8DL-RE`实现高速下载**

**适合爬虫批量下载或实现自动化下载场景使用**

**如果对你有帮助请帮忙 Star 一下**

## 版本记录

- **v2.0** `2024/09/02`
    - 重构下载调用代码逻辑
    - 升级版本N_m3u8DL-RE_v0.2.1-beta
    - 移除对N_m3u8DL-CLI版本支持
    - 移除配置文件支持
    - 实现跨平台调用

- **v1.1** `2023/11/14`
    - 修复特殊字符&引起问题

- **v1.0** `2023/11/06`
    - 发布版本
    - 升级版本N_m3u8DL-RE_v0.2.0-beta
    - 升级版本N_m3u8DL-CLI_v3.0.2
    - 仅支持 Windows


# 使用教程

**发行版:**

【github】 https://github.com/ByteTu/Api-N_m3u8DL/releases

**自己编译:**

``` 
git clone https://github.com/ByteTu/Api-N_m3u8DL.git
cd Api-N_m3u8DL
Windows：
go build -o Api-N_m3u8DL.exe
Liunx：
go build -o Api-N_m3u8DL
``` 

## 请求参数

| 参数           | 必选 | 类型   | 默认值 | 内容                                    |
| -------------- | :--- | :----- | ------ | --------------------------------------- |
| must           | 是   | string |        | m3u8视频url地址(暂不支持其他形式)       |
| save-dir       | 是   | string |        | 保存视频目录                            |
| save-name      | 是   | string |        | 保存视频名称                            |
| base-url       | 否   | string |        | 设置BaseURL                             |
| max-speed      | 否   | string | 不限速 | 设置限速仅支持例如：15M 100K            |
| binary-merge   | 否   | string | False  | 二进制合并 False不合并 True合并         |
| del-after-done | 否   | string | False  | 完成后删除临时文件 False不删除 True删除 |

## 请求示例

``` 
POST http://127.0.0.1:5050/download
# 调用N_m3u8DL-RE
{
    "must": "https://sf1-cdn-tos.huoshanstatic.com/obj/media-fe/xgplayer_doc_video/hls/xgplayer-demo.m3u8",
    "save-name": "这是视频文件名称",
    "save-dir": "这是目录"
}
```

## 返回示例

``` 
{
    "code": 0,
    "console": "./N_m3u8DL-RE.exe https://huoshanstatic.com/demo.m3u8 --tmp-dir ./data/temp_dir/ --save-dir ./data/download/这是目录 --save-name 这是视频文件&Adam&Edavn --del-after-done False",
    "msg": "命令已在后台启动，你可以继续执行其他任务。"
}
```



## N_m3u8DL-RE仅支持部分选项

```
Options:
  --tmp-dir <tmp-dir>                      设置临时文件存储目录
  --save-dir <save-dir>                    设置输出目录
  --save-name <save-name>                  设置保存文件名
  --base-url <base-url>                    设置BaseURL
  -R, --max-speed <SPEED>                  设置限速，单位支持 Mbps 或 Kbps，如：15M 100K
  --binary-merge                           二进制合并 [default: False]
  --del-after-done                         完成后删除临时文件 [default: True]
```


### 感谢作者

**[N_m3u8DL-CLI](https://github.com/nilaoda/N_m3u8DL-CLI)**

**[N_m3u8DL-RE](https://github.com/nilaoda/N_m3u8DL-RE)**
