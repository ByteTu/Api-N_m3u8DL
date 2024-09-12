package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Index 首页
func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "恭喜你服务启动成功,这是首页!!!"})
}

type ReParamsReq struct {
	Must         string `json:"must"`           // 目前仅支持url
	SaveDir      string `json:"save-dir"`       // 视频保存目录
	SaveName     string `json:"save-name"`      // 视频名称
	BaseUrl      string `json:"base-url"`       // 设置BaseURL
	MaxSpeed     string `json:"max-speed"`      // 设置限速 单位支持 Mbps 或 Kbps，如：15M 100K
	DelAfterDone string `json:"del-after-done"` // 完成后删除临时文件  默认: False   F不删除 T删除
	BinaryMerge  string `json:"binary-merge"`   // 二进制合并 默认: False

}

// DownloadVideo 下载视频
func DownloadVideo(c *gin.Context) {
	// 获取不到数据
	var reqData ReParamsReq
	// 尝试绑定接收到的JSON到jsonData变量
	if err := c.BindJSON(&reqData); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "参数获取异常"})
		return
	}
	if reqData.SaveDir == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "save-dir参数不能为空"})
		return
	}
	if reqData.SaveName == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "save-name参数不能为空"})
		return
	}
	// 获取系统版本
	// 根据操作系统选择可执行文件路径
	var rePath string
	sys_os := runtime.GOOS
	switch sys_os {
	case "windows":
		rePath = "./N_m3u8DL-RE.exe"
	case "linux", "darwin":
		rePath = "./N_m3u8DL-RE"
	default:
		log.Fatalf("不支持的操作系统: %s", sys_os)
		return
	}
	cmd := exec.Command(rePath, reqData.Must)
	defaultPath := "./data/download/" // 默认下载目录
	tempPath := "./data/temp_dir/"    // 临时文件目录
	// 临时文件存储
	cmd.Args = append(cmd.Args, "--tmp-dir", tempPath)
	cmd.Args = append(cmd.Args, "--save-dir", defaultPath+reqData.SaveDir)
	cmd.Args = append(cmd.Args, "--save-name", reqData.SaveName)

	if reqData.BaseUrl != "" {
		cmd.Args = append(cmd.Args, "--base-url", reqData.BaseUrl)
	}
	if reqData.MaxSpeed != "" {
		cmd.Args = append(cmd.Args, "--max-speed", reqData.MaxSpeed)
	}
	// 如果为True删除临时下载目录
	if reqData.DelAfterDone != "True" {
		cmd.Args = append(cmd.Args, "--del-after-done", "False")
	}
	// 如果为True使用二进制合并
	if reqData.BinaryMerge == "True" {
		cmd.Args = append(cmd.Args, "--binary-merge", "True")
	}
	// 后台执行命令
	if err := cmd.Start(); err != nil {
		log.Fatalf("命令启动错误: %s\n", err)
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "命令已在后台启动，你可以继续执行其他任务。", "console": cmd.String()})
	return
}

// getCurrentNetworkIP 获取当前服务ip
func getCurrentNetworkIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func main() {
	addr := ":5050"
	serverIP, _ := getCurrentNetworkIP()
	fmt.Println("服务启动中ing.......")
	fmt.Println("服务器地址: http://" + serverIP + addr)
	// 启动服务
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// 设置HTML模板文件目录
	router.GET("/", Index)
	router.POST("/download", DownloadVideo)
	// 启动 Gin 服务器 监听端口
	err := router.Run(addr)
	if err != nil {
		fmt.Println("程序出现异常:" + err.Error())
		fmt.Println("程序将在5秒后自动退出.....")
		time.Sleep(6 * time.Second)
	}
}

// # 设置环境变量
// export GOOS=linux
//# 编译程序
// go build -o Api-N_m3u8DL main.go
// go build -o Api-N_m3u8DL.exe main.go
//
// set CGO_ENABLED=0
// set GOOS=linux
// set GOARCH=amd64
