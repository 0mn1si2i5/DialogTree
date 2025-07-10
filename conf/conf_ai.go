// Path: ./conf/conf_ai.go

package conf

type Ai struct {
	Enable       bool         `yaml:"enable"`
	Nickname     string       `yaml:"nickname"`
	Avatar       string       `yaml:"avatar"`
	Abstract     string       `yaml:"abstract"`
	ChatAnywhere ChatAnywhere `yaml:"chatAnywhere"`
	BackendAi    BackendAi    `yaml:"backendAi"`
}

type ChatAnywhere struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}

type BackendAi struct {
	Model     string `yaml:"model"`
	SecretKey string `yaml:"secretKey"`
}
