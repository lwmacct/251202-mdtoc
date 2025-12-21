package mdtoc

import (
	"bytes"
)

// Cleaner 处理 TOC 块的清理
type Cleaner struct {
	marker []byte
}

// NewCleaner 创建新的清理器
func NewCleaner(marker string) *Cleaner {
	if marker == "" {
		marker = DefaultMarker
	}
	return &Cleaner{marker: []byte(marker)}
}

// Clean 删除所有 TOC 块，返回干净的内容
// TOC 块格式：<!--TOC--> ... <!--TOC-->
// 同时删除 TOC 块后紧跟的一个空行（因为插入时会添加）
func (c *Cleaner) Clean(content []byte) []byte {
	lines := bytes.Split(content, []byte("\n"))

	// 找到 frontmatter 结束位置，跳过 frontmatter
	frontmatterEnd := FindFrontmatterEnd(lines)
	startLine := 0
	if frontmatterEnd >= 0 {
		startLine = frontmatterEnd + 1
	}

	// 找到所有 TOC 块 (成对的 <!--TOC-->)
	var blocks []tocBlockRange
	pendingStart := -1

	for i := startLine; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
		if bytes.Equal(trimmed, c.marker) {
			if pendingStart == -1 {
				pendingStart = i
			} else {
				blocks = append(blocks, tocBlockRange{start: pendingStart, end: i})
				pendingStart = -1
			}
		}
	}

	// 没有 TOC 块，返回原内容
	if len(blocks) == 0 {
		return content
	}

	// 标记要删除的行
	deleteLines := make(map[int]bool)
	for _, block := range blocks {
		// 标记 TOC 块内所有行
		for i := block.start; i <= block.end; i++ {
			deleteLines[i] = true
		}
		// 删除 TOC 块后紧跟的空行
		if block.end+1 < len(lines) && len(bytes.TrimSpace(lines[block.end+1])) == 0 {
			deleteLines[block.end+1] = true
		}
	}

	// 构建干净内容
	cleanedLines := make([][]byte, 0, len(lines)-len(deleteLines))
	for i, line := range lines {
		if !deleteLines[i] {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return bytes.Join(cleanedLines, []byte("\n"))
}

// tocBlockRange 记录 TOC 块的范围
type tocBlockRange struct {
	start int // 开始行 (0-based)
	end   int // 结束行 (0-based)
}

// ReplaceTOCContent 替换现有 TOC 块中的内容
// 保持标记位置不变，只更新标记之间的内容
func (c *Cleaner) ReplaceTOCContent(content []byte, newTOC string) []byte {
	lines := bytes.Split(content, []byte("\n"))

	// 找到 frontmatter 结束位置
	frontmatterEnd := FindFrontmatterEnd(lines)
	startLine := 0
	if frontmatterEnd >= 0 {
		startLine = frontmatterEnd + 1
	}

	// 找到第一对 TOC 标记
	var markerPositions []int
	for i := startLine; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
		if bytes.Equal(trimmed, c.marker) {
			markerPositions = append(markerPositions, i)
			if len(markerPositions) >= 2 {
				break
			}
		}
	}

	// 没有成对标记，返回原内容
	if len(markerPositions) < 2 {
		return content
	}

	startMarker := markerPositions[0]
	endMarker := markerPositions[1]

	// 构建新内容
	result := make([][]byte, 0, len(lines))

	// 添加开始标记之前的内容
	result = append(result, lines[:startMarker]...)

	// 添加开始标记 + 空行 + 新TOC + 空行 + 结束标记
	result = append(result, []byte(DefaultMarker))
	result = append(result, []byte(""))
	result = append(result, []byte(newTOC))
	result = append(result, []byte(""))
	result = append(result, []byte(DefaultMarker))

	// 跳过结束标记后的空行（避免累积）
	nextLine := endMarker + 1
	if nextLine < len(lines) && len(bytes.TrimSpace(lines[nextLine])) == 0 {
		nextLine++
	}

	// 添加结束标记之后的内容
	if nextLine < len(lines) {
		result = append(result, lines[nextLine:]...)
	}

	return bytes.Join(result, []byte("\n"))
}
