package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 定义颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

// Logger 定义日志记录器
type Logger struct {
	logger      *log.Logger
	indent      int
	verboseMode bool // 是否启用详细模式
}

// IncreaseIndent 增加缩进级别
func (l *Logger) IncreaseIndent() {
	l.indent++
}

// DecreaseIndent 减少缩进级别
func (l *Logger) DecreaseIndent() {
	if l.indent > 0 {
		l.indent--
	}
}

// getIndent 获取当前缩进字符串
func (l *Logger) getIndent() string {
	return strings.Repeat("  ", l.indent)
}

// New 创建新的日志记录器
func New() *Logger {
	return &Logger{
		logger:      log.New(os.Stdout, "", 0), // 移除默认时间戳
		indent:      0,
		verboseMode: false,
	}
}

// SetVerboseMode 设置详细模式
func (l *Logger) SetVerboseMode(verbose bool) {
	l.verboseMode = verbose
}

// IsVerboseMode 获取当前是否为详细模式
func (l *Logger) IsVerboseMode() bool {
	return l.verboseMode
}

// formatLog 格式化日志消息
func (l *Logger) formatLog(level, color, msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)
	
	// 为不同级别的日志使用不同的前缀样式
	var prefix string
	if level == "DEBUG" {
		prefix = fmt.Sprintf("%s[%s] [%s] %s:%d:%s", color, timestamp, level, file, line, ColorReset)
	} else if level == "SUCCESS" {
		prefix = fmt.Sprintf("%s[%s] [%s]%s", ColorGreen, timestamp, level, ColorReset)
	} else if level == "ERROR" {
		prefix = fmt.Sprintf("%s[%s] [%s]%s", ColorRed, timestamp, level, ColorReset)
	} else if level == "WARNING" {
		prefix = fmt.Sprintf("%s[%s] [%s]%s", ColorYellow, timestamp, level, ColorReset)
	} else {
		prefix = fmt.Sprintf("%s[%s] [%s]%s", color, timestamp, level, ColorReset)
	}
	
	// 添加固定宽度的格式，使日志对齐
	return fmt.Sprintf("%-40s %s%s%s", prefix, color, msg, ColorReset)
}

// Info 打印信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatLog("INFO", ColorBlue, l.getIndent()+msg))
}

// Success 打印成功日志
func (l *Logger) Success(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatLog("SUCCESS", ColorGreen, l.getIndent()+msg))
}

// Warning 打印警告日志
func (l *Logger) Warning(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatLog("WARNING", ColorYellow, l.getIndent()+msg))
}

// Debug 打印调试日志，仅在详细模式下显示
func (l *Logger) Debug(format string, v ...interface{}) {
	if !l.verboseMode {
		return
	}
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatLog("DEBUG", ColorPurple, l.getIndent()+msg))
}

// Error 打印错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Print(l.formatLog("ERROR", ColorRed, l.getIndent()+msg))
}

// Output 打印命令输出
func (l *Logger) Output(host, taskID, output string) {
	if output == "" {
		return
	}
	// 使用更醒目的颜色和格式来显示主机和任务信息
	hostColor := ColorCyan
	prefix := fmt.Sprintf("%s┌─[%s]%s %s%s%s", hostColor, host, ColorReset, ColorYellow, taskID, ColorReset)
	
	// 打印任务标题行
	l.logger.Printf("%s", prefix)
	
	// 处理并格式化输出内容
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if line != "" {
			// 为每行输出添加缩进和格式
			if i == len(lines)-1 && line == "" {
				break // 跳过末尾的空行
			}
			l.logger.Printf("%s│ %s%s", hostColor, line, ColorReset)
		}
	}
	
	// 添加任务输出结束标记
	l.logger.Printf("%s└─────%s", hostColor, ColorReset)
}