package config

import "errors"

// Config 配置结构
type Config struct {
	MinLevel   int    `koanf:"min-level" yaml:"minLevel"`     // 最小标题层级 (1-6)
	MaxLevel   int    `koanf:"max-level" yaml:"maxLevel"`     // 最大标题层级 (1-6)
	Ordered    bool   `koanf:"ordered" yaml:"ordered"`        // 使用有序列表
	LineNumber bool   `koanf:"line-number" yaml:"lineNumber"` // 显示行号范围
	ShowPath   bool   `koanf:"path" yaml:"path"`              // 显示文件路径
	Global     bool   `koanf:"global" yaml:"global"`          // 全局模式
	Anchor     bool   `koanf:"anchor" yaml:"anchor"`          // 显示锚点链接
	TOCTitle   string `koanf:"toc-title" yaml:"tocTitle"`     // TOC 标题
	Force      bool   `koanf:"force" yaml:"force"`            // 强制生成 TOC（忽略标记检查）
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
