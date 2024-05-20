package conf

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func MustLoad(filePath string, configPtr interface{}) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("初始化加载配置文件出错: %v\n", err)
	}

	if err := yaml.Unmarshal(data, configPtr); err != nil {
		log.Fatalf("解析配置文件出错: %v\n", err)
	}
}
