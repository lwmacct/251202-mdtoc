package mdtoc

// Generator 生成 TOC 字符串
type Generator struct {
	options Options
}

// NewGenerator 创建新的生成器
func NewGenerator(opts Options) *Generator {
	return &Generator{
		options: opts,
	}
}

// Generate 从标题列表生成 TOC 字符串
func (g *Generator) Generate(headers []*Header) string {
	if len(headers) == 0 {
		return ""
	}
	return g.generateTOC(headers, g.options.MinLevel)
}

// GenerateSection 为单个章节生成 TOC (只包含子标题)
// 章节模式下，每个 H1 后面只生成该章节的子目录
// 要求：章节内至少包含一个 H2 才会生成 TOC
func (g *Generator) GenerateSection(section *Section) string {
	if section == nil || len(section.SubHeaders) == 0 {
		return ""
	}

	// 检查是否至少有一个 H2 (章节必须包含 H2 才生成 TOC)
	hasH2 := false
	for _, h := range section.SubHeaders {
		if h.Level == 2 {
			hasH2 = true
			break
		}
	}
	if !hasH2 {
		return ""
	}

	// 筛选符合层级范围的子标题
	var filteredHeaders []*Header
	for _, h := range section.SubHeaders {
		if h.Level >= g.options.MinLevel && h.Level <= g.options.MaxLevel {
			filteredHeaders = append(filteredHeaders, h)
		}
	}

	if len(filteredHeaders) == 0 {
		return ""
	}

	return g.generateTOC(filteredHeaders, minHeaderLevel(filteredHeaders))
}

// generateTOC 生成 TOC 字符串的内部实现
// baseLevel 用于计算缩进的基准层级
func (g *Generator) generateTOC(headers []*Header, baseLevel int) string {
	return renderTOCString(headers, g.options, baseLevel, func(h *Header) string {
		return formatHeaderLineRange(g.options, h)
	})
}
