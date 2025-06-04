# Workflow UI

一个基于 Coze 工作流的图片处理 Web 服务，支持图片上传和工作流自动处理。

## ✨ 功能特性

- 🚀 简单的 CLI 启动方式
- 🔧 灵活的配置文件管理
- 📸 支持图片上传和处理
- 🔄 支持多个 Coze 工作流
- 🌐 友好的 Web UI 界面
- 📝 REST API 接口

## 🚀 快速开始

### 1. 编译项目

```bash
go build -o workflowui main.go workflow_handler.go
```

### 2. 启动服务

```bash
./workflowui --port 8080
```

首次运行会自动创建 `config.json` 配置文件。

### 3. 配置 Coze

编辑 `config.json` 文件，设置您的 Coze API Key 和工作流信息：

```json
{
  "coze": {
    "apikey": "your_coze_api_key",
    "workflows": {
      "topic_ocr": {
        "name": "题目识别",
        "workflow_id": "your_workflow_id",
        "app_id": "your_app_id"
      }
    }
  }
}
```

### 4. 访问服务

打开浏览器访问 `http://localhost:8080`，选择工作流并上传图片进行处理。

## 📖 API 使用

### 调用工作流

```bash
POST /api/workflow/{workflow_key}
```

请求示例：

```bash
curl -X POST http://localhost:8080/api/workflow/topic_ocr \
  -H "Content-Type: application/json" \
  -d '{
    "image_data": "data:image/png;base64,iVBORw0KGg...",
    "parameters": {
      "discipline": "物理",
      "topic_type": "选择题"
    }
  }'
```

## ⚙️ 命令行参数

```bash
./workflowui --help
```

- `--port, -p`: 服务器端口 (默认: 8080)
- `--config, -c`: 配置文件路径 (默认: config.json)  
- `--debug, -d`: 启用调试模式

## 📁 项目结构

```
.
├── main.go              # 主程序入口
├── workflow_handler.go  # 工作流处理逻辑
├── config.json         # 配置文件
├── workflows/          # 工作流定义
├── templates/          # Web 页面模板
└── static/            # 静态资源
```

## 🔧 添加新工作流

在 `config.json` 的 `workflows` 部分添加新配置：

```json
"your_workflow": {
  "name": "自定义工作流",
  "workflow_id": "workflow_id_from_coze",
  "app_id": "app_id_from_coze",
  "parameters": [
    {
      "key": "param1",
      "label": "参数1",
      "type": "text",
      "required": true
    }
  ]
}
```

## 📋 系统要求

- Go 1.24+
- Coze API 账号和密钥

## 🤝 支持的工作流类型

目前支持：
- 题目知识点识别 (topic_ocr)
- 药品说明书识别 (med_instructions)

可通过配置文件扩展更多工作流类型。 