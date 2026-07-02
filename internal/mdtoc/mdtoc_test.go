package mdtoc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lwmacct/251202-mdtoc/internal/mdtoc"
)

// TestTOC_GenerateFromContent 测试从内容生成 TOC 的核心功能
func TestTOC_GenerateFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		opts     mdtoc.Options
		contains []string // TOC 应包含的内容
		excludes []string // TOC 不应包含的内容
	}{
		{
			name: "basic markdown",
			content: `# 项目简介

## 功能特性

### 特性一

### 特性二

## 安装说明

## 使用方法`,
			opts: mdtoc.Options{MinLevel: 1, MaxLevel: 3, ShowAnchor: true},
			contains: []string{
				"[项目简介](#项目简介)",
				"[功能特性](#功能特性)",
				"[特性一](#特性一)",
				"[特性二](#特性二)",
				"[安装说明](#安装说明)",
				"[使用方法](#使用方法)",
			},
		},
		{
			name: "filter by level",
			content: `# Title
## Section 1
### Subsection
#### Deep
## Section 2`,
			opts:     mdtoc.Options{MinLevel: 2, MaxLevel: 2, ShowAnchor: true},
			contains: []string{"[Section 1]", "[Section 2]"},
			excludes: []string{"[Title]", "[Subsection]", "[Deep]"},
		},
		{
			name: "ordered list",
			content: `# Title
## Section 1
## Section 2`,
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2, Ordered: true, ShowAnchor: true},
			contains: []string{"1.", "2."},
			excludes: []string{"- ["},
		},
		{
			name: "with line numbers",
			content: `# Title

## Section 1

Content here

## Section 2`,
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2, LineNumber: true, ShowAnchor: true},
			contains: []string{"`:", "-"},
		},
		{
			name:     "code block headers ignored",
			content:  "# Real Header\n```markdown\n# Fake Header\n```\n## Another Real",
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2, ShowAnchor: true},
			contains: []string{"[Real Header]", "[Another Real]"},
			excludes: []string{"[Fake Header]"},
		},
		{
			name: "duplicate header handling",
			content: `# API
## GET
## POST
## GET
## GET`,
			opts: mdtoc.Options{MinLevel: 1, MaxLevel: 2, ShowAnchor: true},
			contains: []string{
				"(#get)",
				"(#get-1)",
				"(#get-2)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := mdtoc.New(tt.opts)
			got, err := toc.GenerateFromContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("GenerateFromContent() error = %v", err)
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("TOC should contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("TOC should NOT contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// TestTOC_SectionMode 测试章节 TOC 模式
func TestTOC_SectionMode(t *testing.T) {
	content := `# 第一章

## 1.1 概述

## 1.2 详情

# 第二章

## 2.1 功能

### 2.1.1 子功能

# 第三章

只有介绍，没有子标题。
`
	toc := mdtoc.New(mdtoc.Options{
		MinLevel:   2,
		MaxLevel:   3,
		SectionTOC: true,
		ShowAnchor: true,
	})

	preview, err := toc.GenerateSectionTOCsPreview([]byte(content))
	if err != nil {
		t.Fatalf("GenerateSectionTOCsPreview() error = %v", err)
	}

	// 验证章节标题存在
	if !strings.Contains(preview, "第一章") {
		t.Error("Preview should contain chapter 1 title")
	}
	if !strings.Contains(preview, "第二章") {
		t.Error("Preview should contain chapter 2 title")
	}

	// 验证子标题链接
	if !strings.Contains(preview, "[1.1 概述]") {
		t.Error("Preview should contain section 1.1 link")
	}
	if !strings.Contains(preview, "[2.1.1 子功能]") {
		t.Error("Preview should contain subsection 2.1.1 link")
	}

	// 第三章无子标题，不应出现在预览中
	if strings.Contains(preview, "第三章") {
		t.Error("Preview should NOT contain chapter 3 (no sub-headers)")
	}
}

// TestTOC_EmptyAndEdgeCases 测试边缘情况
func TestTOC_EmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		opts        mdtoc.Options
		expectEmpty bool
	}{
		{
			name:        "empty content",
			content:     "",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "no headers",
			content:     "Just some text\n\nMore text",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "only code blocks",
			content:     "```\n# Not a header\n## Also not\n```",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "headers below min level",
			content:     "#### Deep header\n##### Deeper",
			opts:        mdtoc.Options{MinLevel: 1, MaxLevel: 2},
			expectEmpty: true,
		},
		{
			name:        "headers above max level",
			content:     "# Title\n## Section",
			opts:        mdtoc.Options{MinLevel: 3, MaxLevel: 6},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := mdtoc.New(tt.opts)
			got, err := toc.GenerateFromContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("GenerateFromContent() error = %v", err)
			}

			isEmpty := got == ""
			if isEmpty != tt.expectEmpty {
				if tt.expectEmpty {
					t.Errorf("Expected empty TOC, got: %q", got)
				} else {
					t.Error("Expected non-empty TOC, got empty")
				}
			}
		})
	}
}

// TestTOC_FileOperations 测试文件操作功能
//
//nolint:gocyclo,maintidx // Test function with many sub-tests, high complexity is acceptable
func TestTOC_FileOperations(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	t.Run("GenerateFromFile", func(t *testing.T) {
		content := "# Title\n## Section 1\n## Section 2"
		filePath := filepath.Join(tmpDir, "test.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		got, err := toc.GenerateFromFile(filePath)
		if err != nil {
			t.Fatalf("GenerateFromFile() error = %v", err)
		}

		if !strings.Contains(got, "[Title]") {
			t.Error("TOC should contain Title link")
		}
	})

	t.Run("GenerateFromFile_NotExists", func(t *testing.T) {
		toc := mdtoc.New(mdtoc.DefaultOptions())
		_, err := toc.GenerateFromFile(filepath.Join(tmpDir, "nonexistent.md"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("HasMarker", func(t *testing.T) {
		// 有标记的文件
		withMarker := "# Title\n<!--TOC-->\n## Section"
		withPath := filepath.Join(tmpDir, "with_marker.md")
		if err := os.WriteFile(withPath, []byte(withMarker), 0600); err != nil {
			t.Fatal(err)
		}

		// 无标记的文件
		withoutMarker := "# Title\n## Section"
		withoutPath := filepath.Join(tmpDir, "without_marker.md")
		if err := os.WriteFile(withoutPath, []byte(withoutMarker), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())

		has, err := toc.HasMarker(withPath)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Error("HasMarker() should return true for file with marker")
		}

		has, err = toc.HasMarker(withoutPath)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Error("HasMarker() should return false for file without marker")
		}
	})

	t.Run("UpdateFile_WithMarker", func(t *testing.T) {
		content := `# Title

<!--TOC-->

Old TOC content

<!--TOC-->

## Section 1

## Section 2
`
		filePath := filepath.Join(tmpDir, "update_with_marker.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		if !strings.Contains(updatedStr, "[Section 1]") {
			t.Error("Updated file should contain Section 1 link")
		}
		if strings.Contains(updatedStr, "Old TOC content") {
			t.Error("Updated file should NOT contain old TOC content")
		}
	})

	t.Run("UpdateFile_WithoutMarker", func(t *testing.T) {
		content := `# Title

## Section 1

## Section 2
`
		filePath := filepath.Join(tmpDir, "update_without_marker.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 应该自动插入 TOC 标记
		if !strings.Contains(updatedStr, "<!--TOC-->") {
			t.Error("Updated file should contain TOC markers")
		}
		if !strings.Contains(updatedStr, "[Section 1]") {
			t.Error("Updated file should contain Section 1 link")
		}
	})

	t.Run("UpdateFile_SectionMode", func(t *testing.T) {
		content := `# Chapter 1

## Section 1.1

# Chapter 2

## Section 2.1
`
		filePath := filepath.Join(tmpDir, "update_section_mode.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.Options{
			MinLevel:   2,
			MaxLevel:   3,
			SectionTOC: true,
			ShowAnchor: true, // 写入文件时必须显示链接
		})
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 每个章节后应该有独立的 TOC
		if strings.Count(updatedStr, "<!--TOC-->") < 4 {
			t.Error("Section mode should insert multiple TOC marker pairs")
		}
		if !strings.Contains(updatedStr, "[Section 1.1]") {
			t.Error("Updated file should contain Section 1.1 link")
		}
		if !strings.Contains(updatedStr, "[Section 2.1]") {
			t.Error("Updated file should contain Section 2.1 link")
		}
	})

	t.Run("DeleteTOC_WithMarker", func(t *testing.T) {
		content := `# Title

<!--TOC-->

- [Section 1](#section-1) ` + "`:10+6`" + `
- [Section 2](#section-2) ` + "`:16+5`" + `

<!--TOC-->

## Section 1

Content here...

## Section 2

More content...
`
		filePath := filepath.Join(tmpDir, "delete_with_marker.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		deleted, err := toc.DeleteTOC(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !deleted {
			t.Error("DeleteTOC() should return true when TOC was deleted")
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// TOC 标记应该被删除
		if strings.Contains(updatedStr, "<!--TOC-->") {
			t.Error("Deleted file should NOT contain TOC markers")
		}
		// TOC 内容应该被删除
		if strings.Contains(updatedStr, "[Section 1](#section-1)") {
			t.Error("Deleted file should NOT contain TOC links")
		}
		// 标题内容应该保留
		if !strings.Contains(updatedStr, "# Title") {
			t.Error("Deleted file should still contain the title")
		}
		if !strings.Contains(updatedStr, "## Section 1") {
			t.Error("Deleted file should still contain section headers")
		}
	})

	t.Run("DeleteTOC_WithoutMarker", func(t *testing.T) {
		content := `# Title

## Section 1

Content here...
`
		filePath := filepath.Join(tmpDir, "delete_without_marker.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		deleted, err := toc.DeleteTOC(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if deleted {
			t.Error("DeleteTOC() should return false when no TOC to delete")
		}

		// 文件内容应该保持不变
		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		if string(updated) != content {
			t.Error("File without TOC should remain unchanged after DeleteTOC()")
		}
	})

	t.Run("DeleteTOC_MultipleTOCBlocks", func(t *testing.T) {
		// 章节模式的文件有多个 TOC 块
		content := `# Chapter 1

<!--TOC-->

- [Section 1.1](#section-11)

<!--TOC-->

## Section 1.1

Content...

# Chapter 2

<!--TOC-->

- [Section 2.1](#section-21)

<!--TOC-->

## Section 2.1

More content...
`
		filePath := filepath.Join(tmpDir, "delete_multiple_toc.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		deleted, err := toc.DeleteTOC(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !deleted {
			t.Error("DeleteTOC() should return true when TOCs were deleted")
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 所有 TOC 标记都应该被删除
		if strings.Contains(updatedStr, "<!--TOC-->") {
			t.Error("All TOC markers should be deleted")
		}
		// 章节标题应该保留
		if !strings.Contains(updatedStr, "# Chapter 1") {
			t.Error("Chapter 1 should be preserved")
		}
		if !strings.Contains(updatedStr, "# Chapter 2") {
			t.Error("Chapter 2 should be preserved")
		}
		if !strings.Contains(updatedStr, "## Section 1.1") {
			t.Error("Section 1.1 should be preserved")
		}
		if !strings.Contains(updatedStr, "## Section 2.1") {
			t.Error("Section 2.1 should be preserved")
		}
	})
}

// TestTOC_ChineseContent 测试中文内容处理
func TestTOC_ChineseContent(t *testing.T) {
	content := `# 项目介绍

## 功能特性

### 高性能

### 易扩展

## 快速开始

### 安装

### 配置

## 常见问题
`
	toc := mdtoc.New(mdtoc.DefaultOptions())
	got, err := toc.GenerateFromContent([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	// 验证中文标题保留
	expectedLinks := []string{
		"[项目介绍](#项目介绍)",
		"[功能特性](#功能特性)",
		"[高性能](#高性能)",
		"[易扩展](#易扩展)",
		"[快速开始](#快速开始)",
		"[安装](#安装)",
		"[配置](#配置)",
		"[常见问题](#常见问题)",
	}

	for _, link := range expectedLinks {
		if !strings.Contains(got, link) {
			t.Errorf("TOC should contain %q", link)
		}
	}
}

// TestTOC_SpecialCharacters 测试特殊字符处理
func TestTOC_SpecialCharacters(t *testing.T) {
	content := `# Hello, World!

## What's New?

## C++ Guide

## Node.js & npm

## 100% Complete

## Version 2.0.0
`
	toc := mdtoc.New(mdtoc.DefaultOptions())
	got, err := toc.GenerateFromContent([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	// 验证特殊字符被正确处理
	if !strings.Contains(got, "[Hello, World!]") {
		t.Error("Title with comma and exclamation should be preserved")
	}
	if !strings.Contains(got, "[What's New?]") {
		t.Error("Title with apostrophe and question mark should be preserved")
	}
	if !strings.Contains(got, "[C++ Guide]") {
		t.Error("Title with plus signs should be preserved")
	}
}

// TestDefaultOptions 测试默认配置
func TestDefaultOptions(t *testing.T) {
	opts := mdtoc.DefaultOptions()

	if opts.MinLevel != 1 {
		t.Errorf("DefaultOptions().MinLevel = %d, want 1", opts.MinLevel)
	}
	if opts.MaxLevel != 3 {
		t.Errorf("DefaultOptions().MaxLevel = %d, want 3", opts.MaxLevel)
	}
	if opts.Ordered {
		t.Error("DefaultOptions().Ordered should be false")
	}
	if opts.LineNumber {
		t.Error("DefaultOptions().LineNumber should be false")
	}
	if !opts.SectionTOC {
		t.Error("DefaultOptions().SectionTOC should be true (章节模式默认启用)")
	}
}

// BenchmarkTOC_Generate 性能基准测试
func BenchmarkTOC_Generate(b *testing.B) {
	// 构造一个较大的文档
	var sb strings.Builder
	for i := range 100 {
		sb.WriteString("# Chapter ")
		sb.WriteRune(rune('A' + i%26))
		sb.WriteString("\n\n")
		for j := range 10 {
			sb.WriteString("## Section ")
			sb.WriteRune(rune('0' + j))
			sb.WriteString("\n\nContent here...\n\n")
		}
	}
	content := []byte(sb.String())

	toc := mdtoc.New(mdtoc.DefaultOptions())

	b.ResetTimer()
	for b.Loop() {
		_, _ = toc.GenerateFromContent(content)
	}
}

// ==================== YAML Frontmatter Line Number Integration Tests ====================

// TestTOC_LineNumbers_WithFrontmatter 测试带 YAML frontmatter 的文件行号计算
func TestTOC_LineNumbers_WithFrontmatter(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		expectedLineNums []string // 期望的行号格式 (如 `:5+7`, `:7+3`)
	}{
		{
			name: "basic frontmatter",
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
			// Title at line 5, Section 1 at line 7, Section 2 at line 10
			expectedLineNums: []string{`:5+`, `:7+`, `:10+`},
		},
		{
			name: "VitePress style frontmatter with YAML comment",
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
			// Getting Started at line 7, Installation at line 9, Usage at line 12
			expectedLineNums: []string{`:7+`, `:9+`, `:12+`},
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
			// Page Title at line 5, First Section at line 7, Second Section at line 10
			expectedLineNums: []string{`:5+`, `:7+`, `:10+`},
		},
		{
			name: "no frontmatter baseline",
			content: `# Title

## Section 1
Content...

## Section 2
More content...
`,
			// Title at line 1, Section 1 at line 3, Section 2 at line 6
			expectedLineNums: []string{`:1+`, `:3+`, `:6+`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := mdtoc.New(mdtoc.Options{
				MinLevel:   1,
				MaxLevel:   3,
				LineNumber: true,
				ShowAnchor: true,
			})
			got, err := toc.GenerateFromContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("GenerateFromContent() error = %v", err)
			}

			for _, lineNum := range tt.expectedLineNums {
				if !strings.Contains(got, lineNum) {
					t.Errorf("TOC should contain line number %q, got:\n%s", lineNum, got)
				}
			}
		})
	}
}

// TestTOC_UpdateFile_WithFrontmatter 测试带 frontmatter 的文件更新
func TestTOC_UpdateFile_WithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name             string
		content          string
		expectedLineNums []string
	}{
		{
			name: "section mode with frontmatter",
			content: `---
title: Test
layout: page
---
# Chapter 1

## Section 1.1
Content here

## Section 1.2
More content
`,
			// Line numbers reflect FINAL file (after TOC insertion)
			// TOC block adds ~9 lines (空行+标记+空行+内容+空行+标记+空行)
			// Section 1.1 at line 14, Section 1.2 at line 17
			expectedLineNums: []string{`:14+`, `:17+`},
		},
		{
			name: "frontmatter with YAML comment should not affect line numbers",
			content: `---
# This is a YAML comment
title: Test
---
# Real Title

## Section A
First section content

## Section B
Second section content
`,
			// Line numbers reflect FINAL file (after TOC insertion)
			// Section A at line 14, Section B at line 17
			expectedLineNums: []string{`:14+`, `:17+`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tt.name+".md")
			if err := os.WriteFile(filePath, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}

			toc := mdtoc.New(mdtoc.Options{
				MinLevel:   2,
				MaxLevel:   3,
				LineNumber: true,
				SectionTOC: true,
				ShowAnchor: true,
			})
			if err := toc.UpdateFile(filePath); err != nil {
				t.Fatal(err)
			}

			updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
			updatedStr := string(updated)

			for _, lineNum := range tt.expectedLineNums {
				if !strings.Contains(updatedStr, lineNum) {
					t.Errorf("Updated file should contain line number %q, got:\n%s", lineNum, updatedStr)
				}
			}
		})
	}
}

// TestTOC_SectionMode_WithFrontmatter 测试带 frontmatter 的章节模式
func TestTOC_SectionMode_WithFrontmatter(t *testing.T) {
	content := `---
# https://vitepress.dev/reference
title: Documentation
description: Project docs
---
# 第一章

## 1.1 概述

## 1.2 详情

# 第二章

## 2.1 功能

### 2.1.1 子功能

# 第三章

只有介绍，没有子标题。
`
	toc := mdtoc.New(mdtoc.Options{
		MinLevel:   2,
		MaxLevel:   3,
		SectionTOC: true,
		ShowAnchor: true,
		LineNumber: true,
	})

	preview, err := toc.GenerateSectionTOCsPreview([]byte(content))
	if err != nil {
		t.Fatalf("GenerateSectionTOCsPreview() error = %v", err)
	}

	// 验证章节标题存在且行号正确
	// 1.1 概述 at line 8
	if !strings.Contains(preview, "[1.1 概述]") {
		t.Error("Preview should contain section 1.1")
	}
	if !strings.Contains(preview, `:8+`) {
		t.Errorf("Preview should contain line number :8+, got:\n%s", preview)
	}

	// 第三章无子标题，不应出现在预览中
	if strings.Contains(preview, "第三章") {
		t.Error("Preview should NOT contain chapter 3 (no sub-headers)")
	}
}

// TestTOC_FrontmatterPreservedAfterUpdate 测试更新后 frontmatter 保持完整
func TestTOC_FrontmatterPreservedAfterUpdate(t *testing.T) {
	tmpDir := t.TempDir()

	content := `---
title: Test Document
author: Test Author
date: 2024-01-01
tags:
  - test
  - documentation
---
# Main Title

## Section 1

Content here

## Section 2

More content
`
	filePath := filepath.Join(tmpDir, "preserve_frontmatter.md")
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	toc := mdtoc.New(mdtoc.Options{
		MinLevel:   2,
		MaxLevel:   3,
		SectionTOC: true,
		ShowAnchor: true,
		LineNumber: true,
	})
	if err := toc.UpdateFile(filePath); err != nil {
		t.Fatal(err)
	}

	updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
	updatedStr := string(updated)

	// Frontmatter 应该完整保留
	expectedFrontmatterParts := []string{
		"---",
		"title: Test Document",
		"author: Test Author",
		"date: 2024-01-01",
		"tags:",
		"  - test",
		"  - documentation",
	}
	for _, part := range expectedFrontmatterParts {
		if !strings.Contains(updatedStr, part) {
			t.Errorf("Frontmatter should contain %q, got:\n%s", part, updatedStr)
		}
	}

	// TOC 应该正确插入
	if !strings.Contains(updatedStr, "<!--TOC-->") {
		t.Error("Updated file should contain TOC markers")
	}
	if !strings.Contains(updatedStr, "[Section 1]") {
		t.Error("Updated file should contain Section 1 link")
	}
}

// TestTOC_DeleteTOC_WithFrontmatter 测试带 frontmatter 的文件删除 TOC
func TestTOC_DeleteTOC_WithFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	content := `---
title: Test
---
# Title

<!--TOC-->

- [Section 1](#section-1) ` + "`:7+4`" + `

<!--TOC-->

## Section 1

Content
`
	filePath := filepath.Join(tmpDir, "delete_with_frontmatter.md")
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	toc := mdtoc.New(mdtoc.DefaultOptions())
	deleted, err := toc.DeleteTOC(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if !deleted {
		t.Error("DeleteTOC() should return true")
	}

	updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
	updatedStr := string(updated)

	// Frontmatter 应该保留
	if !strings.Contains(updatedStr, "---\ntitle: Test\n---") {
		t.Error("Frontmatter should be preserved after deleting TOC")
	}

	// TOC 应该被删除
	if strings.Contains(updatedStr, "<!--TOC-->") {
		t.Error("TOC markers should be deleted")
	}

	// 标题应该保留
	if !strings.Contains(updatedStr, "# Title") {
		t.Error("Title should be preserved")
	}
	if !strings.Contains(updatedStr, "## Section 1") {
		t.Error("Section headers should be preserved")
	}
}

// ==================== TOC Title Tests ====================

// TestTOC_UpdateFile_WithTOCTitle 测试带 TOC 标题的文件更新
func TestTOC_UpdateFile_WithTOCTitle(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("global mode with TOC title", func(t *testing.T) {
		content := `# Project

<!--TOC-->
<!--TOC-->

## Installation

## Usage
`
		filePath := filepath.Join(tmpDir, "global_toc_title.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.Options{
			MinLevel:   2,
			MaxLevel:   3,
			SectionTOC: false, // 全局模式
			ShowAnchor: true,
			TOCTitle:   "文档目录",
		})
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 验证 TOC 标题存在
		if !strings.Contains(updatedStr, "## 文档目录") {
			t.Errorf("Updated file should contain TOC title '## 文档目录', got:\n%s", updatedStr)
		}

		// 验证 TOC 内容存在
		if !strings.Contains(updatedStr, "[Installation]") {
			t.Error("Updated file should contain Installation link")
		}
		if !strings.Contains(updatedStr, "[Usage]") {
			t.Error("Updated file should contain Usage link")
		}
	})

	t.Run("section mode with TOC title", func(t *testing.T) {
		content := `# Chapter 1

## Section 1.1

## Section 1.2

# Chapter 2

## Section 2.1
`
		filePath := filepath.Join(tmpDir, "section_toc_title.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.Options{
			MinLevel:   2,
			MaxLevel:   3,
			SectionTOC: true,
			ShowAnchor: true,
			TOCTitle:   "目录",
		})
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 章节模式下每个章节都应该有 TOC 标题
		if strings.Count(updatedStr, "## 目录") < 2 {
			t.Errorf("Section mode should have multiple TOC titles, got:\n%s", updatedStr)
		}

		// 验证 TOC 内容存在
		if !strings.Contains(updatedStr, "[Section 1.1]") {
			t.Error("Updated file should contain Section 1.1 link")
		}
		if !strings.Contains(updatedStr, "[Section 2.1]") {
			t.Error("Updated file should contain Section 2.1 link")
		}
	})

	t.Run("empty TOC title should not add title", func(t *testing.T) {
		content := `# Project

<!--TOC-->
<!--TOC-->

## Section 1
`
		filePath := filepath.Join(tmpDir, "empty_toc_title.md")
		if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.Options{
			MinLevel:   2,
			MaxLevel:   3,
			SectionTOC: false,
			ShowAnchor: true,
			TOCTitle:   "", // 空标题
		})
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath) //nolint:gosec // G304: test file path
		updatedStr := string(updated)

		// 不应该有额外的 ## 标题（除了 Section 1）
		// 只检查 TOC 区域内不应该有 ## 标题
		if strings.Contains(updatedStr, "## 目录") || strings.Contains(updatedStr, "## 文档目录") {
			t.Error("Empty TOC title should not add any title")
		}
	})
}
