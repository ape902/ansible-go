package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	logger *log.Logger
	indent int
}

// New 创建新的日志记录器
func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		indent: 0,
	}
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

// Info 打印信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s%s%s", l.getIndent(), ColorBlue, msg, ColorReset)
}

// Success 打印成功日志
func (l *Logger) Success(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s%s%s", l.getIndent(), ColorGreen, msg, ColorReset)
}

// Warning 打印警告日志
func (l *Logger) Warning(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s%s%s", l.getIndent(), ColorYellow, msg, ColorReset)
}

// Error 打印错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s%s%s", l.getIndent(), ColorRed, msg, ColorReset)
}

// Output 打印命令输出
func (l *Logger) Output(host, taskID, output string) {
	if output == "" {
		return
	}
	l.IncreaseIndent()
	defer l.DecreaseIndent()

	l.logger.Printf("%s%s主机 %s 上的任务 %s 的输出:%s", l.getIndent(), ColorCyan, host, taskID, ColorReset)
	l.IncreaseIndent()
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		l.logger.Printf("%s%s%s%s", l.getIndent(), ColorGray, line, ColorReset)
	}
	l.DecreaseIndent()
}