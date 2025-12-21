package mdtoc

import (
	"fmt"
	"regexp"
)

// Finalizer 替换占位符为实际行号
type Finalizer struct {
	parser  *Parser
	options Options
}

// NewFinalizer 创建新的终结器
func NewFinalizer(opts Options) *Finalizer {
	return &Finalizer{
		parser:  NewParser(opts),
		options: opts,
	}
}

// placeholderPattern 匹配占位符 {{LINE:anchor}}
var placeholderPattern = regexp.MustCompile(`\{\{LINE:([^}]+)\}\}`)

// Finalize 解析内容并替换行号占位符
// 如果 LineNumber 选项为 false，跳过处理
func (f *Finalizer) Finalize(content []byte) ([]byte, error) {
	if !f.options.LineNumber {
		return content, nil
	}

	// 解析最终内容中的所有标题（获取实际行号）
	headers, err := f.parser.ParseAllHeaders(content)
	if err != nil {
		return nil, err
	}

	// 构建锚点到行信息的映射
	anchorToLine := make(map[string]lineInfo)
	for _, h := range headers {
		anchorToLine[h.AnchorLink] = lineInfo{
			Line:    h.Line,
			EndLine: h.EndLine,
		}
	}

	// 替换所有占位符
	result := placeholderPattern.ReplaceAllFunc(content, func(match []byte) []byte {
		anchor := extractAnchor(match)
		if info, ok := anchorToLine[anchor]; ok {
			count := info.EndLine - info.Line + 1
			if f.options.ShowPath && f.options.FilePath != "" {
				return fmt.Appendf(nil, "%s:%d+%d", f.options.FilePath, info.Line, count)
			}
			return fmt.Appendf(nil, ":%d+%d", info.Line, count)
		}
		// 未找到锚点，保留占位符（不应该发生）
		return match
	})

	return result, nil
}

// lineInfo 存储行号信息
type lineInfo struct {
	Line    int
	EndLine int
}

// extractAnchor 从占位符中提取锚点名称
// 输入: {{LINE:anchor-name}}
// 输出: anchor-name
func extractAnchor(match []byte) string {
	sub := placeholderPattern.FindSubmatch(match)
	if len(sub) > 1 {
		return string(sub[1])
	}
	return ""
}
