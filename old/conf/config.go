package conf

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func MustLoad(filePath string, configPtr interface{}) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Init MestLoad Config Error: %s\n", err.Error())
	}

	if err := yaml.Unmarshal(data, configPtr); err != nil {
		log.Fatalf("Analyze MestLoad Config Error: %s\n", err.Error())
	}
}
