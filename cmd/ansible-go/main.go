package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ape902/ansible-go/pkg/config"
	"github.com/ape902/ansible-go/pkg/executor"
	"github.com/ape902/ansible-go/pkg/vars"
)

func main() {
	// 命令行参数解析
	configFile := flag.String("config", "", "配置文件路径")
	playbook := flag.String("playbook", "", "playbook文件路径")
	verbose := flag.Bool("verbose", false, "启用详细日志输出")
	flag.Parse()

	// 检查必要参数
	if *configFile == "" && *playbook == "" {
		fmt.Println("错误: 必须提供配置文件或playbook文件路径")
		flag.Usage()
		os.Exit(1)
	}

	// 初始化日志
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化变量存储
	varStore := vars.NewStore()

	// 初始化执行器
	exec, err := executor.NewExecutor(cfg, varStore)
	if err != nil {
		log.Fatalf("初始化执行器失败: %v", err)
	}

	// 执行playbook
	if *playbook != "" {
		err = exec.ExecutePlaybook(*playbook)
		if err != nil {
			log.Fatalf("执行playbook失败: %v", err)
		}
	}

	log.Println("执行完成")
}