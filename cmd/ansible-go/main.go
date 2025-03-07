package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/executor"
	"github.com/ape902/ansible-go/pkg/logger"
)

// 项目模板定义
// configTemplate 配置文件模板
const configTemplate = `# ansible-go 配置文件

# SSH连接配置
ssh:
  # 默认用户名
  user: root
  # 默认密码，建议使用密钥认证
  password: ""
  # 密钥文件路径
  key_file: "~/.ssh/id_rsa"
  # 密钥密码
  key_password: ""
  # 连接超时时间（秒）
  timeout: 10
  # 最大并行执行数
  max_parallel: 5

# 主机组配置
hosts:
  # 示例组
  web_servers:
    # 主机列表
    hosts:
      - "192.168.1.100"
      - "192.168.1.101"
    # 组变量
    vars:
      app_port: 8080
      app_name: "web-app"
  
  # 数据库服务器组
  db_servers:
    hosts:
      - "192.168.1.200"
    vars:
      db_port: 3306
      db_name: "app_db"

# 全局变量
vars:
  env: "production"
  domain: "example.com"
  admin_email: "admin@example.com"
`

// taskTemplate 任务文件模板
const taskTemplate = `# 示例任务文件
name: "示例任务"
description: "这是一个示例任务配置文件"
hosts: ["web_servers"]

# 任务列表
tasks:
  # 安装软件包
  - install_packages:
      name: "安装必要软件包"
      module: "command"
      args:
        cmd: "apt-get install -y nginx curl wget"
        sudo: true
      # 可以添加标签作为变量
      vars:
        tags: ["setup", "packages"]

  # 创建目录
  - create_app_directory:
      name: "创建应用目录"
      module: "file"
      args:
        path: "/opt/{{ .vars.app_name }}"
        state: "directory"
        mode: "0755"
        owner: "www-data"
        group: "www-data"
      vars:
        tags: ["setup", "app"]

  # 复制配置文件
  - copy_app_config:
      name: "复制应用配置"
      module: "copy"
      args:
        src: "app.conf"
        dest: "/etc/nginx/conf.d/{{ .vars.app_name }}.conf"
        mode: "0644"
        owner: "root"
        group: "root"
      vars:
        tags: ["config"]

  # 使用模板生成配置
  - generate_app_config:
      name: "生成应用配置"
      module: "template"
      args:
        src: "app.conf.tmpl"
        dest: "/opt/{{ .vars.app_name }}/config.json"
        mode: "0644"
      vars:
        tags: ["config"]

  # 重启服务
  - restart_nginx:
      name: "重启Nginx服务"
      module: "service"
      args:
        name: "nginx"
        state: "restarted"
      vars:
        tags: ["service"]

# 处理器列表
handlers:
  - name: "reload_nginx"
    module: "service"
    args:
      name: "nginx"
      state: "reloaded"
`

// mainTaskTemplate 主任务文件模板
const mainTaskTemplate = `# 主任务文件
name: "主任务配置"
description: "用于包含和组织其他任务文件"
hosts: ["all"]

# 任务列表
tasks:
  - import_tasks:
      name: "导入示例任务"
      module: "import"
      args:
        file: "example.yaml"

# 可以添加更多导入任务
#  - import_setup:
#      name: "导入设置任务"
#      module: "import"
#      args:
#        file: "setup.yaml"
#
#  - import_deploy:
#      name: "导入部署任务"
#      module: "import"
#      args:
#        file: "deploy.yaml"
#
#  - import_configure:
#      name: "导入配置任务"
#      module: "import"
#      args:
#        file: "configure.yaml"
`

// varsTemplate 变量文件模板
const varsTemplate = `# 变量定义文件
# 这些变量可以在任务中使用

# 应用配置
app:
  name: "my-application"
  version: "1.0.0"
  port: 8080

# 环境配置
environment: "production"

# 数据库配置
database:
  host: "db.example.com"
  port: 3306
  name: "app_db"
  user: "app_user"

# 路径配置
paths:
  app_dir: "/opt/my-application"
  config_dir: "/etc/my-application"
  log_dir: "/var/log/my-application"
`

// exampleFileContent 示例文件内容
const exampleFileContent = `# 示例配置文件
# 这是一个示例配置文件，用于演示文件复制功能

server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    access_log /var/log/nginx/app_access.log;
    error_log /var/log/nginx/app_error.log;
}
`

// CommandFlags 定义命令行参数结构体，包含主命令和子命令的所有参数
type CommandFlags struct {
	// 主命令参数
	ConfigFile string
	Verbose    bool
	Parallel   int
	Tags       string

	// init子命令参数
	ProjectName string
	ProjectPath string
}

// initFlags 初始化并返回主命令和子命令的参数集
// 返回值:
//   - *flag.FlagSet: 主命令参数集
//   - *flag.FlagSet: init子命令参数集
//   - *CommandFlags: 解析后的命令行参数
func initFlags() (*flag.FlagSet, *flag.FlagSet, *CommandFlags) {
	flags := &CommandFlags{}

	// 创建主命令行参数集
	mainFlags := flag.NewFlagSet("ansible-go", flag.ExitOnError)
	mainFlags.StringVar(&flags.ConfigFile, "config", "", "配置文件路径")
	mainFlags.BoolVar(&flags.Verbose, "verbose", false, "启用详细日志输出")
	mainFlags.IntVar(&flags.Parallel, "parallel", 5, "最大并行执行数")
	mainFlags.StringVar(&flags.Tags, "tags", "", "要执行的标签，多个标签用逗号分隔")

	// 创建init子命令
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCmd.StringVar(&flags.ProjectName, "name", "ansible-go-project", "项目名称")
	initCmd.StringVar(&flags.ProjectPath, "path", ".", "项目初始化路径")

	return mainFlags, initCmd, flags
}

// showHelp 显示命令行帮助信息
// 参数:
//   - mainFlags: 主命令参数集，用于显示全局参数说明
func showHelp(mainFlags *flag.FlagSet) {
	fmt.Println("用法: ansible-go <命令> [参数]")
	fmt.Println("\n可用命令:")
	fmt.Println("  init\t初始化新项目")
	fmt.Println("  check\t检查配置文件的合规性")
	fmt.Println("\n子命令参数:")
	fmt.Println("  init:")
	fmt.Println("    -name string\t项目名称 (默认: \"ansible-go-project\")")
	fmt.Println("    -path string\t项目初始化路径 (默认: \".\")")
	fmt.Println("  check:")
	fmt.Println("    -config string\t配置文件路径")
	fmt.Println("\n全局参数:")
	mainFlags.PrintDefaults()
}

// handleCheckCommand 处理check命令，验证配置文件的合规性
// 参数:
//   - configFile: 配置文件路径
//   - mainFlags: 主命令参数集
//   - log: 日志记录器
//
// 处理错误并退出程序
func handleErrorAndExit(log *logger.Logger, format string, args ...interface{}) {
	log.Error(format, args...)
	os.Exit(1)
}

func handleCheckCommand(configFile string, mainFlags *flag.FlagSet, log *logger.Logger) {
	// 检查配置文件
	if configFile == "" {
		handleErrorAndExit(log, "未指定配置文件，请使用 --config 参数指定配置文件路径")
		mainFlags.PrintDefaults()
	}

	// 加载并验证配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		handleErrorAndExit(log, "加载配置失败: %v", err)
	}

	if errors := config.ValidateConfig(cfg); len(errors) > 0 {
		log.Error("配置验证失败:")
		for _, err := range errors {
			log.Error("  - %s: %s", err.Field, err.Message)
		}
		os.Exit(1)
	}

	// 创建配置检查器并检查整个项目
	checker := config.NewConfigChecker(log)
	projectPath := getProjectPath(configFile, mainFlags)
	if err := checker.CheckProject(projectPath); err != nil {
		handleErrorAndExit(log, "项目检查失败: %v", err)
	}

	log.Success("配置文件验证通过")
}

// executeTask 执行ansible任务
// 参数:
//   - configFile: 配置文件路径
//   - flags: 命令行参数
//   - log: 日志记录器
func executeTask(configFile string, flags *CommandFlags, log *logger.Logger) {
	// 加载配置
	if configFile == "" {
		handleErrorAndExit(log, "未指定配置文件，请使用 --config 参数指定配置文件路径")
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		handleErrorAndExit(log, "加载配置失败: %v", err)
	}

	// 验证配置
	if errors := config.ValidateConfig(cfg); len(errors) > 0 {
		log.Error("配置验证失败:")
		for _, err := range errors {
			log.Error("  - %s: %s", err.Field, err.Message)
		}
		os.Exit(1)
	}

	// 创建执行器
	exec := executor.NewExecutor(cfg)

	// 设置verbose模式
	if flags.Verbose {
		exec.SetVerboseMode(true)
	}

	// 设置最大并行执行数
	if flags.Parallel > 0 {
		cfg.SSH.MaxParallel = flags.Parallel
	}

	// 处理标签过滤
	var tagList []string
	if flags.Tags != "" {
		tagList = strings.Split(flags.Tags, ",")
		for i := range tagList {
			tagList[i] = strings.TrimSpace(tagList[i])
		}
		log.Info("应用标签过滤: %v", tagList)
	}

	// 直接执行配置文件中的任务
	log.Info("开始执行配置: %s", configFile)
	// 从配置文件所在目录的tasks目录中查找任务文件
	taskDir := filepath.Join(filepath.Dir(configFile), "tasks")
	taskFile := filepath.Join(taskDir, "main.yaml")

	// 检查任务文件是否存在
	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		log.Error("任务文件不存在: %s", taskFile)
		os.Exit(1)
	}

	// 执行任务，并传入标签列表
	if err := exec.Execute(taskFile); err != nil {
		log.Error("执行任务失败: %v", err)
		os.Exit(1)
	}
	log.Success("执行完成")
}

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

// 创建项目模板文件
func createProjectTemplates(projectRoot string, log *logger.Logger) error {
	// 创建配置文件
	configPath := filepath.Join(projectRoot, "config.yaml")
	if err := createFileIfNotExist(configPath, configTemplate, log); err != nil {
		return fmt.Errorf("创建配置文件模板失败: %w", err)
	}

	// 创建任务文件
	taskPath := filepath.Join(projectRoot, "tasks", "example.yaml")
	if err := createFileIfNotExist(taskPath, taskTemplate, log); err != nil {
		return fmt.Errorf("创建任务文件模板失败: %w", err)
	}

	// 创建main任务文件
	mainTaskPath := filepath.Join(projectRoot, "tasks", "main.yaml")
	if err := createFileIfNotExist(mainTaskPath, mainTaskTemplate, log); err != nil {
		return fmt.Errorf("创建main任务文件模板失败: %w", err)
	}

	// 创建变量文件
	varsPath := filepath.Join(projectRoot, "vars", "main.yaml")
	if err := createFileIfNotExist(varsPath, varsTemplate, log); err != nil {
		return fmt.Errorf("创建变量文件模板失败: %w", err)
	}

	// 创建示例文件
	filePath := filepath.Join(projectRoot, "files", "app.conf")
	if err := createFileIfNotExist(filePath, exampleFileContent, log); err != nil {
		return fmt.Errorf("创建示例文件失败: %w", err)
	}

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
		filepath.Join(projectRoot, "vars"),
		filepath.Join(projectRoot, "executor"),
		filepath.Join(projectRoot, "tasks"),
		filepath.Join(projectRoot, "files"),
	}

	log.IncreaseIndent()
	for _, dir := range dirs {
		if err := createDirIfNotExist(dir, log); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	log.DecreaseIndent()

	// 创建项目模板文件
	if err := createProjectTemplates(projectRoot, log); err != nil {
		return err
	}

	log.Success("项目 %s 初始化完成", name)
	return nil
}

// 获取项目路径
func getProjectPath(configFile string, mainFlags *flag.FlagSet) string {
	projectPath := filepath.Dir(configFile)
	if mainFlags.NArg() > 0 {
		projectPath = mainFlags.Arg(0)
	}
	return projectPath
}

// main 程序入口函数，处理命令行参数并执行相应的命令
func main() {
	// 初始化日志记录器
	log := logger.New()
	// 设置日志级别（logger包中没有LevelInfo常量，所以这里暂时移除该行）

	// 初始化命令行参数
	mainFlags, initCmd, flags := initFlags()

	// 如果没有参数，显示帮助信息
	if len(os.Args) < 2 {
		showHelp(mainFlags)
		os.Exit(0)
	}

	// 根据第一个参数判断是主命令还是子命令
	switch os.Args[1] {
	case "init":
		// 解析init子命令参数
		initCmd.Parse(os.Args[2:])
		// 执行项目初始化
		if err := initProject(flags.ProjectName, flags.ProjectPath, log); err != nil {
			log.Error("项目初始化失败: %v", err)
			os.Exit(1)
		}
		log.Success("项目 %s 初始化完成", flags.ProjectName)

	case "check":
		// 解析主命令参数
		mainFlags.Parse(os.Args[2:])
		// 执行配置检查
		handleCheckCommand(flags.ConfigFile, mainFlags, log)

	case "help", "-h", "--help":
		showHelp(mainFlags)

	default:
		// 解析主命令参数
		mainFlags.Parse(os.Args[1:])
		// 执行任务
		executeTask(flags.ConfigFile, flags, log)
	}
}
