package mdtoc

import (
	"bytes"
	"strings"
)

// Builder 构建包含 TOC 的内容
type Builder struct {
	marker  string
	options Options
}

// NewBuilder 创建新的构建器
func NewBuilder(opts Options) *Builder {
	return &Builder{
		marker:  DefaultMarker,
		options: opts,
	}
}

// Build 在干净内容中插入 TOC 块
// TOC 使用占位符格式的行号：{{LINE:anchor}}
// 这些占位符会在 Finalize 阶段被替换为实际行号
func (b *Builder) Build(cleanContent []byte, doc *Document) []byte {
	if len(doc.InsertionPoints) == 0 {
		return cleanContent
	}

	lines := bytes.Split(cleanContent, []byte("\n"))

	// 创建插入点映射：行号 -> 插入点索引
	insertAt := make(map[int]int)
	for i, point := range doc.InsertionPoints {
		insertAt[point.InsertBeforeLine] = i
	}

	// 标记需要跳过的行（插入点前的空行，因为 TOC 块会自己添加空行）
	skipLines := make(map[int]bool)
	for insertLine := range insertAt {
		// 检查插入点前面的行是否为空行
		if insertLine > 0 && len(bytes.TrimSpace(lines[insertLine-1])) == 0 {
			skipLines[insertLine-1] = true
		}
	}

	// 构建结果 (预分配容量)
	result := make([][]byte, 0, len(lines)+len(doc.InsertionPoints)*10)

	for i, line := range lines {
		// 跳过插入点前的空行
		if skipLines[i] {
			continue
		}

		// 检查是否需要在此行前插入 TOC
		if pointIdx, ok := insertAt[i]; ok {
			tocBlock := b.buildTOCBlock(&doc.InsertionPoints[pointIdx])
			result = append(result, tocBlock...)
		}
		result = append(result, line)
	}

	return bytes.Join(result, []byte("\n"))
}

// BuildPreview 生成预览 TOC（用于 stdout 输出，不带占位符）
func (b *Builder) BuildPreview(doc *Document) string {
	if len(doc.InsertionPoints) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, point := range doc.InsertionPoints {
		// 章节模式显示章节标题
		if point.SectionTitle != "" {
			sb.WriteString("### ")
			sb.WriteString(point.SectionTitle)
			sb.WriteString("\n\n")
		}

		// 生成 TOC 条目（不带占位符）
		entries := b.buildPreviewEntries(point.Headers)
		for _, entry := range entries {
			sb.WriteString(entry)
			sb.WriteString("\n")
		}

		// 章节之间添加空行
		if i < len(doc.InsertionPoints)-1 {
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String())
}

// buildTOCBlock 构建单个 TOC 块
// 格式：空行 + <!--TOC--> + 空行 + [TOC标题] + TOC内容 + 空行 + <!--TOC--> + 空行
func (b *Builder) buildTOCBlock(point *InsertionPoint) [][]byte {
	// 预分配容量：6 个固定行 + TOC 标题可能 2 行 + 标题数量
	block := make([][]byte, 0, 8+len(point.Headers))

	// 空行 + 开始标记 + 空行
	block = append(block, []byte(""))
	block = append(block, []byte(b.marker))
	block = append(block, []byte(""))

	// 可选的 TOC 标题
	if b.options.TOCTitle != "" {
		block = append(block, []byte("## "+b.options.TOCTitle))
		block = append(block, []byte(""))
	}

	// TOC 条目
	tocLines := b.buildTOCEntries(point.Headers)
	for _, line := range tocLines {
		block = append(block, []byte(line))
	}

	// 空行 + 结束标记 + 空行
	block = append(block, []byte(""))
	block = append(block, []byte(b.marker))
	block = append(block, []byte(""))

	return block
}

// buildTOCEntries 构建 TOC 条目列表
// 如果启用行号，使用占位符格式 {{LINE:anchor}}
func (b *Builder) buildTOCEntries(headers []*Header) []string {
	if len(headers) == 0 {
		return nil
	}

	return renderTOCEntries(headers, b.options, minHeaderLevel(headers), func(h *Header) string {
		return "{{LINE:" + h.AnchorLink + "}}"
	})
}

// buildPreviewEntries 构建预览 TOC 条目（带实际行号，不是占位符）
func (b *Builder) buildPreviewEntries(headers []*Header) []string {
	if len(headers) == 0 {
		return nil
	}

	return renderTOCEntries(headers, b.options, minHeaderLevel(headers), func(h *Header) string {
		return formatHeaderLineRange(b.options, h)
	})
}
