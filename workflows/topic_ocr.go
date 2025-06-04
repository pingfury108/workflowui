package workflows

// GetTopicOCRWorkflow 获取题目知识点识别工作流配置
func GetTopicOCRWorkflow() WorkflowConfig {
	return WorkflowConfig{
		Name:       "题目知识点识别",
		WorkflowID: "your_topic_workflow_id_here",
		AppID:      "your_app_id_here",
		Parameters: []WorkflowParameterConfig{
			{
				Key:          "discipline",
				Label:        "学科",
				Type:         "select",
				Required:     true,
				Options:      []string{"物理", "数学", "化学", "生物"},
				DefaultValue: "物理",
			},
			{
				Key:         "topic_type",
				Label:       "题目类型",
				Type:        "text",
				Required:    true,
				Placeholder: "可选，如：选择题、填空题等",
			},
		},
	}
} 

func init() {
	workflows["topic_ocr"] = GetTopicOCRWorkflow()
}