package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"io/fs"

	"embed"
	"workflowui/workflows"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var (
	defaultConfigFile = "config.json"
)

// 使用 workflows 包的类型别名，保持兼容性
type WorkflowConfig = workflows.WorkflowConfig
type WorkflowParameterConfig = workflows.WorkflowParameterConfig

// loadConfig 从配置文件加载配置
func loadConfig(configFile string) (*AppConfig, error) {
	if configFile == "" {
		// 使用当前工作目录下的默认配置文件
		workDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("获取工作目录失败: %v", err)
		}
		configFile = filepath.Join(workDir, defaultConfigFile)
	}

	// 如果文件不存在，创建默认配置文件
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := &AppConfig{
			Coze: CozeConfig{
				APIKey:    "your_coze_api_key_here",
				Workflows: workflows.GetAllWorkflows(),
			},
		}
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("创建默认配置失败: %v", err)
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return nil, fmt.Errorf("写入默认配置文件失败: %v", err)
		}
		log.Printf("已创建默认配置文件: %s", configFile)
		log.Printf("请编辑配置文件并设置正确的 Coze API Key 和工作流 ID")
	}

	// 读取并解析配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// setupRouter 设置路由
func setupRouter(debug bool, config *AppConfig) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.SetHTMLTemplate(template.Must(template.New("").Funcs(r.FuncMap).ParseFS(templateFS, "templates/*")))

	subFS, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/public", http.FS(subFS))

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, authentication, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建工作流处理器
	workflowHandler := NewWorkflowHandler(config)

	// 注册工作流路由（移除认证中间件）
	registerWorkflowRoutes(r, workflowHandler)

	return r
}

// startServer 启动服务器
func startServer(port string, debug bool, configFile string) error {
	config, err := loadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 验证配置
	if config.Coze.APIKey == "" || config.Coze.APIKey == "your_coze_api_key_here" {
		return fmt.Errorf("请在配置文件中设置正确的 Coze API Key")
	}

	if len(config.Coze.Workflows) == 0 {
		return fmt.Errorf("配置文件中没有定义工作流")
	}

	// 验证每个工作流的 AppID
	for key, workflow := range config.Coze.Workflows {
		if workflow.AppID == "" || workflow.AppID == "your_app_id_here" {
			return fmt.Errorf("请在配置文件中为工作流 '%s' 设置正确的 App ID", key)
		}
	}

	r := setupRouter(debug, config)

	log.Printf("服务器启动在端口 %s", port)
	log.Printf("可用的工作流:")
	for key, workflow := range config.Coze.Workflows {
		log.Printf("  - %s: %s (ID: %s, AppID: %s)", key, workflow.Name, workflow.WorkflowID, workflow.AppID)
	}

	return r.Run(":" + port)
}

func main() {
	app := &cli.App{
		Name:  "workflow-ui",
		Usage: "基于 Coze 工作流的 Web UI 服务",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   "8080",
				Usage:   "服务器端口",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.json",
				Usage:   "配置文件路径",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "启用调试模式",
			},
		},
		Action: func(c *cli.Context) error {
			port := c.String("port")
			configFile := c.String("config")
			debug := c.Bool("debug")

			return startServer(port, debug, configFile)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
