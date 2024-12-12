package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config 配置加载逻辑
type Config struct {
	Database struct {
		User     string `yaml:"user"`     // 数据库用户名
		Password string `yaml:"password"` // 数据库密码
		Host     string `yaml:"host"`     // 数据库主机
		Port     string `yaml:"port"`     // 数据库端口
		Name     string `yaml:"name"`     // 数据库名称
		Charset  string `yaml:"charset"`  // 数据库字符集
	} `yaml:"database"`

	Cloudflare struct {
		Email string `yaml:"email"` // cloudflare email
		Key   string `yaml:"key"`   // cloudflare key
	} `yaml:"cloudflare"`

	Telegram struct {
		Id          int64  `yaml:"id"`          // telegram机器人ID
		Token       string `yaml:"token"`       // telegram机器人token
		ApiEndpoint string `yaml:"apiEndpoint"` // telegramAPI
	} `yaml:"telegram"`

	Network struct {
		EnableProxy      bool     `yaml:"enable_proxy"`       // 是否启用telegram代理
		Proxy            string   `yaml:"proxy"`              // telegram网络代理地址
		EnableCheckProxy bool     `yaml:"enable_check_proxy"` // 是否启用检测代理
		CheckProxy       []string `yaml:"check_proxy"`        // 检测代理地址
	} `yaml:"network"`

	Check struct {
		IpCheckTime time.Duration `yaml:"ip_check_time"` // 每秒检测时间
		CheckTime   time.Duration `yaml:"check_time"`    // 每分钟检测时间
	} `yaml:"network"`
}

// LoadConfig 加载 YAML 配置文件
func LoadConfig(filePath string) (*Config, error) {
	// 如果 filePath 为空，则使用默认相对路径
	if filePath == "1" {
		exePath, err := os.Executable() // 获取当前可执行文件路径
		if err != nil {
			log.Printf("无法获取可执行文件路径: %v", err)
			return nil, err
		}
		filePath = filepath.Join(filepath.Dir(exePath), "conf.yaml") // 默认文件名
	}
	workingDir, _ := os.Getwd()
	// 查找项目根目录
	rootDir, err := findProjectRoot(workingDir)
	if err != nil {
		fmt.Println("错误:", err)
		return nil, fmt.Errorf("错误:%s", err)
	}
	//filePath = filepath.Join(filepath.Dir(rootDir), "conf.yaml") // 默认文件名
	//fmt.Println(rootDir + "/conf.yaml") //打印conf.yaml路径情况，进行调试
	file, err := os.Open(rootDir + "/conf.yaml")
	if err != nil {
		log.Printf("无法打开配置文件 %s: %v", filePath, err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Printf("解析配置文件 %s 失败: %v", filePath, err)
		return nil, err
	}

	return &config, nil
}

// 查找项目根目录
func findProjectRoot(startDir string) (string, error) {
	// 假设根目录有一个标志文件，如 go.mod 或 README.md
	// 你可以根据项目的实际情况选择其他标志文件
	for {
		// 检查是否有 go.mod 文件（可以修改为其他标志文件）
		if _, err := os.Stat(filepath.Join(startDir, "go.mod")); err == nil {
			return startDir, nil
		}

		// 向上遍历父目录
		parentDir := filepath.Dir(startDir)
		if parentDir == startDir { // 如果已经到达根目录，停止查找
			break
		}
		startDir = parentDir
	}

	return "", fmt.Errorf("项目根目录未找到")
}
