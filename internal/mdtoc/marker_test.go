package mdtoc

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarkerHandler_FindMarkers(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{
			name:      "no markers",
			content:   "# Title\nSome content",
			wantStart: -1,
			wantEnd:   -1,
			wantFound: false,
		},
		{
			name:      "one marker",
			content:   "# Title\n<!--TOC-->\nSome content",
			wantStart: 1,
			wantEnd:   -1,
			wantFound: true,
		},
		{
			name:      "two markers",
			content:   "# Title\n<!--TOC-->\nTOC content\n<!--TOC-->\nRest",
			wantStart: 1,
			wantEnd:   3,
			wantFound: true,
		},
		{
			name:      "marker with whitespace",
			content:   "# Title\n  <!--TOC-->  \nSome content",
			wantStart: 1,
			wantEnd:   -1,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindMarkers([]byte(tt.content))

			if got.StartLine != tt.wantStart {
				t.Errorf("StartLine = %d, want %d", got.StartLine, tt.wantStart)
			}
			if got.EndLine != tt.wantEnd {
				t.Errorf("EndLine = %d, want %d", got.EndLine, tt.wantEnd)
			}
			if got.Found != tt.wantFound {
				t.Errorf("Found = %v, want %v", got.Found, tt.wantFound)
			}
		})
	}
}

func TestMarkerHandler_InsertTOC(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		expected string
	}{
		{
			name:     "no marker - unchanged",
			content:  "# Title\nContent",
			toc:      "- [Title](#title)",
			expected: "# Title\nContent",
		},
		{
			name:    "one marker - insert",
			content: "# Title\n<!--TOC-->\nContent",
			toc:     "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
		{
			name: "two markers - replace",
			content: `# Title
<!--TOC-->
Old TOC
<!--TOC-->
Content`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOC([]byte(tt.content), tt.toc))

			if got != tt.expected {
				t.Errorf("InsertTOC() =\n%s\n\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_ExtractExistingTOC(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "no markers",
			content:  "# Title\nContent",
			expected: "",
		},
		{
			name:     "one marker",
			content:  "# Title\n<!--TOC-->\nContent",
			expected: "",
		},
		{
			name: "two markers with content",
			content: `# Title
<!--TOC-->

- [Section](#section)

<!--TOC-->
Content`,
			expected: "- [Section](#section)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.ExtractExistingTOC([]byte(tt.content))

			if got != tt.expected {
				t.Errorf("ExtractExistingTOC() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindH1Lines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name:     "no H1",
			content:  "## Section\n### Subsection",
			expected: nil,
		},
		{
			name:     "single H1",
			content:  "# Title\n## Section",
			expected: []int{0},
		},
		{
			name: "multiple H1",
			content: `# Chapter 1
## Section 1.1
# Chapter 2
## Section 2.1
# Chapter 3`,
			expected: []int{0, 2, 4},
		},
		{
			name:     "H1 in code block ignored",
			content:  "# Real H1\n```\n# Not H1\n```\n# Another H1",
			expected: []int{0, 4},
		},
		{
			name:     "H2 not matched",
			content:  "## Not H1\n### Also not H1",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindH1Lines([]byte(tt.content))

			if len(got) != len(tt.expected) {
				t.Fatalf("FindH1Lines() returned %d lines, want %d", len(got), len(tt.expected))
			}
			for i, line := range got {
				if line != tt.expected[i] {
					t.Errorf("FindH1Lines()[%d] = %d, want %d", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestMarkerHandler_InsertSectionTOCs(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionTOCs []SectionTOC
		expected    string
	}{
		{
			name:        "empty section TOCs",
			content:     "# Title\nContent",
			sectionTOCs: []SectionTOC{},
			expected:    "# Title\nContent",
		},
		{
			name:    "single section TOC",
			content: "# Chapter 1\n\nContent...\n\n## Section 1.1\n\nMore content",
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

Content...

## Section 1.1

More content`,
		},
		{
			name: "multiple section TOCs",
			content: `# Chapter 1

## Section 1.1

# Chapter 2

## Section 2.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				{H1Line: 4, TOC: "- [Section 2.1](#section-21)"},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

# Chapter 2

<!--TOC-->

- [Section 2.1](#section-21)

<!--TOC-->

## Section 2.1`,
		},
		{
			name: "section without sub-headers skipped",
			content: `# Chapter 1

## Section 1.1

# Chapter 2

No sub-headers here`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				// Chapter 2 has empty TOC, should be skipped
				{H1Line: 4, TOC: ""},
			},
			expected: `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

# Chapter 2

No sub-headers here`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertSectionTOCs([]byte(tt.content), tt.sectionTOCs))

			if got != tt.expected {
				t.Errorf("InsertSectionTOCs() =\n%s\n\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindH1BlockEnd(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		h1Line   int
		expected int
	}{
		{
			name:     "H1 followed by empty line",
			content:  "# Title\n\nContent",
			h1Line:   0,
			expected: 0, // 空行在第1行，所以 H1 块结束于第0行
		},
		{
			name:     "H1 with badge on next line",
			content:  "# Title\n[![Badge](url)](link)\n\nContent",
			h1Line:   0,
			expected: 1, // 徽标在第1行，空行在第2行，H1 块结束于第1行
		},
		{
			name:     "H1 with description",
			content:  "# Title\nShort description here.\n\n## Section",
			h1Line:   0,
			expected: 1, // 描述在第1行，空行在第2行，H1 块结束于第1行
		},
		{
			name:     "H1 with badge and description",
			content:  "# Title\n[![Badge](url)](link)\nShort description.\n\n## Section",
			h1Line:   0,
			expected: 2, // 徽标第1行、描述第2行、空行第3行，H1 块结束于第2行
		},
		{
			name:     "H1 followed directly by H2 (no empty line)",
			content:  "# Title\n## Section",
			h1Line:   0,
			expected: 0, // H2 在第1行，H1 块结束于第0行（H1 本身）
		},
		{
			name:     "H1 at end of file with content",
			content:  "# Title\nSome content",
			h1Line:   0,
			expected: 1, // 文件末尾没有空行，返回最后一行
		},
		{
			name:     "H1 at end of file alone",
			content:  "# Title",
			h1Line:   0,
			expected: 0, // 只有 H1，返回 H1 行
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			lines := bytes.Split([]byte(tt.content), []byte("\n"))
			got := h.FindH1BlockEnd(lines, tt.h1Line)

			if got != tt.expected {
				t.Errorf("FindH1BlockEnd() = %d, want %d\nContent:\n%s", got, tt.expected, tt.content)
			}
		})
	}
}

func TestMarkerHandler_InsertSectionTOCs_WithBadgeAndDescription(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionTOCs []SectionTOC
		expected    string
	}{
		{
			name: "H1 with badge - TOC after badge",
			content: `# My Project
[![Build](https://img.shields.io/badge/build-passing-green)](https://ci.example.com)

## Installation

Content here`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Installation](#installation)"},
			},
			expected: `# My Project
[![Build](https://img.shields.io/badge/build-passing-green)](https://ci.example.com)

<!--TOC-->

- [Installation](#installation)

<!--TOC-->

## Installation

Content here`,
		},
		{
			name: "H1 with description - TOC after description",
			content: `# My Project
A brief description of this project.

## Getting Started

Content here`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Getting Started](#getting-started)"},
			},
			expected: `# My Project
A brief description of this project.

<!--TOC-->

- [Getting Started](#getting-started)

<!--TOC-->

## Getting Started

Content here`,
		},
		{
			name: "H1 with badge and description - TOC after both",
			content: `# My Project
[![Badge](url)](link)
Short description here.

## Features

Content`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Features](#features)"},
			},
			expected: `# My Project
[![Badge](url)](link)
Short description here.

<!--TOC-->

- [Features](#features)

<!--TOC-->

## Features

Content`,
		},
		{
			name: "multiple H1s with different block sizes",
			content: `# Chapter 1
Badge and description.

## Section 1.1

# Chapter 2

## Section 2.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				{H1Line: 5, TOC: "- [Section 2.1](#section-21)"},
			},
			expected: `# Chapter 1
Badge and description.

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

# Chapter 2

<!--TOC-->

- [Section 2.1](#section-21)

<!--TOC-->

## Section 2.1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertSectionTOCs([]byte(tt.content), tt.sectionTOCs))

			if got != tt.expected {
				t.Errorf("InsertSectionTOCs() =\n%s\n\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindFirstHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "no heading",
			content:  "Just some text\nNo headers here",
			expected: -1,
		},
		{
			name:     "H1 first",
			content:  "# Title\nContent",
			expected: 0,
		},
		{
			name:     "H2 first",
			content:  "Some text\n## Section\nContent",
			expected: 1,
		},
		{
			name:     "heading in code block ignored",
			content:  "```\n# Not heading\n```\n## Real heading",
			expected: 3,
		},
		{
			name:     "fenced code with tilde",
			content:  "~~~\n# Not heading\n~~~\n## Real heading",
			expected: 3,
		},
		{
			name:     "H3 to H6 detection",
			content:  "Some text\n### H3\nContent",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindFirstHeading([]byte(tt.content))

			if got != tt.expected {
				t.Errorf("FindFirstHeading() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_InsertTOCAfterFirstHeading(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		contains []string
	}{
		{
			name:    "insert after H1",
			content: "# Title\n\nContent here",
			toc:     "- [Section](#section)",
			contains: []string{
				"# Title",
				"<!--TOC-->",
				"- [Section](#section)",
				"Content here",
			},
		},
		{
			name:    "insert at beginning when no heading",
			content: "Just some text\nNo headers",
			toc:     "- [Item](#item)",
			contains: []string{
				"<!--TOC-->",
				"- [Item](#item)",
				"Just some text",
			},
		},
		{
			name:    "insert after H2 when no H1",
			content: "Intro\n## Section\nContent",
			toc:     "- [Link](#link)",
			contains: []string{
				"## Section",
				"<!--TOC-->",
				"- [Link](#link)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOCAfterFirstHeading([]byte(tt.content), tt.toc))

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("InsertTOCAfterFirstHeading() should contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

func TestMarkerHandler_UpdateSectionTOCs(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		sectionTOCs []SectionTOC
		contains    []string
		excludes    []string
	}{
		{
			name: "update existing section TOCs",
			content: `# Chapter 1

<!--TOC-->

- [Old Link](#old-link)

<!--TOC-->

## Section 1.1

# Chapter 2

<!--TOC-->

- [Old Link 2](#old-link-2)

<!--TOC-->

## Section 2.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
				{H1Line: 6, TOC: "- [Section 2.1](#section-21)"},
			},
			contains: []string{
				"[Section 1.1](#section-11)",
				"[Section 2.1](#section-21)",
			},
			excludes: []string{
				"[Old Link]",
				"[Old Link 2]",
			},
		},
		{
			name:    "no existing markers - insert new",
			content: "# Chapter 1\n\n## Section 1.1",
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			contains: []string{
				"<!--TOC-->",
				"[Section 1.1](#section-11)",
			},
		},
		{
			name: "single unpaired marker - insert mode",
			content: `# Chapter 1

<!--TOC-->

## Section 1.1`,
			sectionTOCs: []SectionTOC{
				{H1Line: 0, TOC: "- [Section 1.1](#section-11)"},
			},
			contains: []string{
				"[Section 1.1](#section-11)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.UpdateSectionTOCs([]byte(tt.content), tt.sectionTOCs))

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("UpdateSectionTOCs() should contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("UpdateSectionTOCs() should NOT contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

func TestMarkerHandler_CustomMarker(t *testing.T) {
	customMarker := "<!-- TABLE OF CONTENTS -->"
	h := NewMarkerHandler(customMarker)

	content := "# Title\n<!-- TABLE OF CONTENTS -->\nContent"
	markers := h.FindMarkers([]byte(content))

	if !markers.Found {
		t.Error("Should find custom marker")
	}
	if markers.StartLine != 1 {
		t.Errorf("StartLine = %d, want 1", markers.StartLine)
	}
}

func TestMarkerHandler_EmptyMarker(t *testing.T) {
	// Empty marker should default to DefaultMarker
	h := NewMarkerHandler("")

	content := "# Title\n<!--TOC-->\nContent"
	markers := h.FindMarkers([]byte(content))

	if !markers.Found {
		t.Error("Should find default marker when empty string provided")
	}
}

// ==================== YAML Frontmatter Tests ====================

func TestFindFrontmatterEnd(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "no frontmatter",
			content:  "# Title\nContent",
			expected: -1,
		},
		{
			name:     "frontmatter with dash closer",
			content:  "---\ntitle: Test\n---\n# Title",
			expected: 2,
		},
		{
			name:     "frontmatter with dots closer",
			content:  "---\ntitle: Test\n...\n# Title",
			expected: 2,
		},
		{
			name:     "frontmatter with YAML comment",
			content:  "---\n# This is a YAML comment\ntitle: Test\n---\n# Real Title",
			expected: 3,
		},
		{
			name:     "unclosed frontmatter",
			content:  "---\ntitle: Test\n# Not closed",
			expected: -1,
		},
		{
			name:     "dash not on first line - not frontmatter",
			content:  "Some text\n---\ntitle: Test\n---",
			expected: -1,
		},
		{
			name:     "empty content",
			content:  "",
			expected: -1,
		},
		{
			name:     "VitePress style frontmatter",
			content:  "---\n# https://vitepress.dev/reference/default-theme-home-page\nlayout: home\n---\n# Real Title",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			lineBytes := make([][]byte, len(lines))
			for i, l := range lines {
				lineBytes[i] = []byte(l)
			}
			got := FindFrontmatterEnd(lineBytes)
			if got != tt.expected {
				t.Errorf("FindFrontmatterEnd() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindFirstHeading_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name: "frontmatter with YAML comment - should skip",
			content: `---
# This is a YAML comment, not a heading
title: Test
---
# Real H1 Title`,
			expected: 4,
		},
		{
			name: "VitePress frontmatter - should skip YAML comment",
			content: `---
# https://vitepress.dev/reference/default-theme-home-page
layout: home
---
# Real Title`,
			expected: 4,
		},
		{
			name: "frontmatter without YAML comment",
			content: `---
title: Test
layout: page
---
# Title After Frontmatter`,
			expected: 4,
		},
		{
			name:     "no frontmatter - normal behavior",
			content:  "# Title\nContent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindFirstHeading([]byte(tt.content))
			if got != tt.expected {
				t.Errorf("FindFirstHeading() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMarkerHandler_FindMarkers_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantStart int
		wantEnd   int
		wantFound bool
	}{
		{
			name: "marker inside frontmatter should be ignored",
			content: `---
# Comment
<!--TOC-->
title: Test
---
# Title
<!--TOC-->
Content`,
			wantStart: 6,
			wantEnd:   -1,
			wantFound: true,
		},
		{
			name: "markers after frontmatter",
			content: `---
title: Test
---
# Title
<!--TOC-->
TOC content
<!--TOC-->
Rest`,
			wantStart: 4,
			wantEnd:   6,
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindMarkers([]byte(tt.content))

			if got.StartLine != tt.wantStart {
				t.Errorf("StartLine = %d, want %d", got.StartLine, tt.wantStart)
			}
			if got.EndLine != tt.wantEnd {
				t.Errorf("EndLine = %d, want %d", got.EndLine, tt.wantEnd)
			}
			if got.Found != tt.wantFound {
				t.Errorf("Found = %v, want %v", got.Found, tt.wantFound)
			}
		})
	}
}

func TestMarkerHandler_FindH1Lines_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name: "YAML comment should not be detected as H1",
			content: `---
# YAML comment
title: Test
---
# Real H1
## Section`,
			expected: []int{4},
		},
		{
			name: "multiple H1 after frontmatter",
			content: `---
title: Test
---
# Chapter 1
## Section 1.1
# Chapter 2`,
			expected: []int{3, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindH1Lines([]byte(tt.content))

			if len(got) != len(tt.expected) {
				t.Fatalf("FindH1Lines() returned %d lines, want %d. Got: %v", len(got), len(tt.expected), got)
			}
			for i, line := range got {
				if line != tt.expected[i] {
					t.Errorf("FindH1Lines()[%d] = %d, want %d", i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestMarkerHandler_InsertTOCAfterFirstHeading_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		contains []string
		excludes []string
	}{
		{
			name: "should insert after frontmatter, not inside",
			content: `---
# YAML comment
title: Test
---
# Real Title
Content`,
			toc: "- [Section](#section)",
			contains: []string{
				"---\n# YAML comment\ntitle: Test\n---", // frontmatter should be intact
				"# Real Title",
				"<!--TOC-->",
				"- [Section](#section)",
			},
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOCAfterFirstHeading([]byte(tt.content), tt.toc))

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("InsertTOCAfterFirstHeading() should contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("InsertTOCAfterFirstHeading() should NOT contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// ==================== Enhanced Marker Handling Tests ====================

// TestFindAllMarkers 测试查找所有标记
func TestFindAllMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []int
	}{
		{
			name:     "no markers",
			content:  "# Title\nContent",
			expected: []int{},
		},
		{
			name:     "single marker",
			content:  "# Title\n<!--TOC-->\nContent",
			expected: []int{1},
		},
		{
			name:     "multiple markers",
			content:  "# Title\n<!--TOC-->\nContent\n<!--TOC-->\nMore\n<!--TOC-->\nEnd",
			expected: []int{1, 3, 5},
		},
		{
			name:     "markers with whitespace",
			content:  "# Title\n  <!--TOC-->  \n  <!--TOC-->  ",
			expected: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := h.FindAllMarkers([]byte(tt.content))

			if len(got) != len(tt.expected) {
				t.Errorf("FindAllMarkers() returned %d markers, want %d", len(got), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if i < len(got) && got[i] != expected {
					t.Errorf("FindAllMarkers()[%d] = %d, want %d", i, got[i], expected)
				}
			}
		})
	}
}

// TestInsertTOCWithCleanup 测试带清理的插入 TOC
func TestInsertTOCWithCleanup(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		toc      string
		expected string
	}{
		{
			name:     "no markers - unchanged",
			content:  "# Title\nContent",
			toc:      "- [Title](#title)",
			expected: "# Title\nContent",
		},
		{
			name: "single marker - creates pair",
			content: `# Title
<!--TOC-->
Content`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
		{
			name: "two markers - standard replacement",
			content: `# Title
<!--TOC-->
Old TOC
<!--TOC-->
Content`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content`,
		},
		{
			name: "three markers - removes extra",
			content: `# Title
<!--TOC-->
Old TOC 1
<!--TOC-->
Content
<!--TOC-->
More`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content
More`,
		},
		{
			name: "four markers - keeps first two, removes rest",
			content: `# Title
<!--TOC-->
TOC 1
<!--TOC-->
Content 1
<!--TOC-->
TOC 2
<!--TOC-->
Content 2`,
			toc: "- [Title](#title)",
			expected: `# Title
<!--TOC-->

- [Title](#title)

<!--TOC-->
Content 1
Content 2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			got := string(h.InsertTOCWithCleanup([]byte(tt.content), tt.toc))

			if got != tt.expected {
				t.Errorf("\n=== Test failed: %s ===\n\nGot:\n%s\n\nExpected:\n%s", tt.name, got, tt.expected)
			}

			// 验证结果没有孤儿标记
			valid, count, msg := h.ValidateMarkers([]byte(got))
			if !valid {
				t.Errorf("Result has invalid markers: %s (count: %d)", msg, count)
			}
		})
	}
}

// TestValidateMarkers 测试标记验证
func TestValidateMarkers(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectValid bool
		expectCount int
		expectError string
	}{
		{
			name:        "no markers",
			content:     "# Title\nContent",
			expectValid: true,
			expectCount: 0,
			expectError: "",
		},
		{
			name:        "paired markers",
			content:     "# Title\n<!--TOC-->\nTOC\n<!--TOC-->\nContent",
			expectValid: true,
			expectCount: 2,
			expectError: "",
		},
		{
			name:        "odd number of markers",
			content:     "# Title\n<!--TOC-->\nContent\n<!--TOC-->\nMore\n<!--TOC-->",
			expectValid: false,
			expectCount: 3,
			expectError: "文档中有奇数个 TOC 标记，这是无效的",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			valid, count, msg := h.ValidateMarkers([]byte(tt.content))

			if valid != tt.expectValid {
				t.Errorf("ValidateMarkers() validity = %v, want %v", valid, tt.expectValid)
			}
			if count != tt.expectCount {
				t.Errorf("ValidateMarkers() count = %d, want %d", count, tt.expectCount)
			}
			if msg != tt.expectError {
				t.Errorf("ValidateMarkers() error = %q, want %q", msg, tt.expectError)
			}
		})
	}
}

// TestCleanupOrphanMarkers 测试清理孤儿标记
func TestCleanupOrphanMarkers(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedResult  string
		expectedRemoved int
	}{
		{
			name:            "no markers",
			content:         "# Title\nContent",
			expectedResult:  "# Title\nContent",
			expectedRemoved: 0,
		},
		{
			name:            "even markers",
			content:         "# Title\n<!--TOC-->\nTOC\n<!--TOC-->\nContent",
			expectedResult:  "# Title\n<!--TOC-->\nTOC\n<!--TOC-->\nContent",
			expectedRemoved: 0,
		},
		{
			name:            "odd markers - removes last",
			content:         "# Title\n<!--TOC-->\nTOC\n<!--TOC-->\nContent\n<!--TOC-->",
			expectedResult:  "# Title\n<!--TOC-->\nTOC\n<!--TOC-->\nContent",
			expectedRemoved: 1,
		},
		{
			name:            `single marker - removes it`,
			content:         "# Title\nContent\n<!--TOC-->",
			expectedResult:  "# Title\nContent",
			expectedRemoved: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewMarkerHandler(DefaultMarker)
			result, removed := h.CleanupOrphanMarkers([]byte(tt.content))

			if string(result) != tt.expectedResult {
				t.Errorf("CleanupOrphanMarkers() result =\n%s\n\nwant:\n%s", string(result), tt.expectedResult)
			}
			if removed != tt.expectedRemoved {
				t.Errorf("CleanupOrphanMarkers() removed = %d, want %d", removed, tt.expectedRemoved)
			}
		})
	}
}

// TestRealWorldScenario 测试真实场景
func TestRealWorldScenario(t *testing.T) {
	// 模拟用户报告的问题：预先存在单个标记
	content := `# 测试文档

这是第一段内容。

<!--TOC-->

这是第二段内容。

## 章节1

章节1的内容。

## 章节2

章节2的内容。
`

	h := NewMarkerHandler(DefaultMarker)
	toc := "- [章节1](#章节1)\n- [章节2](#章节2)"

	// 使用新的清理方法
	result := string(h.InsertTOCWithCleanup([]byte(content), toc))

	// 验证结果
	if !strings.Contains(result, toc) {
		t.Error("TOC content not found in result")
	}

	// 验证标记有效性
	valid, count, msg := h.ValidateMarkers([]byte(result))
	if !valid {
		t.Errorf("Result has invalid markers: %s (count: %d)", msg, count)
	}

	// 验证格式
	expectedParts := []string{
		"# 测试文档",
		"<!--TOC-->",
		toc,
		"<!--TOC-->",
		"这是第一段内容",
		"这是第二段内容",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Result missing expected part: %s", part)
		}
	}

	// 确保没有多余的标记
	markerCount := strings.Count(result, "<!--TOC-->")
	if markerCount != 2 {
		t.Errorf("Expected 2 markers, got %d", markerCount)
	}
}
