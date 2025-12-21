// Package mdtoc 提供 Markdown 目录生成功能
package mdtoc

// Header 表示一个 Markdown 标题
type Header struct {
	Level      int    // 标题层级 (1-6)
	Text       string // 标题文本 (原始文本，去除 # 和前后空格)
	AnchorLink string // 锚点链接 (GitHub 风格)
	Line       int    // 标题所在行 (1-based)
	EndLine    int    // 内容结束行 (1-based)，下一个标题前一行或文件末尾
}

// Options 配置 TOC 生成选项
type Options struct {
	MinLevel   int    // 最小标题层级 (默认 1)
	MaxLevel   int    // 最大标题层级 (默认 3)
	Ordered    bool   // 使用有序列表
	LineNumber bool   // 显示行号范围 (:start:end)
	ShowPath   bool   // 显示文件路径 (path:start:end)
	FilePath   string // 当前处理的文件路径
	SectionTOC bool   // 章节模式：每个 H1 后生成独立的子目录
	ShowAnchor bool   // 显示锚点链接 [标题](#anchor)，预览默认 false，写入强制 true
	TOCTitle   string // TOC 标题文本 (如 "文档目录")，非空时在 TOC 内容前添加 ## 标题
}

// Section 表示一个章节 (H1 及其子标题)
type Section struct {
	Title      *Header   // H1 标题
	SubHeaders []*Header // 子标题 (H2-H6)
}

// DefaultOptions 返回默认配置
func DefaultOptions() Options {
	return Options{
		MinLevel:   1,
		MaxLevel:   3,
		Ordered:    false,
		SectionTOC: true, // 默认启用章节模式：在每个 H1 下生成独立子目录
		ShowAnchor: true, // 默认生成链接格式 [标题](#anchor)
	}
}

// TOCMarker 表示 TOC 标记位置
type TOCMarker struct {
	StartLine int // 第一个标记所在行号 (0-based)
	EndLine   int // 第二个标记所在行号 (0-based), -1 表示只有一个标记
	Found     bool
}

// DefaultMarker 是默认的 TOC 标记字符串
const DefaultMarker = "<!--TOC-->"
