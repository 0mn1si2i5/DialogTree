// Path: ./core/init_conf.go

package core

import (
	"dialogTree/conf"
	"dialogTree/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// ReadConf 读取 settings.yaml 设置文件并解析配置
// 如果读取或解析过程中出现错误，将会触发panic
func ReadConf(quiet bool) (c *conf.Config) {
	// 从指定的配置文件路径读取内容
	byteData, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	// new 返回的就是指针
	c = new(conf.Config)

	// 将YAML格式的配置文件内容解析到config结构体中
	err = yaml.Unmarshal(byteData, c)
	if err != nil {
		panic(fmt.Sprintln("yaml unmarshal err: ", err))
	}
	if !quiet {
		// 打印配置文件读取成功的消息
		fmt.Printf("configuration of: %s success!\n", "config.yaml")
	}
	return
}

func SetConf() {
	byteData, err := yaml.Marshal(global.Config)
	if err != nil {
		logrus.Errorln("yaml marshal err: ", err)
		return
	}
	err = os.WriteFile("settings.yaml", byteData, 0666)
	if err != nil {
		logrus.Errorln("yaml write err: ", err)
		return
	}
	logrus.Info("settings.yaml write successful")
}
