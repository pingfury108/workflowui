package workflows

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	Name       string                    `json:"name"`
	WorkflowID string                    `json:"workflow_id"`
	AppID      string                    `json:"app_id"`
	Parameters []WorkflowParameterConfig `json:"parameters,omitempty"`
}

// WorkflowParameterConfig 工作流参数配置
type WorkflowParameterConfig struct {
	Key          string   `json:"key"`
	Label        string   `json:"label"`
	Type         string   `json:"type"` // text, select
	Required     bool     `json:"required"`
	Placeholder  string   `json:"placeholder,omitempty"`
	Options      []string `json:"options,omitempty"` // for select type
	DefaultValue string   `json:"default_value,omitempty"`
}

var workflows = make(map[string]WorkflowConfig)

// GetAllWorkflows 获取所有工作流配置
func GetAllWorkflows() map[string]WorkflowConfig {
	return workflows
} 