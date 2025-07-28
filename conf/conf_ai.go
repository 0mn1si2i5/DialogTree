// Path: ./conf/conf_ai.go

package conf

type Ai struct {
	Enable         bool         `yaml:"enable"`
	Nickname       string       `yaml:"nickname"`
	Avatar         string       `yaml:"avatar"`
	Abstract       string       `yaml:"abstract"`
	ContextLayers  int          `yaml:"contextLayers"`
	EmbeddingModel string       `yaml:"embeddingModel"`
	ChatAnywhere   ChatAnywhere `yaml:"chatAnywhere"`
	BackendAi      BackendAi    `yaml:"backendAi"`
	OpenAI         OpenAI       `yaml:"openai"`
	DeepSeek       DeepSeek     `yaml:"deepseek"`
}

type ChatAnywhere struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}

type BackendAi struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}

type OpenAI struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}

type DeepSeek struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}
