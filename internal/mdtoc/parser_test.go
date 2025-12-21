package mdtoc

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		opts     Options
		expected []*Header
	}{
		{
			name: "basic headers",
			content: `# Title
## Section 1
### Subsection 1.1
## Section 2`,
			opts: DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 3, Text: "Subsection 1.1", AnchorLink: "subsection-11"},
				{Level: 2, Text: "Section 2", AnchorLink: "section-2"},
			},
		},
		{
			name: "with min level filter",
			content: `# Title
## Section 1
### Subsection`,
			opts: Options{MinLevel: 2, MaxLevel: 3},
			expected: []*Header{
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
				{Level: 3, Text: "Subsection", AnchorLink: "subsection"},
			},
		},
		{
			name: "with max level filter",
			content: `# Title
## Section 1
### Subsection
#### Deep`,
			opts: Options{MinLevel: 1, MaxLevel: 2},
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section 1", AnchorLink: "section-1"},
			},
		},
		{
			name:    "headers in code block ignored",
			content: "# Real Header\n```\n# Not a header\n```\n## Another Header",
			opts:    DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Real Header", AnchorLink: "real-header"},
				{Level: 2, Text: "Another Header", AnchorLink: "another-header"},
			},
		},
		{
			name: "duplicate headers",
			content: `# Title
## Section
## Section
## Section`,
			opts: DefaultOptions(),
			expected: []*Header{
				{Level: 1, Text: "Title", AnchorLink: "title"},
				{Level: 2, Text: "Section", AnchorLink: "section"},
				{Level: 2, Text: "Section", AnchorLink: "section-1"},
				{Level: 2, Text: "Section", AnchorLink: "section-2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.opts)
			got, err := p.Parse([]byte(tt.content))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("Parse() returned %d headers, want %d", len(got), len(tt.expected))
			}

			for i, h := range got {
				exp := tt.expected[i]
				if h.Level != exp.Level {
					t.Errorf("Header[%d].Level = %d, want %d", i, h.Level, exp.Level)
				}
				if h.Text != exp.Text {
					t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.Text)
				}
				if h.AnchorLink != exp.AnchorLink {
					t.Errorf("Header[%d].AnchorLink = %q, want %q", i, h.AnchorLink, exp.AnchorLink)
				}
			}
		})
	}
}

func TestParser_EmptyContent(t *testing.T) {
	p := NewParser(DefaultOptions())
	got, err := p.Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Parse() returned %d headers for empty content, want 0", len(got))
	}
}

func TestParser_NoHeaders(t *testing.T) {
	p := NewParser(DefaultOptions())
	content := `This is a paragraph.

Another paragraph with some **bold** text.

And a list:
- Item 1
- Item 2
`
	got, err := p.Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Parse() returned %d headers for content without headers, want 0", len(got))
	}
}

func TestParser_LineNumbers(t *testing.T) {
	content := `# Title

## Section 1
Content...

## Section 2
More content...
`
	p := NewParser(DefaultOptions())
	got, err := p.Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []struct {
		text    string
		line    int
		endLine int
	}{
		{"Title", 1, 7},     // H1 包含所有 H2 子内容
		{"Section 1", 3, 5}, // H2 到下一个 H2 前
		{"Section 2", 6, 7}, // H2 到文件末尾
	}

	if len(got) != len(expected) {
		t.Fatalf("Parse() returned %d headers, want %d", len(got), len(expected))
	}

	for i, h := range got {
		exp := expected[i]
		if h.Text != exp.text {
			t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.text)
		}
		if h.Line != exp.line {
			t.Errorf("Header[%d].Line = %d, want %d", i, h.Line, exp.line)
		}
		if h.EndLine != exp.endLine {
			t.Errorf("Header[%d].EndLine = %d, want %d", i, h.EndLine, exp.endLine)
		}
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		content  string
		expected int
	}{
		{"", 0},
		{"line1", 1},
		{"line1\n", 1},
		{"line1\nline2", 2},
		{"line1\nline2\n", 2},
		{"line1\nline2\nline3", 3},
	}

	for _, tt := range tests {
		got := countLines([]byte(tt.content))
		if got != tt.expected {
			t.Errorf("countLines(%q) = %d, want %d", tt.content, got, tt.expected)
		}
	}
}

func TestSplitSections(t *testing.T) {
	tests := []struct {
		name     string
		headers  []*Header
		expected []struct {
			h1Text   string
			subCount int
			subTexts []string
		}
	}{
		{
			name: "multiple H1 with sub-headers",
			headers: []*Header{
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
				{Level: 2, Text: "Section 1.2"},
				{Level: 1, Text: "Chapter 2"},
				{Level: 2, Text: "Section 2.1"},
				{Level: 3, Text: "Subsection 2.1.1"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 2, []string{"Section 1.1", "Section 1.2"}},
				{"Chapter 2", 2, []string{"Section 2.1", "Subsection 2.1.1"}},
			},
		},
		{
			name: "H1 without sub-headers",
			headers: []*Header{
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
				{Level: 1, Text: "Chapter 2"},
				{Level: 1, Text: "Chapter 3"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 1, []string{"Section 1.1"}},
				{"Chapter 2", 0, nil},
				{"Chapter 3", 0, nil},
			},
		},
		{
			name: "headers before first H1 ignored",
			headers: []*Header{
				{Level: 2, Text: "Orphan Section"},
				{Level: 3, Text: "Orphan Subsection"},
				{Level: 1, Text: "Chapter 1"},
				{Level: 2, Text: "Section 1.1"},
			},
			expected: []struct {
				h1Text   string
				subCount int
				subTexts []string
			}{
				{"Chapter 1", 1, []string{"Section 1.1"}},
			},
		},
		{
			name: "no H1 headers",
			headers: []*Header{
				{Level: 2, Text: "Section 1"},
				{Level: 2, Text: "Section 2"},
			},
			expected: nil,
		},
		{
			name:     "empty headers",
			headers:  []*Header{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := SplitSections(tt.headers)

			if tt.expected == nil {
				if len(sections) != 0 {
					t.Errorf("SplitSections() returned %d sections, want 0", len(sections))
				}
				return
			}

			if len(sections) != len(tt.expected) {
				t.Fatalf("SplitSections() returned %d sections, want %d", len(sections), len(tt.expected))
			}

			for i, s := range sections {
				exp := tt.expected[i]
				if s.Title.Text != exp.h1Text {
					t.Errorf("Section[%d].Title.Text = %q, want %q", i, s.Title.Text, exp.h1Text)
				}
				if len(s.SubHeaders) != exp.subCount {
					t.Errorf("Section[%d] has %d sub-headers, want %d", i, len(s.SubHeaders), exp.subCount)
				}
				for j, sub := range s.SubHeaders {
					if j < len(exp.subTexts) && sub.Text != exp.subTexts[j] {
						t.Errorf("Section[%d].SubHeaders[%d].Text = %q, want %q", i, j, sub.Text, exp.subTexts[j])
					}
				}
			}
		})
	}
}

func TestParser_ParseAllHeaders(t *testing.T) {
	content := `# Title
## Section 1
### Subsection 1.1
#### Deep 1
## Section 2
`
	// ParseAllHeaders should return ALL headers regardless of MinLevel/MaxLevel
	p := NewParser(Options{MinLevel: 2, MaxLevel: 2})
	got, err := p.ParseAllHeaders([]byte(content))
	if err != nil {
		t.Fatalf("ParseAllHeaders() error = %v", err)
	}

	// Should return all 5 headers, not filtered by level
	if len(got) != 5 {
		t.Errorf("ParseAllHeaders() returned %d headers, want 5", len(got))
	}

	expectedLevels := []int{1, 2, 3, 4, 2}
	for i, h := range got {
		if h.Level != expectedLevels[i] {
			t.Errorf("Header[%d].Level = %d, want %d", i, h.Level, expectedLevels[i])
		}
	}
}

// ==================== YAML Frontmatter Line Number Tests ====================

func TestParser_LineNumbers_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []struct {
			text    string
			line    int
			endLine int
		}
	}{
		{
			name: "basic frontmatter - line numbers should include frontmatter offset",
			content: `---
title: Test Document
author: Test
---
# Title

## Section 1
Content...

## Section 2
More content...
`,
			expected: []struct {
				text    string
				line    int
				endLine int
			}{
				{"Title", 5, 11},      // H1 at line 5 (after 4 lines of frontmatter)
				{"Section 1", 7, 9},   // H2 at line 7
				{"Section 2", 10, 11}, // H2 at line 10
			},
		},
		{
			name: "frontmatter with YAML comment - should not parse as header",
			content: `---
# This is a YAML comment, not a Markdown header
title: Test
layout: page
---
# Real Title

## Section 1
Content here
`,
			expected: []struct {
				text    string
				line    int
				endLine int
			}{
				{"Real Title", 6, 9}, // H1 at line 6
				{"Section 1", 8, 9},  // H2 at line 8
			},
		},
		{
			name: "VitePress style frontmatter",
			content: `---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
hero:
  name: My Project
---
# Getting Started

## Installation
Install the package...

## Usage
Use the package...
`,
			expected: []struct {
				text    string
				line    int
				endLine int
			}{
				{"Getting Started", 7, 13}, // H1 at line 7
				{"Installation", 9, 11},    // H2 at line 9
				{"Usage", 12, 13},          // H2 at line 12
			},
		},
		{
			name: "Hugo style frontmatter with dots closer",
			content: `---
title: Hugo Page
date: 2024-01-01
...
# Page Title

## First Section
Content

## Second Section
More content
`,
			expected: []struct {
				text    string
				line    int
				endLine int
			}{
				{"Page Title", 5, 11},      // H1 at line 5
				{"First Section", 7, 9},    // H2 at line 7
				{"Second Section", 10, 11}, // H2 at line 10
			},
		},
		{
			name: "no frontmatter - baseline",
			content: `# Title

## Section 1
Content...

## Section 2
More content...
`,
			expected: []struct {
				text    string
				line    int
				endLine int
			}{
				{"Title", 1, 7},     // H1 at line 1
				{"Section 1", 3, 5}, // H2 at line 3
				{"Section 2", 6, 7}, // H2 at line 6
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(DefaultOptions())
			got, err := p.Parse([]byte(tt.content))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("Parse() returned %d headers, want %d\nHeaders: %+v",
					len(got), len(tt.expected), got)
			}

			for i, h := range got {
				exp := tt.expected[i]
				if h.Text != exp.text {
					t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.text)
				}
				if h.Line != exp.line {
					t.Errorf("Header[%d].Line = %d, want %d (text: %q)",
						i, h.Line, exp.line, h.Text)
				}
				if h.EndLine != exp.endLine {
					t.Errorf("Header[%d].EndLine = %d, want %d (text: %q)",
						i, h.EndLine, exp.endLine, h.Text)
				}
			}
		})
	}
}

func TestParser_LineNumbers_FrontmatterWithTOCMarker(t *testing.T) {
	// Test that <!--TOC--> markers inside frontmatter don't affect parsing
	content := `---
title: Test
# Note: <!--TOC--> in frontmatter should be ignored
description: A test file
---
# Main Title

<!--TOC-->

- [Section 1](#section-1)

<!--TOC-->

## Section 1
Content here

## Section 2
More content
`
	p := NewParser(DefaultOptions())
	got, err := p.Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []struct {
		text    string
		line    int
		endLine int
	}{
		{"Main Title", 6, 18}, // H1 at line 6
		{"Section 1", 14, 16}, // H2 at line 14
		{"Section 2", 17, 18}, // H2 at line 17
	}

	if len(got) != len(expected) {
		t.Fatalf("Parse() returned %d headers, want %d", len(got), len(expected))
	}

	for i, h := range got {
		exp := expected[i]
		if h.Text != exp.text {
			t.Errorf("Header[%d].Text = %q, want %q", i, h.Text, exp.text)
		}
		if h.Line != exp.line {
			t.Errorf("Header[%d].Line = %d, want %d (text: %q)",
				i, h.Line, exp.line, h.Text)
		}
		if h.EndLine != exp.endLine {
			t.Errorf("Header[%d].EndLine = %d, want %d (text: %q)",
				i, h.EndLine, exp.endLine, h.Text)
		}
	}
}

func TestParser_CountLines_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "with frontmatter",
			content: `---
title: Test
---
# Title
Content`,
			expected: 5,
		},
		{
			name: "with frontmatter and trailing newline",
			content: `---
title: Test
---
# Title
Content
`,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLines([]byte(tt.content))
			if got != tt.expected {
				t.Errorf("countLines() = %d, want %d", got, tt.expected)
			}
		})
	}
}
