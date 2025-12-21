package mdtoc

import (
	"os"
	"strings"
)

// TOC 是主要的门面结构，封装所有 TOC 生成功能
type TOC struct {
	parser    *Parser
	generator *Generator
	marker    *MarkerHandler
	options   Options
}

// New 创建新的 TOC 实例
func New(opts Options) *TOC {
	return &TOC{
		parser:    NewParser(opts),
		generator: NewGenerator(opts),
		marker:    NewMarkerHandler(DefaultMarker),
		options:   opts,
	}
}

// GenerateFromFile 从文件生成 TOC 字符串
func (t *TOC) GenerateFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return "", err
	}
	return t.GenerateFromContent(content)
}

// GenerateFromContent 从内容生成 TOC 字符串
func (t *TOC) GenerateFromContent(content []byte) (string, error) {
	headers, err := t.parser.Parse(content)
	if err != nil {
		return "", err
	}
	return t.generator.Generate(headers), nil
}

// GenerateSectionTOCs 生成章节模式的 TOC (每个 H1 有独立的子目录)
func (t *TOC) GenerateSectionTOCs(content []byte) ([]SectionTOC, error) {
	// 解析所有标题
	headers, err := t.parser.ParseAllHeaders(content)
	if err != nil {
		return nil, err
	}

	// 按 H1 分割成章节
	sections := SplitSections(headers)

	// 为每个章节生成 TOC
	var sectionTOCs []SectionTOC
	for _, section := range sections {
		toc := t.generator.GenerateSection(section)
		if toc != "" {
			sectionTOCs = append(sectionTOCs, SectionTOC{
				H1Line: section.Title.Line - 1, // 转换为 0-based
				TOC:    toc,
			})
		}
	}

	return sectionTOCs, nil
}

// GenerateSectionTOCsWithOffset 生成章节模式的 TOC，预计算偏移量使行号正确
// 在干净内容（已移除旧 TOC）上解析标题，计算 TOC 插入后的正确行号
func (t *TOC) GenerateSectionTOCsWithOffset(cleanContent []byte) ([]SectionTOC, error) {
	// 在干净内容上解析所有标题（基准行号）
	headers, err := t.parser.ParseAllHeaders(cleanContent)
	if err != nil {
		return nil, err
	}

	// 按 H1 分割成章节
	sections := SplitSections(headers)

	// 第一遍：临时禁用行号，计算每个 TOC 块的行数
	origLineNumber := t.options.LineNumber
	t.options.LineNumber = false
	tempGenerator := NewGenerator(t.options)

	type sectionInfo struct {
		section      *Section
		tocLines     int // TOC 块总行数
		originalLine int // H1 在干净内容中的原始行号 (0-based)
	}
	var infos []sectionInfo

	for _, section := range sections {
		toc := tempGenerator.GenerateSection(section)
		if toc != "" {
			tocBlockLines := CalcTOCBlockLines(toc)
			infos = append(infos, sectionInfo{
				section:      section,
				tocLines:     tocBlockLines,
				originalLine: section.Title.Line - 1, // 转换为 0-based
			})
		}
	}

	// 恢复行号设置
	t.options.LineNumber = origLineNumber

	// 第二遍：计算累积偏移量并生成带正确行号的 TOC
	var sectionTOCs []SectionTOC
	cumulativeOffset := 0

	for _, info := range infos {
		// H1 标题使用累积偏移量（之前所有 TOC 块的影响）
		// 子标题使用累积偏移量 + 当前 TOC 块的行数（因为 TOC 插入在 H1 后、子标题前）
		subHeaderOffset := cumulativeOffset + info.tocLines

		adjustedSection := &Section{
			Title:      adjustHeader(info.section.Title, cumulativeOffset),
			SubHeaders: make([]*Header, len(info.section.SubHeaders)),
		}
		for i, h := range info.section.SubHeaders {
			adjustedSection.SubHeaders[i] = adjustHeader(h, subHeaderOffset)
		}

		// 用调整后的行号生成 TOC
		toc := t.generator.GenerateSection(adjustedSection)
		if toc != "" {
			sectionTOCs = append(sectionTOCs, SectionTOC{
				H1Line: info.originalLine, // 使用原始行号定位插入位置
				TOC:    toc,
			})
		}

		// 累加偏移量（当前 TOC 块会增加的行数）
		cumulativeOffset += info.tocLines
	}

	return sectionTOCs, nil
}

// adjustHeader 创建调整行号后的 Header 副本
func adjustHeader(h *Header, offset int) *Header {
	return &Header{
		Level:      h.Level,
		Text:       h.Text,
		AnchorLink: h.AnchorLink,
		Line:       h.Line + offset,
		EndLine:    h.EndLine + offset,
	}
}

// GenerateSectionTOCsPreview 生成章节模式的 TOC 预览 (用于 stdout 输出)
func (t *TOC) GenerateSectionTOCsPreview(content []byte) (string, error) {
	// 解析所有标题
	headers, err := t.parser.ParseAllHeaders(content)
	if err != nil {
		return "", err
	}

	// 按 H1 分割成章节
	sections := SplitSections(headers)

	var sb strings.Builder
	for i, section := range sections {
		toc := t.generator.GenerateSection(section)
		if toc != "" {
			sb.WriteString("### ")
			sb.WriteString(section.Title.Text)
			sb.WriteString("\n\n")
			sb.WriteString(toc)
			if i < len(sections)-1 {
				sb.WriteString("\n\n")
			}
		}
	}

	return sb.String(), nil
}

// UpdateFile 原地更新文件中的 TOC
// 如果文件没有 TOC 标记，会自动在第一个标题后插入
//
//nolint:nestif // Complex nesting is clearer here for section/global mode branching
func (t *TOC) UpdateFile(filename string) error {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return err
	}

	var newContent []byte

	if t.options.SectionTOC {
		// 章节模式：在每个 H1 后插入独立的子目录
		// 先清理现有 TOC 块，获取干净内容
		cleanContent, _ := t.marker.CleanTOCBlocks(content)

		// 使用预计算偏移量的方法生成 TOC
		sectionTOCs, err := t.GenerateSectionTOCsWithOffset(cleanContent)
		if err != nil {
			return err
		}

		// 在干净内容上插入新的 TOC
		newContent = t.marker.InsertSectionTOCs(cleanContent, sectionTOCs)
	} else {
		// 普通模式：在 <!--TOC--> 标记处插入完整 TOC
		toc, err := t.GenerateFromContent(content)
		if err != nil {
			return err
		}

		markers := t.marker.FindMarkers(content)
		if markers.Found {
			newContent = t.marker.InsertTOC(content, toc)
		} else {
			newContent = t.marker.InsertTOCAfterFirstHeading(content, toc)
		}
	}

	return os.WriteFile(filename, newContent, 0600)
}

// HasMarker 检查文件是否包含 TOC 标记
func (t *TOC) HasMarker(filename string) (bool, error) {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return false, err
	}
	markers := t.marker.FindMarkers(content)
	return markers.Found, nil
}

// DeleteTOC 删除文件中的所有 TOC 块
// 返回是否有内容被删除
func (t *TOC) DeleteTOC(filename string) (bool, error) {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return false, err
	}

	cleanContent, blockInfos := t.marker.CleanTOCBlocks(content)

	// 没有 TOC 块被删除
	if len(blockInfos) == 0 {
		return false, nil
	}

	// 写入清理后的内容
	if err := os.WriteFile(filename, cleanContent, 0600); err != nil {
		return false, err
	}

	return true, nil
}
