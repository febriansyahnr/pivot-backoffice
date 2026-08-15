package config

import "strings"

type AIConfig struct {
	Provider string        `mapstructure:"PROVIDER"`
	LLMs     []AILLMConfig `mapstructure:"LLMS"`
}

func (c *AIConfig) GetLLM(name string) *AILLMConfig {
	for _, llm := range c.LLMs {
		if strings.EqualFold(llm.Name, name) {
			return &llm
		}
	}
	return nil
}

type AILLMConfig struct {
	Name                            string        `mapstructure:"NAME"`
	BaseURL                         string        `mapstructure:"BASE_URL"`
	Default                         AIModelConfig `mapstructure:"DEFAULT"`
	ClickupCommandParsing           AIModelConfig `mapstructure:"CLICKUP_COMMAND_PARSING"`
	SlackConversation               AIModelConfig `mapstructure:"SLACK_CONVERSATION"`
	DeploymentDescriptionGeneration AIModelConfig `mapstructure:"DEPLOYMENT_DESCRIPTION_GENERATION"`
	GithubAnalysis                  AIModelConfig `mapstructure:"GITHUB_ANALYSIS"`
	SqlGeneration                   AIModelConfig `mapstructure:"SQL_GENERATION"`
	LogAnalysis                     AIModelConfig `mapstructure:"LOG_ANALYSIS"`
}

type AIModelConfig struct {
	Model       string  `mapstructure:"MODEL"`
	MaxToken    int     `mapstructure:"MAX_TOKEN"`
	Temperature float64 `mapstructure:"TEMPERATURE"`
}
