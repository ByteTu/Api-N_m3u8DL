package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"sync"
	"time"
)

// 版本信息
const AppVersion = "v3.0"

// stripAnsi 去除 ANSI 转义码
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)


func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// gbkToUtf8 将 GBK 编码转换为 UTF-8 字符串（如果数据已经是 UTF-8 则不转换）
func gbkToUtf8(data []byte) string {
	if isUtf8(data) {
		return string(data)
	}
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return string(data)
	}
	return string(result)
}

// TaskInfo 下载任务信息
type TaskInfo struct {
	TaskID    string    `json:"taskId"`
	PID       int       `json:"pid"`
	Name      string    `json:"name"`
	SaveDir   string    `json:"saveDir"`
	SaveName  string    `json:"saveName"`
	Url       string    `json:"url"`
	Command   string    `json:"command"`
	Running   bool      `json:"running"`
	Paused    bool      `json:"paused"`
	ExitCode  int       `json:"exitCode"`
	Logs      []string  `json:"logs"`
	StartTime time.Time `json:"startTime"`
}

// TaskHistory 历史任务（用于持久化，不含 Logs）
type TaskHistory struct {
	TaskID    string    `json:"taskId"`
	Name      string    `json:"name"`
	SaveDir   string    `json:"saveDir"`
	SaveName  string    `json:"saveName"`
	Url       string    `json:"url"`
	Command   string    `json:"command"`
	Running   bool      `json:"running"`
	Paused    bool      `json:"paused"`
	ExitCode  int       `json:"exitCode"`
	StartTime time.Time `json:"startTime"`
	LogFile   string    `json:"logFile"`
	RawParams string    `json:"rawParams"` // 原始请求参数 JSON，用于重新发起
}

var (
	tasks       = make(map[string]*TaskInfo)
	tasksMu     sync.RWMutex
	taskIDCnt   int64
	taskIDMu    sync.Mutex
	taskHistory []TaskHistory
	historyFile = "./data/task_history.json"
)

func nextTaskID() string {
	taskIDMu.Lock()
	taskIDCnt++
	id := taskIDCnt
	taskIDMu.Unlock()
	return fmt.Sprintf("T%06d", id)
}

// loadTaskHistory 从文件加载历史任务
func loadTaskHistory() {
	data, err := os.ReadFile(historyFile)
	if err != nil {
		taskHistory = make([]TaskHistory, 0)
		return
	}
	err = json.Unmarshal(data, &taskHistory)
	if err != nil {
		taskHistory = make([]TaskHistory, 0)
		return
	}
	// 更新 taskIDCnt 使新 ID 不冲突
	for _, h := range taskHistory {
		var num int64
		fmt.Sscanf(h.TaskID, "T%06d", &num)
		if num > taskIDCnt {
			taskIDCnt = num
		}
	}
	// 修复残留的 running 状态：如果任务标记为运行中但进程已不存在，修正为失败
	needSave := false
	for i := range taskHistory {
		if taskHistory[i].Running {
			taskHistory[i].Running = false
			taskHistory[i].ExitCode = -1 // 进程已不存在，视为异常退出
			needSave = true
		}
	}
	if needSave {
		saveTaskHistory()
	}
}

// saveTaskHistory 保存历史任务到文件
func saveTaskHistory() {
	os.MkdirAll("./data", 0755)
	data, err := json.MarshalIndent(taskHistory, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(historyFile, data, 0644)
}

// addTaskHistory 添加一条历史记录
func addTaskHistory(task *TaskInfo, logFile string, rawParams string) {
	h := TaskHistory{
		TaskID:    task.TaskID,
		Name:      task.Name,
		SaveDir:   task.SaveDir,
		SaveName:  task.SaveName,
		Url:       task.Url,
		Command:   task.Command,
		Running:   task.Running,
		Paused:    task.Paused,
		ExitCode:  task.ExitCode,
		StartTime: task.StartTime,
		LogFile:   logFile,
		RawParams: rawParams,
	}
	taskHistory = append(taskHistory, h)
	saveTaskHistory()
}

// updateTaskHistoryExitCode 更新历史任务的退出码
func updateTaskHistoryExitCode(taskID string, exitCode int) {
	for i := range taskHistory {
		if taskHistory[i].TaskID == taskID {
			taskHistory[i].Running = false
			taskHistory[i].ExitCode = exitCode
			saveTaskHistory()
			return
		}
	}
}

// Index 首页
func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// ReParamsReq 下载参数请求体（对应 N_m3u8DL-RE 的命令行参数）
type ReParamsReq struct {
	// 必填
	Must     string `json:"must"`      // m3u8 URL
	SaveDir  string `json:"save-dir"`  // 保存目录
	SaveName string `json:"save-name"` // 保存文件名

	// 基础选项
	BaseUrl         string `json:"base-url"`          // BaseURL
	MaxSpeed        string `json:"max-speed"`         // 限速，如 15M / 100K
	LogFilePath     string `json:"log-file-path"`     // 日志文件路径
	ThreadCount     int    `json:"thread-count"`      // 下载线程数
	Header          string `json:"header"`            // 自定义 HTTP 请求头

	// 合并选项
	BinaryMerge           bool `json:"binary-merge"`            // 二进制合并
	UseFfmpegConcatDemuxer bool `json:"use-ffmpeg-concat-demuxer"` // ffmpeg concat demuxer
	DelAfterDone          bool `json:"del-after-done"`         // 完成后删除临时文件
	NoDateInfo            bool `json:"no-date-info"`           // 混流时不写入日期信息
	WriteMetaJson         bool `json:"write-meta-json"`        // 输出 meta json 文件

	// 高级选项
	SkipMerge              bool   `json:"skip-merge"`               // 跳过合并
	SkipDownload           bool   `json:"skip-download"`            // 跳过下载（仅探测）
	CheckSegmentsCount      bool   `json:"check-segments-count"`     // 检测分片数量
	AutoSelect             bool   `json:"auto-select"`             // 自动选择最佳轨道
	NoAnsiColor            bool   `json:"no-ansi-color"`            // 去除 ANSI 颜色
	NoLog                  bool   `json:"no-log"`                  // 关闭日志文件输出
	SubOnly                bool   `json:"sub-only"`                // 仅选取字幕轨道
	AppendUrlParams        bool   `json:"append-url-params"`       // 将 URL 参数追加到分片
	ConcurrentDownload     bool   `json:"concurrent-download"`     // 并发下载音视频字幕

	// 字幕选项
	SubFormat          string `json:"sub-format"`           // 字幕格式 SRT|VTT
	AutoSubtitleFix    bool   `json:"auto-subtitle-fix"`    // 自动修正字幕

	// 选择流
	SelectVideo   string `json:"select-video"`   // 正则选择视频流
	SelectAudio   string `json:"select-audio"`   // 正则选择音频流
	SelectSubtitle string `json:"select-subtitle"` // 正则选择字幕流

	// 网络选项
	CustomProxy    string `json:"custom-proxy"`     // 自定义代理
	CustomRange    string `json:"custom-range"`     // 下载分片范围
	TaskStartAt    string `json:"task-start-at"`   // 任务开始时间
	HttpTimeout    int    `json:"http-timeout"`     // HTTP 超时秒数
	DownloadRetry  int    `json:"download-retry"`   // 分片下载重试次数

	// ffmpeg / 解密
	FfmpegBinaryPath        string `json:"ffmpeg-binary-path"`         // ffmpeg 路径
	DecryptionBinaryPath   string `json:"decryption-binary-path"`    // mp4decrypt 路径

	// 其他
	UiLanguage      string `json:"ui-language"`      // UI 语言 zh-CN|zh-TW|en-US
	LogLevel        string `json:"log-level"`        // 日志级别 DEBUG|ERROR|INFO|WARN|OFF
	DisableUpdate   bool   `json:"disable-update"`   // 禁用版本更新检测
}

// DownloadVideo 下载视频
func DownloadVideo(c *gin.Context) {
	// 保存原始请求体用于重新发起
	rawBody, _ := c.GetRawData()
	var reqData ReParamsReq
	if err := json.Unmarshal(rawBody, &reqData); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "参数获取异常"})
		return
	}
	rawParams := string(rawBody)
	if reqData.SaveDir == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "save-dir 参数不能为空"})
		return
	}
	if reqData.SaveName == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "save-name 参数不能为空"})
		return
	}

	// 根据操作系统选择可执行文件路径
	var rePath string
	sysOs := runtime.GOOS
	switch sysOs {
	case "windows":
		rePath = "./N_m3u8DL-RE.exe"
	case "linux", "darwin":
		rePath = "./N_m3u8DL-RE"
	default:
		c.JSON(200, gin.H{"code": 1, "msg": "不支持的操作系统: " + sysOs})
		return
	}

	cmd := exec.Command(rePath, reqData.Must)

	tempPath := "./data/temp_dir/"
	// 保存路径: data/videos/用户输入的目录名
	savePath := "./data/videos/" + reqData.SaveDir

	// 确保保存目录存在
	os.MkdirAll(savePath, 0755)

	// === 基础参数 ===
	cmd.Args = append(cmd.Args, "--tmp-dir", tempPath)
	cmd.Args = append(cmd.Args, "--save-dir", savePath)
	cmd.Args = append(cmd.Args, "--save-name", reqData.SaveName)

	if reqData.BaseUrl != "" {
		cmd.Args = append(cmd.Args, "--base-url", reqData.BaseUrl)
	}
	if reqData.MaxSpeed != "" {
		cmd.Args = append(cmd.Args, "--max-speed", reqData.MaxSpeed)
	}

	// 确保 Logs 目录存在，N_m3u8DL-RE 写入日志文件
	logsDir := "./Logs"
	os.MkdirAll(logsDir, 0755)
	logFileName := fmt.Sprintf("%s_%s.log", reqData.SaveName, time.Now().Format("20060102_150405"))
	logFilePath := logsDir + "/" + logFileName
	cmd.Args = append(cmd.Args, "--log-file-path", logFilePath)
	if reqData.ThreadCount > 0 {
		cmd.Args = append(cmd.Args, "--thread-count", fmt.Sprintf("%d", reqData.ThreadCount))
	}
	if reqData.Header != "" {
		cmd.Args = append(cmd.Args, "-H", reqData.Header)
	}

	// === 合并参数 ===
	if reqData.BinaryMerge {
		cmd.Args = append(cmd.Args, "--binary-merge")
	}
	if reqData.UseFfmpegConcatDemuxer {
		cmd.Args = append(cmd.Args, "--use-ffmpeg-concat-demuxer")
	}
	if !reqData.DelAfterDone {
		cmd.Args = append(cmd.Args, "--del-after-done")
	}
	if reqData.NoDateInfo {
		cmd.Args = append(cmd.Args, "--no-date-info")
	}
	if !reqData.WriteMetaJson {
		cmd.Args = append(cmd.Args, "--write-meta-json=false")
	}

	// === 高级参数 ===
	if reqData.SkipMerge {
		cmd.Args = append(cmd.Args, "--skip-merge")
	}
	if reqData.SkipDownload {
		cmd.Args = append(cmd.Args, "--skip-download")
	}
	if reqData.CheckSegmentsCount {
		cmd.Args = append(cmd.Args, "--check-segments-count")
	}
	if reqData.AutoSelect {
		cmd.Args = append(cmd.Args, "--auto-select")
	}
	if reqData.NoAnsiColor {
		cmd.Args = append(cmd.Args, "--no-ansi-color")
	}
	if reqData.NoLog {
		cmd.Args = append(cmd.Args, "--no-log")
	}
	if reqData.SubOnly {
		cmd.Args = append(cmd.Args, "--sub-only")
	}
	if reqData.AppendUrlParams {
		cmd.Args = append(cmd.Args, "--append-url-params")
	}
	if reqData.ConcurrentDownload {
		cmd.Args = append(cmd.Args, "-mt")
	}

	// === 字幕参数 ===
	if reqData.SubFormat != "" {
		cmd.Args = append(cmd.Args, "--sub-format", reqData.SubFormat)
	}
	if reqData.AutoSubtitleFix {
		cmd.Args = append(cmd.Args, "--auto-subtitle-fix")
	}

	// === 流选择 ===
	if reqData.SelectVideo != "" {
		cmd.Args = append(cmd.Args, "-sv", reqData.SelectVideo)
	}
	if reqData.SelectAudio != "" {
		cmd.Args = append(cmd.Args, "-sa", reqData.SelectAudio)
	}
	if reqData.SelectSubtitle != "" {
		cmd.Args = append(cmd.Args, "-ss", reqData.SelectSubtitle)
	}

	// === 网络参数 ===
	if reqData.CustomProxy != "" {
		cmd.Args = append(cmd.Args, "--custom-proxy", reqData.CustomProxy)
	}
	if reqData.CustomRange != "" {
		cmd.Args = append(cmd.Args, "--custom-range", reqData.CustomRange)
	}
	if reqData.TaskStartAt != "" {
		cmd.Args = append(cmd.Args, "--task-start-at", reqData.TaskStartAt)
	}
	if reqData.HttpTimeout > 0 {
		cmd.Args = append(cmd.Args, "--http-request-timeout", fmt.Sprintf("%d", reqData.HttpTimeout))
	}
	if reqData.DownloadRetry > 0 {
		cmd.Args = append(cmd.Args, "--download-retry-count", fmt.Sprintf("%d", reqData.DownloadRetry))
	}

	// === ffmpeg / 解密 ===
	if reqData.FfmpegBinaryPath != "" {
		cmd.Args = append(cmd.Args, "--ffmpeg-binary-path", reqData.FfmpegBinaryPath)
	}
	if reqData.DecryptionBinaryPath != "" {
		cmd.Args = append(cmd.Args, "--decryption-binary-path", reqData.DecryptionBinaryPath)
	}

	// === 其他 ===
	if reqData.UiLanguage != "" {
		cmd.Args = append(cmd.Args, "--ui-language", reqData.UiLanguage)
	}
	if reqData.LogLevel != "" {
		cmd.Args = append(cmd.Args, "--log-level", reqData.LogLevel)
	}
	if reqData.DisableUpdate {
		cmd.Args = append(cmd.Args, "--disable-update-check")
	}

	// 捕获标准输出和标准错误
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// 启动命令
	if err := cmd.Start(); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "命令启动错误: " + err.Error()})
		return
	}

	pid := cmd.Process.Pid

	// 创建任务记录
	taskID := nextTaskID()
	task := &TaskInfo{
		TaskID:    taskID,
		PID:       pid,
		Name:      reqData.SaveName,
		SaveDir:   reqData.SaveDir,
		SaveName:  reqData.SaveName,
		Url:       reqData.Must,
		Command:   cmd.String(),
		Running:   true,
		ExitCode:  -1,
		Logs:      []string{},
		StartTime: time.Now(),
	}

	tasksMu.Lock()
	tasks[taskID] = task
	tasksMu.Unlock()

	// 保存到历史记录
	addTaskHistory(task, logFilePath, rawParams)

	// 实时捕获 stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := stripAnsi(gbkToUtf8(scanner.Bytes()))
			tasksMu.Lock()
			if t, ok := tasks[taskID]; ok {
				t.Logs = append(t.Logs, line)
			}
			tasksMu.Unlock()
		}
	}()

	// 实时捕获 stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := stripAnsi(gbkToUtf8(scanner.Bytes()))
			tasksMu.Lock()
			if t, ok := tasks[taskID]; ok {
				t.Logs = append(t.Logs, "[ERR] "+line)
			}
			tasksMu.Unlock()
		}
	}()

	// 等待命令完成
	go func() {
		err := cmd.Wait()
		tasksMu.Lock()
		var exitCode int = -1
		if t, ok := tasks[taskID]; ok {
			t.Running = false
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
					t.ExitCode = exitCode
				} else {
					t.ExitCode = -1
				}
			} else {
				exitCode = 0
				t.ExitCode = 0
			}
			updateTaskHistoryExitCode(taskID, t.ExitCode)
			delete(tasks, taskID) // 完成后从 map 中移除
		}
		tasksMu.Unlock()
	}()

	c.JSON(200, gin.H{
		"code":    0,
		"msg":     "任务已在后台启动，日志面板实时显示下载进度",
		"console": cmd.String(),
		"taskId":  taskID,
		"pid":     pid,
	})
}

// TaskStatus 查询任务状态
func TaskStatus(c *gin.Context) {
	taskID := c.Query("taskId")
	if taskID == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "taskId 参数无效"})
		return
	}

	tasksMu.RLock()
	task, ok := tasks[taskID]
	tasksMu.RUnlock()

	if !ok {
		c.JSON(200, gin.H{"code": 1, "msg": "任务不存在", "running": false})
		return
	}

	c.JSON(200, gin.H{
		"code":     0,
		"taskId":   task.TaskID,
		"pid":      task.PID,
		"name":     task.Name,
		"running":  task.Running,
		"paused":   task.Paused,
		"exitCode": task.ExitCode,
		"logs":     task.Logs,
	})
}

// TaskList 获取所有任务列表
func TaskList(c *gin.Context) {
	tasksMu.RLock()
	defer tasksMu.RUnlock()

	list := make([]gin.H, 0, len(tasks))
	for _, task := range tasks {
		list = append(list, gin.H{
			"taskId":    task.TaskID,
			"pid":       task.PID,
			"name":      task.Name,
			"url":       task.Url,
			"running":   task.Running,
			"paused":    task.Paused,
			"exitCode":  task.ExitCode,
			"startTime": task.StartTime.Format("2006-01-02 15:04:05"),
			"logCount":  len(task.Logs),
		})
	}
	c.JSON(200, gin.H{"code": 0, "tasks": list})
}

// TaskLogs 获取任务日志（支持增量获取）
func TaskLogs(c *gin.Context) {
	taskID := c.Query("taskId")
	if taskID == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "taskId 参数无效"})
		return
	}

	var offset int
	if _, err := fmt.Sscanf(c.DefaultQuery("offset", "0"), "%d", &offset); err != nil {
		offset = 0
	}

	tasksMu.RLock()
	task, ok := tasks[taskID]
	tasksMu.RUnlock()

	if !ok {
		c.JSON(200, gin.H{"code": 1, "msg": "任务不存在"})
		return
	}

	total := len(task.Logs)
	if offset > total {
		offset = total
	}
	newLogs := task.Logs[offset:]

	c.JSON(200, gin.H{
		"code":    0,
		"taskId":  task.TaskID,
		"pid":     task.PID,
		"total":   total,
		"offset":  offset,
		"logs":    newLogs,
		"running": task.Running,
	})
}

// PauseTask 暂停任务
func PauseTask(c *gin.Context) {
	taskID := c.PostForm("taskId")
	if taskID == "" {
		taskID = c.Query("taskId")
	}
	if taskID == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "taskId 参数无效"})
		return
	}

	tasksMu.Lock()
	task, ok := tasks[taskID]
	if !ok {
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务已结束，无需暂停"})
		return
	}
	if task.Paused {
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务已暂停"})
		return
	}
	// 保存PID用于后续检查
	pid := task.PID
	tasksMu.Unlock()

	// 先检查进程是否还在运行
	if !isProcessRunning(pid) {
		c.JSON(200, gin.H{"code": 0, "msg": "任务进程已结束"})
		return
	}

	// 暂停主进程
	if err := suspendProcess(pid); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "暂停失败: " + err.Error()})
		return
	}

	tasksMu.Lock()
	task.Paused = true
	tasksMu.Unlock()

	c.JSON(200, gin.H{"code": 0, "msg": "任务已暂停"})
}

// ResumeTask 继续任务
func ResumeTask(c *gin.Context) {
	taskID := c.PostForm("taskId")
	if taskID == "" {
		taskID = c.Query("taskId")
	}
	if taskID == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "taskId 参数无效"})
		return
	}

	tasksMu.Lock()
	task, ok := tasks[taskID]
	if !ok {
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务已结束，无法继续"})
		return
	}
	if !task.Paused {
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务已在运行中"})
		return
	}
	tasksMu.Unlock()

	if err := resumeProcess(task.PID); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "继续失败: " + err.Error()})
		return
	}

	tasksMu.Lock()
	task.Paused = false
	tasksMu.Unlock()

	c.JSON(200, gin.H{"code": 0, "msg": "任务已继续"})
}

// StopTask 终止任务
func StopTask(c *gin.Context) {
	taskID := c.PostForm("taskId")
	if taskID == "" {
		taskID = c.Query("taskId")
	}
	if taskID == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "taskId 参数无效"})
		return
	}

	tasksMu.Lock()
	task, ok := tasks[taskID]
	if !ok {
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务已结束"})
		return
	}
	pid := task.PID
	tasksMu.Unlock()

	// 先检查进程是否还在运行
	if !isProcessRunning(pid) {
		updateTaskHistoryExitCode(taskID, 0)
		tasksMu.Lock()
		delete(tasks, taskID)
		tasksMu.Unlock()
		c.JSON(200, gin.H{"code": 0, "msg": "任务进程已结束"})
		return
	}

	if err := terminateProcess(pid, 1); err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "终止失败: " + err.Error()})
		return
	}

	updateTaskHistoryExitCode(taskID, -2) // -2 表示手动终止
	tasksMu.Lock()
	delete(tasks, taskID)
	tasksMu.Unlock()

	c.JSON(200, gin.H{"code": 0, "msg": "任务已终止"})
}

// getCurrentNetworkIP 获取当前服务 ip
func getCurrentNetworkIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// LogList 获取日志文件列表
func LogList(c *gin.Context) {
	logDir := "./Logs"
	entries, err := os.ReadDir(logDir)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "无法读取日志目录", "logs": []gin.H{}})
		return
	}

	var logs []gin.H
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			logs = append(logs, gin.H{
				"name":    entry.Name(),
				"size":    info.Size(),
				"modTime": info.ModTime().Format("2006-01-02 15:04:05"),
				"rawTime": info.ModTime().Unix(),
			})
		}
	}

	// 按修改时间倒序排列（最新在前）
	sort.Slice(logs, func(i, j int) bool {
		return logs[i]["rawTime"].(int64) > logs[j]["rawTime"].(int64)
	})

	c.JSON(200, gin.H{"code": 0, "logs": logs})
}

// LogContent 获取日志文件内容
func LogContent(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(400, gin.H{"code": 1, "msg": "文件名不能为空"})
		return
	}

	// 安全检查：防止路径遍历
	if containsDotDot(filename) {
		c.JSON(400, gin.H{"code": 1, "msg": "非法文件名"})
		return
	}

	logPath := "./Logs/" + filename
	data, err := os.ReadFile(logPath)
	if err != nil {
		c.JSON(404, gin.H{"code": 1, "msg": "日志文件不存在"})
		return
	}

	content := gbkToUtf8(data)

	c.JSON(200, gin.H{"code": 0, "content": content, "filename": filename})
}

// containsDotDot 检查字符串是否包含 .. 路径遍历
func containsDotDot(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '.' && s[i+1] == '.' {
			return true
		}
	}
	return false
}

// isUtf8 检查字节数组是否为 UTF-8 编码
func isUtf8(data []byte) bool {
	// 简单检查：看是否包含非法字节序列
	for i := 0; i < len(data); {
		if data[i] < 0x80 {
			i++
		} else if data[i]&0xE0 == 0xC0 && i+1 < len(data) && data[i+1]&0xC0 == 0x80 {
			i += 2
		} else if data[i]&0xF0 == 0xE0 && i+2 < len(data) && data[i+1]&0xC0 == 0x80 && data[i+2]&0xC0 == 0x80 {
			i += 3
		} else if data[i]&0xF8 == 0xF0 && i+3 < len(data) && data[i+1]&0xC0 == 0x80 && data[i+2]&0xC0 == 0x80 && data[i+3]&0xC0 == 0x80 {
			i += 4
		} else {
			return false
		}
	}
	return true
}

// ClearOldLogs 删除1天前的日志文件
func ClearOldLogs(c *gin.Context) {
	logDir := "./Logs"
	entries, err := os.ReadDir(logDir)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "无法读取日志目录"})
		return
	}

	oneDayAgo := time.Now().Add(-24 * time.Hour)
	deleted := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(oneDayAgo) {
			err := os.Remove(logDir + "/" + entry.Name())
			if err == nil {
				deleted++
			}
		}
	}

	c.JSON(200, gin.H{"code": 0, "deleted": deleted, "msg": "已删除 " + fmt.Sprintf("%d", deleted) + " 个日志文件"})
}

// TaskHistoryList 获取所有历史任务
func TaskHistoryList(c *gin.Context) {
	tasksMu.RLock()
	history := make([]gin.H, 0, len(taskHistory))
	// 倒序：最新的在前
	for i := len(taskHistory) - 1; i >= 0; i-- {
		h := taskHistory[i]
		statusLabel := ""
		statusClass := ""
		if h.Running {
			statusLabel = "进行中"
			statusClass = "running"
		} else if h.ExitCode == 0 {
			statusLabel = "已完成"
			statusClass = "done"
		} else if h.ExitCode == -2 {
			statusLabel = "已终止"
			statusClass = "stopped"
		} else {
			statusLabel = "失败"
			statusClass = "failed"
		}
		history = append(history, gin.H{
			"taskId":      h.TaskID,
			"name":        h.Name,
			"saveDir":     h.SaveDir,
			"url":         h.Url,
			"command":     h.Command,
			"running":     h.Running,
			"exitCode":    h.ExitCode,
			"startTime":   h.StartTime.Format("2006-01-02 15:04:05"),
			"logFile":     h.LogFile,
			"statusLabel": statusLabel,
			"statusClass": statusClass,
			"rawParams":   h.RawParams,
		})
	}
	tasksMu.RUnlock()

	c.JSON(200, gin.H{"code": 0, "tasks": history})
}

// TaskVerify 批量验证任务状态
func TaskVerify(c *gin.Context) {
	var req struct {
		TaskIDs []string `json:"taskIds"`
	}
	if err := c.BindJSON(&req); err != nil || len(req.TaskIDs) == 0 {
		c.JSON(200, gin.H{"code": 0, "tasks": []gin.H{}})
		return
	}

	tasksMu.Lock()
	defer tasksMu.Unlock()

	result := make([]gin.H, 0, len(req.TaskIDs))
	for _, taskID := range req.TaskIDs {
		task, ok := tasks[taskID]
		if !ok {
			// 任务不在内存中，从历史记录查找，并修复残留的 running 状态
			exitCode := -1
			running := false
			for i := range taskHistory {
				if taskHistory[i].TaskID == taskID {
					if taskHistory[i].Running {
						// 历史记录标记为运行中，但任务不在内存中 -> 实际已结束
						taskHistory[i].Running = false
						taskHistory[i].ExitCode = -1 // 进程异常退出
						exitCode = -1
						running = false
						saveTaskHistory()
					} else {
						exitCode = taskHistory[i].ExitCode
						running = false
					}
					break
				}
			}
			result = append(result, gin.H{
				"taskId":   taskID,
				"running":  running,
				"exitCode": exitCode,
				"exists":   false,
			})
			continue
		}
		// 任务在内存中且标记为运行中，实际检查进程是否存活
		if task.Running {
			if !isProcessRunning(task.PID) {
				task.Running = false
				task.ExitCode = -1
				// 同步更新历史记录
				updateTaskHistoryExitCode(taskID, -1)
				// 从内存中移除已结束的任务
				delete(tasks, taskID)
				result = append(result, gin.H{
					"taskId":   taskID,
					"running":  false,
					"exitCode": -1,
					"exists":   false,
				})
				continue
			}
		}
		result = append(result, gin.H{
			"taskId":   taskID,
			"running":  task.Running,
			"paused":   task.Paused,
			"exitCode": task.ExitCode,
			"exists":   true,
		})
	}
	c.JSON(200, gin.H{"code": 0, "tasks": result})
}

// ClearDoneTaskHistory 清空已完成的历史任务记录
func ClearDoneTaskHistory(c *gin.Context) {
	tasksMu.Lock()
	defer tasksMu.Unlock()

	deleted := 0
	newHistory := make([]TaskHistory, 0)
	for i := range taskHistory {
		if taskHistory[i].ExitCode == 0 && !taskHistory[i].Running {
			deleted++
		} else {
			newHistory = append(newHistory, taskHistory[i])
		}
	}
	taskHistory = newHistory
	saveTaskHistory()

	c.JSON(200, gin.H{"code": 0, "deleted": deleted, "msg": "已清空 " + fmt.Sprintf("%d", deleted) + " 条已完成记录"})
}

// VideosDirList 获取 /data/videos 下的目录列表
func VideosDirList(c *gin.Context) {
	videosDir := "./data/videos"
	entries, err := os.ReadDir(videosDir)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "无法读取目录", "dirs": []gin.H{}})
		return
	}

	var dirs []gin.H
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			dirs = append(dirs, gin.H{
				"name":    entry.Name(),
				"modTime": info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(200, gin.H{"code": 0, "dirs": dirs})
}

// VideosFileList 获取指定子目录下的视频文件列表
func VideosFileList(c *gin.Context) {
	subDir := c.Query("dir")
	if subDir == "" {
		c.JSON(200, gin.H{"code": 1, "msg": "dir 参数不能为空", "files": []gin.H{}})
		return
	}

	// 安全检查：防止路径遍历
	if containsDotDot(subDir) {
		c.JSON(200, gin.H{"code": 1, "msg": "非法目录名"})
		return
	}

	targetDir := "./data/videos/" + subDir
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "目录不存在或无法读取", "files": []gin.H{}})
		return
	}

	var files []gin.H
	videoExts := []string{".mp4", ".mkv", ".avi", ".flv", ".mov", ".wmv", ".ts", ".webm", ".mpg", ".mpeg", ".m4v"}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		isVideo := false
		for _, ext := range videoExts {
			if len(name) >= len(ext) && name[len(name)-len(ext):] == ext {
				isVideo = true
				break
			}
		}
		if !isVideo {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, gin.H{
			"name":    name,
			"size":    info.Size(),
			"modTime": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(200, gin.H{"code": 0, "files": files, "dir": subDir})
}

func main() {
	// 加载历史任务
	loadTaskHistory()

	addr := ":5050"
	serverIP, _ := getCurrentNetworkIP()
	fmt.Println("服务启动中.......")
	fmt.Println("服务器地址: http://" + serverIP + addr)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "./static")

	router.GET("/", Index)
	router.POST("/download", DownloadVideo)
	router.GET("/task-status", TaskStatus)
	router.GET("/task-list", TaskList)
	router.GET("/task-logs", TaskLogs)
	router.POST("/task-pause", PauseTask)
	router.POST("/task-resume", ResumeTask)
	router.POST("/task-stop", StopTask)
	router.GET("/api/logs", LogList)
	router.GET("/api/logs/*filename", LogContent)
	router.POST("/api/logs/clear-old", ClearOldLogs)
	router.GET("/api/task-history", TaskHistoryList)
	router.POST("/api/task-verify", TaskVerify)
	router.POST("/api/task-history/clear-done", ClearDoneTaskHistory)
	router.GET("/api/videos/dirs", VideosDirList)
	router.GET("/api/videos/files", VideosFileList)
	router.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": AppVersion})
	})

	err := router.Run(addr)
	if err != nil {
		fmt.Println("程序出现异常:" + err.Error())
		fmt.Println("程序将在5秒后自动退出.....")
		time.Sleep(6 * time.Second)
	}
}
