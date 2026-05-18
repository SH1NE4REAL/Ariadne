package model

type MemoryRecord struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	SessionID    string  `json:"session_id"`
	MemoryType   string  `json:"memory_type"`
	Text         string  `json:"text"`
	Source       string  `json:"source"`
	Importance   int     `json:"importance"`
	CreatedAt    string  `json:"created_at"`
	MetadataJSON string  `json:"metadata_json"`
	Score        float64 `json:"score"`
}

type MemoryWriteRequest struct {
	UserID       string
	SessionID    string
	MemoryType   string
	Text         string
	Source       string
	Importance   int
	MetadataJSON string
}

type PreferenceConstraint struct {
	Domain string `json:"domain"` // hotel / transport / attraction / route / food / general

	// 旧版兼容：关键词
	ExcludeKeywords []string `json:"exclude_keywords"`
	PreferKeywords  []string `json:"prefer_keywords"`

	// 新版核心：标签
	PreferTags []string `json:"prefer_tags"`
	AvoidTags  []string `json:"avoid_tags"`

	Strength string `json:"strength"` // hard / soft

	// 优先级：当前请求 > 长期记忆 > 默认规则
	Priority int    `json:"priority"`
	Source   string `json:"source"` // current_request / long_term_memory / system_default

	Reason         string `json:"reason"`
	SourceMemoryID string `json:"source_memory_id"`
}
