package mdtoc

import (
	"bytes"
)

// Analyzer 分析文档结构，找到 TOC 插入点
type Analyzer struct {
	parser  *Parser
	options Options
}

// NewAnalyzer 创建新的分析器
func NewAnalyzer(opts Options) *Analyzer {
	return &Analyzer{
		parser:  NewParser(opts),
		options: opts,
	}
}

// Analyze 分析文档，返回文档结构和插入点
// 统一逻辑：TOC 插入在第一个 H2 之前
func (a *Analyzer) Analyze(content []byte) (*Document, error) {
	lines := bytes.Split(content, []byte("\n"))

	// 检测 frontmatter
	frontmatter := a.detectFrontmatter(lines)

	// 解析所有标题
	headers, err := a.parser.ParseAllHeaders(content)
	if err != nil {
		return nil, err
	}

	// 根据模式找到插入点
	var insertPoints []InsertionPoint
	if a.options.SectionTOC {
		insertPoints = a.findSectionInsertPoints(lines, headers)
	} else {
		insertPoints = a.findGlobalInsertPoint(lines, headers, frontmatter)
	}

	return &Document{
		Frontmatter:     frontmatter,
		InsertionPoints: insertPoints,
	}, nil
}

// detectFrontmatter 检测 YAML frontmatter
func (a *Analyzer) detectFrontmatter(lines [][]byte) FrontmatterInfo {
	endLine := FindFrontmatterEnd(lines)
	return FrontmatterInfo{
		Exists:  endLine >= 0,
		EndLine: endLine,
	}
}

// findGlobalInsertPoint 找到全局模式的插入点（第一个 H2 之前）
func (a *Analyzer) findGlobalInsertPoint(lines [][]byte, headers []*Header, fm FrontmatterInfo) []InsertionPoint {
	startLine := 0
	if fm.Exists {
		startLine = fm.EndLine + 1
	}

	// 找到第一个 H2
	firstH2Line := a.findFirstH2InRange(headers, startLine, len(lines))
	if firstH2Line == -1 {
		// 没有 H2，不插入 TOC
		return nil
	}

	// 筛选符合层级范围的标题
	filteredHeaders := a.filterHeaders(headers)
	if len(filteredHeaders) == 0 {
		return nil
	}

	return []InsertionPoint{{
		InsertBeforeLine: firstH2Line,
		SectionTitle:     "",
		Headers:          filteredHeaders,
	}}
}

// findSectionInsertPoints 找到章节模式的插入点（每个 H1 后的第一个 H2 之前）
func (a *Analyzer) findSectionInsertPoints(lines [][]byte, headers []*Header) []InsertionPoint {
	// 按 H1 分割成章节
	sections := SplitSections(headers)

	insertPoints := make([]InsertionPoint, 0, len(sections))

	for i, section := range sections {
		// 跳过没有子标题的章节
		if len(section.SubHeaders) == 0 {
			continue
		}

		// 检查是否至少有一个 H2
		hasH2 := false
		for _, h := range section.SubHeaders {
			if h.Level == 2 {
				hasH2 = true
				break
			}
		}
		if !hasH2 {
			continue
		}

		// 确定章节范围
		sectionStart := section.Title.Line - 1 // 转换为 0-based
		sectionEnd := len(lines)
		if i+1 < len(sections) {
			sectionEnd = sections[i+1].Title.Line - 1 // 下一个 H1 行之前
		}

		// 找到该章节内的第一个 H2
		firstH2Line := a.findFirstH2InRange(section.SubHeaders, sectionStart, sectionEnd)
		if firstH2Line == -1 {
			continue
		}

		// 筛选符合层级范围的子标题
		filteredHeaders := a.filterSectionHeaders(section.SubHeaders)
		if len(filteredHeaders) == 0 {
			continue
		}

		insertPoints = append(insertPoints, InsertionPoint{
			InsertBeforeLine: firstH2Line,
			SectionTitle:     section.Title.Text,
			Headers:          filteredHeaders,
		})
	}

	return insertPoints
}

// findFirstH2InRange 在指定范围内找到第一个 H2
// headers 是已解析的标题列表，startLine/endLine 是 0-based 行号范围
// 返回 H2 所在行号 (0-based)，未找到返回 -1
func (a *Analyzer) findFirstH2InRange(headers []*Header, startLine, endLine int) int {
	for _, h := range headers {
		line := h.Line - 1 // 转换为 0-based
		if line >= startLine && line < endLine && h.Level == 2 {
			return line
		}
	}
	return -1
}

// filterHeaders 筛选符合层级范围的标题（全局模式）
func (a *Analyzer) filterHeaders(headers []*Header) []*Header {
	var filtered []*Header
	for _, h := range headers {
		if h.Level >= a.options.MinLevel && h.Level <= a.options.MaxLevel {
			filtered = append(filtered, h)
		}
	}
	return filtered
}

// filterSectionHeaders 筛选符合层级范围的章节子标题
func (a *Analyzer) filterSectionHeaders(headers []*Header) []*Header {
	var filtered []*Header
	for _, h := range headers {
		if h.Level >= a.options.MinLevel && h.Level <= a.options.MaxLevel {
			filtered = append(filtered, h)
		}
	}
	return filtered
}
