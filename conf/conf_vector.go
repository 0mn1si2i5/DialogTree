// Path: ./conf/conf_vector.go

package conf

type Vector struct {
	Enable              bool    `yaml:"enable"`
	Provider            string  `yaml:"provider"`
	Qdrant              Qdrant  `yaml:"qdrant"`
	TopK                int     `yaml:"topK"`
	SimilarityThreshold float64 `yaml:"similarityThreshold"`
}

type Qdrant struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Collection string `yaml:"collection"`
	ApiKey     string `yaml:"apiKey"`
}