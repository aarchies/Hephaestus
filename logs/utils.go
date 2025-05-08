package logs

import (
	"fmt"
	"path/filepath"
)

// getShortFile 获取文件的短路径
func getShortFile(file string) string {
	return filepath.Base(file)
}

// formatCaller 格式化调用者信息
func formatCaller(caller *CallerInfo, fullPath bool) string {
	if caller == nil {
		return "???:0"
	}

	file := caller.File
	if !fullPath {
		file = getShortFile(file)
	}

	return fmt.Sprintf("%s:%d", file, caller.Line)
}
