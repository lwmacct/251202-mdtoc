package mdtoc

import (
	"os"
)

// TOC 是主要的门面结构，封装所有 TOC 生成功能
// 使用 4 步管道：Clean → Analyze → Build → Finalize
type TOC struct {
	cleaner   *Cleaner
	analyzer  *Analyzer
	builder   *Builder
	finalizer *Finalizer
	marker    *MarkerHandler // 保留用于兼容性
	options   Options
}

// New 创建新的 TOC 实例
func New(opts Options) *TOC {
	return &TOC{
		cleaner:   NewCleaner(DefaultMarker),
		analyzer:  NewAnalyzer(opts),
		builder:   NewBuilder(opts),
		finalizer: NewFinalizer(opts),
		marker:    NewMarkerHandler(DefaultMarker),
		options:   opts,
	}
}

// GenerateFromFile 从文件生成 TOC 字符串（预览用）
func (t *TOC) GenerateFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return "", err
	}
	return t.GenerateFromContent(content)
}

// GenerateFromContent 从内容生成 TOC 字符串（预览用）
// 预览模式始终使用全局模式，显示所有标题
func (t *TOC) GenerateFromContent(content []byte) (string, error) {
	// 预览模式临时禁用章节模式，显示所有标题
	previewOpts := t.options
	previewOpts.SectionTOC = false

	analyzer := NewAnalyzer(previewOpts)
	doc, err := analyzer.Analyze(content)
	if err != nil {
		return "", err
	}

	builder := NewBuilder(previewOpts)
	return builder.BuildPreview(doc), nil
}

// UpdateFile 原地更新文件中的 TOC
// 4 步管道：Clean → Analyze → Build → Finalize
func (t *TOC) UpdateFile(filename string) error {
	content, err := os.ReadFile(filename) //nolint:gosec // G304: file path from user input is intentional
	if err != nil {
		return err
	}

	// Step 1: Clean - 删除现有 TOC 块
	cleanContent := t.cleaner.Clean(content)

	// Step 2: Analyze - 分析文档结构，找到插入点
	doc, err := t.analyzer.Analyze(cleanContent)
	if err != nil {
		return err
	}

	// 如果没有插入点，保持内容不变
	if len(doc.InsertionPoints) == 0 {
		return os.WriteFile(filename, cleanContent, 0600)
	}

	// Step 3: Build - 插入带占位符的 TOC
	withTOC := t.builder.Build(cleanContent, doc)

	// Step 4: Finalize - 替换占位符为实际行号
	finalContent, err := t.finalizer.Finalize(withTOC)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, finalContent, 0600)
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

	// 使用新的 cleaner
	cleanContent := t.cleaner.Clean(content)

	// 检查是否有内容被删除
	if len(cleanContent) == len(content) {
		return false, nil
	}

	// 写入清理后的内容
	if err := os.WriteFile(filename, cleanContent, 0600); err != nil {
		return false, err
	}

	return true, nil
}

// GenerateSectionTOCsPreview 生成章节模式的 TOC 预览 (用于 stdout 输出)
func (t *TOC) GenerateSectionTOCsPreview(content []byte) (string, error) {
	// 临时启用章节模式
	originalSectionTOC := t.options.SectionTOC
	t.options.SectionTOC = true
	defer func() { t.options.SectionTOC = originalSectionTOC }()

	// 创建临时分析器
	analyzer := NewAnalyzer(t.options)
	doc, err := analyzer.Analyze(content)
	if err != nil {
		return "", err
	}

	// 生成预览
	builder := NewBuilder(t.options)
	return builder.BuildPreview(doc), nil
}
