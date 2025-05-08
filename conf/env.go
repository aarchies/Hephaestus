package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// EnvPrefix 环境变量前缀
const EnvPrefix = "NETLEAP_"

// LoadEnv 从环境变量加载配置
func LoadEnv(config interface{}) error {
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return loadEnvRecursive(val, "")
}

// loadEnvRecursive 递归加载环境变量
func loadEnvRecursive(val reflect.Value, prefix string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 获取环境变量名
		envTag := fieldType.Tag.Get("env")
		if envTag == "-" {
			continue
		}
		if envTag == "" {
			envTag = strings.ToUpper(fieldType.Name)
		}

		envName := EnvPrefix + prefix + envTag

		// 处理嵌套结构
		if field.Kind() == reflect.Struct {
			newPrefix := prefix
			if envTag != "" {
				newPrefix += envTag + "_"
			}
			if err := loadEnvRecursive(field, newPrefix); err != nil {
				return err
			}
			continue
		}

		// 获取环境变量值
		envValue := os.Getenv(envName)
		if envValue == "" {
			continue
		}

		// 设置字段值
		if err := setFieldValue(field, envValue); err != nil {
			return fmt.Errorf("set field %s failed: %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

type EnvFlag struct {
	Name    string
	AltName string
}

func NewEnvFlag(name string) EnvFlag {
	return EnvFlag{
		Name:    name,
		AltName: NormalizeEnvName(name),
	}
}

func (f EnvFlag) GetValue(defaultValue func() string) string {
	if v, found := os.LookupEnv(f.Name); found {
		return v
	}
	if len(f.AltName) > 0 {
		if v, found := os.LookupEnv(f.AltName); found {
			return v
		}
	}

	return defaultValue()
}

func (f EnvFlag) GetValueAsInt(defaultValue int) int {
	useDefaultValue := false
	s := f.GetValue(func() string {
		useDefaultValue = true
		return ""
	})
	if useDefaultValue {
		return defaultValue
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(v)
}

func NormalizeEnvName(name string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(name)), ".", "_")
}

func getExecutableDir() string {
	exec, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exec)
}

func getExecutableSubDir(dir string) func() string {
	return func() string {
		return filepath.Join(getExecutableDir(), dir)
	}
}

func GetPluginDirectory() string {
	const name = "netleap.location.module"
	pluginDir := NewEnvFlag(name).GetValue(getExecutableSubDir("plugins"))
	return pluginDir
}

func GetConfigurationPath() string {
	const name = "netleap.location.config"
	configPath := NewEnvFlag(name).GetValue(getExecutableDir)
	return filepath.Join(configPath, "config.json")
}

func GetConfDirPath() string {
	const name = "netleap.location.confdir"
	configPath := NewEnvFlag(name).GetValue(func() string { return "" })
	return configPath
}
