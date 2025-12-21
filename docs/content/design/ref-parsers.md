# Markdown 解析器调研

<!--TOC-->

## Table of Contents

- [1. 规范标准](#1-规范标准) `:26+24`
  - [1.1 CommonMark](#11-commonmark) `:28+10`
  - [1.2 GFM (GitHub Flavored Markdown)](#12-gfm-github-flavored-markdown) `:38+12`
- [2. 主流解析器对比](#2-主流解析器对比) `:50+38`
  - [2.1 按语言分类](#21-按语言分类) `:52+18`
  - [2.2 框架使用情况](#22-框架使用情况) `:70+18`
- [3. Go 生态解析器详细对比](#3-go-生态解析器详细对比) `:88+45`
  - [3.1 goldmark vs blackfriday](#31-goldmark-vs-blackfriday) `:90+11`
  - [3.2 goldmark 性能](#32-goldmark-性能) `:101+10`
  - [3.3 goldmark 扩展](#33-goldmark-扩展) `:111+22`
- [4. 我们的选择](#4-我们的选择) `:133+48`
  - [4.1 决策：goldmark](#41-决策goldmark) `:135+11`
  - [4.2 GitHub Anchor Link 规则](#42-github-anchor-link-规则) `:146+27`
  - [4.3 参考实现](#43-参考实现) `:173+8`
- [5. 未来扩展](#5-未来扩展) `:181+27`
  - [5.1 VitePress 支持 (P2)](#51-vitepress-支持-p2) `:183+11`
  - [5.2 Hugo 支持 (P2)](#52-hugo-支持-p2) `:194+14`
- [6. 参考资料](#6-参考资料) `:208+7`

<!--TOC-->

## 1. 规范标准

### 1.1 CommonMark

[CommonMark](https://commonmark.org/) 是 Markdown 的标准化规范，解决了原始 Markdown 语法歧义问题。

| 版本   | 发布日期   | 说明           |
| ------ | ---------- | -------------- |
| 0.31.2 | 2024-01-28 | 最新稳定版     |
| 0.30   | 2021-06-19 | 主要更新       |
| 0.29   | 2019-04-06 | GFM 基于此版本 |

### 1.2 GFM (GitHub Flavored Markdown)

[GFM](https://github.github.com/gfm/) 是 CommonMark 的超集，添加了以下扩展：

| 扩展      | 语法示例              | 说明           |
| --------- | --------------------- | -------------- |
| 表格      | `\| a \| b \|`        | 管道符分隔     |
| 任务列表  | `- [x] done`          | 复选框         |
| 删除线    | `~~text~~`            | 双波浪线       |
| 自动链接  | `https://example.com` | 无需 `<>` 包裹 |
| 禁用 HTML | `<script>` 等         | 安全过滤       |

## 2. 主流解析器对比

### 2.1 按语言分类

| 语言       | 解析器         | Stars | 特点                       |
| ---------- | -------------- | ----- | -------------------------- |
| **C**      | cmark          | ~1.7k | CommonMark 官方参考实现    |
|            | cmark-gfm      | ~900  | GitHub 官方 GFM 实现       |
|            | MD4C           | ~700  | SAX 风格，极快             |
| **Go**     | goldmark       | ~3.5k | Hugo 默认，CommonMark 兼容 |
|            | blackfriday    | ~5.4k | 老牌，非 CommonMark 兼容   |
|            | Lute           | ~1k   | 中文优化                   |
| **JS**     | markdown-it    | ~18k  | VitePress/Docusaurus 使用  |
|            | commonmark.js  | ~1.5k | 官方 JS 参考实现           |
|            | remark         | ~7k   | Gatsby 使用，AST 生态      |
| **Rust**   | comrak         | ~1.1k | 基于 cmark-gfm             |
|            | pulldown-cmark | ~2k   | 高性能                     |
| **Python** | markdown-it-py | ~700  | markdown-it 移植           |
|            | mistletoe      | ~800  | 纯 Python，最快            |

### 2.2 框架使用情况

```
┌─────────────────┬──────────────────┬─────────────┐
│ 框架/平台        │ 解析器            │ 语言        │
├─────────────────┼──────────────────┼─────────────┤
│ GitHub.com      │ cmark-gfm        │ C           │
│ GitLab          │ commonmarker     │ Ruby (C绑定) │
│ Hugo            │ goldmark         │ Go          │
│ VitePress       │ markdown-it      │ JavaScript  │
│ Docusaurus      │ markdown-it      │ JavaScript  │
│ Gatsby          │ remark           │ JavaScript  │
│ Jekyll          │ kramdown         │ Ruby        │
│ Astro           │ remark           │ JavaScript  │
│ MkDocs          │ Python-Markdown  │ Python      │
└─────────────────┴──────────────────┴─────────────┘
```

## 3. Go 生态解析器详细对比

### 3.1 goldmark vs blackfriday

| 特性            | goldmark      | blackfriday    |
| --------------- | ------------- | -------------- |
| CommonMark 兼容 | ✅ 完全兼容   | ❌ 不兼容      |
| GFM 支持        | ✅ 内置扩展   | ✅ 部分支持    |
| 可扩展性        | ✅ 接口设计   | ❌ struct 设计 |
| 性能            | ⭐⭐⭐⭐⭐    | ⭐⭐⭐⭐       |
| 维护状态        | 活跃          | 维护模式       |
| 使用者          | Hugo, glamour | 旧项目         |

### 3.2 goldmark 性能

与 cmark (C 参考实现) 对比：

```
Benchmark (50 iterations on same file):
cmark:    0.0044073057 sec
goldmark: 0.0041611990 sec  ← 略快
```

### 3.3 goldmark 扩展

```go
import (
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
)

md := goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,           // 表格、删除线、自动链接、任务列表
        extension.Table,         // 单独启用表格
        extension.Strikethrough, // 单独启用删除线
        extension.TaskList,      // 单独启用任务列表
        extension.Linkify,       // 单独启用自动链接
        extension.DefinitionList,// 定义列表
        extension.Footnote,      // 脚注
        extension.Typographer,   // 排版优化 ("--" → "—")
    ),
)
```

## 4. 我们的选择

### 4.1 决策：goldmark

**选择理由**：

1. **Go 原生** - 无需 CGO，单一二进制部署
2. **CommonMark 0.31 兼容** - 符合最新规范
3. **GFM 扩展内置** - 开箱即用
4. **性能出色** - 与 C 实现相当
5. **生态成熟** - Hugo/glamour/GitHub CLI 验证
6. **可扩展** - 便于未来添加 VitePress 等支持

### 4.2 GitHub Anchor Link 规则

goldmark 不直接提供 GitHub 风格的 anchor link 生成，需要自行实现：

```
输入: "Hello, World! 你好"

处理步骤:
1. 小写化        → "hello, world! 你好"
2. 移除 HTML     → "hello, world! 你好"
3. 移除强调符号   → "hello, world! 你好"
4. 移除标点      → "hello world 你好"   (保留 \w\- 空格)
5. 空格转连字符   → "hello-world-你好"

输出: "hello-world-你好"
```

**重复标题处理**：

```markdown
# Title → #title

# Title → #title-1

# Title → #title-2
```

### 4.3 参考实现

| 项目                   | 文件            | 说明                  |
| ---------------------- | --------------- | --------------------- |
| `vendor/gh-md-toc-go/` | `ghtoc.go`      | Go 实现 GitHub anchor |
| `vendor/md-toc/`       | `md_toc/api.py` | Python 详细算法       |
| `vendor/goldmark-toc/` | `toc.go`        | goldmark AST 遍历     |

## 5. 未来扩展

### 5.1 VitePress 支持 (P2)

VitePress 使用 markdown-it，主要差异：

| 特性          | GitHub   | VitePress           |
| ------------- | -------- | ------------------- |
| Frontmatter   | 不处理   | 跳过 YAML           |
| 自定义容器    | 无       | `:::` 语法          |
| 自定义标题 ID | 自动生成 | 支持 `{#custom-id}` |
| 代码组        | 无       | 多 tab 切换         |

### 5.2 Hugo 支持 (P2)

Hugo 同样使用 goldmark，但有额外配置：

```toml
# hugo.toml
[markup.goldmark]
  [markup.goldmark.renderer]
    unsafe = true  # 允许原始 HTML
  [markup.goldmark.extensions]
    definitionList = true
    footnote = true
```

## 6. 参考资料

- [CommonMark Spec 0.31.2](https://spec.commonmark.org/0.31.2/)
- [GitHub Flavored Markdown Spec](https://github.github.com/gfm/)
- [goldmark GitHub](https://github.com/yuin/goldmark)
- [CommonMark Implementations](https://github.com/commonmark/commonmark-spec/wiki/List-of-CommonMark-Implementations)
- [VitePress Markdown Extensions](https://vitepress.dev/guide/markdown)
