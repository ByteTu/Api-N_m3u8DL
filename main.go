package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Index 测试
func Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "恭喜你访问成功,这是首页!"})
}

// DownloadVideo 下载视频
func DownloadVideo(c *gin.Context) {
	// 获取不到数据
	var requestData map[string]interface{}
	err := c.ShouldBind(&requestData)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "参数获取异常"})
		return
	}
	// 将请求数据转换为JSON字符串
	jsonStrByte, _ := json.Marshal(requestData)
	fmt.Println("请求参数--->", string(jsonStrByte))
	must, _ := requestData["must"].(string)
	app, _ := requestData["app"].(string)
	var exe_path string
	if app == "cli" {
		exe_path = viper.GetString("cli_path")
	} else {
		exe_path = viper.GetString("re_path")
	}
	// 解码 JSON 数据到 map
	var data map[string]interface{}
	json.Unmarshal(jsonStrByte, &data)
	delete(data, "must")
	delete(data, "app")
	// 创建一个空的 []string 来存储命令
	commands := []string{}
	// 遍历 JSON 对象中的字段
	for key, value := range data {
		commands = append(commands, key)
		// 根据字段类型添加字段值到命令
		switch v := value.(type) {
		case string:
			commands = append(commands, v)
		case bool:
			if v {
				commands = append(commands, "true")
			}
		}
	}
	// 下载终端是否自动退出
	out_commands := []string{must}
	out_commands = append(out_commands, commands...)
	auto_exit := viper.GetBool("auto_exit")
	var cmdArgs []string
	if auto_exit == true {
		cmdArgs = append([]string{"/C", "start", exe_path}, out_commands...)
	} else {
		cmdArgs = append([]string{"/C", "start", "cmd", "/K", exe_path}, out_commands...)
	}
	cmd := exec.Command("cmd", cmdArgs...)
	options_line := exe_path + " " + strings.Join(out_commands, " ")
	fmt.Println("调用命令--->", options_line)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 避免过快启动终端日志异常
	time.Sleep(400 * time.Millisecond)
	err_cmd := cmd.Run()
	if err_cmd != nil {
		fmt.Println(err_cmd)
	}
	c.JSON(200, gin.H{"code": 0, "msg": "运行成功", "cmd": options_line})
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

// InitConfig 初始化配置文件
func InitConfig() {
	viper.SetConfigFile("api_config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Fatal error config file", err)
	}
}

func main() {
	// 初始化配置文件
	InitConfig()
	addr := viper.GetString("port")
	serverIP, _ := getCurrentNetworkIP()
	fmt.Println("服务启动中ing.......")
	fmt.Println("服务器地址: http://" + serverIP + addr)
	//启动服务
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/", Index)
	router.POST("/download", DownloadVideo)
	// 启动 Gin 服务器 监听端口
	err := router.Run(addr)
	if err != nil {
		fmt.Println("程序出现异常:" + err.Error())
		fmt.Println("请检查或修改http_config.toml中信息")
		fmt.Println("程序将在5秒后自动退出.....")
		time.Sleep(6 * time.Second)
	}
}
