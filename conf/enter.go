// Path: ./conf/enter.go

package conf

type Config struct {
	System System `yaml:"system"`
	Logrus Logrus `yaml:"logrus"`
	DB     DB     `yaml:"db"`
	Redis  Redis  `yaml:"redis"`
	Ai     Ai     `yaml:"ai"`
	Vector Vector `yaml:"vector"`
}
