package mdtoc

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// 预编译的正则表达式（程序启动时只编译一次）
var (
	htmlTagRe     = regexp.MustCompile(`<[^>]*>`)
	hyphensRe     = regexp.MustCompile(`-+`)
	boldDoubleRe  = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicStarRe  = regexp.MustCompile(`\*(.+?)\*`)
	boldUnderRe   = regexp.MustCompile(`__(.+?)__`)
	italicUnderRe = regexp.MustCompile(`(?:^|[\s])_([^_]+?)_(?:[\s]|$)`)
	strikeRe      = regexp.MustCompile(`~~(.+?)~~`)
	codeRe        = regexp.MustCompile("`(.+?)`")
	linkRe        = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	imgRe         = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
)

// AnchorGenerator 生成 GitHub 风格的 anchor link
type AnchorGenerator struct {
	counter map[string]int // 重复标题计数器
}

// NewAnchorGenerator 创建新的锚点生成器
func NewAnchorGenerator() *AnchorGenerator {
	return &AnchorGenerator{
		counter: make(map[string]int),
	}
}

// Reset 重置计数器
func (g *AnchorGenerator) Reset() {
	g.counter = make(map[string]int)
}

// Generate 生成 anchor link
// 规则 (参考 GitHub):
// 1. 转小写
// 2. 移除 HTML 标签
// 3. 移除强调符号 (*, _, ~)
// 4. 保留 Unicode 字母、数字、连字符、空格
// 5. 空格转连字符
// 6. 处理重复标题 (添加 -1, -2, ...)
func (g *AnchorGenerator) Generate(text string) string {
	// 1. 转小写
	anchor := strings.ToLower(text)

	// 2. 移除 HTML 标签
	anchor = removeHTMLTags(anchor)

	// 3. 移除强调符号和代码标记
	anchor = removeEmphasis(anchor)

	// 4. 保留 Unicode 字母、数字、连字符、空格
	anchor = filterCharacters(anchor)

	// 5. 空格转连字符，合并多个连字符
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = mergeHyphens(anchor)
	anchor = strings.Trim(anchor, "-")

	// 6. 处理重复标题
	anchor = g.handleDuplicate(anchor)

	return anchor
}

// removeHTMLTags 移除 HTML 标签
func removeHTMLTags(s string) string {
	return htmlTagRe.ReplaceAllString(s, "")
}

// removeEmphasis 移除 Markdown 强调符号
func removeEmphasis(s string) string {
	result := s

	// 按顺序应用各个正则替换
	result = boldDoubleRe.ReplaceAllString(result, "$1")
	result = italicStarRe.ReplaceAllString(result, "$1")
	result = boldUnderRe.ReplaceAllString(result, "$1")
	result = italicUnderRe.ReplaceAllString(result, "$1")
	result = strikeRe.ReplaceAllString(result, "$1")
	result = codeRe.ReplaceAllString(result, "$1")

	// 移除链接语法 [text](url) -> text
	result = linkRe.ReplaceAllString(result, "$1")

	// 移除图片语法 ![alt](url) -> alt
	result = imgRe.ReplaceAllString(result, "$1")

	return result
}

// filterCharacters 保留 Unicode 字母、数字、连字符、下划线、空格
func filterCharacters(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// mergeHyphens 合并多个连续的连字符为一个
func mergeHyphens(s string) string {
	return hyphensRe.ReplaceAllString(s, "-")
}

// handleDuplicate 处理重复标题
func (g *AnchorGenerator) handleDuplicate(anchor string) string {
	count, exists := g.counter[anchor]
	g.counter[anchor] = count + 1

	if exists {
		return anchor + "-" + strconv.Itoa(count)
	}
	return anchor
}
