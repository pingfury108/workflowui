package workflows

// GetMedInstructionsWorkflow 获取药品说明书识别工作流配置
func GetMedInstructionsWorkflow() WorkflowConfig {
	return WorkflowConfig{
		Name:       "药品说明书识别",
		WorkflowID: "your_med_workflow_id_here",
		AppID:      "your_app_id_here",
		Parameters: []WorkflowParameterConfig{},
	}
} 

func init() {
	workflows["med_instructions"] = GetMedInstructionsWorkflow()
}