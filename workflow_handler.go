package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coze-dev/coze-go"
	"github.com/gin-gonic/gin"

	"workflowui/workflows"
)

// CozeConfig Coze 配置
type CozeConfig struct {
	APIKey    string                              `json:"apikey"`
	Workflows map[string]workflows.WorkflowConfig `json:"workflows"`
}

// AppConfig 应用配置
type AppConfig struct {
	Coze CozeConfig `json:"coze"`
}

// WorkflowResponseData 工作流响应数据结构
type WorkflowResponseData struct {
	Text   string      `json:"text"`
	Output string      `json:"output"`
	Result interface{} `json:"result,omitempty"`
}

// WorkflowHandler 工作流处理器
type WorkflowHandler struct {
	cozeClient *coze.CozeAPI
	config     *AppConfig
}

// NewWorkflowHandler 创建新的工作流处理器
func NewWorkflowHandler(config *AppConfig) *WorkflowHandler {
	// 初始化 Coze 客户端
	authCli := coze.NewTokenAuth(config.Coze.APIKey)

	// 创建自定义 HTTP 客户端并设置超时
	httpClient := &http.Client{
		Timeout: 6000 * time.Second,
	}

	// 创建 Coze 客户端
	cozeClient := coze.NewCozeAPI(
		authCli,
		coze.WithBaseURL(coze.CnBaseURL),
		coze.WithHttpClient(httpClient),
	)

	return &WorkflowHandler{
		cozeClient: &cozeClient,
		config:     config,
	}
}

// ProcessWorkflow 处理工作流请求的通用方法
func (h *WorkflowHandler) ProcessWorkflow(c *gin.Context, workflowKey string) {
	// 检查工作流是否存在
	workflow, exists := h.config.Coze.Workflows[workflowKey]
	if !exists {
		c.JSON(400, gin.H{
			"error": "工作流不存在: " + workflowKey,
		})
		return
	}

	var request struct {
		ImageData  string                 `json:"image_data" binding:"required"`
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}

	// JSON绑定请求数据
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": "无效请求: " + err.Error(),
		})
		return
	}

	// 从 base64 字符串中提取实际的图片数据
	imageData := request.ImageData
	if strings.Contains(imageData, "base64,") {
		imageData = strings.Split(imageData, "base64,")[1]
	}

	// 解码 base64 数据
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "无效的图片数据: " + err.Error(),
		})
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "image-*.png")
	if err != nil {
		c.JSON(500, gin.H{
			"error": "创建临时文件失败: " + err.Error(),
		})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 写入图片数据到临时文件
	if _, err := tempFile.Write(imgBytes); err != nil {
		c.JSON(500, gin.H{
			"error": "写入临时文件失败: " + err.Error(),
		})
		return
	}
	tempFile.Close() // 关闭文件以确保数据被写入

	// 上传文件到 Coze
	file, err := os.Open(tempFile.Name())
	if err != nil {
		c.JSON(500, gin.H{
			"error": "打开临时文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	uploadReq := &coze.UploadFilesReq{
		File: file,
	}
	uploadResp, err := h.cozeClient.Files.Upload(c.Request.Context(), uploadReq)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "上传文件失败: " + err.Error(),
		})
		return
	}

	// 构建工作流参数
	data := map[string]any{
		"img": map[string]any{
			"type":    "image",
			"file_id": uploadResp.FileInfo.ID,
		},
	}

	// 添加额外参数
	if request.Parameters != nil {
		for key, value := range request.Parameters {
			data[key] = value
		}
	}

	workflowReq := &coze.RunWorkflowsReq{
		AppID:      workflow.AppID,
		WorkflowID: workflow.WorkflowID,
		Parameters: data,
	}

	// 执行工作流
	resp, err := h.cozeClient.Workflows.Runs.Create(context.Background(), workflowReq)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "工作流执行失败: " + err.Error(),
		})
		return
	}

	// 尝试解析工作流响应
	var workflowData WorkflowResponseData
	err = json.Unmarshal([]byte(resp.Data), &workflowData)
	if err != nil {
		// 如果解析失败，将整个响应作为输出
		workflowData.Output = resp.Data
		workflowData.Text = ""
	}

	// 如果Result字段为空，尝试解析整个响应数据
	if workflowData.Result == nil {
		var resultData interface{}
		if err := json.Unmarshal([]byte(resp.Data), &resultData); err == nil {
			workflowData.Result = resultData
		}
	}

	// 返回结果
	c.JSON(200, gin.H{
		"text":   workflowData.Text,
		"output": workflowData.Output,
		"result": workflowData.Result,
	})
}

// registerWorkflowRoutes 注册工作流路由
func registerWorkflowRoutes(r *gin.Engine, handler *WorkflowHandler) {
	// 主页面路由 - 工作流选择页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "workflow.html", nil)
	})

	// 工作流执行页面路由
	r.GET("/workflow/:workflow_key", func(c *gin.Context) {
		workflowKey := c.Param("workflow_key")

		// 检查工作流是否存在
		workflow, exists := handler.config.Coze.Workflows[workflowKey]
		if !exists {
			c.HTML(http.StatusNotFound, "workflow.html", gin.H{
				"error": "工作流不存在: " + workflowKey,
			})
			return
		}

		// 渲染工作流执行页面
		c.HTML(http.StatusOK, "workflow_execute.html", gin.H{
			"workflowKey":  workflowKey,
			"workflowName": workflow.Name,
		})
	})

	// 获取工作流配置信息
	r.GET("/api/workflows", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"workflows": handler.config.Coze.Workflows,
		})
	})

	s := r.Group("/api/workflow")
	{
		// 通用工作流处理路由
		s.POST("/:workflow_key", func(c *gin.Context) {
			workflowKey := c.Param("workflow_key")
			handler.ProcessWorkflow(c, workflowKey)
		})
	}
}
