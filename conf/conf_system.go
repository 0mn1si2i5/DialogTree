// Path: ./conf/conf_system.go

package conf

type System struct {
	Mode    string `yaml:"mode"`
	Ip      string `yaml:"ip"`
	Port    string `yaml:"port"`
	Env     string `yaml:"env"`
	GinMode string `yaml:"ginMode"`
}

func (s System) Addr() string {
	host := s.Ip
	if host == "" {
		host = "localhost"
	}
	return host + ":" + s.Port
}
