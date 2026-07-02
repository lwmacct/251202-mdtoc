package config

import "errors"

// Config 配置结构
type Config struct {
	MinLevel   int    `json:"min-level"   desc:"最小标题层级 (1-6)"`
	MaxLevel   int    `json:"max-level"   desc:"最大标题层级 (1-6)"`
	Ordered    bool   `json:"ordered"     desc:"使用有序列表 (1. 2. 3.)"`
	LineNumber bool   `json:"line-number" desc:"显示行号范围 (:start+count=end)"`
	ShowPath   bool   `json:"path"        desc:"显示文件路径 (path:start+count=end)"`
	Global     bool   `json:"global"      desc:"全局模式: 生成完整文档的单一目录 (默认为章节模式)"`
	Anchor     bool   `json:"anchor"      desc:"预览时显示锚点链接 [标题](#anchor)"`
	TOCTitle   string `json:"toc-title"   desc:"TOC 标题 (如 '文档目录'，将在 TOC 内生成 ## 文档目录，设为空则不生成标题)"`
	Force      bool   `json:"force"       desc:"强制生成 TOC，即使文件中没有 <!--TOC--> 标记"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MinLevel:   1,
		MaxLevel:   3,
		Ordered:    false,
		LineNumber: true,
		ShowPath:   false,
		Global:     false,
		Anchor:     false,
		TOCTitle:   "Table of Contents",
	}
}

// Validate 校验配置有效性
func (c *Config) Validate() error {
	if c.MinLevel < 1 || c.MinLevel > 6 {
		return errors.New("min-level 必须在 1-6 之间")
	}
	if c.MaxLevel < 1 || c.MaxLevel > 6 {
		return errors.New("max-level 必须在 1-6 之间")
	}
	if c.MinLevel > c.MaxLevel {
		return errors.New("min-level 不能大于 max-level")
	}
	return nil
}
