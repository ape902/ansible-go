package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/executor"
	"github.com/ape902/ansible-go/pkg/logger"
)

// 创建目录，如果不存在则创建
func createDirIfNotExist(path string, log *logger.Logger) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("创建目录: %s", path)
		return os.MkdirAll(path, 0755)
	} else if err != nil {
		return err
	}
	log.Info("目录已存在: %s", path)
	return nil
}

// 创建文件，如果不存在则创建
func createFileIfNotExist(path string, content string, log *logger.Logger) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info("创建文件: %s", path)
		return os.WriteFile(path, []byte(content), 0644)
	} else if err != nil {
		return err
	}
	log.Info("文件已存在: %s", path)
	return nil
}

// 初始化项目
func initProject(name string, path string, log *logger.Logger) error {
	// 创建项目根目录
	projectRoot := filepath.Join(path, name)
	if err := createDirIfNotExist(projectRoot, log); err != nil {
		return fmt.Errorf("创建项目根目录失败: %w", err)
	}

	// 创建子目录
	dirs := []string{
		filepath.Join(projectRoot, "pkg"),
		filepath.Join(projectRoot, "pkg", "config"),
		filepath.Join(projectRoot, "pkg", "executor"),
		filepath.Join(projectRoot, "pkg", "vars"),
	}

	log.IncreaseIndent()
	for _, dir := range dirs {
		if err := createDirIfNotExist(dir, log); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	log.DecreaseIndent()

	// 创建配置文件模板
	configTemplate := `inventory:
  localhost: [127.0.0.1]
vars:
  example_var: "value"
`
	configPath := filepath.Join(projectRoot, "pkg", "config", "config.yaml")
	if err := createFileIfNotExist(configPath, configTemplate, log); err != nil {
		return fmt.Errorf("创建配置文件模板失败: %w", err)
	}

	log.Success("项目 %s 初始化完成", name)
	return nil
}

func main() {
	// 命令行参数解析
	configFile := flag.String("config", "", "配置文件路径")
	playbook := flag.String("playbook", "", "playbook文件路径")
	verbose := flag.Bool("verbose", false, "启用详细日志输出")

	// 添加初始化相关参数
	initFlag := flag.Bool("init", false, "初始化项目")
	projectName := flag.String("name", "ansible-go-project", "项目名称")
	projectPath := flag.String("path", ".", "项目初始化路径")

	flag.Parse()

	// 初始化日志
	log := logger.New()
	// verbose变量已使用，传递给执行器

	// 处理初始化命令
	if *initFlag {
		if err := initProject(*projectName, *projectPath, log); err != nil {
			log.Error("初始化项目失败: %v", err)
			os.Exit(1)
		}
		return
	}

	// 检查必要参数
	if *configFile == "" && *playbook == "" {
		log.Error("错误: 必须提供配置文件或playbook文件路径")
		flag.Usage()
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Error("加载配置失败: %v", err)
		os.Exit(1)
	}

	// 初始化执行器
	exec := executor.NewExecutor(cfg)
	// 设置verbose模式
	if *verbose {
		exec.SetVerboseMode(true)
	}

	// 执行playbook
	if *playbook != "" {
		err = exec.Execute(*playbook)
		if err != nil {
			log.Error("执行playbook失败: %v", err)
			os.Exit(1)
		}
	}

	log.Success("执行完成")
}
