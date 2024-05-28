package message

type Metadata map[string]string

// Get 返回指定键的元数据值
func (m Metadata) Get(key string) string {
	if v, ok := m[key]; ok {
		return v
	}

	return ""
}

// Set 设置元数据值
func (m Metadata) Set(key, value string) {
	m[key] = value
}
